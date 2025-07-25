// Package auth provides authentication, token management, validation, and session utilities for the ecom-backend project.
package auth

import (
	"errors"
	"strings"
	"testing"
)

// session_test.go: Tests for GenerateState function with controlled random reader success and failure scenarios.

type failReader struct{}

// failReader always returns an error to simulate read failure.
func (f *failReader) Read(_ []byte) (int, error) {
	return 0, errors.New("fail")
}

// TestGenerateState_Success verifies GenerateState returns a non-empty, non-default string when RandomReader works.
func TestGenerateState_Success(t *testing.T) {
	old := RandomReader
	defer func() { RandomReader = old }()
	RandomReader = strings.NewReader("abcdefghijklmnopABCDEFGHIJKLMNOP")

	state := GenerateState()
	if state == "" || state == "default_state" {
		t.Errorf("expected non-empty, non-default state, got %q", state)
	}
}

// TestGenerateState_Fail verifies GenerateState returns "default_state" if RandomReader fails.
func TestGenerateState_Fail(t *testing.T) {
	old := RandomReader
	defer func() { RandomReader = old }()
	RandomReader = &failReader{}

	state := GenerateState()
	if state != "default_state" {
		t.Errorf("expected default_state, got %q", state)
	}
}
