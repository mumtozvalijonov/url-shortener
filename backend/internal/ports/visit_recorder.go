package ports

import (
	"context"

	"github.com/mumtozvalijonov/url-shortener/internal/domain"
)

type ShortURLVisitedRecorder interface {
	RecordVisits(ctx context.Context, visits []domain.ShortURLVisited) error
}
