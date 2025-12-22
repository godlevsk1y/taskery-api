package models

import (
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/vo"
	"github.com/google/uuid"
)

// Task is a model that represents a task.
// It includes the task's ID, title, description, completion status, and deadline.
type Task struct {
	id    uuid.UUID
	owner uuid.UUID

	title       vo.Title
	description vo.Description

	isCompleted bool
	deadline    *vo.Deadline
}

func (t *Task) ID() uuid.UUID               { return t.id }
func (t *Task) Owner() uuid.UUID            { return t.owner }
func (t *Task) Title() vo.Title             { return t.title }
func (t *Task) Description() vo.Description { return t.description }
func (t *Task) IsCompleted() bool           { return t.isCompleted }
func (t *Task) Deadline() *vo.Deadline      { return t.deadline }

func NewTask(title string, description string, owner uuid.UUID) (*Task, error) {
	titleVO, err := vo.NewTitle(title)
	if err != nil {
		return nil, err
	}

	descriptionVO, err := vo.NewDescription(description)
	if err != nil {
		return nil, err
	}

	return &Task{
		id:    uuid.New(),
		owner: owner,

		title:       titleVO,
		description: descriptionVO,

		deadline:    nil,
		isCompleted: false,
	}, nil
}

func NewTaskWithDeadline(title string, description string, owner uuid.UUID, deadline time.Time) (*Task, error) {
	task, err := NewTask(title, description, owner)
	if err != nil {
		return nil, err
	}

	deadlineVO, err := vo.NewDeadline(deadline)
	if err != nil {
		return nil, err
	}

	task.deadline = &deadlineVO
	return task, nil
}
