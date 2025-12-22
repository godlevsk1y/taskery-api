package models

import (
	"errors"
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

	deadline *vo.Deadline

	isCompleted bool
	completedAt *time.Time
}

func (t *Task) ID() uuid.UUID               { return t.id }
func (t *Task) Owner() uuid.UUID            { return t.owner }
func (t *Task) Title() vo.Title             { return t.title }
func (t *Task) Description() vo.Description { return t.description }
func (t *Task) Deadline() *vo.Deadline      { return t.deadline }
func (t *Task) IsCompleted() bool           { return t.isCompleted }

// CompletedAt returns the timestamp when the task was completed.
//
// If the task is not completed, it returns nil. The returned value
// is a copy of the internal timestamp, so modifying it does not
// affect the internal state of the task.
//
// This ensures that the task's completion time can be safely read
// without allowing external code to mutate it.
func (t *Task) CompletedAt() *time.Time {
	if t.completedAt == nil {
		return nil
	}

	copy := *t.completedAt
	return &copy
}

// NewTask creates a new Task instance with the given title, description, and owner. It does not set a deadline.
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

		deadline: nil,

		isCompleted: false,
		completedAt: nil,
	}, nil
}

// NewTaskWithDeadline creates a new Task instance with the given title, description, owner, and deadline. It sets the deadline.
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

// ChangeTitle changes the title of the task.
func (t *Task) ChangeTitle(newTitle string) error {
	newTitleVO, err := vo.NewTitle(newTitle)
	if err != nil {
		return err
	}
	t.title = newTitleVO
	return nil
}

// ChangeDescription changes the description of the task.
func (t *Task) ChangeDescription(newDescription string) error {
	newDescriptionVO, err := vo.NewDescription(newDescription)
	if err != nil {
		return err
	}
	t.description = newDescriptionVO
	return nil
}

// SetDeadline sets the deadline in case it is not already set or edits the deadline.
func (t *Task) SetDealine(deadline time.Time) error {
	deadlineVO, err := vo.NewDeadline(deadline)
	if err != nil {
		return err
	}

	t.deadline = &deadlineVO
	return nil
}

// IsDeadlineSet checks if the task has a deadline set.
func (t *Task) HasDeadline() bool {
	return t.deadline != nil
}

// ErrDeadlineNotSet represents a case when the caller tries to perform operations on a deadline that is not set.
var ErrDeadlineNotSet = errors.New("deadline not set")

// RemoveDeadline removes the deadline from the task if it is set.
func (t *Task) RemoveDeadline() error {
	if t.deadline == nil {
		return ErrDeadlineNotSet
	}

	t.deadline = nil
	return nil
}

// IsOverdue checks if the task is overdue.
// It returns a boolean indicating whether the task is overdue and an error if the deadline is not set.
func (t *Task) IsOverdue() (bool, error) {
	if t.deadline == nil {
		return false, ErrDeadlineNotSet
	}

	return t.deadline.IsOverdue(), nil
}
