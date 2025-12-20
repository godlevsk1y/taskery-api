package vo_test

import (
    "testing"

    "github.com/stretchr/testify/require"

    "github.com/cyberbrain-dev/taskery-api/internal/domain/user/vo"
)

func TestNewEmail(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expected    string
        expectedErr error
    }{
        {name: "valid email", input: "user@example.com", expected: "user@example.com", expectedErr: nil},
        {name: "valid email with spaces", input: "  user@example.com  ", expected: "user@example.com", expectedErr: nil},
        {name: "empty", input: "", expectedErr: vo.ErrEmailEmpty},
        {name: "spaces only", input: "   ", expectedErr: vo.ErrEmailEmpty},
        {name: "invalid format", input: "not-an-email", expectedErr: vo.ErrEmailInvalid},
        {name: "missing domain", input: "user@", expectedErr: vo.ErrEmailInvalid},
        {name: "missing local part", input: "@example.com", expectedErr: vo.ErrEmailInvalid},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            e, err := vo.NewEmail(tt.input)
            if tt.expectedErr != nil {
                require.ErrorIs(t, err, tt.expectedErr)
                return
            }

            require.NoError(t, err)
            require.Equal(t, tt.expected, e.String())
        })
    }
}
