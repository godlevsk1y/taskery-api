package services_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/models"
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

func TestTaskService_ChangeTitle(t *testing.T) {
	realTaskID := uuid.New()
	realOwnerID := uuid.New()
	newTitle := "New Task Title"

	tests := []struct {
		name     string
		id       string
		ownerID  string
		newTitle string

		wantErr error

		mocksSetup func(repo *mocks.TaskRepository, taskToReturn *models.Task)
	}{
		{
			name:     "success",
			id:       realTaskID.String(),
			ownerID:  realOwnerID.String(),
			newTitle: newTitle,
			wantErr:  nil,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(nil)
			},
		},
		{
			name:     "task not found",
			id:       realTaskID.String(),
			ownerID:  realOwnerID.String(),
			newTitle: newTitle,
			wantErr:  services.ErrTaskNotFound,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(nil, services.ErrTaskRepoNotFound)
			},
		},
		{
			name:     "access denied",
			id:       realTaskID.String(),
			ownerID:  uuid.New().String(),
			newTitle: newTitle,
			wantErr:  services.ErrTaskAccessDenied,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)
			},
		},
		{
			name:     "change title validation error",
			id:       realTaskID.String(),
			ownerID:  realOwnerID.String(),
			newTitle: "",
			wantErr:  vo.ErrTitleEmpty,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)
			},
		},
		{
			name:     "update repository error",
			id:       realTaskID.String(),
			ownerID:  realOwnerID.String(),
			newTitle: newTitle,
			wantErr:  services.ErrTaskChangeTitleFailed,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(errors.New("repo failure"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskToReturn, err := models.NewTaskFromDB(models.TaskFromDBParams{
				ID:          realTaskID.String(),
				OwnerID:     realOwnerID.String(),
				Title:       "Old Title",
				Description: "Some Description",
				Deadline:    nil,
				IsCompleted: false,
				CompletedAt: nil,
			})
			require.NoError(t, err)
			require.NotNil(t, taskToReturn)

			repo := new(mocks.TaskRepository)
			if tt.mocksSetup != nil {
				tt.mocksSetup(repo, taskToReturn)
			}

			service, err := services.NewTaskService(repo)
			require.NoError(t, err)
			require.NotNil(t, service)

			ctx := context.Background()
			err = service.ChangeTitle(ctx, tt.id, tt.ownerID, tt.newTitle)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.Equal(t, tt.newTitle, taskToReturn.Title().String())
			require.NoError(t, err)
		})
	}
}

func TestTaskService_ChangeDescription(t *testing.T) {
	realTaskID := uuid.New()
	realOwnerID := uuid.New()
	newDescription := "Updated description"

	tests := []struct {
		name        string
		id          string
		ownerID     string
		newDesc     string
		taskOwnerID uuid.UUID

		wantErr error

		mocksSetup func(repo *mocks.TaskRepository, taskToReturn *models.Task)
	}{
		{
			name:        "success",
			id:          realTaskID.String(),
			ownerID:     realOwnerID.String(),
			newDesc:     newDescription,
			taskOwnerID: realOwnerID,

			wantErr: nil,

			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(nil)
			},
		},
		{
			name:        "task not found",
			id:          realTaskID.String(),
			ownerID:     realOwnerID.String(),
			newDesc:     newDescription,
			taskOwnerID: realOwnerID,

			wantErr: services.ErrTaskNotFound,

			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(nil, services.ErrTaskRepoNotFound)
			},
		},
		{
			name:        "access denied",
			id:          realTaskID.String(),
			ownerID:     uuid.New().String(), // different owner
			newDesc:     newDescription,
			taskOwnerID: realOwnerID,

			wantErr: services.ErrTaskAccessDenied,

			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)
			},
		},
		{
			name:        "update fails",
			id:          realTaskID.String(),
			ownerID:     realOwnerID.String(),
			newDesc:     newDescription,
			taskOwnerID: realOwnerID,

			wantErr: services.ErrTaskChangeDescriptionFailed,

			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(fmt.Errorf("some error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskToReturn, err := models.NewTaskFromDB(models.TaskFromDBParams{
				ID:          realTaskID.String(),
				OwnerID:     tt.taskOwnerID.String(),
				Title:       "Some Title",
				Description: "Some Description",
				Deadline:    nil,
				IsCompleted: false,
				CompletedAt: nil,
			})
			require.NoError(t, err)
			require.NotNil(t, taskToReturn)

			repo := new(mocks.TaskRepository)
			if tt.mocksSetup != nil {
				tt.mocksSetup(repo, taskToReturn)
			}

			service, err := services.NewTaskService(repo)
			require.NoError(t, err)
			require.NotNil(t, service)

			ctx := context.Background()
			err = service.ChangeDescription(ctx, tt.id, tt.ownerID, tt.newDesc)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.Equal(t, tt.newDesc, taskToReturn.Description().String())
			require.NoError(t, err)
		})
	}
}

func TestTaskService_SetDeadline(t *testing.T) {
	realTaskID := uuid.New()
	realOwnerID := uuid.New()
	validDeadline := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name             string
		id               string
		previousDeadline *time.Time
		deadline         time.Time

		wantErr error

		mocksSetup func(repo *mocks.TaskRepository, taskToReturn *models.Task)
	}{
		{
			name:             "success without deadline before",
			id:               realTaskID.String(),
			previousDeadline: nil,
			deadline:         time.Now().Add(1 * time.Hour),

			wantErr: nil,

			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(nil)
			},
		},
		{
			name:             "success with existing deadline before",
			id:               realTaskID.String(),
			previousDeadline: &validDeadline,
			deadline:         time.Now().Add(1 * time.Hour),

			wantErr: nil,

			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(nil)
			},
		},
		{
			name:     "invalid deadline",
			id:       realTaskID.String(),
			deadline: time.Now().Add(-1 * time.Hour),

			wantErr: vo.ErrDeadlineBeforeNow,

			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskToReturn, err := models.NewTaskFromDB(models.TaskFromDBParams{
				ID:          realTaskID.String(),
				OwnerID:     realOwnerID.String(),
				Title:       "Some Title",
				Description: "Some Description",
				Deadline:    tt.previousDeadline,
				IsCompleted: false,
				CompletedAt: nil,
			})
			require.NoError(t, err)
			require.NotNil(t, taskToReturn)

			repo := new(mocks.TaskRepository)
			if tt.mocksSetup != nil {
				tt.mocksSetup(repo, taskToReturn)
			}

			service, err := services.NewTaskService(repo)
			require.NoError(t, err)
			require.NotNil(t, service)

			ctx := context.Background()
			err = service.SetDeadline(ctx, tt.id, realOwnerID.String(), tt.deadline)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTaskService_Complete(t *testing.T) {
	realTaskID := uuid.New()
	realOwnerID := uuid.New()
	notRealTaskID := uuid.New()

	realCompletedAt := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name    string
		id      string
		ownerID string

		wasCompleted bool

		wantErr error

		mocksSetup func(repo *mocks.TaskRepository, taskToReturn *models.Task)
	}{
		{
			name:         "success if not completed",
			id:           realTaskID.String(),
			ownerID:      realOwnerID.String(),
			wasCompleted: false,
			wantErr:      nil,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(nil)
			},
		},
		{
			name:         "success if already completed",
			id:           realTaskID.String(),
			ownerID:      realOwnerID.String(),
			wasCompleted: true,
			wantErr:      nil,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Maybe().
					Return(nil)
			},
		},
		{
			name:         "task not found",
			id:           notRealTaskID.String(),
			ownerID:      realOwnerID.String(),
			wasCompleted: false,
			wantErr:      services.ErrTaskNotFound,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, notRealTaskID.String()).
					Once().
					Return(nil, services.ErrTaskRepoNotFound)
			},
		},
		{
			name:         "access denied",
			id:           realTaskID.String(),
			ownerID:      uuid.New().String(),
			wasCompleted: true,
			wantErr:      services.ErrTaskAccessDenied,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)
			},
		},
		{
			name:         "update failed",
			id:           realTaskID.String(),
			ownerID:      realOwnerID.String(),
			wasCompleted: false,
			wantErr:      services.ErrTaskCompleteFailed,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(errors.New("failed to connect to db"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var completedAt *time.Time = nil
			if tt.wasCompleted {
				completedAt = &realCompletedAt
			}

			taskToReturn, err := models.NewTaskFromDB(models.TaskFromDBParams{
				ID:          realTaskID.String(),
				OwnerID:     realOwnerID.String(),
				Title:       "Some Title",
				Description: "Some Description",
				Deadline:    nil,
				IsCompleted: tt.wasCompleted,
				CompletedAt: completedAt,
			})
			require.NoError(t, err)
			require.NotNil(t, taskToReturn)

			repo := new(mocks.TaskRepository)
			if tt.mocksSetup != nil {
				tt.mocksSetup(repo, taskToReturn)
			}

			service, err := services.NewTaskService(repo)
			require.NoError(t, err)
			require.NotNil(t, service)

			ctx := context.Background()
			err = service.Complete(ctx, tt.id, tt.ownerID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.True(t, taskToReturn.IsCompleted())
			require.NoError(t, err)
		})
	}
}

func TestTaskService_Reopen(t *testing.T) {
	realTaskID := uuid.New()
	realOwnerID := uuid.New()
	notRealTaskID := uuid.New()

	realCompletedAt := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name    string
		id      string
		ownerID string

		wasCompleted bool

		wantErr error

		mocksSetup func(repo *mocks.TaskRepository, taskToReturn *models.Task)
	}{
		{
			name:         "success if already completed",
			id:           realTaskID.String(),
			ownerID:      realOwnerID.String(),
			wasCompleted: true,
			wantErr:      nil,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(nil)
			},
		},
		{
			name:         "success if wasn't completed",
			id:           realTaskID.String(),
			ownerID:      realOwnerID.String(),
			wasCompleted: false,
			wantErr:      nil,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Maybe().
					Return(nil)
			},
		},
		{
			name:         "task not found",
			id:           notRealTaskID.String(),
			ownerID:      realOwnerID.String(),
			wasCompleted: false,
			wantErr:      services.ErrTaskNotFound,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, notRealTaskID.String()).
					Once().
					Return(nil, services.ErrTaskRepoNotFound)
			},
		},
		{
			name:         "access denied",
			id:           realTaskID.String(),
			ownerID:      uuid.New().String(),
			wasCompleted: true,
			wantErr:      services.ErrTaskAccessDenied,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)
			},
		},
		{
			name:         "update failed",
			id:           realTaskID.String(),
			ownerID:      realOwnerID.String(),
			wasCompleted: true,
			wantErr:      services.ErrTaskReopenFailed,
			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(errors.New("failed to connect to db"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var completedAt *time.Time = nil
			if tt.wasCompleted {
				completedAt = &realCompletedAt
			}

			taskToReturn, err := models.NewTaskFromDB(models.TaskFromDBParams{
				ID:          realTaskID.String(),
				OwnerID:     realOwnerID.String(),
				Title:       "Some Title",
				Description: "Some Description",
				Deadline:    nil,
				IsCompleted: tt.wasCompleted,
				CompletedAt: completedAt,
			})
			require.NoError(t, err)
			require.NotNil(t, taskToReturn)

			repo := new(mocks.TaskRepository)
			if tt.mocksSetup != nil {
				tt.mocksSetup(repo, taskToReturn)
			}

			service, err := services.NewTaskService(repo)
			require.NoError(t, err)
			require.NotNil(t, service)

			ctx := context.Background()
			err = service.Reopen(ctx, tt.id, tt.ownerID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.False(t, taskToReturn.IsCompleted())
			require.NoError(t, err)
		})
	}
}

func TestTaskService_FindByOwner(t *testing.T) {
	realOwnerID := uuid.New()

	tests := []struct {
		name       string
		ownerID    string
		wantErr    error
		wantLen    int
		mocksSetup func(repo *mocks.TaskRepository)
	}{
		{
			name:    "success with 1 element",
			ownerID: realOwnerID.String(),
			wantErr: nil,
			wantLen: 1,

			mocksSetup: func(repo *mocks.TaskRepository) {
				t.Helper()

				task1, err := models.NewTask("title", "some description", realOwnerID)
				require.NoError(t, err)

				sliceToReturn := []*models.Task{
					task1,
				}

				repo.On("FindByOwner", mock.Anything, realOwnerID.String()).
					Once().
					Return(sliceToReturn, nil)
			},
		},
		{
			name:    "success with more elements",
			ownerID: realOwnerID.String(),
			wantErr: nil,
			wantLen: 3,

			mocksSetup: func(repo *mocks.TaskRepository) {
				t.Helper()

				task1, err := models.NewTask("title", "some description", realOwnerID)
				require.NoError(t, err)

				task2, err := models.NewTask("title2", "some description2", realOwnerID)
				require.NoError(t, err)

				task3, err := models.NewTask("title3", "", realOwnerID)
				require.NoError(t, err)

				sliceToReturn := []*models.Task{
					task1,
					task2,
					task3,
				}

				repo.On("FindByOwner", mock.Anything, realOwnerID.String()).
					Once().
					Return(sliceToReturn, nil)
			},
		},
		{
			name:    "success with no elements",
			ownerID: realOwnerID.String(),
			wantErr: nil,
			wantLen: 0,

			mocksSetup: func(repo *mocks.TaskRepository) {
				t.Helper()

				repo.On("FindByOwner", mock.Anything, realOwnerID.String()).
					Once().
					Return([]*models.Task{}, nil)
			},
		},
		{
			name:    "internal db error",
			ownerID: realOwnerID.String(),
			wantErr: services.ErrTaskFindByOwnerFailed,

			mocksSetup: func(repo *mocks.TaskRepository) {
				t.Helper()

				repo.On("FindByOwner", mock.Anything, realOwnerID.String()).
					Once().
					Return(nil, errors.New("failed to connect to db"))
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

			ctx := context.Background()

			result, err := service.FindByOwner(ctx, tt.ownerID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, result, tt.wantLen)
		})
	}
}
