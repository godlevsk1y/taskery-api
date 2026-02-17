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
		return ErrorResponse{Errors: makeValidationErrors(vErrs)}
	}

	return ErrorResponse{Error: err.Error()}
}

func makeValidationErrors(errs validator.ValidationErrors) []FieldError {
	fieldErrors := make([]FieldError, 0, len(errs))

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			fieldErrors = append(fieldErrors, FieldError{Field: err.Field(), Error: "field is required"})

		case "email":
			fieldErrors = append(fieldErrors, FieldError{Field: err.Field(), Error: "field is not a valid email"})

		case "printascii":
			fieldErrors = append(
				fieldErrors,
				FieldError{Field: err.Field(), Error: "field contains invalid characters"},
			)

		default:
			fieldErrors = append(fieldErrors, FieldError{Field: err.Field(), Error: "field is invalid"})
		}
	}

	return fieldErrors
}
