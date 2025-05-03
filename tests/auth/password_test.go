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
		hash     string
		wantErr  bool
	}{
		{
			name:     "Matching password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Non-matching password",
			password: "wrongpassword",
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  true,
		},
		{
			name:     "Empty hash",
			password: "password123",
			hash:     "",
			wantErr:  true,
		},
		{
			name:     "Invalid hash format",
			password: "password123",
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hash string
			var err error

			// ใช้ hash ที่ generate เฉพาะในกรณีที่ไม่ได้กำหนดไว้
			if tt.hash == "" && tt.name != "Empty hash" {
				hash, err = auth.HashPassword("password123")
				if err != nil {
					t.Fatalf("Failed to hash password: %v", err)
				}
			} else {
				hash = tt.hash
			}

			err = auth.CheckPasswordHash(tt.password, hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
