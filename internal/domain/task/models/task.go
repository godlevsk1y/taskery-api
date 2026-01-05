package models

import (
	"errors"
	"fmt"
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

	completedAtCopy := *t.completedAt
	return &completedAtCopy
}

var ErrTaskFailedCreateFromDB = errors.New("failed to create task from DB")

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

// TaskFromDBParams contains raw task data loaded from the database.
//
// This struct is used as an input for constructing a domain Task
// from persisted storage. All fields represent database values
// and may require validation or transformation before being used
// inside the domain model.
type TaskFromDBParams struct {
	ID    uuid.UUID
	Owner uuid.UUID

	Title       string
	Description string

	Deadline *time.Time

	IsCompleted bool
	CompletedAt *time.Time
}

// NewTaskFromDB creates a Task from database parameters.
// It validates that the completedAt and isCompleted fields are consistent:
// completedAt must be non-nil if and only if isCompleted is true.
// It returns an error if any of the value objects (title, description, deadline) fail to be created.
//
// If p.Deadline is not nil, the deadline field of the Task will be set.
func NewTaskFromDB(p TaskFromDBParams) (*Task, error) {
	if (p.IsCompleted && p.CompletedAt == nil) || (!p.IsCompleted && p.CompletedAt != nil) {
		return nil, fmt.Errorf("%w: %s", ErrTaskFailedCreateFromDB, "completedAt and isCompleted fields contradict")
	}

	titleVO, err := vo.NewTitle(p.Title)
	if err != nil {
		return nil, err
	}

	descriptionVO, err := vo.NewDescription(p.Description)
	if err != nil {
		return nil, err
	}

	task := &Task{
		id:          p.ID,
		owner:       p.Owner,
		title:       titleVO,
		description: descriptionVO,

		deadline: nil,

		isCompleted: p.IsCompleted,
		completedAt: p.CompletedAt,
	}

	if p.Deadline != nil {
		deadlineVO, err := vo.NewDeadline(*p.Deadline)
		if err != nil {
			return nil, err
		}

		task.deadline = &deadlineVO
	}

	return task, nil
}

// NewTaskWithDeadline creates a new Task instance with the given title, description, owner, and deadline.
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
func (t *Task) SetDeadline(deadline time.Time) error {
	deadlineVO, err := vo.NewDeadline(deadline)
	if err != nil {
		return err
	}

	t.deadline = &deadlineVO
	return nil
}

// HasDeadline checks if the task has a deadline set.
func (t *Task) HasDeadline() bool {
	return t.deadline != nil
}

// RemoveDeadline removes the deadline from the task if it is set.
func (t *Task) RemoveDeadline() {
	t.deadline = nil
}

// IsOverdue checks if the task is overdue.
// It returns a boolean indicating whether the task is overdue.
//
// If the task is completed, it returns false.
func (t *Task) IsOverdue() bool {
	if t.deadline == nil {
		return false
	}

	return t.deadline.IsOverdue() && !t.isCompleted
}

// Complete marks the task as complete and sets the time of completion.
func (t *Task) Complete() {
	if !t.isCompleted {
		t.isCompleted = true
		now := time.Now()
		t.completedAt = &now
	}
}

// Reopen marks the task as incomplete and clears the time of completion.
func (t *Task) Reopen() {
	t.isCompleted = false
	t.completedAt = nil
}
