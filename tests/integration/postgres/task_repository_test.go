//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	taskModels "github.com/cyberbrain-dev/taskery-api/internal/domain/task/models"
	userModels "github.com/cyberbrain-dev/taskery-api/internal/domain/user/models"
	"github.com/cyberbrain-dev/taskery-api/internal/services"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/database/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func migrateTasks(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		CREATE TABLE tasks (
			id UUID PRIMARY KEY,
			owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		
			title TEXT NOT NULL,
			description TEXT NOT NULL,
		
			deadline TIMESTAMPTZ NULL,
		
			is_completed BOOLEAN NOT NULL DEFAULT FALSE,
			completed_at TIMESTAMPTZ NULL
		);
	`)

	require.NoError(t, err)
}

func TestTaskRepository_Create(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	migrateUsers(t, db)
	migrateTasks(t, db)

	userRepo, err := postgres.NewUserRepository(db)
	require.NoError(t, err)

	taskRepo, err := postgres.NewTaskRepository(db)
	require.NoError(t, err)

	realUser, err := userModels.NewUserFromDB(userModels.UserFromDBParams{
		ID:           uuid.New().String(),
		Username:     "Test User",
		Email:        "test@example.com",
		PasswordHash: "password123",
	})
	require.NoError(t, err)

	ctx := context.Background()

	err = userRepo.Create(ctx, realUser)
	require.NoError(t, err)

	validTask, err := taskModels.NewTask("title", "no description", realUser.ID())
	require.NoError(t, err)

	t.Run("success with nil deadline", func(t *testing.T) {
		err = taskRepo.Create(ctx, validTask)
		require.NoError(t, err)

		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&count)
		require.NoError(t, err)

		require.Equal(t, 1, count)
	})

	t.Run("success with deadline", func(t *testing.T) {
		deadline := time.Now().Add(24 * time.Hour)
		task, err := taskModels.NewTaskWithDeadline("other title", "some description", realUser.ID(), deadline)
		require.NoError(t, err)

		err = taskRepo.Create(ctx, task)
		require.NoError(t, err)

		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&count)
		require.NoError(t, err)

		require.Equal(t, 2, count) // this db already contains user from prev test
	})

	t.Run("not existing owner", func(t *testing.T) {
		task, err := taskModels.NewTask("not exist", "", uuid.New())
		require.NoError(t, err)

		err = taskRepo.Create(ctx, task)
		require.ErrorIs(t, err, services.ErrTaskRepoOwnerNotFound)

		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&count)
		require.NoError(t, err)

		require.Equal(t, 2, count) // no more new tasks
	})

	t.Run("task already exists", func(t *testing.T) {
		err = taskRepo.Create(ctx, validTask)
		require.ErrorIs(t, err, services.ErrTaskRepoExists)

		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&count)
		require.NoError(t, err)

		require.Equal(t, 2, count) // no more new tasks
	})
}

func TestTaskRepository_FindByID(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	migrateUsers(t, db)
	migrateTasks(t, db)

	userRepo, err := postgres.NewUserRepository(db)
	require.NoError(t, err)

	taskRepo, err := postgres.NewTaskRepository(db)
	require.NoError(t, err)

	realUser, err := userModels.NewUserFromDB(userModels.UserFromDBParams{
		ID:           uuid.New().String(),
		Username:     "Test User",
		Email:        "test@example.com",
		PasswordHash: "password123",
	})
	require.NoError(t, err)

	ctx := context.Background()

	err = userRepo.Create(ctx, realUser)
	require.NoError(t, err)

	validTask, err := taskModels.NewTask("title", "no description", realUser.ID())
	require.NoError(t, err)

	err = taskRepo.Create(ctx, validTask)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		taskFromDB, err := taskRepo.FindByID(ctx, validTask.ID().String())
		require.NoError(t, err)

		require.Equal(t, *validTask, *taskFromDB)
	})

	t.Run("not found", func(t *testing.T) {
		taskFromDB, err := taskRepo.FindByID(ctx, uuid.New().String())
		require.ErrorIs(t, err, services.ErrTaskRepoNotFound)
		require.Nil(t, taskFromDB)
	})
}

func TestTaskRepository_Update(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	migrateUsers(t, db)
	migrateTasks(t, db)

	userRepo, err := postgres.NewUserRepository(db)
	require.NoError(t, err)

	taskRepo, err := postgres.NewTaskRepository(db)
	require.NoError(t, err)

	realUser, err := userModels.NewUserFromDB(userModels.UserFromDBParams{
		ID:           uuid.New().String(),
		Username:     "Test User",
		Email:        "test@example.com",
		PasswordHash: "password123",
	})
	require.NoError(t, err)

	ctx := context.Background()

	err = userRepo.Create(ctx, realUser)
	require.NoError(t, err)

	validTask, err := taskModels.NewTask("title", "no description", realUser.ID())
	require.NoError(t, err)

	err = taskRepo.Create(ctx, validTask)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		err := validTask.ChangeDescription("new description")
		require.NoError(t, err)

		err = taskRepo.Update(ctx, validTask)
		require.NoError(t, err)

		var taskFromDB *taskModels.Task
		taskFromDB, err = taskRepo.FindByID(ctx, validTask.ID().String())
		require.NoError(t, err)

		require.Equal(t, *validTask, *taskFromDB)
	})

	t.Run("task not found", func(t *testing.T) {
		notExistingTask, err := taskModels.NewTask("not existing title", "no description", realUser.ID())
		require.NoError(t, err)

		err = taskRepo.Update(ctx, notExistingTask)
		require.ErrorIs(t, err, services.ErrTaskRepoNotFound)
	})
}

func TestTaskRepository_Delete(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	migrateUsers(t, db)
	migrateTasks(t, db)

	userRepo, err := postgres.NewUserRepository(db)
	require.NoError(t, err)

	taskRepo, err := postgres.NewTaskRepository(db)
	require.NoError(t, err)

	realUser, err := userModels.NewUserFromDB(userModels.UserFromDBParams{
		ID:           uuid.New().String(),
		Username:     "Test User",
		Email:        "test@example.com",
		PasswordHash: "password123",
	})
	require.NoError(t, err)

	ctx := context.Background()

	err = userRepo.Create(ctx, realUser)
	require.NoError(t, err)

	validTask, err := taskModels.NewTask("title", "no description", realUser.ID())
	require.NoError(t, err)

	err = taskRepo.Create(ctx, validTask)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		err := taskRepo.Delete(ctx, validTask.ID().String())
		require.NoError(t, err)

		taskFromDB, err := taskRepo.FindByID(ctx, validTask.ID().String())
		require.ErrorIs(t, err, services.ErrTaskRepoNotFound)
		require.Nil(t, taskFromDB)
	})

	t.Run("task not found", func(t *testing.T) {
		err = taskRepo.Delete(ctx, uuid.New().String())
		require.ErrorIs(t, err, services.ErrTaskRepoNotFound)
	})
}

func TestTaskRepository_FindByOwner(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	migrateUsers(t, db)
	migrateTasks(t, db)

	userRepo, err := postgres.NewUserRepository(db)
	require.NoError(t, err)

	taskRepo, err := postgres.NewTaskRepository(db)
	require.NoError(t, err)

	realUser, err := userModels.NewUserFromDB(userModels.UserFromDBParams{
		ID:           uuid.New().String(),
		Username:     "Test User",
		Email:        "test@example.com",
		PasswordHash: "password123",
	})
	require.NoError(t, err)

	ctx := context.Background()

	err = userRepo.Create(ctx, realUser)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		task1, err := taskModels.NewTask("first task", "no description", realUser.ID())
		require.NoError(t, err)
		err = taskRepo.Create(ctx, task1)
		require.NoError(t, err)

		task2, err := taskModels.NewTask("second task", "other description", realUser.ID())
		require.NoError(t, err)
		err = taskRepo.Create(ctx, task2)
		require.NoError(t, err)

		tasksFromDB, err := taskRepo.FindByOwner(ctx, realUser.ID().String())
		require.NoError(t, err)

		require.Equal(t, 2, len(tasksFromDB))
		require.Equal(t, *task1, *(tasksFromDB[0]))
		require.Equal(t, *task2, *(tasksFromDB[1]))
	})
	t.Run("empty slice", func(t *testing.T) {
		tasks, err := taskRepo.FindByOwner(ctx, realUser.ID().String())
		require.NoError(t, err)
		require.NotNil(t, tasks)
		require.Equal(t, 0, len(tasks))
	})
}
