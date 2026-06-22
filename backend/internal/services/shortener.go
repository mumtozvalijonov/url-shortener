package services

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mumtozvalijonov/url-shortener/internal/domain"
	"github.com/mumtozvalijonov/url-shortener/internal/ports"
)

type ShortenerService struct {
	shortURLRepo  ports.ShortURLRepository
	codeGenerator ports.CodeGenerator
}

func NewShortenerService(
	shortURLRepo ports.ShortURLRepository,
	codeGenerator ports.CodeGenerator,
) *ShortenerService {
	return &ShortenerService{
		shortURLRepo:  shortURLRepo,
		codeGenerator: codeGenerator,
	}
}

func (s *ShortenerService) Shorten(ctx context.Context, target url.URL) (domain.ShortURL, error) {
	code := s.codeGenerator.Generate()
	result, err := s.shortURLRepo.Insert(ctx, code, target)
	if err != nil {
		return domain.ShortURL{}, fmt.Errorf("shorten url: %w", err)
	}
	return result, nil
}

func (s *ShortenerService) Retrieve(ctx context.Context, code string) (domain.ShortURL, error) {
	result, err := s.shortURLRepo.GetByCode(ctx, code)
	if err != nil {
		return domain.ShortURL{}, fmt.Errorf("retrieve url: %w", err)
	}
	return result, nil
}

func (s *ShortenerService) Update(ctx context.Context, code string, target url.URL) (domain.ShortURL, error) {
	result, err := s.shortURLRepo.UpdateByCode(ctx, code, target)
	if err != nil {
		return domain.ShortURL{}, fmt.Errorf("retrieve url: %w", err)
	}
	return result, nil
}

func (s *ShortenerService) Delete(ctx context.Context, code string) error {
	err := s.shortURLRepo.DeleteByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("delete url: %w", err)
	}
	return nil
}

func (s *ShortenerService) GetStatistics(ctx context.Context, code string) (domain.ShortURLStatistics, error) {
	result, err := s.shortURLRepo.GetStatistics(ctx, code)
	if err != nil {
		return domain.ShortURLStatistics{}, fmt.Errorf("get url statistics: %w", err)
	}
	return result, nil
}
