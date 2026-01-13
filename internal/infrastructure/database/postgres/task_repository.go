package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/models"
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

func (t TaskRepository) Create(ctx context.Context, task *models.Task) error {
	//TODO implement me
	panic("implement me")
}

func (t TaskRepository) FindByID(ctx context.Context, id string) (*models.Task, error) {
	//TODO implement me
	panic("implement me")
}

func (t TaskRepository) Update(ctx context.Context, task *models.Task) error {
	//TODO implement me
	panic("implement me")
}

func (t TaskRepository) Delete(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}
