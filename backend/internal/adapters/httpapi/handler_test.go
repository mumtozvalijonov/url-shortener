package httpapi

import (
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

	service := mocks.NewMockShortenerService(t)
	service.EXPECT().
		Retrieve(mock.Anything, "abcde").
		Return(domain.ShortURL{TargetURL: *target}, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/abcde", nil)
	req.SetPathValue("shortCode", "abcde")

	recorder := httptest.NewRecorder()
	handler := NewHandler(service)

	handler.redirectFromShortCode(recorder, req)

	require.Equal(t, http.StatusTemporaryRedirect, recorder.Code)
	require.Equal(t, target.String(), recorder.Header().Get("Location"))
}
