package task_test

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
	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/models"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/task"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/task/mocks"
	myMw "github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/middleware"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindByOwnerHandler(t *testing.T) {
	validUserID := gofakeit.UUID()
	validTaskID := gofakeit.UUID()

	tests := []struct {
		name         string
		expectedCode int
		expectedBody string
		userID       string
		mockSetup    func(finder *mocks.Finder)
	}{
		{
			name:         "empty user id",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"bad request"}`,
			userID:       "",
			mockSetup:    nil,
		},
		{
			name:         "internal error",
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"internal server error"}`,
			userID:       validUserID,
			mockSetup: func(finder *mocks.Finder) {
				finder.On("FindByOwner", mock.Anything, validUserID).
					Return(nil, services.ErrTaskFindByOwnerFailed)
			},
		},
		{
			name:         "success with tasks",
			expectedCode: http.StatusOK,
			expectedBody: func() string {
				taskDTOs, _ := json.Marshal([]task.TaskDTO{
					{
						ID:          validTaskID,
						Title:       "Test task",
						Description: "Test description",
						Deadline:    nil,
						IsCompleted: false,
						CompletedAt: nil,
					},
				})
				return `{"owner_id":"` + validUserID + `","tasks":` + string(taskDTOs) + `}`
			}(),
			userID: validUserID,
			mockSetup: func(finder *mocks.Finder) {
				params := models.TaskFromDBParams{
					ID:          validTaskID,
					OwnerID:     validUserID,
					Title:       "Test task",
					Description: "Test description",
					Deadline:    nil,
					IsCompleted: false,
					CompletedAt: nil,
				}
				task, err := models.NewTaskFromDB(params)
				require.NoError(t, err)
				finder.On("FindByOwner", mock.Anything, validUserID).
					Return([]*models.Task{task}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(
				context.WithValue(context.Background(), myMw.UserIDKey, tt.userID),
				http.MethodGet,
				"/task/owner",
				bytes.NewBufferString(""),
			)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			finder := new(mocks.Finder)
			if tt.mockSetup != nil {
				tt.mockSetup(finder)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

			h := task.NewFindByOwnerHandler(finder, 4*time.Second, logger, validator.New())
			h.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedCode, rr.Code)
			require.JSONEq(t, tt.expectedBody, rr.Body.String())

			if tt.mockSetup != nil {
				finder.AssertExpectations(t)
			}
		})
	}
}
