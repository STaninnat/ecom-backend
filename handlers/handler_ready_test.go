package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerReadiness(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HandlerReadiness(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "ecom-backend", response["service"])
}

func TestHandlerError(t *testing.T) {
	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	HandlerError(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "Internal server error", response["error"])
	assert.Equal(t, "INTERNAL_ERROR", response["code"])
	assert.Equal(t, "An unexpected error occurred. Please try again later.", response["message"])
}

func TestHandlerHealth(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	HandlerHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "ecom-backend", response["service"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "2024-01-01T00:00:00Z", response["timestamp"])
}

func TestHandlerReadiness_ResponseStructure(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HandlerReadiness(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check that response has exactly the expected fields
	expectedKeys := []string{"status", "service"}
	for _, key := range expectedKeys {
		assert.Contains(t, response, key)
	}
	assert.Len(t, response, len(expectedKeys))
}

func TestHandlerError_ResponseStructure(t *testing.T) {
	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	HandlerError(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check that response has exactly the expected fields
	expectedKeys := []string{"error", "code", "message"}
	for _, key := range expectedKeys {
		assert.Contains(t, response, key)
	}
	assert.Len(t, response, len(expectedKeys))
}

func TestHandlerHealth_ResponseStructure(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	HandlerHealth(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check that response has exactly the expected fields
	expectedKeys := []string{"status", "service", "version", "timestamp"}
	for _, key := range expectedKeys {
		assert.Contains(t, response, key)
	}
	assert.Len(t, response, len(expectedKeys))
}

func TestHandlerReadiness_DifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/healthz", nil)
			w := httptest.NewRecorder()

			HandlerReadiness(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

func TestHandlerError_DifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/error", nil)
			w := httptest.NewRecorder()

			HandlerError(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

func TestHandlerHealth_DifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health", nil)
			w := httptest.NewRecorder()

			HandlerHealth(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}
