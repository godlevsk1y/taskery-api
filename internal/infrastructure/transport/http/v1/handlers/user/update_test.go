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
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateHandler(t *testing.T) {
	correctUsername := "whose_hair"
	correctEmail := gofakeit.Email()
	correctPassword := gofakeit.Password(true, true, true, true, false, 16)
	userID := gofakeit.UUID()

	tests := []struct {
		name         string
		payload      user.UpdateRequest
		expectedCode int
		expectedBody string

		mockSetup func(u *mocks.Updater)
	}{
		{
			name: "update username success",
			payload: user.UpdateRequest{
				Username: correctUsername,
				Password: correctPassword,
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"username":"` + correctUsername + `"}`,
			mockSetup: func(u *mocks.Updater) {
				u.On("ChangeUsername", mock.Anything, userID, correctUsername, correctPassword).
					Return(nil)
			},
		},
		{
			name: "update email success",
			payload: user.UpdateRequest{
				Email:    correctEmail,
				Password: correctPassword,
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"email":"` + correctEmail + `"}`,
			mockSetup: func(u *mocks.Updater) {
				u.On("ChangeEmail", mock.Anything, userID, correctEmail, correctPassword).
					Return(nil)
			},
		},
		{
			name: "both username and email update",
			payload: user.UpdateRequest{
				Username: correctUsername,
				Email:    correctEmail,
				Password: correctPassword,
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"username":"` + correctUsername + `","email":"` + correctEmail + `"}`,
			mockSetup: func(u *mocks.Updater) {
				u.On("ChangeUsername", mock.Anything, userID, correctUsername, correctPassword).
					Return(nil)
				u.On("ChangeEmail", mock.Anything, userID, correctEmail, correctPassword).
					Return(nil)
			},
		},
		{
			name: "username change fails: user not found",
			payload: user.UpdateRequest{
				Username: correctUsername,
				Password: correctPassword,
			},
			expectedCode: http.StatusNotFound,
			expectedBody: `{"error":"user not found"}`,
			mockSetup: func(u *mocks.Updater) {
				u.On("ChangeUsername", mock.Anything, userID, correctUsername, correctPassword).
					Return(services.ErrUserNotFound)
			},
		},
		{
			name: "email change fails: already taken",
			payload: user.UpdateRequest{
				Email:    correctEmail,
				Password: correctPassword,
			},
			expectedCode: http.StatusConflict,
			expectedBody: `{"error":"email already taken"}`,
			mockSetup: func(u *mocks.Updater) {
				u.On("ChangeEmail", mock.Anything, userID, correctEmail, correctPassword).
					Return(services.ErrUserEmailAlreadyTaken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/user/", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			req = req.WithContext(context.WithValue(req.Context(), "userID", userID))

			rr := httptest.NewRecorder()

			updater := new(mocks.Updater)
			tt.mockSetup(updater)

			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

			h := user.NewUpdateHandler(updater, 4*time.Second, logger, validator.New())
			h.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedCode, rr.Code)
			if tt.expectedBody != "" {
				require.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}
