package cockroach_test

import (
	"context"
	"database/sql"
	"net/url"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mumtozvalijonov/url-shortener/internal/adapters/persistence/cockroach"
	"github.com/mumtozvalijonov/url-shortener/internal/domain"
	"github.com/mumtozvalijonov/url-shortener/migrations"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func startCockroach(ctx context.Context, t *testing.T) *sql.DB {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	t.Cleanup(cancel)

	container, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v25.2")
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, container)

	dsn, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	err = db.PingContext(ctx)
	require.NoError(t, err)

	migrationsFS, err := migrations.FS()
	require.NoError(t, err)

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrationsFS,
	)
	require.NoError(t, err)

	_, err = provider.Up(ctx)
	require.NoError(t, err)

	return db
}

func TestShortURLRepository_Insert(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := startCockroach(ctx, t)
	repo := cockroach.NewShortURLRepository(db)

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
}

func TestShortURLRepository_GetByCode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := startCockroach(ctx, t)
	repo := cockroach.NewShortURLRepository(db)

	target, err := url.Parse("https://www.google.com/")
	require.NoError(t, err)

	inserted, err := repo.Insert(ctx, "abcde", *target)
	require.NoError(t, err)

	got, err := repo.GetByCode(ctx, inserted.ShortCode)
	require.NoError(t, err)

	require.Equal(t, inserted, got)
}

func TestShortURLRepository_GetByCode_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := startCockroach(ctx, t)
	repo := cockroach.NewShortURLRepository(db)

	target, err := url.Parse("https://www.google.com/")
	require.NoError(t, err)

	_, err = repo.Insert(ctx, "abcde", *target)
	require.NoError(t, err)

	_, err = repo.GetByCode(ctx, "qwert")
	require.ErrorIs(t, err, domain.ErrShortURLNotFound)
}

func TestShortURLRepository_UpdateByCode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := startCockroach(ctx, t)
	repo := cockroach.NewShortURLRepository(db)

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
	require.Equal(t, updated.TargetURL, *newTarget)
	require.NotEqual(t, inserted.UpdatedAt, updated.UpdatedAt)
}

func TestShortURLRepository_DeleteByCode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := startCockroach(ctx, t)
	repo := cockroach.NewShortURLRepository(db)

	target, err := url.Parse("https://www.google.com/")
	require.NoError(t, err)

	inserted, err := repo.Insert(ctx, "abcde", *target)
	require.NoError(t, err)

	err = repo.DeleteByCode(ctx, inserted.ShortCode)
	require.NoError(t, err)

	_, err = repo.GetByCode(ctx, inserted.ShortCode)
	require.ErrorIs(t, err, domain.ErrShortURLNotFound)
}

func TestShortURLRepository_DeleteByCode_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := startCockroach(ctx, t)
	repo := cockroach.NewShortURLRepository(db)

	target, err := url.Parse("https://www.google.com/")
	require.NoError(t, err)

	_, err = repo.Insert(ctx, "abcde", *target)
	require.NoError(t, err)

	err = repo.DeleteByCode(ctx, "qwert")
	require.ErrorIs(t, err, domain.ErrShortURLNotFound)
}
