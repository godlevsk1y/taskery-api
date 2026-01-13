package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/models"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/lib/pq"
)

// TaskRepository represents a repository of tasks in PostgreSQL database
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new TaskRepository using the provided sql.DB.
// It returns an error if the db argument is nil.
func NewTaskRepository(db *sql.DB) (*TaskRepository, error) {
	const op = "postgres.TaskRepository.NewTaskRepository"

	if db == nil {
		return nil, fmt.Errorf("%s: db is nil", op)
	}

	return &TaskRepository{db: db}, nil
}

// Create inserts a new task into the database.
// It returns an error if the insertion fails.
//
// Special cases:
//
// Returns services.ErrTaskRepoExists if a task with the same ID already exists.
// Returns services.ErrTaskRepoOwnerNotFound if the owner of the task does not exist.
//
// The task's deadline is optional; if nil, it is stored as NULL in the database.
// Completed tasks can have a completion timestamp, which is also stored in the database.
func (tr *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	const op = "postgres.TaskRepository.Create"

	const query = `INSERT INTO tasks (
        id,
		owner_id,
		title,
		description,
		deadline,
		is_completed,
		completed_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	var deadlineToInsert *time.Time = nil
	if task.Deadline() != nil {
		deadlineTime := task.Deadline().Time()
		deadlineToInsert = &deadlineTime
	}

	_, err := tr.db.ExecContext(
		ctx,
		query,
		task.ID().String(),
		task.OwnerID().String(),
		task.Title().String(),
		task.Description().String(),
		deadlineToInsert,
		task.IsCompleted(),
		task.CompletedAt(),
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505": // unique constraint
				return services.ErrTaskRepoExists

			case "23503": // foreign key constraint
				return services.ErrTaskRepoOwnerNotFound

			default:
				return fmt.Errorf("%s: create task: %w", op, err)
			}
		} else {
			return fmt.Errorf("%s: create task: %w", op, err)
		}
	}

	return nil
}

func (tr *TaskRepository) FindByID(ctx context.Context, id string) (*models.Task, error) {
	//TODO implement me
	panic("implement me")
}

func (tr *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	//TODO implement me
	panic("implement me")
}

func (tr *TaskRepository) Delete(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}
