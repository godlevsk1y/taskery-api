package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/models"
	"github.com/google/uuid"
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

// ErrTaskRepositoryNil is an error that indicates that the task repository
// that is passed to NewTaskService is nil.
var ErrTaskRepositoryNil = errors.New("task repository is nil")

// Repository-level errors
var (
	// ErrTaskRepoExists is returned by repository if the task
	// that is to be added to the repository already exists there
	ErrTaskRepoExists = errors.New("task already exists in the repository")

	// ErrTaskRepoNotFound is returned by repository if the task was not found there
	ErrTaskRepoNotFound = errors.New("user was not found in the repository")

	// ErrTaskRepoOwnerNotFound is returned by repository
	// when the owner with the given ID does not exist there.
	ErrTaskRepoOwnerNotFound = errors.New("owner was not found in the repository")
)

// Application-level errors
var (
	// ErrTaskExists is returned by TaskService if the task that is to be added already exists
	ErrTaskExists = errors.New("task already exists")

	// ErrTaskOwnerNotFound is returned by TaskService
	// when an operation cannot be completed because the owner with the given ID does not exist.
	ErrTaskOwnerNotFound = errors.New("owner was not found")

	// ErrTaskCreateFailed is returned by TaskService if an internal error occurred during creation
	ErrTaskCreateFailed = errors.New("failed to create task")

	// ErrTaskChangeTitleFailed is returned by TaskService if an internal error occurred during editing the title
	ErrTaskChangeTitleFailed = errors.New("change title failed")
)

// NewTaskService creates a new TaskService instance.
// It returns nil and error if the task repository is nil
func NewTaskService(tasksRepo TaskRepository) (*TaskService, error) {
	if tasksRepo == nil {
		return nil, ErrTaskRepositoryNil
	}

	return &TaskService{tasksRepo}, nil
}

// CreateTaskCommand contains all data required to create a new Task.
// It is used as input for TaskService.Create method.
//
// Title, Description, and OwnerID are required. Deadline is optional;
// if no deadline is needed, set it to nil.
type CreateTaskCommand struct {
	Title       string
	Description string
	OwnerID     uuid.UUID
	Deadline    *time.Time
}

// Create creates a new task with the given title, description, and owner.
// If cmd.Deadline is provided, the task will have a deadline; otherwise, it will be created without one.
//
// Create returns ErrTaskExists if the task already exists,
// ErrTaskOwnerNotFound if the specified owner does not exist,
// or ErrTaskCreateFailed if the repository fails to create the task.
func (ts *TaskService) Create(ctx context.Context, cmd CreateTaskCommand) error {
	var task *models.Task
	var err error

	if cmd.Deadline == nil {
		task, err = models.NewTask(cmd.Title, cmd.Description, cmd.OwnerID)
	} else {
		task, err = models.NewTaskWithDeadline(cmd.Title, cmd.Description, cmd.OwnerID, *cmd.Deadline)
	}
	if err != nil {
		return err
	}

	if err := ts.tasksRepo.Create(ctx, task); err != nil {
		if errors.Is(err, ErrTaskRepoExists) {
			return ErrTaskExists
		}

		if errors.Is(err, ErrTaskRepoOwnerNotFound) {
			return ErrTaskOwnerNotFound
		}

		return fmt.Errorf("%w: %s", ErrTaskCreateFailed, err)
	}

	return nil
}

// ChangeTitle changes the title of the task with the given id.
//
// It looks up the task in the repository, validates and applies the new
// title, and persists the updated task back to the repository.
//
// If the task with the given id does not exist, ChangeTitle returns
// ErrTaskRepoNotFound. If updating the task fails due to a repository
// error, it returns ErrTaskChangeTitleFailed wrapping the underlying error.
//
// Validation errors returned by task.ChangeTitle are propagated as-is.
func (ts *TaskService) ChangeTitle(ctx context.Context, id string, new string) error {
	task, err := ts.tasksRepo.FindByID(ctx, id)
	if errors.Is(err, ErrTaskRepoNotFound) {
		return ErrTaskRepoNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrTaskChangeTitleFailed, err)
	}

	if err := task.ChangeTitle(new); err != nil {
		return err
	}

	if err := ts.tasksRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("%w: %s", ErrTaskChangeTitleFailed, err)
	}

	return nil
}
