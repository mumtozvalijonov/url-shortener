package httpapi

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/mumtozvalijonov/url-shortener/internal/domain"
	"github.com/mumtozvalijonov/url-shortener/internal/ports/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_RedirectFromShortCode(t *testing.T) {
	target, err := url.Parse("https://www.google.com/")
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		service := mocks.NewMockShortenerService(t)
		handler := NewHandler(service)
		router := http.NewServeMux()
		handler.RegisterRoutes(router)

		service.EXPECT().
			Retrieve(mock.Anything, "abcde").
			Return(domain.ShortURL{TargetURL: *target}, nil).
			Once()

		req := httptest.NewRequest(http.MethodGet, "/abcde", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
		require.Equal(t, target.String(), recorder.Header().Get("Location"))
	})
	t.Run("url not found", func(t *testing.T) {
		t.Parallel()
		service := mocks.NewMockShortenerService(t)
		handler := NewHandler(service)
		router := http.NewServeMux()
		handler.RegisterRoutes(router)

		service.EXPECT().
			Retrieve(mock.Anything, "qwert").
			Return(domain.ShortURL{}, domain.ErrShortURLNotFound).
			Once()

		req := httptest.NewRequest(http.MethodGet, "/qwert", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNotFound, recorder.Code)
	})
}

func TestHandler_CreateShortURL(t *testing.T) {
	target, err := url.Parse("https://www.google.com/")
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		service := mocks.NewMockShortenerService(t)
		handler := NewHandler(service)
		router := http.NewServeMux()
		handler.RegisterRoutes(router)

		service.EXPECT().
			Shorten(mock.Anything, *target).
			Return(
				domain.ShortURL{
					ID: 1, ShortCode: "abcde",
					TargetURL: *target,
				},
				nil,
			).
			Once()

		req := httptest.NewRequest(
			http.MethodPost, "/",
			bytes.NewReader([]byte(`{"target_url": "https://www.google.com/"}`)),
		)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusCreated, recorder.Code)
	})
	t.Run("invalid url", func(t *testing.T) {
		t.Parallel()
		service := mocks.NewMockShortenerService(t)
		handler := NewHandler(service)
		router := http.NewServeMux()
		handler.RegisterRoutes(router)

		req := httptest.NewRequest(
			http.MethodPost, "/",
			bytes.NewReader([]byte(`{"target_url": "noprotocol://definitely not a url"}`)),
		)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}

func TestHandler_UpdateShortURL(t *testing.T) {
	target, err := url.Parse("https://www.google.com/")
	require.NoError(t, err)

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		service := mocks.NewMockShortenerService(t)
		handler := NewHandler(service)
		router := http.NewServeMux()
		handler.RegisterRoutes(router)

		service.EXPECT().
			Update(mock.Anything, "abcde", *target).
			Return(
				domain.ShortURL{
					ID: 1, ShortCode: "abcde",
					TargetURL: *target,
				},
				nil,
			).
			Once()

		req := httptest.NewRequest(
			http.MethodPatch, "/abcde",
			bytes.NewReader([]byte(`{"target_url": "https://www.google.com/"}`)),
		)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
	})
	t.Run("url not found", func(t *testing.T) {
		t.Parallel()
		service := mocks.NewMockShortenerService(t)
		handler := NewHandler(service)
		router := http.NewServeMux()
		handler.RegisterRoutes(router)

		service.EXPECT().
			Update(mock.Anything, "qwert", *target).
			Return(domain.ShortURL{}, domain.ErrShortURLNotFound)

		req := httptest.NewRequest(
			http.MethodPatch, "/qwert",
			bytes.NewReader([]byte(`{"target_url": "https://www.google.com/"}`)),
		)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNotFound, recorder.Code)
	})
	t.Run("unexpected error", func(t *testing.T) {
		t.Parallel()
		service := mocks.NewMockShortenerService(t)
		handler := NewHandler(service)
		router := http.NewServeMux()
		handler.RegisterRoutes(router)

		service.EXPECT().
			Update(mock.Anything, "qwert", *target).
			Return(domain.ShortURL{}, errors.New("unexpected error"))

		req := httptest.NewRequest(
			http.MethodPatch, "/qwert",
			bytes.NewReader([]byte(`{"target_url": "https://www.google.com/"}`)),
		)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusInternalServerError, recorder.Code)
	})
}
