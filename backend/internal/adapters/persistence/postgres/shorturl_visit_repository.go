package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mumtozvalijonov/url-shortener/internal/domain"
)

type shortURLVisitRepository struct {
	db *sql.DB
}

func NewShortURLVisitRepository(db *sql.DB) *shortURLVisitRepository {
	return &shortURLVisitRepository{db: db}
}

func isFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func (r *shortURLVisitRepository) InsertVisits(ctx context.Context, visits []domain.ShortURLVisited) error {
	if len(visits) == 0 {
		return nil
	}

	values := make([]string, 0, len(visits))
	args := make([]any, 0, len(visits)*2)

	for i, visit := range visits {
		values = append(values, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		args = append(args, visit.ShortCode, visit.VisitedAt)
	}

	query := fmt.Sprintf(`
		INSERT INTO short_url_visits (short_code, visited_at)
		VALUES %s
	`, strings.Join(values, ","))

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		switch {
		case isFKViolation(err):
			return fmt.Errorf("insert visits: %w", domain.ErrShortURLNotFound)
		default:
			return fmt.Errorf("insert visits: %v", err)
		}
	}
	return nil
}
