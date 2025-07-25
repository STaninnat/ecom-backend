// Package handlers provides core interfaces, configurations, middleware, and utilities to support HTTP request handling, authentication, logging, and user management in the ecom-backend project.
package handlers

import (
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/middlewares"
)

// handler_ready.go: Provides basic HTTP handlers for service readiness, health status, and error responses.

// HandlerReadiness handles health check requests and returns a simple status response.
// @Summary      Service readiness
// @Description  Returns a simple readiness status for health checks and monitoring
// @Tags         infrastructure
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /v1/readiness [get]
func HandlerReadiness(w http.ResponseWriter, _ *http.Request) {
	// Use a more efficient response structure
	response := map[string]any{
		"status":  "ok",
		"service": "ecom-backend",
	}
	middlewares.RespondWithJSON(w, http.StatusOK, response)
}

// HandlerError handles error requests and returns a standard error response with details.
// @Summary      Error simulation
// @Description  Returns a simulated error response for testing error handling
// @Tags         infrastructure
// @Produce      json
// @Success      500  {object}  map[string]interface{}
// @Router       /v1/errorz [get]
func HandlerError(w http.ResponseWriter, _ *http.Request) {
	response := map[string]any{
		"error":   "Internal server error",
		"code":    "INTERNAL_ERROR",
		"message": "An unexpected error occurred. Please try again later.",
	}
	middlewares.RespondWithJSON(w, http.StatusInternalServerError, response)
}

// HandlerHealth provides a more detailed health check response, including service version and timestamp.
// @Summary      Service health (detailed)
// @Description  Returns a detailed health status including version and timestamp
// @Tags         infrastructure
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /v1/health [get]
func HandlerHealth(w http.ResponseWriter, _ *http.Request) {
	response := map[string]any{
		"status":    "healthy",
		"service":   "ecom-backend",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	middlewares.RespondWithJSON(w, http.StatusOK, response)
}
