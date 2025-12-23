package vo

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Password is a VO that represents a password hash.
type Password struct {
	value []byte
}

const (
	PasswordMinLength = 8
	PasswordMaxLength = 72
)

var (
	ErrPasswordEmpty    = errors.New("password is empty")
	ErrPasswordTooShort = errors.New("password is too short")
	ErrPasswordTooLong  = errors.New("password is too long")
	ErrPasswordInvalid  = errors.New("password is invalid")
	ErrPasswordHashing  = errors.New("failed to hash password")

	ErrPasswordVerifyFailed = errors.New("failed to verify the password")
	ErrPassowrdNotMatch     = errors.New("password does not match the hash")
)

// NewPassword creates a new Password instance.
func NewPassword(raw string) (Password, error) {
	if raw == "" {
		return Password{}, ErrPasswordEmpty
	}

	if !isASCII(raw) {
		return Password{}, ErrPasswordInvalid
	}

	if len(raw) < PasswordMinLength {
		return Password{}, ErrPasswordTooShort
	}

	if len(raw) > PasswordMaxLength {
		return Password{}, ErrPasswordTooLong
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, ErrPasswordHashing
	}

	return Password{value: hash}, nil
}

// NewPasswordFromHash creates a new Password instance from an existing hash.
func NewPasswordFromHash(hash []byte) Password {
	hashCopy := make([]byte, len(hash))
	copy(hashCopy, hash)

	return Password{value: hashCopy}
}

// Verify checks if the provided raw password matches the hash.
func (p Password) Verify(raw string) error {
	err := bcrypt.CompareHashAndPassword(p.value, []byte(raw))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrPassowrdNotMatch
	}
	if err != nil {
		return fmt.Errorf("%w: %s", ErrPasswordVerifyFailed, err)
	}

	return nil
}

// String returns the password hash as string.
func (p Password) String() string {
	return string(p.value)
}

// IsASCII checks if the string includes only ASCII characters.
func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}

	return true
}
