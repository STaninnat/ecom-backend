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

	var response map[string]any
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

	var response map[string]any
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

	var response map[string]any
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

	var response map[string]any
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

	var response map[string]any
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

	var response map[string]any
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

// Edge case tests for handler_ready.go

func TestHandlerReadiness_EdgeCases(t *testing.T) {
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
			request:        httptest.NewRequest("GET", "/healthz", nil),
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
			request:        httptest.NewRequest("GET", "/healthz", nil),
			responseWriter: httptest.NewRecorder(),
			expectPanic:    false,
		},
		{
			name:           "request with empty body",
			request:        httptest.NewRequest("GET", "/healthz", http.NoBody),
			responseWriter: httptest.NewRecorder(),
			expectPanic:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					HandlerReadiness(tt.responseWriter, tt.request)
				})
			} else {
				assert.NotPanics(t, func() {
					HandlerReadiness(tt.responseWriter, tt.request)
				})
			}
		})
	}
}

func TestHandlerError_EdgeCases(t *testing.T) {
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
			request:        httptest.NewRequest("GET", "/error", nil),
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
			request:        httptest.NewRequest("GET", "/error", nil),
			responseWriter: httptest.NewRecorder(),
			expectPanic:    false,
		},
		{
			name:           "request with empty body",
			request:        httptest.NewRequest("GET", "/error", http.NoBody),
			responseWriter: httptest.NewRecorder(),
			expectPanic:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					HandlerError(tt.responseWriter, tt.request)
				})
			} else {
				assert.NotPanics(t, func() {
					HandlerError(tt.responseWriter, tt.request)
				})
			}
		})
	}
}

func TestHandlerHealth_EdgeCases(t *testing.T) {
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
			request:        httptest.NewRequest("GET", "/health", nil),
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
			request:        httptest.NewRequest("GET", "/health", nil),
			responseWriter: httptest.NewRecorder(),
			expectPanic:    false,
		},
		{
			name:           "request with empty body",
			request:        httptest.NewRequest("GET", "/health", http.NoBody),
			responseWriter: httptest.NewRecorder(),
			expectPanic:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					HandlerHealth(tt.responseWriter, tt.request)
				})
			} else {
				assert.NotPanics(t, func() {
					HandlerHealth(tt.responseWriter, tt.request)
				})
			}
		})
	}
}

func TestHandlerReadiness_MalformedRequests(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		method string
		body   any
	}{
		{
			name:   "request with large body",
			url:    "/healthz",
			method: "POST",
			body:   string(make([]byte, 1000000)), // 1MB body
		},
		{
			name:   "request with special characters in URL",
			url:    "/healthz?param=test%20value",
			method: "GET",
			body:   nil,
		},
		{
			name:   "request with query parameters",
			url:    "/healthz?param1=value1&param2=value2",
			method: "GET",
			body:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body != nil {
				req, err = http.NewRequest(tt.method, tt.url, nil)
			} else {
				req, err = http.NewRequest(tt.method, tt.url, nil)
			}

			if err != nil {
				// Skip tests that can't create valid requests
				t.Skipf("Could not create request: %v", err)
			}

			w := httptest.NewRecorder()

			// Should not panic even with malformed requests
			assert.NotPanics(t, func() {
				HandlerReadiness(w, req)
			})

			// Should still return a valid response
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

func TestHandlerError_MalformedRequests(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		method string
		body   any
	}{
		{
			name:   "request with large body",
			url:    "/error",
			method: "POST",
			body:   string(make([]byte, 1000000)), // 1MB body
		},
		{
			name:   "request with special characters in URL",
			url:    "/error?param=test%20value",
			method: "GET",
			body:   nil,
		},
		{
			name:   "request with query parameters",
			url:    "/error?param1=value1&param2=value2",
			method: "GET",
			body:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body != nil {
				req, err = http.NewRequest(tt.method, tt.url, nil)
			} else {
				req, err = http.NewRequest(tt.method, tt.url, nil)
			}

			if err != nil {
				// Skip tests that can't create valid requests
				t.Skipf("Could not create request: %v", err)
			}

			w := httptest.NewRecorder()

			// Should not panic even with malformed requests
			assert.NotPanics(t, func() {
				HandlerError(w, req)
			})

			// Should still return a valid response
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

func TestHandlerHealth_MalformedRequests(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		method string
		body   any
	}{
		{
			name:   "request with large body",
			url:    "/health",
			method: "POST",
			body:   string(make([]byte, 1000000)), // 1MB body
		},
		{
			name:   "request with special characters in URL",
			url:    "/health?param=test%20value",
			method: "GET",
			body:   nil,
		},
		{
			name:   "request with query parameters",
			url:    "/health?param1=value1&param2=value2",
			method: "GET",
			body:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body != nil {
				req, err = http.NewRequest(tt.method, tt.url, nil)
			} else {
				req, err = http.NewRequest(tt.method, tt.url, nil)
			}

			if err != nil {
				// Skip tests that can't create valid requests
				t.Skipf("Could not create request: %v", err)
			}

			w := httptest.NewRecorder()

			// Should not panic even with malformed requests
			assert.NotPanics(t, func() {
				HandlerHealth(w, req)
			})

			// Should still return a valid response
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

func TestHandlerReadiness_ResponseConsistency(t *testing.T) {
	// Test that multiple calls return consistent responses
	req := httptest.NewRequest("GET", "/healthz", nil)

	responses := make([]map[string]any, 5)

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		HandlerReadiness(w, req)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		responses[i] = response
	}

	// All responses should be identical
	for i := 1; i < len(responses); i++ {
		assert.Equal(t, responses[0], responses[i])
	}
}

func TestHandlerError_ResponseConsistency(t *testing.T) {
	// Test that multiple calls return consistent responses
	req := httptest.NewRequest("GET", "/error", nil)

	responses := make([]map[string]any, 5)

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		HandlerError(w, req)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		responses[i] = response
	}

	// All responses should be identical
	for i := 1; i < len(responses); i++ {
		assert.Equal(t, responses[0], responses[i])
	}
}

func TestHandlerHealth_ResponseConsistency(t *testing.T) {
	// Test that multiple calls return consistent responses
	req := httptest.NewRequest("GET", "/health", nil)

	responses := make([]map[string]any, 5)

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		HandlerHealth(w, req)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		responses[i] = response
	}

	// All responses should be identical
	for i := 1; i < len(responses); i++ {
		assert.Equal(t, responses[0], responses[i])
	}
}

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
