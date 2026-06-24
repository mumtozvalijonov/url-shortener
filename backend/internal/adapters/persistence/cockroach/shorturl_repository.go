package cockroach

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mumtozvalijonov/url-shortener/internal/domain"
)

type ShortURLRepository struct {
	db *sql.DB
}

func NewShortURLRepository(db *sql.DB) *ShortURLRepository {
	return &ShortURLRepository{db: db}
}

func (r *ShortURLRepository) Insert(ctx context.Context, code string, target url.URL) (domain.ShortURL, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
			INSERT INTO short_urls (target_url, short_code)
			VALUES ($1, $2)
			RETURNING id, target_url, short_code, created_at, updated_at
		`,
		target.String(),
		code,
	)

	var rawTarget string
	var shortURL domain.ShortURL

	err := row.Scan(
		&shortURL.ID,
		&rawTarget,
		&shortURL.ShortCode,
		&shortURL.CreatedAt,
		&shortURL.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ShortURL{}, fmt.Errorf("insert short url: %w", domain.ErrShortCodeTaken)
		}

		return domain.ShortURL{}, fmt.Errorf("insert short url: %v", err)
	}

	parsedTarget, err := url.Parse(rawTarget)
	if err != nil {
		return domain.ShortURL{}, domain.ErrInvalidURL
	}

	shortURL.TargetURL = *parsedTarget
	return shortURL, nil
}

func (r *ShortURLRepository) GetByCode(ctx context.Context, code string) (domain.ShortURL, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
			SELECT id, target_url, short_code, created_at, updated_at
			FROM short_urls
			WHERE short_code = $1
		`,
		code,
	)

	var rawTarget string
	var shortURL domain.ShortURL

	err := row.Scan(
		&shortURL.ID,
		&rawTarget,
		&shortURL.ShortCode,
		&shortURL.CreatedAt,
		&shortURL.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ShortURL{}, fmt.Errorf("get short url by code %q: %w", code, domain.ErrShortURLNotFound)
		}
		return domain.ShortURL{}, fmt.Errorf("get short url by code %q: %v", code, err)
	}

	parsedTarget, err := url.Parse(rawTarget)
	if err != nil {
		return domain.ShortURL{}, domain.ErrInvalidURL
	}

	shortURL.TargetURL = *parsedTarget
	return shortURL, nil
}

func (r *ShortURLRepository) UpdateByCode(ctx context.Context, code string, target url.URL) (domain.ShortURL, error) {
	panic("unimplemented")
}

func (r *ShortURLRepository) DeleteByCode(ctx context.Context, code string) error {
	panic("unimplemented")
}

func (r *ShortURLRepository) GetStatistics(ctx context.Context, code string) (domain.ShortURLStatistics, error) {
	panic("unimplemented")
}
