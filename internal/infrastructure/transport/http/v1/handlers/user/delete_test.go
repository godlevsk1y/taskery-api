package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/user"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/user/mocks"
	myMw "github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/middleware"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteHandler(t *testing.T) {
	correctUserID := gofakeit.UUID()
	correctPassword := gofakeit.Password(true, true, true, true, false, 16)

	tests := []struct {
		name string

		userID  string
		payload user.DeleteRequest

		expectedCode int
		expectedBody string

		mockSetup func(d *mocks.Deleter)
	}{
		{
			name: "success",

			userID: correctUserID,
			payload: user.DeleteRequest{
				Password: correctPassword,
			},

			expectedCode: http.StatusNoContent,
			expectedBody: "",

			mockSetup: func(d *mocks.Deleter) {
				d.On("Delete", mock.Anything, correctUserID, correctPassword).
					Return(nil)
			},
		},
		{
			name: "user not found",

			userID: gofakeit.UUID(),
			payload: user.DeleteRequest{
				Password: correctPassword,
			},

			expectedCode: http.StatusNotFound,
			expectedBody: `{"error":"user not found"}`,

			mockSetup: func(d *mocks.Deleter) {
				d.On("Delete", mock.Anything, mock.AnythingOfType("string"), correctPassword).
					Return(services.ErrUserNotFound)
			},
		},
		{
			name: "invalid password",

			userID: correctUserID,
			payload: user.DeleteRequest{
				Password: gofakeit.Password(true, true, true, true, false, 16),
			},

			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"unauthorized"}`,

			mockSetup: func(d *mocks.Deleter) {
				d.On("Delete", mock.Anything, correctUserID, mock.AnythingOfType("string")).
					Return(services.ErrUserUnauthorized)
			},
		},
		{
			name: "internal server error",

			userID: correctUserID,
			payload: user.DeleteRequest{
				Password: correctPassword,
			},

			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"internal server error"}`,

			mockSetup: func(d *mocks.Deleter) {
				d.On("Delete", mock.Anything, correctUserID, correctPassword).
					Return(services.ErrUserDeleteFailed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodDelete, "/user/", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			req = req.WithContext(context.WithValue(req.Context(), myMw.UserIDKey, tt.userID))

			rr := httptest.NewRecorder()

			deleter := new(mocks.Deleter)
			tt.mockSetup(deleter)

			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

			h := user.NewDeleteHandler(deleter, 4*time.Second, logger, validator.New())
			h.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedCode, rr.Code)
			if tt.expectedBody != "" {
				require.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}
