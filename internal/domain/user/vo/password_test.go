package vo_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/user/vo"
)

func TestNewPassword(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{name: "empty", input: "", expectedErr: vo.ErrPasswordEmpty},
		{name: "too short", input: mkString('a', vo.PasswordMinLength-1), expectedErr: vo.ErrPasswordTooShort},
		{name: "too long", input: mkString('a', vo.PasswordMaxLength+1), expectedErr: vo.ErrPasswordTooLong},
		{name: "non-ascii invalid", input: "пароль123", expectedErr: vo.ErrPasswordInvalid},
		{name: "valid ascii min length", input: mkString('a', vo.PasswordMinLength), expectedErr: nil},
		{name: "valid ascii typical", input: "Password123!", expectedErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := vo.NewPassword(tt.input)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.NoError(t, p.Verify(tt.input))
			require.NotEmpty(t, p.String())
			require.NotEqual(t, tt.input, p.String())
		})
	}
}

func TestNewPasswordFromHashAndVerify(t *testing.T) {
	raw := "Password123!"
	p1, err := vo.NewPassword(raw)
	require.NoError(t, err)

	require.NoError(t, p1.Verify(raw))

	p2 := vo.NewPasswordFromHash([]byte(p1.String()))

	require.NoError(t, p2.Verify(raw))

	require.Equal(t, p1.String(), p2.String())
}

func mkString(ch rune, n int) string {
	b := make([]rune, n)
	for i := range n {
		b[i] = ch
	}
	return string(b)
}
