package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/vo"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/cyberbrain-dev/taskery-api/internal/services/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewTaskService(t *testing.T) {
	tests := []struct {
		name      string
		tasksRepo services.TaskRepository
		wantErr   error
	}{
		{
			name:      "success",
			tasksRepo: new(mocks.TaskRepository),
			wantErr:   nil,
		},
		{
			name:      "nil tasks repo",
			tasksRepo: nil,
			wantErr:   services.ErrTaskRepositoryNil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := services.NewTaskService(tt.tasksRepo)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, service)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

func TestTaskService_Create(t *testing.T) {
	realUserID := uuid.New()
	validDeadline := time.Now().Add(1 * time.Hour)
	invalidDeadline := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name    string
		cmd     services.CreateTaskCommand
		wantErr error

		mocksSetup func(repo *mocks.TaskRepository)
	}{
		{
			name: "success without deadline",
			cmd: services.CreateTaskCommand{
				Title:       "title",
				Description: "description",
				OwnerID:     realUserID,
				Deadline:    nil,
			},
			wantErr: nil,

			mocksSetup: func(repo *mocks.TaskRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(nil)
			},
		},
		{
			name: "success with deadline",
			cmd: services.CreateTaskCommand{
				Title:       "title",
				Description: "description",
				OwnerID:     realUserID,
				Deadline:    &validDeadline,
			},
			wantErr: nil,

			mocksSetup: func(repo *mocks.TaskRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(nil)
			},
		},
		{
			name: "task already exists",
			cmd: services.CreateTaskCommand{
				Title:       "title",
				Description: "description",
				OwnerID:     realUserID,
				Deadline:    &validDeadline,
			},
			wantErr: services.ErrTaskExists,

			mocksSetup: func(repo *mocks.TaskRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(services.ErrTaskRepoExists)
			},
		},
		{
			name: "invalid deadline",
			cmd: services.CreateTaskCommand{
				Title:       "title",
				Description: "description",
				OwnerID:     realUserID,
				Deadline:    &invalidDeadline,
			},
			wantErr: vo.ErrDeadlineBeforeNow,
		},
		{
			name: "internal error",
			cmd: services.CreateTaskCommand{
				Title:       "title",
				Description: "description",
				OwnerID:     realUserID,
				Deadline:    &validDeadline,
			},
			wantErr: services.ErrTaskCreateFailed,

			mocksSetup: func(repo *mocks.TaskRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(errors.New("failed to load db"))
			},
		},
		{
			name: "owner not found",
			cmd: services.CreateTaskCommand{
				Title:       "title",
				Description: "description",
				OwnerID:     uuid.New(),
				Deadline:    &validDeadline,
			},
			wantErr: services.ErrTaskOwnerNotFound,

			mocksSetup: func(repo *mocks.TaskRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(services.ErrTaskRepoOwnerNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mocks.TaskRepository)
			if tt.mocksSetup != nil {
				tt.mocksSetup(repo)
			}

			service, err := services.NewTaskService(repo)
			require.NoError(t, err)
			require.NotNil(t, service)

			ctx := context.Background()
			err = service.Create(ctx, tt.cmd)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
