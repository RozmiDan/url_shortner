package update_handler_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	update_handler "github.com/RozmiDan/url_shortener/internal/http-server/handlers/update"
	"github.com/RozmiDan/url_shortener/internal/storage"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockURLUpdater struct {
	mock.Mock
}

func (m *MockURLUpdater) UpdateURL(currAlias string, newAlias string) error {
	args := m.Called(currAlias, newAlias)
	return args.Error(0)
}

func TestUpdateHandlerIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	testCases := []struct {
		name             string
		currAlias        string
		newAlias         string
		mockErr          error
		expectedStatus   int
		expectedContains string
		expectUpdateCall bool
	}{
		{
			name:             "successful update",
			currAlias:        "oldAlias",
			newAlias:         "newAlias",
			mockErr:          nil,
			expectedStatus:   http.StatusOK,
			expectedContains: `"status":"OK"`,
			expectUpdateCall: true,
		},
		{
			name:             "alias already exists",
			currAlias:        "alias1",
			newAlias:         "alias2",
			mockErr:          storage.ErrAliasExists,
			expectedStatus:   http.StatusConflict,
			expectedContains: `"error":"alias already exists"`,
			expectUpdateCall: true,
		},
		{
			name:             "alias not found",
			currAlias:        "notfound",
			newAlias:         "updateMe",
			mockErr:          storage.ErrAliasNotFound,
			expectedStatus:   http.StatusNotFound,
			expectedContains: `"error":"alias not found"`,
			expectUpdateCall: true,
		},
		{
			name:             "empty newAlias",
			currAlias:        "aliasX",
			newAlias:         "",
			mockErr:          nil,
			expectedStatus:   http.StatusBadRequest,
			expectedContains: `"error":"new alias is required"`,
			expectUpdateCall: false,
		},
		{
			name:             "same alias",
			currAlias:        "same",
			newAlias:         "same",
			mockErr:          nil,
			expectedStatus:   http.StatusBadRequest,
			expectedContains: `"error":"new alias must be different"`,
			expectUpdateCall: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUpdater := new(MockURLUpdater)
			handler := update_handler.NewUpdateHandler(logger, mockUpdater)

			// Настраиваем ожидания мока только если вызов ожидается
			if tc.expectUpdateCall {
				mockUpdater.On("UpdateURL", tc.currAlias, tc.newAlias).Return(tc.mockErr)
			}

			input := fmt.Sprintf(`{"newAlias": "%s"}`, tc.newAlias)

			// Создаём маршрутизатор и регистрируем хендлер
			r := chi.NewRouter()
			r.Put("/url/{alias}", handler)

			req := httptest.NewRequest(http.MethodPut, "/url/"+tc.currAlias, bytes.NewReader([]byte(input)))
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expectedContains)

			// Проверяем вызовы мока только если они ожидаются
			if tc.expectUpdateCall {
				mockUpdater.AssertCalled(t, "UpdateURL", tc.currAlias, tc.newAlias)
			} else {
				mockUpdater.AssertNotCalled(t, "UpdateURL")
			}
		})
	}
}
