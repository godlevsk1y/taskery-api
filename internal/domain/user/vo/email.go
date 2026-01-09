package vo

import (
	"errors"
	"net/mail"
	"strings"
)

// Email is a value object (VO) that represents a validated email address.
// It ensures the email is not empty and conforms to the standard email format.
type Email struct {
	value string
}

var (
	// ErrEmailEmpty is returned when an empty string is provided.
	ErrEmailEmpty = errors.New("email is empty")

	// ErrEmailInvalid is returned when the email format is invalid.
	ErrEmailInvalid = errors.New("email is invalid")
)

// NewEmail creates a new Email instance after validating the input string.
// It trims whitespace and checks for emptiness and proper email format.
// Returns an error if validation fails.
func NewEmail(value string) (Email, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return Email{}, ErrEmailEmpty
	}

	_, err := mail.ParseAddress(value)
	if err != nil {
		return Email{}, ErrEmailInvalid
	}

	return Email{value: value}, nil
}

func (e Email) String() string {
	return e.value
}
