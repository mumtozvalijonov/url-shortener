package services

import (
	"context"
	"log"

	"github.com/mumtozvalijonov/url-shortener/internal/domain"
)

type statisticsService struct{}

func NewStatisticsService() *statisticsService {
	return &statisticsService{}
}

func (s *statisticsService) RecordVisits(ctx context.Context, visits []domain.ShortURLVisited) error {
	log.Printf("visits: %v\n", visits)
	// TODO: push batch into short_url_visits table in postgres
	return nil
}
