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
	"github.com/google/uuid"
)

type Creator interface {
	Create(ctx context.Context, cmd services.CreateTaskCommand) error
}

type CreateHandler struct {
	creator  Creator
	timeout  time.Duration
	logger   *slog.Logger
	validate *validator.Validate
}

func NewCreateHandler(
	creator Creator,
	timeout time.Duration,
	logger *slog.Logger,
	validate *validator.Validate) *CreateHandler {

	return &CreateHandler{
		creator:  creator,
		timeout:  timeout,
		logger:   logger,
		validate: validate,
	}
}

func (h *CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.Task.Create"

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	req, ok := handlers.DecodeAndValidate[CreateRequest](w, r, logger, h.validate)
	if !ok {
		return
	}

	// TODO: move the owner ID extraction to a separate function (like chi middleware.GetReqID does)
	ownerIDStr := r.Context().Value(myMw.UserContextKey).(string)

	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		logger.Error("failed to parse owner ID", slog.String("err", err.Error()))
		handlers.WriteError(w, http.StatusBadRequest, errors.New("bad request"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	err = h.creator.Create(ctx, services.CreateTaskCommand{
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     ownerID,
		Deadline:    req.Deadline,
	})
	if err != nil {
		logger.Error("failed to create task", slog.String("err", err.Error()))

		if errors.Is(err, services.ErrTaskExists) {
			handlers.WriteError(w, http.StatusConflict, errors.New("task already exists"))
			return
		}

		if errors.Is(err, services.ErrTaskOwnerNotFound) {
			handlers.WriteError(w, http.StatusNotFound, errors.New("task owner not found"))
			return
		}

		if errors.Is(err, services.ErrTaskCreateFailed) {
			handlers.WriteError(w, http.StatusInternalServerError, errors.New("task creation failed"))
			return
		}

		handlers.WriteError(w, http.StatusBadRequest, err)
		return
	}

	handlers.WriteJSON(w, http.StatusCreated, CreateResponse{
		Title: req.Title,
	})
}
