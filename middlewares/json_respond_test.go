package middlewares

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRespondWithError tests error response generation with various scenarios
// It verifies that error responses have correct status codes, JSON structure, and optional error codes
func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		msg        string
		code       []string
		expectCode string
	}{
		{"basic error", 400, "bad request", nil, ""},
		{"with code", 401, "unauthorized", []string{"AUTH_FAIL"}, "AUTH_FAIL"},
		{"server error", 500, "internal", nil, ""},
		{"with empty code", 403, "forbidden", []string{""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			RespondWithError(rr, tt.status, tt.msg, tt.code...)
			if rr.Code != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, rr.Code)
			}
			var resp struct {
				Error string `json:"error"`
				Code  string `json:"code,omitempty"`
			}
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if resp.Error != tt.msg {
				t.Errorf("expected error %q, got %q", tt.msg, resp.Error)
			}
			if resp.Code != tt.expectCode {
				t.Errorf("expected code %q, got %q", tt.expectCode, resp.Code)
			}
			if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
				t.Errorf("expected Content-Type application/json, got %q", ct)
			}
		})
	}
}

// TestRespondWithJSON tests JSON response generation with valid payloads
// It verifies that JSON responses have correct status codes, content type, and marshaled data
func TestRespondWithJSON(t *testing.T) {
	type payload struct {
		Msg string `json:"msg"`
	}
	rr := httptest.NewRecorder()
	data := payload{"hello"}
	RespondWithJSON(rr, 201, data)
	if rr.Code != 201 {
		t.Errorf("expected status 201, got %d", rr.Code)
	}
	var got payload
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got != data {
		t.Errorf("expected body %+v, got %+v", data, got)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

// TestRespondWithJSON_MarshalError tests JSON response generation when marshaling fails
// It verifies that the function handles marshaling errors gracefully and returns 500 status
func TestRespondWithJSON_MarshalError(t *testing.T) {
	rr := httptest.NewRecorder()
	// channels cannot be marshaled to JSON
	RespondWithJSON(rr, 200, make(chan int))
	if rr.Code != 500 {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Internal server error") {
		t.Errorf("expected internal server error message, got %q", rr.Body.String())
	}
}
