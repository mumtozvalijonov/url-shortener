package httpapi

import (
	"errors"
	"net/http"

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
	router.HandleFunc("GET /{shortCode}", h.redirectFromShortCode)
}

func (h *Handler) redirectFromShortCode(w http.ResponseWriter, r *http.Request) {
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
}
