package http

import (
	"encoding/json"
	"net/http"
)

// WriteJSONResponse writes a JSON response with the given status code and data
func WriteJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteErrorResponse writes a JSON error response with the given status code and message
func WriteErrorResponse(w http.ResponseWriter, status int, message string) {
	WriteJSONResponse(w, status, map[string]interface{}{
		"error":  message,
		"status": status,
	})
}
