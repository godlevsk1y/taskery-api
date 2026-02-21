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
	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/vo"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/task"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers/task/mocks"
	myMw "github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/middleware"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateHandler(t *testing.T) {
	validDeadline := time.Now().Add(time.Hour)
	validUserID := gofakeit.UUID()

	tests := []struct {
		name         string
		payload      task.CreateRequest
		expectedCode int
		expectedBody string

		userID string

		mockSetup func(creator *mocks.Creator)
	}{
		{
			name: "success without validDeadline",
			payload: task.CreateRequest{
				Title:       "Do homework",
				Description: "Some description",
				Deadline:    nil,
			},

			expectedCode: http.StatusCreated,
			expectedBody: `{"title":"Do homework"}`,

			userID: validUserID,

			mockSetup: func(creator *mocks.Creator) {
				creator.On("Create", mock.Anything, mock.AnythingOfType("services.CreateTaskCommand")).
					Return(nil)
			},
		},
		{
			name: "success without deadline",
			payload: task.CreateRequest{
				Title:       "Do homework",
				Description: "Some description",
				Deadline:    &validDeadline,
			},

			expectedCode: http.StatusCreated,
			expectedBody: `{"title":"Do homework"}`,

			userID: validUserID,

			mockSetup: func(creator *mocks.Creator) {
				creator.On("Create", mock.Anything, mock.AnythingOfType("services.CreateTaskCommand")).
					Return(nil)
			},
		},
		{
			name: "task already exists",
			payload: task.CreateRequest{
				Title:       "Do homework",
				Description: "Some description",
				Deadline:    nil,
			},
			expectedCode: http.StatusConflict,
			expectedBody: `{"error":"task already exists"}`,

			userID: validUserID,

			mockSetup: func(creator *mocks.Creator) {
				creator.On("Create", mock.Anything, mock.AnythingOfType("services.CreateTaskCommand")).
					Return(services.ErrTaskExists)
			},
		},
		{
			name: "empty title",
			payload: task.CreateRequest{
				Title:       "",
				Description: "Some description",
				Deadline:    nil,
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errors":[{"field":"Title","error":"field is required"}]}`,

			userID: validUserID,

			mockSetup: func(creator *mocks.Creator) {
				creator.On("Create", mock.Anything, mock.AnythingOfType("services.CreateTaskCommand")).
					Return(vo.ErrTitleEmpty)
			},
		},
		{
			name: "user not exists",
			payload: task.CreateRequest{
				Title:       "Do homework",
				Description: "Some description",
				Deadline:    nil,
			},

			expectedCode: http.StatusNotFound,
			expectedBody: `{"error":"task owner not found"}`,

			userID: gofakeit.UUID(),

			mockSetup: func(creator *mocks.Creator) {
				creator.On("Create", mock.Anything, mock.AnythingOfType("services.CreateTaskCommand")).
					Return(services.ErrTaskOwnerNotFound)
			},
		},
		{
			name: "internal server error",
			payload: task.CreateRequest{
				Title:       "Do homework",
				Description: "Some description",
				Deadline:    nil,
			},

			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"task creation failed"}`,

			userID: validUserID,

			mockSetup: func(creator *mocks.Creator) {
				creator.On("Create", mock.Anything, mock.AnythingOfType("services.CreateTaskCommand")).
					Return(services.ErrTaskCreateFailed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequestWithContext(
				context.WithValue(context.Background(), myMw.UserContextKey, tt.userID),
				http.MethodPost, "/task",
				bytes.NewBuffer(body),
			)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			creator := new(mocks.Creator)
			if tt.mockSetup != nil {
				tt.mockSetup(creator)
			}

			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

			h := task.NewCreateHandler(creator, 4*time.Second, logger, validator.New())
			h.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedCode, rr.Code)
			require.Equal(t, tt.expectedBody, rr.Body.String())
		})
	}
}
