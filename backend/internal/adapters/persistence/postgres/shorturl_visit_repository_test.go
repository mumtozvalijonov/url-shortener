package postgres_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	postgresadapter "github.com/mumtozvalijonov/url-shortener/internal/adapters/persistence/postgres"
	"github.com/mumtozvalijonov/url-shortener/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestShortURLVisitRepository(t *testing.T) {
	server := startPostgres(context.Background(), t)

	t.Run("InsertVisits: happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db := server.newDB(ctx, t)
		targetURL, err := url.Parse("https://www.google.com/")
		require.NoError(t, err)
		seedShortURL(ctx, t, db, []domain.ShortURL{
			{
				ShortCode: "abcde",
				TargetURL: *targetURL,
			},
		})
		repo := postgresadapter.NewShortURLVisitRepository(db)

		visits := []domain.ShortURLVisited{
			{
				ShortCode: "abcde",
				VisitedAt: time.Now(),
			},
		}
		err = repo.InsertVisits(ctx, visits)
		require.NoError(t, err)
	})

	t.Run("InsertVisits: FK violation", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		db := server.newDB(ctx, t)
		repo := postgresadapter.NewShortURLVisitRepository(db)

		visits := []domain.ShortURLVisited{
			{
				ShortCode: "abcde",
				VisitedAt: time.Now(),
			},
		}
		err := repo.InsertVisits(ctx, visits)
		require.ErrorIs(t, err, domain.ErrShortURLNotFound)
	})
}
