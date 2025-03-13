package middlewares

import (
	"encoding/json"
	"log"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, status int, msg string) {
	if status > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	RespondWithJSON(w, status, errorResponse{
		Error: msg,
	})
}

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
