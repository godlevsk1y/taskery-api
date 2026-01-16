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

	// FindByOwner fetches all tasks for the given ownerID.
	// Returns a slice of tasks and a nil error if tasks exist,
	// an empty slice and nil if no tasks are found,
	// or nil and an error if something goes wrong.
	FindByOwner(ctx context.Context, ownerID string) ([]*models.Task, error)

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
	ErrTaskRepoNotFound = errors.New("task was not found in the repository")

	// ErrTaskRepoOwnerNotFound is returned by repository
	// when the owner with the given ID does not exist there.
	ErrTaskRepoOwnerNotFound = errors.New("owner was not found in the repository")
)

// Application-level errors
var (
	// ErrTaskExists is returned by TaskService if the task that is to be added already exists
	ErrTaskExists = errors.New("task already exists")

	// ErrTaskNotFound is returned by TaskService if the task was not found in the repository
	ErrTaskNotFound = errors.New("user was not found")

	// ErrTaskOwnerNotFound is returned by TaskService
	// when an operation cannot be completed because the owner with the given ID does not exist.
	ErrTaskOwnerNotFound = errors.New("owner was not found")

	// ErrTaskCreateFailed is returned by TaskService if an internal error occurred during creation
	ErrTaskCreateFailed = errors.New("failed to create task")

	// ErrTaskChangeTitleFailed is returned by TaskService if an internal error occurred during editing the title
	ErrTaskChangeTitleFailed = errors.New("failed to change title")

	// ErrTaskChangeDescriptionFailed is returned by TaskService
	// if an internal error occurred during editing the description
	ErrTaskChangeDescriptionFailed = errors.New("failed to change description")

	// ErrTaskSetDeadlineFailed is returned by TaskService
	// if an internal error occurred during setting the deadline
	ErrTaskSetDeadlineFailed = errors.New("failed to set deadline")

	// ErrTaskRemoveDeadlineFailed is returned by TaskService
	// if an internal error occurred during removing the deadline
	ErrTaskRemoveDeadlineFailed = errors.New("failed to remove deadline")

	// ErrTaskCompleteFailed is returned by TaskService
	// if an internal error occurred during completion of the task
	ErrTaskCompleteFailed = errors.New("failed to complete task")

	// ErrTaskReopenFailed is returned by TaskService
	// if an internal error occurred during reopening of the task
	ErrTaskReopenFailed = errors.New("failed to reopen task")

	ErrTaskFindByOwnerFailed = errors.New("failed to find by owner")

	// ErrTaskAccessDenied is returned when an operation on a task is not allowed
	// because the caller does not have permission to access the task.
	ErrTaskAccessDenied = errors.New("task access denied")
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
func (ts *TaskService) ChangeTitle(ctx context.Context, id string, ownerID string, new string) error {
	task, err := ts.tasksRepo.FindByID(ctx, id)
	if errors.Is(err, ErrTaskRepoNotFound) {
		return ErrTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrTaskChangeTitleFailed, err)
	}

	if task.OwnerID().String() != ownerID {
		return ErrTaskAccessDenied
	}

	if err := task.ChangeTitle(new); err != nil {
		return err
	}

	if err := ts.tasksRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("%w: %s", ErrTaskChangeTitleFailed, err)
	}

	return nil
}

// ChangeDescription updates the description of the task identified by id.
//
// It loads the task from the repository, applies the new description,
// and persists the updated task.
//
// ChangeDescription returns ErrTaskRepoNotFound if no task with the given
// id exists. It returns ErrTaskChangeDescriptionFailed if the description
// cannot be changed or if the updated task cannot be saved.
//
// The operation respects the provided context ctx.
func (ts *TaskService) ChangeDescription(ctx context.Context, id string, ownerID string, new string) error {
	task, err := ts.tasksRepo.FindByID(ctx, id)
	if errors.Is(err, ErrTaskRepoNotFound) {
		return ErrTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrTaskChangeDescriptionFailed, err)
	}

	if task.OwnerID().String() != ownerID {
		return ErrTaskAccessDenied
	}

	if err := task.ChangeDescription(new); err != nil {
		return err
	}

	if err := ts.tasksRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("%w: %s", ErrTaskChangeDescriptionFailed, err)
	}

	return nil
}

// SetDeadline sets or updates the deadline of the task with the given id.
//
// It loads the task from the repository, applies the new deadline value,
// and persists the updated task.
//
// SetDeadline returns ErrTaskRepoNotFound if the task does not exist.
// It returns an error if the deadline value is invalid or if the task
// cannot be updated in the repository.
func (ts *TaskService) SetDeadline(ctx context.Context, id string, ownerID string, deadline time.Time) error {
	task, err := ts.tasksRepo.FindByID(ctx, id)
	if errors.Is(err, ErrTaskRepoNotFound) {
		return ErrTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrTaskSetDeadlineFailed, err)
	}

	if task.OwnerID().String() != ownerID {
		return ErrTaskAccessDenied
	}

	if err := task.SetDeadline(deadline); err != nil {
		return err
	}

	if err := ts.tasksRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("%w: %s", ErrTaskSetDeadlineFailed, err)
	}

	return nil
}

// RemoveDeadline removes the deadline from the task with the given id.
//
// It looks up the task in the repository, clears its deadline, and
// persists the updated task.
//
// RemoveDeadline returns ErrTaskRepoNotFound if a task with the given
// id does not exist. It returns a wrapped error if fetching or updating
// the task fails for any other reason.
func (ts *TaskService) RemoveDeadline(ctx context.Context, id string, ownerID string) error {
	task, err := ts.tasksRepo.FindByID(ctx, id)
	if errors.Is(err, ErrTaskRepoNotFound) {
		return ErrTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrTaskRemoveDeadlineFailed, err)
	}

	if task.OwnerID().String() != ownerID {
		return ErrTaskAccessDenied
	}

	task.RemoveDeadline()

	if err := ts.tasksRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("%w: %s", ErrTaskRemoveDeadlineFailed, err)
	}

	return nil
}

// Complete marks the task with the given id as completed.
//
// It returns ErrTaskNotFound if the task does not exist.
// If the task is owned by a different user than ownerID,
// Complete returns ErrTaskAccessDenied.
//
// If updating the task fails, Complete returns ErrTaskCompleteFailed
func (ts *TaskService) Complete(ctx context.Context, id string, ownerID string) error {
	task, err := ts.tasksRepo.FindByID(ctx, id)
	if errors.Is(err, ErrTaskRepoNotFound) {
		return ErrTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrTaskCompleteFailed, err)
	}

	if task.OwnerID().String() != ownerID {
		return ErrTaskAccessDenied
	}

	if task.IsCompleted() {
		return nil
	}

	task.Complete()

	if err := ts.tasksRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("%w: %s", ErrTaskCompleteFailed, err)
	}

	return nil
}

// Reopen marks a completed task as not completed.
// Returns ErrTaskNotFound if the task does not exist.
// Returns ErrTaskAccessDenied if the ownerID does not match.
// Returns ErrTaskReopenFailed if updating the task fails.
func (ts *TaskService) Reopen(ctx context.Context, id string, ownerID string) error {
	task, err := ts.tasksRepo.FindByID(ctx, id)
	if errors.Is(err, ErrTaskRepoNotFound) {
		return ErrTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrTaskReopenFailed, err)
	}

	if task.OwnerID().String() != ownerID {
		return ErrTaskAccessDenied
	}

	if !task.IsCompleted() {
		return nil
	}

	task.Reopen()

	if err := ts.tasksRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("%w: %s", ErrTaskReopenFailed, err)
	}

	return nil
}

// FindByOwner returns all tasks that belong to the given ownerID.
// If no tasks are found, it returns an empty slice. If an error occurred
func (ts *TaskService) FindByOwner(ctx context.Context, ownerID string) ([]*models.Task, error) {
	tasks, err := ts.tasksRepo.FindByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTaskFindByOwnerFailed, err)
	}

	return tasks, nil
}
