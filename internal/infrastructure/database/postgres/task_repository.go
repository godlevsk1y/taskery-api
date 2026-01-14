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

// FindByID returns the task with the given id.
//
// If no task with the specified id exists, FindByID returns
// services.ErrTaskRepoNotFound.
//
// Other errors may be returned if the query fails or if the task
// cannot be restored from the database representation.
func (tr *TaskRepository) FindByID(ctx context.Context, id string) (*models.Task, error) {
	const op = "postgres.TaskRepository.FindByID"

	const query = `
		SELECT id, owner_id, title, description, deadline, is_completed, completed_at 
		FROM tasks WHERE id = $1`

	row := tr.db.QueryRowContext(ctx, query, id)

	var (
		userID      string
		ownerId     string
		title       string
		description string
		deadline    *time.Time
		isCompleted bool
		completedAt *time.Time
	)

	err := row.Scan(&userID, &ownerId, &title, &description, &deadline, &isCompleted, &completedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrTaskRepoNotFound
		}

		return nil, fmt.Errorf("%s: find by id: %w", op, err)
	}

	task, err := models.NewTaskFromDB(models.TaskFromDBParams{
		ID:          userID,
		OwnerID:     ownerId,
		Title:       title,
		Description: description,
		Deadline:    deadline,
		IsCompleted: isCompleted,
		CompletedAt: completedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: restore task: %w", op, err)
	}

	return task, nil
}

func (tr *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	const op = "postgres.TaskRepository.Update"

	const query = `
		UPDATE tasks SET 
			 title = $1, 
			 description = $2, 
			 deadline = $3, 
			 is_completed = $4, 
			 completed_at = $5
		WHERE id = $6`

	var deadlineToUpdate *time.Time = nil
	if task.Deadline() != nil {
		deadlineTime := task.Deadline().Time()
		deadlineToUpdate = &deadlineTime
	}

	res, err := tr.db.ExecContext(
		ctx,
		query,
		task.Title().String(),
		task.Description().String(),
		deadlineToUpdate,
		task.IsCompleted(),
		task.CompletedAt(),
	)
	if err != nil {
		return fmt.Errorf("%s: update task: %w", op, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: get affected rows: %w", op, err)
	}

	if affected == 0 {
		return services.ErrTaskRepoNotFound
	}

	return nil
}

func (tr *TaskRepository) Delete(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}
