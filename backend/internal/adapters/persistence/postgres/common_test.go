package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mumtozvalijonov/url-shortener/internal/domain"
	"github.com/mumtozvalijonov/url-shortener/migrations"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const templateDatabase = "url_shortener_template"

type postgresTestServer struct {
	adminDSN string
	adminDB  *sql.DB
	nextID   atomic.Uint64
}

func startPostgres(ctx context.Context, t *testing.T) *postgresTestServer {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	t.Cleanup(cancel)

	container, err := postgres.Run(
		ctx,
		"",
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context:    "../../../../..",
			Dockerfile: "docker/postgres/Dockerfile",
			Repo:       "url-shortener-postgres",
			Tag:        "18.4-test",
			KeepImage:  true,
		}),
		postgres.WithDatabase(templateDatabase),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithCmd(
			"postgres",
			"-c", "fsync=off",
			"-c", "shared_preload_libraries=pg_cron",
			"-c", "cron.database_name="+templateDatabase,
			"-c", "cron.timezone=UTC",
			"-c", "cron.use_background_workers=on",
			"-c", "timezone=UTC",
		),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, container)

	templateDSN, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	templateDB, err := sql.Open("pgx", templateDSN)
	require.NoError(t, err)
	require.NoError(t, templateDB.PingContext(ctx))

	migrationsFS, err := migrations.FS()
	require.NoError(t, err)

	provider, err := goose.NewProvider(goose.DialectPostgres, templateDB, migrationsFS)
	require.NoError(t, err)

	_, err = provider.Up(ctx)
	require.NoError(t, err)
	require.NoError(t, templateDB.Close())

	adminDSN, err := dsnForDatabase(templateDSN, "postgres")
	require.NoError(t, err)

	adminDB, err := sql.Open("pgx", adminDSN)
	require.NoError(t, err)
	require.NoError(t, adminDB.PingContext(ctx))

	_, err = adminDB.ExecContext(
		ctx,
		fmt.Sprintf("ALTER DATABASE %s IS_TEMPLATE true", templateDatabase),
	)
	require.NoError(t, err)
	_, err = adminDB.ExecContext(
		ctx,
		fmt.Sprintf("ALTER DATABASE %s ALLOW_CONNECTIONS false", templateDatabase),
	)
	require.NoError(t, err)
	_, err = adminDB.ExecContext(
		ctx,
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1",
		templateDatabase,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, adminDB.Close())
	})

	return &postgresTestServer{
		adminDSN: adminDSN,
		adminDB:  adminDB,
	}
}

func dsnForDatabase(dsn, database string) (string, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("parse PostgreSQL DSN: %w", err)
	}

	parsed.Path = "/" + database
	return parsed.String(), nil
}

func (s *postgresTestServer) newDB(
	ctx context.Context,
	t *testing.T,
) *sql.DB {
	t.Helper()

	database := fmt.Sprintf("test_%d", s.nextID.Add(1))
	_, err := s.adminDB.ExecContext(
		ctx,
		fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s", database, templateDatabase),
	)
	require.NoError(t, err)

	dsn, err := dsnForDatabase(s.adminDSN, database)
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	require.NoError(t, db.PingContext(ctx))

	t.Cleanup(func() {
		require.NoError(t, db.Close())
		_, err := s.adminDB.ExecContext(
			context.Background(),
			fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", database),
		)
		require.NoError(t, err)
	})

	return db
}

func seedShortURL(
	ctx context.Context,
	t *testing.T,
	db *sql.DB,
	shortURLs []domain.ShortURL,
) {
	if len(shortURLs) == 0 {
		return
	}

	t.Helper()

	values := make([]string, 0, len(shortURLs))
	args := make([]any, 0, len(shortURLs)*2)

	for i, shortURL := range shortURLs {
		values = append(values, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		args = append(args, shortURL.TargetURL.String(), shortURL.ShortCode)
	}

	query := fmt.Sprintf(`
		INSERT INTO short_urls (target_url, short_code)
		VALUES %s
	`, strings.Join(values, ","))

	_, err := db.ExecContext(ctx, query, args...)
	require.NoError(t, err)
}
