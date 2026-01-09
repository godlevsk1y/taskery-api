package vo_test

import (
	"testing"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/vo"
	"github.com/stretchr/testify/require"
)

func TestNewDescription(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError error
	}{
		{
			name:      "empty description",
			input:     "",
			wantError: nil,
		},
		{
			name:      "valid description",
			input:     "This is a valid task description.",
			wantError: nil,
		},
		{
			name:      "description too long",
			input:     string(make([]rune, vo.DescriptionMaxLength+1)),
			wantError: vo.ErrDescriptionTooLong,
		},
		{
			name:      "description exactly max length",
			input:     string(make([]rune, vo.DescriptionMaxLength)),
			wantError: nil,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			desc, err := vo.NewDescription(tt.input)
			if tt.wantError != nil {
				require.ErrorIs(t, err, tt.wantError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.input, desc.String())
			}
		})
	}
}
