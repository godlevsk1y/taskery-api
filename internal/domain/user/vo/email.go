package vo

import (
	"errors"
	"net/mail"
	"strings"
)

// Email is a VO that represents an email address.
type Email struct {
	value string
}

var (
	ErrEmailEmpty   = errors.New("email is empty")
	ErrEmailInvalid = errors.New("email is invalid")
)

// NewEmail creates a new Email instance.
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
