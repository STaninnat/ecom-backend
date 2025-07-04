package handlers

import (
	"net/http"

	"github.com/STaninnat/ecom-backend/middlewares"
)

// HandlerReadiness handles health check requests
func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	// Use a more efficient response structure
	response := map[string]any{
		"status":  "ok",
		"service": "ecom-backend",
	}
	middlewares.RespondWithJSON(w, http.StatusOK, response)
}

// HandlerError handles error requests with improved error details
func HandlerError(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"error":   "Internal server error",
		"code":    "INTERNAL_ERROR",
		"message": "An unexpected error occurred. Please try again later.",
	}
	middlewares.RespondWithJSON(w, http.StatusInternalServerError, response)
}

// HandlerHealth provides a more detailed health check
func HandlerHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"status":    "healthy",
		"service":   "ecom-backend",
		"version":   "1.0.0",                // TODO: Get from build info
		"timestamp": "2024-01-01T00:00:00Z", // TODO: Get current timestamp
	}
	middlewares.RespondWithJSON(w, http.StatusOK, response)
}
