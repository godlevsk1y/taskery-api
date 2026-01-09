package vo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/vo"
)

func TestNewUsername(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{name: "empty", input: "", expectedErr: vo.ErrUsernameEmpty},
		{name: "too short - 1 char", input: "a", expectedErr: vo.ErrUsernameTooShort},
		{name: "min length - 2 chars", input: "ab", expected: "ab", expectedErr: nil},
		{name: "normal length", input: "john_doe", expected: "john_doe", expectedErr: nil},
		{name: "max length", input: mkString('a', vo.UsernameMaxLength), expected: mkString('a', vo.UsernameMaxLength), expectedErr: nil},
		{name: "too long", input: mkString('a', vo.UsernameMaxLength+1), expectedErr: vo.ErrUsernameTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := vo.NewUsername(tt.input)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, u.String())
		})
	}
}
