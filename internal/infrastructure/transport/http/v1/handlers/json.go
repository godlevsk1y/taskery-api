package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
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
	_, _ = w.Write(data)
}
