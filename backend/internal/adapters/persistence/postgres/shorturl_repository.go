package postgres

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

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func rowToShortURL(row *sql.Row) (domain.ShortURL, error) {
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
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return domain.ShortURL{}, fmt.Errorf("row to short url: %w", domain.ErrShortURLNotFound)
		case isUniqueViolation(err):
			return domain.ShortURL{}, fmt.Errorf("row to short url: %w", domain.ErrShortCodeTaken)
		default:
			return domain.ShortURL{}, fmt.Errorf("row to short url: %v", err)
		}
	}

	parsedTarget, err := url.Parse(rawTarget)
	if err != nil {
		return domain.ShortURL{}, domain.ErrInvalidURL
	}

	shortURL.TargetURL = *parsedTarget
	return shortURL, nil
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
	return rowToShortURL(row)
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
	return rowToShortURL(row)
}

func (r *ShortURLRepository) UpdateByCode(ctx context.Context, code string, target url.URL) (domain.ShortURL, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
			UPDATE short_urls SET target_url = $1
			WHERE short_code = $2
			RETURNING id, target_url, short_code, created_at, updated_at
		`,
		target.String(),
		code,
	)
	return rowToShortURL(row)
}

func (r *ShortURLRepository) DeleteByCode(ctx context.Context, code string) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			DELETE from short_urls
			WHERE short_code = $1
		`,
		code,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrShortURLNotFound
	}

	return nil
}

func (r *ShortURLRepository) GetStatistics(ctx context.Context, code string) (domain.ShortURLStatistics, error) {
	panic("unimplemented")
}
