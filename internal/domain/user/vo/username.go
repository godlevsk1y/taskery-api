package vo

import "errors"

// Username is a value object (VO) that represents a validated username.
// It ensures the username is not empty and conforms to all the rules.
type Username struct {
	value string
}

const (
	UsernameMinLength = 2
	UsernameMaxLength = 30
)

var (
	ErrUsernameEmpty    = errors.New("username is empty")
	ErrUsernameTooShort = errors.New("username is too short")
	ErrUsernameTooLong  = errors.New("username is too long")
)

// NewUsername creates a new Username instance.
func NewUsername(value string) (Username, error) {
	if value == "" {
		return Username{}, ErrUsernameEmpty
	}

	if len([]rune(value)) < UsernameMinLength {
		return Username{}, ErrUsernameTooShort
	}

	if len([]rune(value)) > UsernameMaxLength {
		return Username{}, ErrUsernameTooLong
	}

	return Username{value: value}, nil
}

func (n Username) String() string {
	return n.value
}
