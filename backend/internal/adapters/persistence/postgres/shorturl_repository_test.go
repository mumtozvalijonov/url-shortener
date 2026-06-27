package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	postgresadapter "github.com/mumtozvalijonov/url-shortener/internal/adapters/persistence/postgres"
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
		"postgres:18-alpine",
		postgres.WithDatabase(templateDatabase),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
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

func (s *postgresTestServer) newRepository(
	ctx context.Context,
	t *testing.T,
) *postgresadapter.ShortURLRepository {
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

	return postgresadapter.NewShortURLRepository(db)
}

func TestShortURLRepository(t *testing.T) {
	server := startPostgres(context.Background(), t)

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		repo := server.newRepository(ctx, t)

		target, err := url.Parse("https://www.google.com/")
		require.NoError(t, err)

		inserted, err := repo.Insert(ctx, "abcde", *target)
		require.NoError(t, err)

		require.Equal(t, "abcde", inserted.ShortCode)
		require.Equal(t, *target, inserted.TargetURL)
		require.NotZero(t, inserted.ID)
		require.False(t, inserted.CreatedAt.IsZero())
		require.False(t, inserted.UpdatedAt.IsZero())

		_, err = repo.Insert(ctx, inserted.ShortCode, *target)
		require.ErrorIs(t, err, domain.ErrShortCodeTaken)
	})

	t.Run("GetByCode", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		repo := server.newRepository(ctx, t)

		target, err := url.Parse("https://www.google.com/")
		require.NoError(t, err)

		inserted, err := repo.Insert(ctx, "abcde", *target)
		require.NoError(t, err)

		got, err := repo.GetByCode(ctx, inserted.ShortCode)
		require.NoError(t, err)
		require.Equal(t, inserted, got)
	})

	t.Run("GetByCode/not found", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		repo := server.newRepository(ctx, t)

		_, err := repo.GetByCode(ctx, "qwert")
		require.ErrorIs(t, err, domain.ErrShortURLNotFound)
	})

	t.Run("UpdateByCode", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		repo := server.newRepository(ctx, t)

		target, err := url.Parse("https://www.google.com/")
		require.NoError(t, err)

		inserted, err := repo.Insert(ctx, "abcde", *target)
		require.NoError(t, err)

		newTarget, err := url.Parse("https://www.amazon.com")
		require.NoError(t, err)

		updated, err := repo.UpdateByCode(ctx, inserted.ShortCode, *newTarget)
		require.NoError(t, err)

		require.Equal(t, inserted.ID, updated.ID)
		require.Equal(t, inserted.ShortCode, updated.ShortCode)
		require.Equal(t, *newTarget, updated.TargetURL)
		require.NotEqual(t, inserted.UpdatedAt, updated.UpdatedAt)
	})

	t.Run("DeleteByCode", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		repo := server.newRepository(ctx, t)

		target, err := url.Parse("https://www.google.com/")
		require.NoError(t, err)

		inserted, err := repo.Insert(ctx, "abcde", *target)
		require.NoError(t, err)

		require.NoError(t, repo.DeleteByCode(ctx, inserted.ShortCode))

		_, err = repo.GetByCode(ctx, inserted.ShortCode)
		require.ErrorIs(t, err, domain.ErrShortURLNotFound)
	})

	t.Run("DeleteByCode/not found", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		repo := server.newRepository(ctx, t)

		err := repo.DeleteByCode(ctx, "qwert")
		require.ErrorIs(t, err, domain.ErrShortURLNotFound)
	})
}
