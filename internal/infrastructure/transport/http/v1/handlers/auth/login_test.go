package auth_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/auth"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/auth/mocks"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLoginHandler(t *testing.T) {
	correctEmail := gofakeit.Email()
	correctPassword := gofakeit.Password(true, true, true, true, false, 16)
	token := "some.jwt.token"

	tests := []struct {
		name         string
		payload      auth.LoginRequest
		expectedCode int
		expectedBody string

		mockSetup func(a *mocks.Authenticator)
	}{
		{
			name: "success",
			payload: auth.LoginRequest{
				Email:    correctEmail,
				Password: correctPassword,
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"token":"` + token + `"}`,
			mockSetup: func(a *mocks.Authenticator) {
				a.On("Login", mock.Anything, correctEmail, correctPassword).
					Return(token, nil)
			},
		},
		{
			name: "validation error",
			payload: auth.LoginRequest{
				Email:    "invalid_email",
				Password: correctPassword,
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":[{"field":"Email","error":"field is not a valid email"}]}`,
			mockSetup: func(a *mocks.Authenticator) {
				a.On("Login", mock.Anything, "invalid_email", correctPassword).
					Return("", nil)
			},
		},
		{
			name: "user not found",
			payload: auth.LoginRequest{
				Email:    correctEmail,
				Password: correctPassword,
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"invalid credentials"}`,
			mockSetup: func(a *mocks.Authenticator) {
				a.On("Login", mock.Anything, correctEmail, correctPassword).
					Return("", services.ErrUserNotFound)
			},
		},
		{
			name: "user unauthorized",
			payload: auth.LoginRequest{
				Email:    correctEmail,
				Password: correctPassword,
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"invalid credentials"}`,
			mockSetup: func(a *mocks.Authenticator) {
				a.On("Login", mock.Anything, correctEmail, correctPassword).
					Return("", services.ErrUserUnauthorized)
			},
		},
		{
			name: "internal error",
			payload: auth.LoginRequest{
				Email:    correctEmail,
				Password: correctPassword,
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"login failed"}`,
			mockSetup: func(a *mocks.Authenticator) {
				a.On("Login", mock.Anything, correctEmail, correctPassword).
					Return("", services.ErrUserLoginFailed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			authenticator := new(mocks.Authenticator)
			tt.mockSetup(authenticator)

			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
			h := auth.NewLoginHandler(authenticator, 4*time.Second, logger, validator.New())
			h.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedCode, rr.Code)
			if tt.expectedBody != "" {
				require.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}
