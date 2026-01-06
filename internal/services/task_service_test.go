package services_test

import (
	"testing"

	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/cyberbrain-dev/taskery-api/internal/services/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewTaskService(t *testing.T) {
	tests := []struct {
		name      string
		tasksRepo services.TaskRepository
		wantErr   error
	}{
		{
			name:      "success",
			tasksRepo: new(mocks.TaskRepository),
			wantErr:   nil,
		},
		{
			name:      "nil tasks repo",
			tasksRepo: nil,
			wantErr:   services.ErrTaskRepositoryNil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := services.NewTaskService(tt.tasksRepo)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, service)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}
