package delete_handler_test

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	delete_handler "github.com/RozmiDan/url_shortener/internal/http-server/handlers/delete"
	"github.com/RozmiDan/url_shortener/internal/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/stretchr/testify/mock"
	"gopkg.in/go-playground/assert.v1"
)

type MockURLDeleter struct {
	mock.Mock
}

func (m *MockURLDeleter) DeleteURL(alias string) error {
	args := m.Called(alias)
	return args.Error(0)
}

func TestDeleteHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))

	testCases := []struct {
		name             string
		alias            string
		mockErr          error
		expectedStatus   int
		expectedResponse delete_handler.Response
		expectCall       bool
	}{
		{
			name:           "success",
			alias:          "valid-alias",
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedResponse: delete_handler.Response{
				Status: "OK",
			},
			expectCall: true,
		},
		{
			name:           "not found",
			alias:          "non-existent",
			mockErr:        storage.ErrAliasNotFound,
			expectedStatus: http.StatusNotFound,
			expectedResponse: delete_handler.Response{
				Status: "Error",
				Error:  "alias not found",
			},
			expectCall: true,
		},
		{
			name:           "internal error",
			alias:          "error-alias",
			mockErr:        errors.New("some internal error"),
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: delete_handler.Response{
				Status: "Error",
				Error:  "internal error",
			},
			expectCall: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDeleter := new(MockURLDeleter)
			handler := delete_handler.NewDeleteHandler(logger, mockDeleter)

			if tc.expectCall {
				mockDeleter.On("DeleteURL", tc.alias).Return(tc.mockErr)
			}

			r := chi.NewRouter()
			r.Delete("/url/{alias}", handler)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/url/%s", tc.alias), nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			var response delete_handler.Response
			render.DecodeJSON(rec.Body, &response)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			assert.Equal(t, tc.expectedResponse, response)

			if tc.expectCall {
				mockDeleter.AssertCalled(t, "DeleteURL", tc.alias)
			} else {
				mockDeleter.AssertNotCalled(t, "DeleteURL")
			}
		})
	}
}
