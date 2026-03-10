package task

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/handlers"
	myMw "github.com/cyberbrain-dev/taskery-api/internal/infrastructure/transport/http/v1/middleware"
	"github.com/cyberbrain-dev/taskery-api/internal/services"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
)

type Completer interface {
	Complete(ctx context.Context, id string, ownerID string) error
}

type CompleteHandler struct {
	completer Completer
	timeout   time.Duration
	logger    *slog.Logger
	validate  *validator.Validate
}

func NewCompleteHandler(
	completer Completer,
	timeout time.Duration,
	logger *slog.Logger,
	validate *validator.Validate,
) *CompleteHandler {
	return &CompleteHandler{
		completer: completer,
		timeout:   timeout,
		logger:    logger,
		validate:  validate,
	}
}

// @Summary Complete a task
// @Description Marks a task as completed for the authenticated user
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body CompleteRequest true "Task completion request"
// @Success 204 {object} nil
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 403 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /tasks/complete [patch]
func (h *CompleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.Task.Complete"

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	req, ok := handlers.DecodeAndValidate[CompleteRequest](w, r, logger, h.validate)
	if !ok {
		return
	}

	userID := myMw.GetUserID(r.Context())
	if userID == "" {
		logger.Error("failed to extract owner id")
		handlers.WriteError(w, http.StatusBadRequest, errors.New("bad request"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	err := h.completer.Complete(ctx, req.TaskID, userID)
	if err != nil {
		logger.Error("failed to complete task", slog.String("err", err.Error()))

		if errors.Is(err, services.ErrTaskNotFound) {
			handlers.WriteError(w, http.StatusNotFound, errors.New("task not found"))
			return
		}

		if errors.Is(err, services.ErrTaskAccessDenied) {
			handlers.WriteError(w, http.StatusForbidden, errors.New("access denied"))
			return
		}

		handlers.WriteError(w, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	logger.Info("task completed")
	handlers.WriteJSON(w, http.StatusNoContent, nil)
}
