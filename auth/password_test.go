// Package auth provides authentication, token management, validation, and session utilities for the ecom-backend project.
package auth

import (
	"testing"
)

// password_test.go: Tests for password hashing and verification using HashPassword and CheckPasswordHash.

// TestHashPassword verifies password hashing behavior, including error on short input.
func TestHashPassword(t *testing.T) {
	t.Run("valid password", func(t *testing.T) {
		hash, err := HashPassword("longenoughpassword")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hash) == 0 {
			t.Error("expected non-empty hash")
		}
	})

	t.Run("too short", func(t *testing.T) {
		_, err := HashPassword("short")
		if err == nil {
			t.Error("expected error for short password")
		}
	})
}

// TestCheckPasswordHash verifies that password hashes can be correctly validated,
// including failure cases like wrong password or invalid hash format.
func TestCheckPasswordHash(t *testing.T) {
	password := "longenoughpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("correct password", func(t *testing.T) {
		err := CheckPasswordHash(password, hash)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		err := CheckPasswordHash("wrongpassword", hash)
		if err == nil || err.Error() != "password mismatch" {
			t.Errorf("expected password mismatch error, got %v", err)
		}
	})

	t.Run("invalid hash", func(t *testing.T) {
		err := CheckPasswordHash(password, "notahash")
		if err == nil {
			t.Error("expected error for invalid hash")
		}
	})
}
