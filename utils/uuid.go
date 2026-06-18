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

// IsValidUUID checks if the given string is a valid UUID.
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

// IsValidObjectID checks if the given string is a valid MongoDB ObjectID (24 hex characters).
func IsValidObjectID(u string) bool {
	if len(u) != 24 {
		return false
	}
	for _, c := range u {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
