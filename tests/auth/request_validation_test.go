package auth_test

import (
	"bytes"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/STaninnat/ecom-backend/auth"
)

type DummySignUpParameters struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type DummySignInParameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func TestDecodeAndValidate(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		expectError   bool
		expectedValue any
	}{
		{
			name:        "Valid SignUp JSON",
			body:        `{"name":"John","email":"john@example.com","password":"password123"}`,
			expectError: false,
			expectedValue: &DummySignUpParameters{
				Name:     "John",
				Email:    "john@example.com",
				Password: "password123",
			},
		},
		{
			name:        "Valid SignIn JSON",
			body:        `{"email":"john@example.com","password":"password123"}`,
			expectError: false,
			expectedValue: &DummySignInParameters{
				Email:    "john@example.com",
				Password: "password123",
			},
		},
		{
			name:        "Invalid JSON",
			body:        `{"name":"John","email"}`,
			expectError: true,
		},
		{
			name:        "Missing fields",
			body:        `{"name":""}`,
			expectError: true,
		},
		{
			name:        "Empty body",
			body:        ``,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an HTTP request with the test body
			req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(tt.body)))
			w := httptest.NewRecorder()

			var result any
			var ok bool

			if tt.name == "Valid SignUp JSON" || tt.name == "Invalid JSON" {
				result, ok = auth.DecodeAndValidate[DummySignUpParameters](w, req)
			} else if tt.name == "Valid SignIn JSON" {
				result, ok = auth.DecodeAndValidate[DummySignInParameters](w, req)
			}

			if (ok == false) != tt.expectError {
				t.Errorf("Expected error status: %v, got: %v", tt.expectError, ok)
			}

			if !tt.expectError && !reflect.DeepEqual(result, tt.expectedValue) {
				t.Errorf("Expected value: %v, got: %v", tt.expectedValue, result)
			}
		})
	}
}
