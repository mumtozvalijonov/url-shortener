package ports

import (
	"context"
	"net/url"

	"github.com/mumtozvalijonov/url-shortener/internal/domain"
)

type ShortenerService interface {
	Shorten(ctx context.Context, url url.URL) (domain.ShortURL, error)
	Retrieve(ctx context.Context, code string) (domain.ShortURL, error)
	Update(ctx context.Context, code string, url url.URL) (domain.ShortURL, error)
	Delete(ctx context.Context, code string) error
	GetStatistics(ctx context.Context, code string) (domain.ShortURLStatistics, error)
}
