package vo

import "errors"

// Description is a VO that represents a description of the task.
type Description struct {
	value string
}

const DescriptionMaxLength = 1000

var ErrDescriptionTooLong = errors.New("description is too long")

// NewDescription creates a new Description instance.
func NewDescription(value string) (Description, error) {
	if len([]rune(value)) > DescriptionMaxLength {
		return Description{}, ErrDescriptionTooLong
	}

	return Description{value: value}, nil
}

func (d Description) String() string {
	return d.value
}
