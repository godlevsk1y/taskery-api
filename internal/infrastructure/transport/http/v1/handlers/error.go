package handlers

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"error"`
}

// WriteError writes given code and error to the HTTP response
func WriteError(w http.ResponseWriter, code int, err error) {
	resp := ErrorResponse{Message: err.Error()}

	data, mErr := json.Marshal(resp)
	if mErr != nil {
		http.Error(w, "failed to encode error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(data)
}
