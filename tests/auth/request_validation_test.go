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
		targetType    string
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
			targetType: "signup",
		},
		{
			name:        "Valid SignIn JSON",
			body:        `{"email":"john@example.com","password":"password123"}`,
			expectError: false,
			expectedValue: &DummySignInParameters{
				Email:    "john@example.com",
				Password: "password123",
			},
			targetType: "signin",
		},
		{
			name:        "Invalid JSON",
			body:        `{"name":"John","email"}`,
			expectError: true,
			targetType:  "signup",
		},
		{
			name:        "Missing fields",
			body:        `{"name":""}`,
			expectError: true,
			targetType:  "signup",
		},
		{
			name:        "Empty body",
			body:        ``,
			expectError: true,
			targetType:  "signup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(tt.body)))
			w := httptest.NewRecorder()

			var err error
			var result any

			switch tt.targetType {
			case "signup":
				result, err = auth.DecodeAndValidate[DummySignUpParameters](w, req)
			case "signin":
				result, err = auth.DecodeAndValidate[DummySignInParameters](w, req)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got error: %v", tt.expectError, err)
			}

			if !tt.expectError && !reflect.DeepEqual(result, tt.expectedValue) {
				t.Errorf("Expected value: %+v, got: %+v", tt.expectedValue, result)
			}
		})
	}
}
