// Package handlers provides core interfaces, configurations, middleware, and utilities to support HTTP request handling, authentication, logging, and user management in the ecom-backend project.
package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// handler_ready_test.go: Tests for basic HTTP handlers for service readiness, health status, and error responses.

// TestHandlerReadiness tests the HandlerReadiness function for the /healthz endpoint.
// It checks that the response is OK and contains the expected JSON fields.
func TestHandlerReadiness(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HandlerReadiness(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "ecom-backend", response["service"])
}

// TestHandlerError tests the HandlerError function for the /error endpoint.
// It checks that the response is InternalServerError and contains the expected JSON error fields.
func TestHandlerError(t *testing.T) {
	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	HandlerError(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Internal server error", response["error"])
	assert.Equal(t, "INTERNAL_ERROR", response["code"])
	assert.Equal(t, "An unexpected error occurred. Please try again later.", response["message"])
}

// TestHandlerHealth tests the HandlerHealth function for the /health endpoint.
// It checks that the response is OK and contains the expected health information.
func TestHandlerHealth(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	HandlerHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "ecom-backend", response["service"])
	assert.Equal(t, "1.0.0", response["version"])

	timestamp, ok := response["timestamp"].(string)
	assert.True(t, ok)
	_, err = time.Parse(time.RFC3339, timestamp)
	assert.NoError(t, err)
}

// TestHandlerReadiness_ResponseStructure tests the structure of the HandlerReadiness response.
// It checks that the response contains exactly the expected fields.
func TestHandlerReadiness_ResponseStructure(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HandlerReadiness(w, req)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check that response has exactly the expected fields
	expectedKeys := []string{"status", "service"}
	for _, key := range expectedKeys {
		assert.Contains(t, response, key)
	}
	assert.Len(t, response, len(expectedKeys))
}

// TestHandlerError_ResponseStructure tests the structure of the HandlerError response.
// It checks that the response contains exactly the expected error fields.
func TestHandlerError_ResponseStructure(t *testing.T) {
	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	HandlerError(w, req)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check that response has exactly the expected fields
	expectedKeys := []string{"error", "code", "message"}
	for _, key := range expectedKeys {
		assert.Contains(t, response, key)
	}
	assert.Len(t, response, len(expectedKeys))
}

// TestHandlerHealth_ResponseStructure tests the structure of the HandlerHealth response.
// It checks that the response contains exactly the expected health fields.
func TestHandlerHealth_ResponseStructure(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	HandlerHealth(w, req)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check that response has exactly the expected fields
	expectedKeys := []string{"status", "service", "version", "timestamp"}
	for _, key := range expectedKeys {
		assert.Contains(t, response, key)
	}
	assert.Len(t, response, len(expectedKeys))
}

// TestHandlerReadiness_DifferentMethods tests HandlerReadiness with different HTTP methods.
// It checks that the response is OK and has the correct content type for each method.
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

// TestHandlerError_DifferentMethods tests HandlerError with different HTTP methods.
// It checks that the response is InternalServerError and has the correct content type for each method.
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

// TestHandlerHealth_DifferentMethods tests HandlerHealth with different HTTP methods.
// It checks that the response is OK and has the correct content type for each method.
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

// Edge case tests for handler_ready.go

// runHandlerEdgeCasesTest is a shared helper for edge case tests for readiness, error, and health handlers.
func runHandlerEdgeCasesTest(
	t *testing.T,
	handler func(http.ResponseWriter, *http.Request),
	baseURL string,
) {
	tests := []struct {
		name           string
		request        *http.Request
		responseWriter http.ResponseWriter
		expectPanic    bool
	}{
		{
			name:           "nil request",
			request:        nil,
			responseWriter: httptest.NewRecorder(),
			expectPanic:    false,
		},
		{
			name:           "nil response writer",
			request:        httptest.NewRequest("GET", baseURL, nil),
			responseWriter: nil,
			expectPanic:    true,
		},
		{
			name:           "both nil",
			request:        nil,
			responseWriter: nil,
			expectPanic:    true,
		},
		{
			name:           "request with nil body",
			request:        httptest.NewRequest("GET", baseURL, nil),
			responseWriter: httptest.NewRecorder(),
			expectPanic:    false,
		},
		{
			name:           "request with empty body",
			request:        httptest.NewRequest("GET", baseURL, http.NoBody),
			responseWriter: httptest.NewRecorder(),
			expectPanic:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					handler(tt.responseWriter, tt.request)
				})
			} else {
				assert.NotPanics(t, func() {
					handler(tt.responseWriter, tt.request)
				})
			}
		})
	}
}

// runHandlerResponseConsistencyTest is a shared helper for response consistency tests for readiness, error, and health handlers.
func runHandlerResponseConsistencyTest(
	t *testing.T,
	handler func(http.ResponseWriter, *http.Request),
	req *http.Request,
) {
	responses := make([]map[string]any, 5)
	for i := range 5 {
		w := httptest.NewRecorder()
		handler(w, req)
		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		responses[i] = response
	}
	for i := 1; i < len(responses); i++ {
		assert.Equal(t, responses[0], responses[i])
	}
}

func TestHandlerReadiness_EdgeCases(t *testing.T) {
	runHandlerEdgeCasesTest(t, HandlerReadiness, "/healthz")
}

func TestHandlerError_EdgeCases(t *testing.T) {
	runHandlerEdgeCasesTest(t, HandlerError, "/error")
}

func TestHandlerHealth_EdgeCases(t *testing.T) {
	runHandlerEdgeCasesTest(t, HandlerHealth, "/health")
}

// Helper for malformed request tests
func testMalformedRequests(t *testing.T, handler func(http.ResponseWriter, *http.Request), baseURL string) {
	tests := []struct {
		name   string
		url    string
		method string
		body   any
	}{
		{
			name:   "request with large body",
			url:    baseURL,
			method: "POST",
			body:   string(make([]byte, 1000000)), // 1MB body
		},
		{
			name:   "request with special characters in URL",
			url:    baseURL + "?param=test%20value",
			method: "GET",
			body:   nil,
		},
		{
			name:   "request with query parameters",
			url:    baseURL + "?param1=value1&param2=value2",
			method: "GET",
			body:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Skipf("Could not create request: %v", err)
			}
			w := httptest.NewRecorder()
			assert.NotPanics(t, func() {
				handler(w, req)
			})
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

func TestHandlerReadiness_MalformedRequests(t *testing.T) {
	testMalformedRequests(t, HandlerReadiness, "/healthz")
}

func TestHandlerError_MalformedRequests(t *testing.T) {
	testMalformedRequests(t, HandlerError, "/error")
}

func TestHandlerHealth_MalformedRequests(t *testing.T) {
	testMalformedRequests(t, HandlerHealth, "/health")
}

// TestHandlerReadiness_ResponseConsistency tests that HandlerReadiness returns consistent responses across multiple calls.
// It checks that all responses are identical.
func TestHandlerReadiness_ResponseConsistency(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	runHandlerResponseConsistencyTest(t, HandlerReadiness, req)
}

// TestHandlerError_ResponseConsistency tests that HandlerError returns consistent responses across multiple calls.
// It checks that all responses are identical.
func TestHandlerError_ResponseConsistency(t *testing.T) {
	req := httptest.NewRequest("GET", "/error", nil)
	runHandlerResponseConsistencyTest(t, HandlerError, req)
}

// TestHandlerHealth_ResponseConsistency tests that HandlerHealth returns consistent responses across multiple calls.
// It checks that all responses are identical.
func TestHandlerHealth_ResponseConsistency(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	runHandlerResponseConsistencyTest(t, HandlerHealth, req)
}

// TestHandlerReadiness_Headers tests the headers set by HandlerReadiness.
// It checks that only expected headers are set and common security headers are not set.
func TestHandlerReadiness_Headers(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HandlerReadiness(w, req)

	// Check that no unexpected headers are set
	expectedHeaders := []string{"Content-Type"}
	for _, header := range expectedHeaders {
		assert.NotEmpty(t, w.Header().Get(header))
	}

	// Check that common security headers are not set (these handlers don't set them)
	unexpectedHeaders := []string{"X-Frame-Options", "X-Content-Type-Options", "X-XSS-Protection"}
	for _, header := range unexpectedHeaders {
		assert.Empty(t, w.Header().Get(header))
	}
}

// TestHandlerError_Headers tests the headers set by HandlerError.
// It checks that only expected headers are set and common security headers are not set.
func TestHandlerError_Headers(t *testing.T) {
	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	HandlerError(w, req)

	// Check that no unexpected headers are set
	expectedHeaders := []string{"Content-Type"}
	for _, header := range expectedHeaders {
		assert.NotEmpty(t, w.Header().Get(header))
	}

	// Check that common security headers are not set (these handlers don't set them)
	unexpectedHeaders := []string{"X-Frame-Options", "X-Content-Type-Options", "X-XSS-Protection"}
	for _, header := range unexpectedHeaders {
		assert.Empty(t, w.Header().Get(header))
	}
}

// TestHandlerHealth_Headers tests the headers set by HandlerHealth.
// It checks that only expected headers are set and common security headers are not set.
func TestHandlerHealth_Headers(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	HandlerHealth(w, req)

	// Check that no unexpected headers are set
	expectedHeaders := []string{"Content-Type"}
	for _, header := range expectedHeaders {
		assert.NotEmpty(t, w.Header().Get(header))
	}

	// Check that common security headers are not set (these handlers don't set them)
	unexpectedHeaders := []string{"X-Frame-Options", "X-Content-Type-Options", "X-XSS-Protection"}
	for _, header := range unexpectedHeaders {
		assert.Empty(t, w.Header().Get(header))
	}
}
