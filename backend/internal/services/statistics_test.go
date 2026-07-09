package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/mumtozvalijonov/url-shortener/internal/domain"
	"github.com/mumtozvalijonov/url-shortener/internal/ports/mocks"
	"github.com/mumtozvalijonov/url-shortener/internal/services"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStatisticsService(t *testing.T) {
	t.Run("RecordVisits: happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		shortURLVisitRepo := mocks.NewMockShortURLVisitRepository(t)
		batch := []domain.ShortURLVisited{
			{
				ShortCode: "abcde",
				VisitedAt: time.Now(),
			},
		}
		shortURLVisitRepo.EXPECT().InsertVisits(mock.Anything, batch).Return(nil)

		service := services.NewStatisticsService(shortURLVisitRepo)
		err := service.RecordVisits(ctx, batch)
		require.NoError(t, err)
	})
	t.Run("RecordVisits: on code not found", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		shortURLVisitRepo := mocks.NewMockShortURLVisitRepository(t)
		batch := []domain.ShortURLVisited{
			{
				ShortCode: "abcde",
				VisitedAt: time.Now(),
			},
		}
		shortURLVisitRepo.EXPECT().InsertVisits(mock.Anything, batch).Return(domain.ErrShortURLNotFound)

		service := services.NewStatisticsService(shortURLVisitRepo)
		err := service.RecordVisits(ctx, batch)
		require.ErrorIs(t, err, domain.ErrShortURLNotFound)
	})
}
