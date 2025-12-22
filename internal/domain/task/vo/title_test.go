package vo_test

import (
	"testing"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/vo"
	"github.com/stretchr/testify/require"
)

func TestNewTitle(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   error
		wantValue string
	}{
		{
			name:      "valid title",
			input:     "My Task",
			wantErr:   nil,
			wantValue: "My Task",
		},
		{
			name:      "empty string",
			input:     "",
			wantErr:   vo.ErrTitleEmpty,
			wantValue: "",
		},
		{
			name:      "string with only spaces",
			input:     "   ",
			wantErr:   vo.ErrTitleEmpty,
			wantValue: "",
		},
		{
			name:      "title too long",
			input:     "This title is definitely way too long to be accepted by the VO",
			wantErr:   vo.ErrTitleTooLong,
			wantValue: "",
		},
		{
			name:      "title with leading and trailing spaces",
			input:     "   Task with spaces   ",
			wantErr:   nil,
			wantValue: "Task with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, err := vo.NewTitle(tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantValue, title.String())
			}
		})
	}
}
