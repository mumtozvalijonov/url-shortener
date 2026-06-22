package ports

import (
	"context"
	"net/url"

	"github.com/mumtozvalijonov/url-shortener/internal/domain"
)

type ShortURLRepository interface {
	Insert(ctx context.Context, code string, target url.URL) (domain.ShortURL, error)
	GetByCode(ctx context.Context, code string) (domain.ShortURL, error)
	UpdateByCode(ctx context.Context, code string, target url.URL) (domain.ShortURL, error)
	DeleteByCode(ctx context.Context, code string) error
	GetStatistics(ctx context.Context, code string) (domain.ShortURLStatistics, error)
}
