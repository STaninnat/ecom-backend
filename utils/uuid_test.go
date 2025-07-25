// Package utils provides utility functions and helpers used throughout the ecom-backend project.
package utils

import (
	"testing"

	"github.com/google/uuid"
)

// uuid_test.go: Tests for UUID helper functions in utils, ensuring the UUIDs generated are valid and non-nil.

// TestNewUUIDString verifies that NewUUIDString returns a valid UUID string.
func TestNewUUIDString(t *testing.T) {
	id := NewUUIDString()

	if _, err := uuid.Parse(id); err != nil {
		t.Errorf("NewUUIDString returned invalid UUID string: %v", err)
	}
}

// TestNewUUID verifies that NewUUID returns a non-nil UUID object.
func TestNewUUID(t *testing.T) {
	id := NewUUID()

	if id == uuid.Nil {
		t.Error("NewUUID returned Nil UUID")
	}
}
