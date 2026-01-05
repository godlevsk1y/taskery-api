package models_test

import (
	"testing"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/models"
	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/vo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewTask(t *testing.T) {
	t.Parallel()

	validOwner := uuid.New()

	tests := []struct {
		name        string
		title       string
		description string
		owner       uuid.UUID

		expectedError error
	}{
		{
			name:          "success",
			title:         "Valid title",
			description:   "Valid description",
			owner:         validOwner,
			expectedError: nil,
		},
		{
			name:          "empty title",
			title:         "",
			description:   "Valid description",
			owner:         validOwner,
			expectedError: vo.ErrTitleEmpty,
		},
		{
			name:          "too long title",
			title:         string(make([]rune, vo.TitleMaxLength+1)),
			description:   "Valid description",
			owner:         validOwner,
			expectedError: vo.ErrTitleTooLong,
		},
		{
			name:          "too long description",
			title:         "Valid title",
			description:   string(make([]rune, vo.DescriptionMaxLength+1)),
			owner:         validOwner,
			expectedError: vo.ErrDescriptionTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskEntity, err := models.NewTask(tt.title, tt.description, tt.owner)

			if tt.expectedError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedError)
				require.Nil(t, taskEntity)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, taskEntity)

			require.NotEqual(t, uuid.Nil, taskEntity.ID())
			require.Equal(t, tt.owner, taskEntity.Owner())

			require.Equal(t, tt.title, taskEntity.Title().String())
			require.Equal(t, tt.description, taskEntity.Description().String())

			require.False(t, taskEntity.IsCompleted())
			require.Nil(t, taskEntity.Deadline())
			require.Nil(t, taskEntity.CompletedAt())
		})
	}
}

func TestNewTaskWithDeadline(t *testing.T) {
	t.Parallel()

	validOwner := uuid.New()
	validDeadline := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name        string
		title       string
		description string
		owner       uuid.UUID
		deadline    time.Time

		expectedErr error
	}{
		{
			name:        "success",
			title:       "Test title",
			description: "Test description",
			owner:       validOwner,
			deadline:    validDeadline,
			expectedErr: nil,
		},
		{
			name:        "invalid title",
			title:       "",
			description: "Test description",
			owner:       validOwner,
			deadline:    validDeadline,
			expectedErr: vo.ErrTitleEmpty,
		},
		{
			name:        "invalid deadline",
			title:       "Test title",
			description: "Test description",
			owner:       validOwner,
			deadline:    time.Now().Add(-24 * time.Hour),
			expectedErr: vo.ErrDeadlineBeforeNow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := models.NewTaskWithDeadline(
				tt.title,
				tt.description,
				tt.owner,
				tt.deadline,
			)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
				require.Nil(t, task)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, task)

			deadline := task.Deadline()
			require.NotNil(t, deadline)
			require.True(t, tt.deadline.Equal(deadline.Time()))
		})
	}
}

func TestTask_SetDeadline(t *testing.T) {
	validDeadline := time.Now().Add(24 * time.Hour)
	invalidDeadline := time.Time{}

	tests := []struct {
		name     string
		deadline time.Time

		expectedErr error
	}{
		{
			name:        "success",
			deadline:    validDeadline,
			expectedErr: nil,
		},
		{
			name:        "invalid deadline",
			deadline:    invalidDeadline,
			expectedErr: vo.ErrDeadlineBeforeNow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &models.Task{}

			err := task.SetDeadline(tt.deadline)

			if tt.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedErr)
				require.Nil(t, task.Deadline())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, task.Deadline())
			require.Equal(t, tt.deadline, task.Deadline().Time())
		})
	}
}
