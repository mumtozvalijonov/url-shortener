package ports

import (
	"context"

	"github.com/mumtozvalijonov/url-shortener/internal/domain"
)

type ShortURLVisitedPublisher interface {
	Publish(ctx context.Context, event domain.ShortURLVisited) error
}
