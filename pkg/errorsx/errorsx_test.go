package errorsx_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cyberbrain-dev/taskery-api/pkg/errorsx"
	"github.com/stretchr/testify/require"
)

func TestIsAny(t *testing.T) {
	fooErr := errors.New("foo")
	barErr := errors.New("bar")

	tests := []struct {
		name    string
		err     error
		targets []error
		want    bool
	}{
		{
			name:    "direct match",
			err:     fooErr,
			targets: []error{fooErr},
			want:    true,
		},
		{
			name:    "wrapped match",
			err:     fmt.Errorf("wrapped foo: %w", fooErr),
			targets: []error{fooErr},
			want:    true,
		},
		{
			name:    "not in list",
			err:     fooErr,
			targets: []error{barErr},
			want:    false,
		},
		{
			name:    "empty list",
			err:     fooErr,
			targets: []error{},
			want:    false,
		},
		{
			name:    "multiple targets",
			err:     fmt.Errorf("wrapped foo: %w", fooErr),
			targets: []error{fooErr, fmt.Errorf("wrapped foo: %w", fooErr)},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.True(t, tt.want == errorsx.IsAny(tt.err, tt.targets...))
		})
	}
}
