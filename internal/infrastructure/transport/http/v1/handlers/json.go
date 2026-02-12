package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// WriteJSON writes a code and an object that can be serialized to JSON to the HTTP response
func WriteJSON(w http.ResponseWriter, code int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, errors.New("failed to encode JSON"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if v != nil {
		_, _ = w.Write(data)
	}
}

// DecodeAndValidate decodes the JSON request body into a value of type T
// and validates it using validate.
//
// If decoding or validation fails, DecodeAndValidate writes an HTTP 400
// error response, logs the error using logger, and reports failure.
// A request with an empty body is treated as an error.
//
// It returns a pointer to the decoded value and reports whether decoding
// and validation succeeded.
func DecodeAndValidate[T any](
	w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
	validate *validator.Validate) (*T, bool) {

	var req T
	err := json.NewDecoder(r.Body).Decode(&req)
	if errors.Is(err, io.EOF) {
		logger.Error("request body is empty")
		WriteError(w, http.StatusBadRequest, errors.New("request body is empty"))
		return nil, false
	}
	if err != nil {
		logger.Error("failed to decode request body")
		WriteError(w, http.StatusBadRequest, errors.New("failed to decode request body"))
		return nil, false
	}

	if err := validate.Struct(req); err != nil {
		logger.Error("failed to validate request body", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, err)
		return nil, false
	}

	return &req, true
}
