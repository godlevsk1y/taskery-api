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

func TestUpdateHandler(t *testing.T) {
	validUserID := gofakeit.UUID()
	validTaskID := gofakeit.UUID()

	validTitle := gofakeit.Sentence(10)
	validDescription := gofakeit.Paragraph(1, 5, 10, " ")
	validDeadline := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name         string
		payload      task.UpdateRequest
		expectedCode int
		expectedBody string
		userID       string
		mockSetup    func(updater *mocks.Updater)
	}{
		{
			name: "successful update with all fields",
			payload: task.UpdateRequest{
				TaskID:      validTaskID,
				Title:       &validTitle,
				Description: &validDescription,
				Deadline:    &validDeadline,
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"task_id":"` + validTaskID + `"}`,
			userID:       validUserID,
			mockSetup: func(updater *mocks.Updater) {
				updater.On("Update", mock.Anything, validTaskID, validUserID, mock.AnythingOfType("services.UpdateTaskCommand")).
					Return(nil)
			},
		},
		{
			name: "successful update with one field",
			payload: task.UpdateRequest{
				TaskID: validTaskID,
				Title:  &validTitle,
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"task_id":"` + validTaskID + `"}`,
			userID:       validUserID,
			mockSetup: func(updater *mocks.Updater) {
				updater.On("Update", mock.Anything, validTaskID, validUserID, services.UpdateTaskCommand{
					Title: &validTitle,
				}).Return(nil)
			},
		},
		{
			name: "task not found",
			payload: task.UpdateRequest{
				TaskID:      validTaskID,
				Title:       &validTitle,
				Description: &validDescription,
				Deadline:    &validDeadline,
			},
			expectedCode: http.StatusNotFound,
			expectedBody: `{"error":"task was not found"}`,
			userID:       validUserID,
			mockSetup: func(updater *mocks.Updater) {
				updater.On("Update", mock.Anything, validTaskID, validUserID, mock.Anything).
					Return(services.ErrTaskNotFound)
			},
		},
		{
			name: "access forbidden",
			payload: task.UpdateRequest{
				TaskID:      validTaskID,
				Title:       &validTitle,
				Description: &validDescription,
				Deadline:    &validDeadline,
			},
			expectedCode: http.StatusForbidden,
			expectedBody: `{"error":"task access denied"}`,
			userID:       validUserID,
			mockSetup: func(updater *mocks.Updater) {
				updater.On("Update", mock.Anything, validTaskID, validUserID, mock.Anything).
					Return(services.ErrTaskAccessDenied)
			},
		},
		{
			name: "internal error",
			payload: task.UpdateRequest{
				TaskID:      validTaskID,
				Title:       &validTitle,
				Description: &validDescription,
				Deadline:    &validDeadline,
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"internal server error"}`,
			userID:       validUserID,
			mockSetup: func(updater *mocks.Updater) {
				updater.On("Update", mock.Anything, validTaskID, validUserID, mock.Anything).
					Return(services.ErrTaskUpdateFailed)
			},
		},
		{
			name: "validation error",
			payload: task.UpdateRequest{
				TaskID: "",
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":[{"field":"TaskID","error":"field is required"}]}`,
			userID:       validUserID,
			mockSetup:    nil,
		},
		{
			name: "domain error - empty title",
			payload: task.UpdateRequest{
				TaskID: validTaskID,
				Title:  new(""),
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"title cannot be empty"}`,
			userID:       validUserID,
			mockSetup: func(updater *mocks.Updater) {
				updater.On("Update", mock.Anything, validTaskID, validUserID, mock.AnythingOfType("services.UpdateTaskCommand")).
					Return(errors.New("title cannot be empty"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequestWithContext(
				context.WithValue(context.Background(), myMw.UserIDKey, tt.userID),
				http.MethodPut,
				"/task",
				bytes.NewBuffer(body),
			)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			updater := new(mocks.Updater)
			if tt.mockSetup != nil {
				tt.mockSetup(updater)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

			h := task.NewUpdateHandler(updater, 4*time.Second, logger, validator.New())
			h.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedBody != "" {
				require.Equal(t, tt.expectedBody, rr.Body.String())
			}

			if tt.mockSetup != nil {
				updater.AssertExpectations(t)
			}
		})
	}
}
