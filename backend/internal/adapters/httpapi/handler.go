package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/mumtozvalijonov/url-shortener/internal/domain"
	"github.com/mumtozvalijonov/url-shortener/internal/ports"
)

type Handler struct {
	shortURLService ports.ShortenerService
}

func NewHandler(shortURLService ports.ShortenerService) *Handler {
	return &Handler{shortURLService: shortURLService}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.Handle("POST /", h.handleCreateShortURL())
	router.Handle("GET /{shortCode}", h.handleRedirectFromShortCode())
	router.Handle("PATCH /{shortCode}", h.handleUpdateShortURL())
}

func (h *Handler) handleCreateShortURL() http.Handler {
	type requestDTO struct {
		TargetURL string `json:"target_url"`
	}
	type responseDTO struct {
		ShortCode string `json:"short_code"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var request requestDTO
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&request); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			targetURL, err := url.Parse(request.TargetURL)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			shortURL, err := h.shortURLService.Shorten(r.Context(), *targetURL)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			data := responseDTO{
				ShortCode: shortURL.ShortCode,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(data)
		},
	)
}

func (h *Handler) handleRedirectFromShortCode() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			shortCode := r.PathValue("shortCode")

			shortURL, err := h.shortURLService.Retrieve(r.Context(), shortCode)
			if err != nil {
				switch {
				case errors.Is(err, domain.ErrShortURLNotFound):
					w.WriteHeader(http.StatusNotFound)
				default:
					w.WriteHeader(http.StatusInternalServerError)
				}
				return
			}

			http.Redirect(w, r, shortURL.TargetURL.String(), http.StatusTemporaryRedirect)
		},
	)
}

func (h *Handler) handleUpdateShortURL() http.Handler {
	type requestDTO struct {
		TargetURL string `json:"target_url"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			shortCode := r.PathValue("shortCode")

			var request requestDTO
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&request); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			targetURL, err := url.Parse(request.TargetURL)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if _, err := h.shortURLService.Update(r.Context(), shortCode, *targetURL); err != nil {
				switch {
				case errors.Is(err, domain.ErrShortURLNotFound):
					w.WriteHeader(http.StatusNotFound)
				default:
					w.WriteHeader(http.StatusInternalServerError)
				}
				return
			}

			w.WriteHeader(http.StatusOK)
		},
	)
}
