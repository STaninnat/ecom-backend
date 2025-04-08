package middlewares__test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/middlewares"
)

func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		msg        string
	}{
		{
			name:       "Server error (5xx)",
			statusCode: http.StatusInternalServerError,
			msg:        "Internal server error",
		},
		{
			name:       "Client error (4xx)",
			statusCode: http.StatusBadRequest,
			msg:        "Bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new recorder to capture the response
			rr := httptest.NewRecorder()

			_, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			middlewares.RespondWithError(rr, tt.statusCode, tt.msg)

			if rr.Code != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, rr.Code)
			}

			expectedResponse := `{"error":"` + tt.msg + `"}`
			if rr.Body.String() != expectedResponse {
				t.Errorf("expected body %s, got %s", expectedResponse, rr.Body.String())
			}
		})
	}
}

func TestRespondWithJSON(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		payload    any
		expected   string
	}{
		{
			name:       "Valid response",
			statusCode: http.StatusOK,
			payload: map[string]string{
				"message": "Success",
			},
			expected: `{"message":"Success"}`,
		},
		{
			name:       "Empty response",
			statusCode: http.StatusOK,
			payload:    map[string]string{},
			expected:   `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new recorder to capture the response
			rr := httptest.NewRecorder()

			_, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			middlewares.RespondWithJSON(rr, tt.statusCode, tt.payload)

			if rr.Code != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, rr.Code)
			}

			if rr.Body.String() != tt.expected {
				t.Errorf("expected body %s, got %s", tt.expected, rr.Body.String())
			}
		})
	}
}
