package services_test

import (
	"context"
	"errors"
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

		isTaskIDExpected bool

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

			isTaskIDExpected: true,

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

			isTaskIDExpected: true,

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

			isTaskIDExpected: false,

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

			isTaskIDExpected: false,
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

			isTaskIDExpected: false,

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

			isTaskIDExpected: false,

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
			taskID, err := service.Create(ctx, tt.cmd)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			if tt.isTaskIDExpected {
				require.NotEmpty(t, taskID)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTaskService_Update(t *testing.T) {
	realTaskID := uuid.New()
	realOwnerID := uuid.New()

	tests := []struct {
		name        string
		id          string
		cmd         services.UpdateTaskCommand
		expectedErr error

		mocksSetup func(repo *mocks.TaskRepository, taskToReturn *models.Task)
	}{
		{
			name: "success with all fields",
			id:   realTaskID.String(),
			cmd: services.UpdateTaskCommand{
				Title:       new("new title"),
				Description: new("some new description"),
				Deadline:    new(time.Now().Add(2 * time.Hour)),
			},
			expectedErr: nil,

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
			name: "success with one field",
			id:   realTaskID.String(),
			cmd: services.UpdateTaskCommand{
				Title:       new("new title"),
				Description: nil,
				Deadline:    nil,
			},

			expectedErr: nil,

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
			name: "success with no fields",
			id:   realTaskID.String(),
			cmd: services.UpdateTaskCommand{
				Title:       nil,
				Description: nil,
				Deadline:    nil,
			},

			expectedErr: nil,
			mocksSetup:  nil,
		},
		{
			name: "task not found",
			id:   uuid.New().String(),
			cmd: services.UpdateTaskCommand{
				Title:       new("new title"),
				Description: nil,
				Deadline:    nil,
			},

			expectedErr: services.ErrTaskNotFound,

			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, mock.AnythingOfType("string")).
					Once().
					Return(nil, services.ErrTaskRepoNotFound)
			},
		},
		{
			name: "internal update error",
			id:   realTaskID.String(),
			cmd: services.UpdateTaskCommand{
				Title:       new("new title"),
				Description: nil,
				Deadline:    nil,
			},

			expectedErr: services.ErrTaskUpdateFailed,

			mocksSetup: func(repo *mocks.TaskRepository, taskToReturn *models.Task) {
				repo.On("FindByID", mock.Anything, realTaskID.String()).
					Once().
					Return(taskToReturn, nil)

				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Task")).
					Once().
					Return(errors.New("internal db error"))
			},
		},
		{
			name: "invalid field",
			id:   realTaskID.String(),
			cmd: services.UpdateTaskCommand{
				Title:       new(""),
				Description: new("some new description"),
				Deadline:    new(time.Now().Add(2 * time.Hour)),
			},

			expectedErr: vo.ErrTitleEmpty,

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
				Title:       "Old Title",
				Description: "Some Description",
				Deadline:    new(time.Now().Add(time.Hour)),
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
			err = service.Update(ctx, tt.id, realOwnerID.String(), tt.cmd)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			if tt.cmd.Title != nil {
				require.Equal(t, *tt.cmd.Title, taskToReturn.Title().String())
			}

			if tt.cmd.Description != nil {
				require.Equal(t, *tt.cmd.Description, taskToReturn.Description().String())
			}

			if tt.cmd.Deadline != nil {
				require.Equal(t, *tt.cmd.Deadline, taskToReturn.Deadline().Time())
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

func TestTaskService_Delete(t *testing.T) {
	validTaskID := uuid.New()
	validOwnerID := uuid.New()

	tests := []struct {
		name    string
		taskID  string
		ownerID string
		wantErr error

		mocksSetup func(repo *mocks.TaskRepository)
	}{
		{
			name:    "success",
			taskID:  validTaskID.String(),
			ownerID: validOwnerID.String(),
			wantErr: nil,

			mocksSetup: func(repo *mocks.TaskRepository) {
				repo.On("FindByID", mock.Anything, validTaskID.String()).
					Once().
					Return(models.NewTaskFromDB(models.TaskFromDBParams{
						ID:          validTaskID.String(),
						OwnerID:     validOwnerID.String(),
						Title:       "some title",
						Description: "some description",
						Deadline:    nil,
						IsCompleted: false,
						CompletedAt: nil,
					}))

				repo.On("Delete", mock.Anything, validTaskID.String()).
					Once().
					Return(nil)
			},
		},
		{
			name:    "access denied",
			taskID:  validTaskID.String(),
			ownerID: uuid.New().String(),
			wantErr: services.ErrTaskAccessDenied,

			mocksSetup: func(repo *mocks.TaskRepository) {
				repo.On("FindByID", mock.Anything, validTaskID.String()).
					Once().
					Return(models.NewTaskFromDB(models.TaskFromDBParams{
						ID:          validTaskID.String(),
						OwnerID:     validOwnerID.String(),
						Title:       "some title",
						Description: "some description",
						Deadline:    nil,
						IsCompleted: false,
						CompletedAt: nil,
					}))
			},
		},
		{
			name:    "internal db error",
			taskID:  validTaskID.String(),
			ownerID: validOwnerID.String(),
			wantErr: services.ErrTaskDeleteFailed,

			mocksSetup: func(repo *mocks.TaskRepository) {
				repo.On("FindByID", mock.Anything, validTaskID.String()).
					Once().
					Return(models.NewTaskFromDB(models.TaskFromDBParams{
						ID:          validTaskID.String(),
						OwnerID:     validOwnerID.String(),
						Title:       "some title",
						Description: "some description",
						Deadline:    nil,
						IsCompleted: false,
						CompletedAt: nil,
					}))

				repo.On("Delete", mock.Anything, validTaskID.String()).
					Once().
					Return(errors.New("failed to connect to db"))
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
			err = service.Delete(ctx, tt.taskID, tt.ownerID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
