package task

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/models"
	"github.com/cyberbrain-dev/taskery-api/internal/domain/task/vo"
	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers"
	myMw "github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/middleware"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

type Finder interface {
	FindByOwner(ctx context.Context, ownerID string) ([]*models.Task, error)
}

type FindByOwnerHandler struct {
	finder   Finder
	timeout  time.Duration
	logger   *slog.Logger
	validate *validator.Validate
}

func NewFindByOwnerHandler(
	finder Finder,
	timeout time.Duration,
	logger *slog.Logger,
	validate *validator.Validate,
) *FindByOwnerHandler {
	return &FindByOwnerHandler{
		finder:   finder,
		timeout:  timeout,
		logger:   logger,
		validate: validate,
	}
}

// @Summary List tasks by owner
// @Description Retrieves all tasks for the authenticated user
// @Tags tasks
// @Produce json
// @Success 200 {object} FindByOwnerResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /tasks [get]
func (h *FindByOwnerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.Task.FindByOwner"

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	userID := myMw.GetUserID(r.Context())
	if userID == "" {
		logger.Error("failed to extract owner id")
		handlers.WriteError(w, http.StatusBadRequest, errors.New("bad request"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	tasks, err := h.finder.FindByOwner(ctx, userID)
	if err != nil {
		logger.Error("failed to find tasks by owner", slog.String("err", err.Error()))

		if errors.Is(err, services.ErrTaskFindByOwnerFailed) {
			handlers.WriteError(w, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}

		handlers.WriteError(w, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	taskDTOs := make([]TaskDTO, len(tasks))
	for i, task := range tasks {
		taskDTOs[i] = TaskDTO{
			ID:          task.ID().String(),
			Title:       task.Title().String(),
			Description: task.Description().String(),
			Deadline:    convertDeadline(task.Deadline()),
			IsCompleted: task.IsCompleted(),
			CompletedAt: task.CompletedAt(),
		}
	}

	handlers.WriteJSON(w, http.StatusOK, FindByOwnerResponse{
		OwnerID: userID,
		Tasks:   taskDTOs,
	})
}

func convertDeadline(deadline *vo.Deadline) *time.Time {
	if deadline == nil {
		return nil
	}
	t := deadline.Time()
	return &t
}
