package redirect_handler_test

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	redirect_handler "github.com/RozmiDan/url_shortener/internal/http-server/handlers/redirect"
	"github.com/RozmiDan/url_shortener/internal/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/stretchr/testify/mock"
	"gopkg.in/go-playground/assert.v1"
)

type MockURLGetter struct {
	mock.Mock
}

func (m *MockURLGetter) GetURL(alias string) (string, error) {
	args := m.Called(alias)
	return args.Get(0).(string), args.Error(1)
}

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func TestGetHandler(t *testing.T) {

	testCases := []struct {
		name             string
		alias            string
		mockURL          string
		mockErr          error
		expectedStatus   int
		expectedResponse redirect_handler.Response
		expectCall       bool
	}{
		{
			name:           "success",
			alias:          "valid-alias",
			mockURL:        "https://google.com",
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedResponse: redirect_handler.Response{
				Status: "OK",
				URL:    "https://google.com",
			},
			expectCall: true,
		},
		{
			name:           "not found",
			alias:          "non-existent",
			mockURL:        "",
			mockErr:        storage.ErrURLNotFound,
			expectedStatus: http.StatusNotFound,
			expectedResponse: redirect_handler.Response{
				Status: "Error",
				Error:  "URL not found",
			},
			expectCall: true,
		},
		{
			name:           "internal error",
			alias:          "error-alias",
			mockURL:        "",
			mockErr:        errors.New("some internal error"),
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: redirect_handler.Response{
				Status: "Error",
				Error:  "internal error",
			},
			expectCall: true,
		},
	}

	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockGetter := new(MockURLGetter)
			handler := redirect_handler.NewRedirectHandler(logger, mockGetter)

			if tc.expectCall {
				mockGetter.On("GetURL", tc.alias).Return(tc.mockURL, tc.mockErr)
			}

			r := chi.NewRouter()
			r.Get("/{alias}", handler)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", tc.alias), nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			var response redirect_handler.Response
			render.DecodeJSON(rec.Body, &response)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			assert.Equal(t, tc.expectedResponse, response)

			if tc.expectCall {
				mockGetter.AssertCalled(t, "GetURL", tc.alias)
			} else {
				mockGetter.AssertNotCalled(t, "GetURL")
			}
		})
	}
}
