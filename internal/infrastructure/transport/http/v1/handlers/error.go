package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type ErrorResponse struct {
	Error  string       `json:"error,omitempty"`
	Errors []FieldError `json:"errors,omitempty"`
}

// WriteError writes given code and error to the HTTP response.
// WriteError handles validation errors as well.
func WriteError(w http.ResponseWriter, code int, err error) {
	resp := makeResponse(err)

	data, mErr := json.Marshal(resp)
	if mErr != nil {
		http.Error(w, "failed to encode error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(data)
}

func makeResponse(err error) ErrorResponse {
	if vErrs, ok := err.(validator.ValidationErrors); ok {
		fieldErrs := make([]FieldError, 0, len(vErrs))
		for _, vErr := range vErrs {
			fieldErrs = append(fieldErrs, FieldError{Field: vErr.Field(), Error: vErr.Error()})
		}

		return ErrorResponse{Errors: fieldErrs}
	}

	return ErrorResponse{Error: err.Error()}
}
