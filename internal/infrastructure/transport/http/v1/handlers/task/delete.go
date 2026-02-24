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

type Deleter interface {
	Delete(ctx context.Context, id string, ownerID string) error
}

type DeleteHandler struct {
	deleter  Deleter
	timeout  time.Duration
	logger   *slog.Logger
	validate *validator.Validate
}

func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.Task.Delete"

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	req, ok := handlers.DecodeAndValidate[DeleteRequest](w, r, logger, h.validate)
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

	err := h.deleter.Delete(ctx, req.TaskID, userID)
	if err != nil {
		logger.Error("failed to delete task", slog.String("err", err.Error()))

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

	logger.Info("task deleted")
	handlers.WriteJSON(w, http.StatusNoContent, nil)
}
