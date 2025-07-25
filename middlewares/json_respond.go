// Package middlewares provides HTTP middleware components for request processing in the ecom-backend project.
package middlewares

import (
	"encoding/json"
	"log"
	"net/http"
)

// json_respond.go: Utilities for consistent JSON HTTP responses in middleware.

// RespondWithError writes an error response with the given status code and message, and an optional error code.
// Logs 5XX errors and uses RespondWithJSON for consistent formatting.
func RespondWithError(w http.ResponseWriter, status int, msg string, code ...string) {
	if status > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}

	type errorResponse struct {
		Error string `json:"error"`
		Code  string `json:"code,omitempty"`
	}

	errResp := errorResponse{Error: msg}
	if len(code) > 0 && code[0] != "" {
		errResp.Code = code[0]
	}

	RespondWithJSON(w, status, errResp)
}

// RespondWithJSON writes a JSON response with the given status code and payload.
// Sets Content-Type to application/json and handles marshaling and write errors.
func RespondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %s", err)
		http.Error(w, `{"Error": "Internal server error"}`, http.StatusInternalServerError)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
