package auth_test

import (
	"testing"

	"github.com/STaninnat/ecom-backend/auth"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectedErr bool
	}{
		{"Empty password", "", true},                                                  // Test case with an empty password
		{"Short password", "123", true},                                               // Test case with a short password
		{"Common password", "password123", false},                                     // Test case with a common password
		{"Long password", "aVeryLongPasswordThatShouldStillBeHashedCorrectly", false}, // Test case with a long password
		{"Password with special characters", "P@ssw0rd!#", false},                     // Test case with special characters
		{"Password with spaces", "p a s s w o r d", false},                            // Test case with spaces in pass
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := auth.HashPassword(tt.password)
			if (err != nil) != tt.expectedErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.expectedErr)
			}

			if err == nil {
				if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(tt.password)); err != nil {
					t.Errorf("HashPassword() generated an incorrect hash for password: %v", tt.password)
				}
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		match    bool
		wantErr  bool
	}{
		{"Matching password", "password123", true, false},       // Test with a matching password
		{"Non-matching password", "wrongpassword", false, true}, // Test with a non-matching password
		{"Empty password", "", false, true},                     // Test with an empty password
		{"Empty hash", "password123", false, true},              // Test with an empty hash
		{"Invalid hash format", "password123", false, true},     // Test with an invalid hash format
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hash string
			var err error

			// Only hash the password if it's needed for the test
			if tt.match || tt.password != "" {
				hash, err = auth.HashPassword("password123")
				if err != nil {
					t.Fatalf("Failed to hash password: %v", err)
				}
			}

			if tt.name == "Invalid hash format" {
				hash = "invalidhash"
			}

			if tt.name == "Empty hash" {
				hash = ""
			}

			match, err := auth.CheckPasswordHash(tt.password, hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if match != tt.match {
				t.Errorf("CheckPasswordHash() match = %v, want %v", match, tt.match)
			}
		})
	}
}
