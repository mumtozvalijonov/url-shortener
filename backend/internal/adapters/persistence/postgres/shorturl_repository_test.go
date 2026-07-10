package postgres_test

import (
	"context"
	"net/url"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	postgresadapter "github.com/mumtozvalijonov/url-shortener/internal/adapters/persistence/postgres"
	"github.com/mumtozvalijonov/url-shortener/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestShortURLRepository(t *testing.T) {
	server := startPostgres(context.Background(), t)

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		repo := postgresadapter.NewShortURLRepository(server.newDB(ctx, t))

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
		repo := postgresadapter.NewShortURLRepository(server.newDB(ctx, t))

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
		repo := postgresadapter.NewShortURLRepository(server.newDB(ctx, t))

		_, err := repo.GetByCode(ctx, "qwert")
		require.ErrorIs(t, err, domain.ErrShortURLNotFound)
	})

	t.Run("UpdateByCode", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		repo := postgresadapter.NewShortURLRepository(server.newDB(ctx, t))

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
		repo := postgresadapter.NewShortURLRepository(server.newDB(ctx, t))

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
		repo := postgresadapter.NewShortURLRepository(server.newDB(ctx, t))

		err := repo.DeleteByCode(ctx, "qwert")
		require.ErrorIs(t, err, domain.ErrShortURLNotFound)
	})
}
