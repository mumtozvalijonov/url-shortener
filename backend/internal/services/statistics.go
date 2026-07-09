package services

import (
	"context"
	"fmt"

	"github.com/mumtozvalijonov/url-shortener/internal/domain"
	"github.com/mumtozvalijonov/url-shortener/internal/ports"
)

type statisticsService struct {
	shortURLVisitRepo ports.ShortURLVisitRepository
}

func NewStatisticsService(shortURLVisitRepo ports.ShortURLVisitRepository) *statisticsService {
	return &statisticsService{
		shortURLVisitRepo: shortURLVisitRepo,
	}
}

func (s *statisticsService) RecordVisits(ctx context.Context, visits []domain.ShortURLVisited) error {
	if err := s.shortURLVisitRepo.InsertVisits(ctx, visits); err != nil {
		return fmt.Errorf("record visits: %w", err)
	}
	return nil
}
