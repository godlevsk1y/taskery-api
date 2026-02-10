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

func TestRegisterHandler(t *testing.T) {
	correctUsername := gofakeit.Username()
	correctEmail := gofakeit.Email()
	correctPassword := gofakeit.Password(true, true, true, true, false, 16)

	tests := []struct {
		name         string
		payload      auth.RegisterRequest
		expectedCode int
		expectedBody string

		mockSetup func(r *mocks.Registrar)
	}{
		{
			name: "success",
			payload: auth.RegisterRequest{
				Username: correctUsername,
				Email:    correctEmail,
				Password: correctPassword,
			},

			expectedCode: http.StatusCreated,
			expectedBody: `{"username":"` + correctUsername + `","email":"` + correctEmail + `"}`,

			mockSetup: func(r *mocks.Registrar) {
				r.On("Register", mock.Anything, correctUsername, correctEmail, correctPassword).
					Return(nil)
			},
		},
		{
			name: "validation error",
			payload: auth.RegisterRequest{
				Username: correctUsername,
				Email:    "not_correct",
				Password: correctPassword,
			},

			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":[{"field":"Email","error":"field is not a valid email"}]}`,

			mockSetup: func(r *mocks.Registrar) {
				r.On("Register", mock.Anything, correctUsername, "not_correct", correctPassword).
					Return(nil)
			},
		},
		{
			name: "user exists",
			payload: auth.RegisterRequest{
				Username: correctUsername,
				Email:    correctEmail,
				Password: correctPassword,
			},

			expectedCode: http.StatusConflict,
			expectedBody: `{"error":"user already exists"}`,

			mockSetup: func(r *mocks.Registrar) {
				r.On("Register", mock.Anything, correctUsername, correctEmail, correctPassword).
					Return(services.ErrUserExists)
			},
		},
		{
			name: "internal server error",
			payload: auth.RegisterRequest{
				Username: correctUsername,
				Email:    correctEmail,
				Password: correctPassword,
			},

			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"register failed"}`,

			mockSetup: func(r *mocks.Registrar) {
				r.On("Register", mock.Anything, correctUsername, correctEmail, correctPassword).
					Return(services.ErrUserRegisterFailed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			registrar := new(mocks.Registrar)
			tt.mockSetup(registrar)

			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

			h := auth.NewRegisterHandler(registrar, 4*time.Second, logger, validator.New())
			h.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedBody != "" {
				require.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}
