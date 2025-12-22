package vo

import (
	"errors"
	"time"
)

// Deadline is a VO that represents a deadline for the task.
type Deadline struct {
	value time.Time
}

var ErrDeadlineBeforeNow = errors.New("deadline is in the past")

// NewDeadline creates a new Deadline instance.
func NewDeadline(value time.Time) (Deadline, error) {
	if value.Before(time.Now()) {
		return Deadline{}, ErrDeadlineBeforeNow
	}

	return Deadline{value: value}, nil
}

func (d Deadline) Time() time.Time {
	return d.value
}

// IsBefore checks if the deadline is before the given time.
func (d Deadline) IsBefore(t time.Time) bool {
	return d.value.Before(t)
}

// IsAfter checks if the deadline is after the given time.
func (d Deadline) IsAfter(t time.Time) bool {
	return d.value.After(t)
}

// IsOverdue checks if the deadline is overdue.
func (d Deadline) IsOverdue() bool {
	return d.value.Before(time.Now())
}
