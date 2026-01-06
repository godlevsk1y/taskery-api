package services

import (
	"context"
	"errors"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/models"
)

// TaskService is a service that handles task operations.
type TaskService struct {
	tasksRepo TaskRepository
}

// TaskRepository defines the methods for managing task data in a persistent storage.
// It provides basic CRUD operations for the Task model.
type TaskRepository interface {
	// Create saves a new task in the repository.
	// Returns an error if the operation fails.
	Create(ctx context.Context, task *models.Task) error

	// FindByID retrieves a task by its unique identifier.
	// Returns the task and nil error if found, otherwise returns nil and an error.
	FindByID(ctx context.Context, id string) (*models.Task, error)

	// Update modifies an existing task's data in the repository.
	// Returns an error if the operation fails or the task does not exist.
	Update(ctx context.Context, task *models.Task) error

	// Delete removes a user from the repository by their unique identifier.
	// Returns an error if the operation fails or the user does not exist.
	Delete(ctx context.Context, id string) error
}

var (
	ErrTaskRepositoryNil = errors.New("tasks repository is nil")
)

// NewTaskService creates a new TaskService instance.
// It returns nil and error if the task repository is nil
func NewTaskService(tasksRepo TaskRepository) (*TaskService, error) {
	if tasksRepo == nil {
		return nil, ErrTaskRepositoryNil
	}

	return &TaskService{tasksRepo}, nil
}
