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

type Updater interface {
	Update(ctx context.Context, id string, ownerID string, cmd services.UpdateTaskCommand) error
}

type UpdateHandler struct {
	updater  Updater
	timeout  time.Duration
	logger   *slog.Logger
	validate *validator.Validate
}

func NewUpdateHandler(
	updater Updater,
	timeout time.Duration,
	logger *slog.Logger,
	validate *validator.Validate) *UpdateHandler {

	return &UpdateHandler{
		updater:  updater,
		timeout:  timeout,
		logger:   logger,
		validate: validate,
	}
}

func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.Task.Update"

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	req, ok := handlers.DecodeAndValidate[UpdateRequest](w, r, logger, h.validate)
	if !ok {
		return
	}

	taskID := req.TaskID
	ownerID := myMw.GetUserID(r.Context())
	if ownerID == "" {
		logger.Error("failed to extract owner id")
		handlers.WriteError(w, http.StatusBadRequest, errors.New("bad request"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	err := h.updater.Update(ctx, taskID, ownerID, services.UpdateTaskCommand{
		Title:       req.Title,
		Description: req.Description,
		Deadline:    req.Deadline,
	})
	if err != nil {
		logger.Error("failed to update task", slog.String("err", err.Error()))

		if errors.Is(err, services.ErrTaskNotFound) {
			handlers.WriteError(w, http.StatusNotFound, errors.New("task was not found"))
			return
		}

		if errors.Is(err, services.ErrTaskAccessDenied) {
			handlers.WriteError(w, http.StatusForbidden, errors.New("task access denied"))
			return
		}

		if errors.Is(err, services.ErrTaskUpdateFailed) {
			handlers.WriteError(w, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}

		handlers.WriteError(w, http.StatusBadRequest, err)
		return
	}

	handlers.WriteJSON(w, http.StatusOK, UpdateResponse{
		TaskID: taskID,
	})
}
