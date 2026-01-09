package vo

import (
	"errors"
	"strings"
)

// Title is a VO that represents a title of the task.
type Title struct {
	value string
}

const TitleMaxLength = 50

var (
	ErrTitleEmpty   = errors.New("title is empty")
	ErrTitleTooLong = errors.New("title is too long")
)

// NewTitle creates a new Title instance.
func NewTitle(value string) (Title, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return Title{}, ErrTitleEmpty
	}

	if len([]rune(value)) > TitleMaxLength {
		return Title{}, ErrTitleTooLong
	}

	return Title{value: value}, nil
}

func (t Title) String() string {
	return t.value
}
