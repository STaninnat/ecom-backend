// Package utils provides utility functions and helpers used throughout the ecom-backend project.
package utils

import "github.com/google/uuid"

// uuid.go: This file provides helper functions for generating UUIDs, both as raw UUID objects and as string representations.

// NewUUIDString returns a newly generated UUID as a string.
// It wraps uuid.New().String() from the google/uuid package.
func NewUUIDString() string {
	return uuid.New().String()
}

// NewUUID returns a newly generated UUID as a uuid.UUID object.
// It wraps uuid.New() from the google/uuid package.
func NewUUID() uuid.UUID {
	return uuid.New()
}
