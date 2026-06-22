package services_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/mumtozvalijonov/url-shortener/internal/domain"
	"github.com/mumtozvalijonov/url-shortener/internal/services"
)

type fakeCodeGenerator struct {
	code string
}

func (f fakeCodeGenerator) Generate() string {
	return f.code
}

type fakeShortURLRepository struct {
	onInsert        func(ctx context.Context, code string, target url.URL) (domain.ShortURL, error)
	onGetByCode     func(ctx context.Context, code string) (domain.ShortURL, error)
	onUpdateByCode  func(ctx context.Context, code string, target url.URL) (domain.ShortURL, error)
	onDeleteByCode  func(ctx context.Context, code string) error
	onGetStatistics func(ctx context.Context, code string) (domain.ShortURLStatistics, error)
}

func requireFunc[T any](name string, fn T) T {
	if reflect.ValueOf(fn).IsNil() {
		panic(fmt.Sprintf("unexpected %s call", name))
	}
	return fn
}

func (f fakeShortURLRepository) Insert(ctx context.Context, code string, target url.URL) (domain.ShortURL, error) {
	return requireFunc("Insert", f.onInsert)(ctx, code, target)
}

func (f fakeShortURLRepository) GetByCode(ctx context.Context, code string) (domain.ShortURL, error) {
	return requireFunc("GetByCode", f.onGetByCode)(ctx, code)
}

func (f fakeShortURLRepository) UpdateByCode(ctx context.Context, code string, url url.URL) (domain.ShortURL, error) {
	return requireFunc("UpdateByCode", f.onUpdateByCode)(ctx, code, url)
}

func (f fakeShortURLRepository) DeleteByCode(ctx context.Context, code string) error {
	return requireFunc("DeleteByCode", f.onDeleteByCode)(ctx, code)
}

func (f fakeShortURLRepository) GetStatistics(ctx context.Context, code string) (domain.ShortURLStatistics, error) {
	return requireFunc("GetStatistics", f.onGetStatistics)(ctx, code)
}

func TestShortenerService_Insert_OnSuccess(t *testing.T) {
	ctx := context.Background()
	longURL, err := url.Parse("https://www.google.com/")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)
	shortURLRepo := fakeShortURLRepository{
		onInsert: func(ctx context.Context, code string, target url.URL) (domain.ShortURL, error) {
			return domain.ShortURL{
				ID:        1,
				TargetURL: target,
				ShortCode: code,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	code := "abcde"
	codeGenerator := fakeCodeGenerator{
		code: code,
	}

	want := domain.ShortURL{
		ID:        1,
		TargetURL: *longURL,
		ShortCode: code,
		CreatedAt: now,
		UpdatedAt: now,
	}

	service := services.NewShortenerService(shortURLRepo, codeGenerator)

	got, err := service.Shorten(ctx, *longURL)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestShortenerService_Insert_OnError(t *testing.T) {
	tests := []struct {
		name      string
		repoErr   error
		targetErr error
	}{
		{
			name:      "on ErrShortCodeTaken",
			repoErr:   domain.ErrShortCodeTaken,
			targetErr: domain.ErrShortCodeTaken,
		},
		{
			name:    "on generic error",
			repoErr: errors.New("db unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			shortURLRepo := fakeShortURLRepository{
				onInsert: func(ctx context.Context, code string, target url.URL) (domain.ShortURL, error) {
					return domain.ShortURL{}, tt.repoErr
				},
			}
			codeGenerator := fakeCodeGenerator{
				code: "abcde",
			}

			service := services.NewShortenerService(shortURLRepo, codeGenerator)
			longURL, err := url.Parse("https://www.google.com/")
			if err != nil {
				t.Fatal(err)
			}

			_, err = service.Shorten(ctx, *longURL)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Fatalf("expected %v, got %v", tt.targetErr, err)
			}
		})
	}
}

func TestShortenerService_Retrieve_OnSuccess(t *testing.T) {
	ctx := context.Background()
	longURL, err := url.Parse("https://www.google.com/")
	if err != nil {
		t.Fatal(err)
	}

	code := "abcde"
	codeGenerator := fakeCodeGenerator{
		code: code,
	}

	now := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)
	want := domain.ShortURL{
		ID:        1,
		TargetURL: *longURL,
		ShortCode: code,
		CreatedAt: now,
		UpdatedAt: now,
	}
	shortURLRepo := fakeShortURLRepository{
		onGetByCode: func(ctx context.Context, code string) (domain.ShortURL, error) {
			return want, nil
		},
	}

	service := services.NewShortenerService(shortURLRepo, codeGenerator)

	got, err := service.Retrieve(ctx, code)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestShortenerService_Retrieve_OnError(t *testing.T) {
	tests := []struct {
		name      string
		repoErr   error
		targetErr error
	}{
		{
			name:      "on ErrShortURLNotFound",
			repoErr:   domain.ErrShortURLNotFound,
			targetErr: domain.ErrShortURLNotFound,
		},
		{
			name:    "on generic error",
			repoErr: errors.New("db unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			shortURLRepo := fakeShortURLRepository{
				onGetByCode: func(ctx context.Context, code string) (domain.ShortURL, error) {
					return domain.ShortURL{}, domain.ErrShortURLNotFound
				},
			}

			code := "abcde"
			codeGenerator := fakeCodeGenerator{
				code: code,
			}

			service := services.NewShortenerService(shortURLRepo, codeGenerator)

			_, err := service.Retrieve(ctx, code)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Fatalf("expected %v, got %v", tt.targetErr, err)
			}
		})
	}

}

func TestShortenerService_Update_OnSuccess(t *testing.T) {
	ctx := context.Background()
	longURL, err := url.Parse("https://www.google.com/")
	if err != nil {
		t.Fatal(err)
	}

	code := "abcde"
	codeGenerator := fakeCodeGenerator{
		code: code,
	}

	now := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)
	want := domain.ShortURL{
		ID:        1,
		TargetURL: *longURL,
		ShortCode: code,
		CreatedAt: now,
		UpdatedAt: now,
	}
	shortURLRepo := fakeShortURLRepository{
		onUpdateByCode: func(ctx context.Context, code string, target url.URL) (domain.ShortURL, error) {
			return want, nil
		},
	}

	service := services.NewShortenerService(shortURLRepo, codeGenerator)

	got, err := service.Update(ctx, code, *longURL)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestShortenerService_Update_OnError(t *testing.T) {
	tests := []struct {
		name      string
		repoErr   error
		targetErr error
	}{
		{
			name:      "on ErrShortURLNotFound",
			repoErr:   domain.ErrShortURLNotFound,
			targetErr: domain.ErrShortURLNotFound,
		},
		{
			name:    "on generic error",
			repoErr: errors.New("db unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			longURL, err := url.Parse("https://www.google.com/")
			if err != nil {
				t.Fatal(err)
			}

			code := "abcde"
			codeGenerator := fakeCodeGenerator{
				code: code,
			}

			shortURLRepo := fakeShortURLRepository{
				onUpdateByCode: func(ctx context.Context, code string, target url.URL) (domain.ShortURL, error) {
					return domain.ShortURL{}, tt.repoErr
				},
			}

			service := services.NewShortenerService(shortURLRepo, codeGenerator)

			_, err = service.Update(ctx, code, *longURL)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Fatalf("expected %v, got %v", tt.targetErr, err)
			}
		})
	}
}

func TestShortenerService_Delete_OnSuccess(t *testing.T) {
	ctx := context.Background()
	shortURLRepo := fakeShortURLRepository{
		onDeleteByCode: func(ctx context.Context, code string) error {
			return nil
		},
	}
	code := "abcde"
	codeGenerator := fakeCodeGenerator{
		code: code,
	}

	service := services.NewShortenerService(shortURLRepo, codeGenerator)

	err := service.Delete(ctx, code)
	if err != nil {
		t.Fatal(err)
	}
}

func TestShortenerService_Delete_OnError(t *testing.T) {
	tests := []struct {
		name      string
		repoErr   error
		targetErr error
	}{
		{
			name:      "on ErrShortURLNotFound",
			repoErr:   domain.ErrShortURLNotFound,
			targetErr: domain.ErrShortURLNotFound,
		},
		{
			name:    "on generic error",
			repoErr: errors.New("db unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			shortURLRepo := fakeShortURLRepository{
				onDeleteByCode: func(ctx context.Context, code string) error {
					return tt.repoErr
				},
			}
			code := "abcde"
			codeGenerator := fakeCodeGenerator{
				code: code,
			}

			service := services.NewShortenerService(shortURLRepo, codeGenerator)

			err := service.Delete(ctx, code)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Fatalf("expected %v, got %v", tt.targetErr, err)
			}
		})
	}
}

func TestShortenerService_GetStatistics_OnSuccess(t *testing.T) {
	ctx := context.Background()
	longURL, err := url.Parse("https://www.google.com/")
	if err != nil {
		t.Fatal(err)
	}

	code := "abcde"
	codeGenerator := fakeCodeGenerator{
		code: code,
	}

	now := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)
	want := domain.ShortURLStatistics{
		ShortURL: domain.ShortURL{
			ID:        1,
			TargetURL: *longURL,
			ShortCode: code,
			CreatedAt: now,
			UpdatedAt: now,
		},
		AccessCount: 1,
	}
	shortURLRepo := fakeShortURLRepository{
		onGetStatistics: func(ctx context.Context, code string) (domain.ShortURLStatistics, error) {
			return want, nil
		},
	}

	service := services.NewShortenerService(shortURLRepo, codeGenerator)

	got, err := service.GetStatistics(ctx, code)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestShortenerService_GetStatistics_OnError(t *testing.T) {
	tests := []struct {
		name      string
		repoErr   error
		targetErr error
	}{
		{
			name:      "on ErrShortURLNotFound",
			repoErr:   domain.ErrShortURLNotFound,
			targetErr: domain.ErrShortURLNotFound,
		},
		{
			name:    "on generic error",
			repoErr: errors.New("db unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			code := "abcde"
			codeGenerator := fakeCodeGenerator{
				code: code,
			}

			shortURLRepo := fakeShortURLRepository{
				onGetStatistics: func(ctx context.Context, code string) (domain.ShortURLStatistics, error) {
					return domain.ShortURLStatistics{}, tt.repoErr
				},
			}

			service := services.NewShortenerService(shortURLRepo, codeGenerator)

			_, err := service.GetStatistics(ctx, code)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.targetErr != nil && !errors.Is(err, tt.targetErr) {
				t.Fatalf("expected %v, got %v", tt.targetErr, err)
			}
		})
	}
}
