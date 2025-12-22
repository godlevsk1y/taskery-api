package vo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewDeadline(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		input     time.Time
		wantError bool
	}{
		{"Future date", now.Add(time.Hour), false},
		{"Now", now, true},
		{"Past date", now.Add(-time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDeadline(tt.input)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.input, d.Time())
			}
		})
	}
}

func TestDeadline_IsBefore_IsAfter(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	d, err := NewDeadline(future)
	require.NoError(t, err)

	tests := []struct {
		name             string
		checkTime        time.Time
		isBeforeDeadline bool
		isAfterDeadline  bool
	}{
		{"Check past", past, false, true},
		{"Check now", now, false, true},
		{"Check future", future, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.isBeforeDeadline, d.IsBefore(tt.checkTime))
			require.Equal(t, tt.isAfterDeadline, d.IsAfter(tt.checkTime))
		})
	}
}

func TestDeadline_IsOverdue(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    time.Time
		wantOver bool
	}{
		{"Future deadline", now.Add(time.Hour), false},
		{"Past deadline", now.Add(-time.Hour), true},
		{"Now", now, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Deadline{value: tt.input}
			require.Equal(t, tt.wantOver, d.IsOverdue())
		})
	}
}
