package task_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/task"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/task/mocks"
	myMw "github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/middleware"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteHandler(t *testing.T) {
	validUserID := gofakeit.UUID()
	validTaskID := gofakeit.UUID()

	tests := []struct {
		name         string
		payload      task.DeleteRequest
		expectedCode int
		expectedBody string

		userID string

		mockSetup func(deleter *mocks.Deleter)
	}{
		{
			name: "success",
			payload: task.DeleteRequest{
				TaskID: validTaskID,
			},
			expectedCode: http.StatusNoContent,
			expectedBody: "",

			userID: validUserID,

			mockSetup: func(deleter *mocks.Deleter) {
				deleter.
					On("Delete", mock.Anything, validTaskID, validUserID).
					Return(nil)
			},
		},
		{
			name: "task not found",
			payload: task.DeleteRequest{
				TaskID: validTaskID,
			},
			expectedCode: http.StatusNotFound,
			expectedBody: `{"error":"task not found"}`,

			userID: validUserID,

			mockSetup: func(deleter *mocks.Deleter) {
				deleter.
					On("Delete", mock.Anything, validTaskID, validUserID).
					Return(services.ErrTaskNotFound)
			},
		},
		{
			name: "access denied",
			payload: task.DeleteRequest{
				TaskID: validTaskID,
			},
			expectedCode: http.StatusForbidden,
			expectedBody: `{"error":"access denied"}`,

			userID: validUserID,

			mockSetup: func(deleter *mocks.Deleter) {
				deleter.
					On("Delete", mock.Anything, validTaskID, validUserID).
					Return(services.ErrTaskAccessDenied)
			},
		},
		{
			name: "internal server error",
			payload: task.DeleteRequest{
				TaskID: validTaskID,
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"internal server error"}`,

			userID: validUserID,

			mockSetup: func(deleter *mocks.Deleter) {
				deleter.
					On("Delete", mock.Anything, validTaskID, validUserID).
					Return(errors.New("unexpected error"))
			},
		},
		{
			name: "empty user id",
			payload: task.DeleteRequest{
				TaskID: validTaskID,
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"bad request"}`,

			userID: "",

			mockSetup: nil,
		},
		{
			name: "validation error - empty task id",
			payload: task.DeleteRequest{
				TaskID: "",
			},
			expectedCode: http.StatusBadRequest,
			// зависит от твоего DecodeAndValidate, при необходимости скорректируй
			expectedBody: `{"errors":[{"field":"TaskID","error":"field is required"}]}`,

			userID: validUserID,

			mockSetup: func(deleter *mocks.Deleter) {
				// Delete не должен вызываться
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequestWithContext(
				context.WithValue(context.Background(), myMw.UserIDKey, tt.userID),
				http.MethodDelete,
				"/task",
				bytes.NewBuffer(body),
			)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			deleter := new(mocks.Deleter)
			if tt.mockSetup != nil {
				tt.mockSetup(deleter)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

			h := task.NewDeleteHandler(deleter, 4*time.Second, logger, validator.New())
			h.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedBody != "" {
				require.Equal(t, tt.expectedBody, rr.Body.String())
			}

			deleter.AssertExpectations(t)
		})
	}
}
