// Package auth provides authentication, token management, validation, and session utilities for the ecom-backend project.
package auth

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

// request_validation_test.go: Tests for DecodeAndValidate generic function handling JSON decoding and validation of request bodies.

type testReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// TestDecodeAndValidate_Valid ensures decoding and validation succeed for valid input.
func TestDecodeAndValidate_Valid(t *testing.T) {
	body := `{"email":"test@example.com","password":"longenough"}`
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	params, err := DecodeAndValidate[testReq](w, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.Email != "test@example.com" || params.Password != "longenough" {
		t.Errorf("unexpected params: %+v", params)
	}
}

// TestDecodeAndValidate_InvalidJSON expects error on malformed JSON input.
func TestDecodeAndValidate_InvalidJSON(t *testing.T) {
	body := `{"email":"test@example.com",` // malformed
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	_, err := DecodeAndValidate[testReq](w, r)
	if err == nil || err.Error() != "invalid request format" {
		t.Errorf("expected invalid request format error, got %v", err)
	}
}

// TestDecodeAndValidate_ValidationError expects error when validation rules fail.
func TestDecodeAndValidate_ValidationError(t *testing.T) {
	body := `{"email":"notanemail","password":"short"}`
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	_, err := DecodeAndValidate[testReq](w, r)
	if err == nil || err.Error() == "invalid request format" {
		t.Errorf("expected validation error, got %v", err)
	}
}

// TestDecodeAndValidate_UnknownField expects error if JSON contains unexpected fields.
func TestDecodeAndValidate_UnknownField(t *testing.T) {
	body := `{"email":"test@example.com","password":"longenough","extra":"field"}`
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	_, err := DecodeAndValidate[testReq](w, r)
	if err == nil || !strings.Contains(err.Error(), "invalid request format") {
		t.Errorf("expected invalid request format error for unknown field, got %v", err)
	}
}

// TestDecodeAndValidate_MissingField expects validation failure if required fields are missing.
func TestDecodeAndValidate_MissingField(t *testing.T) {
	body := `{"email":"test@example.com"}` // missing password
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	_, err := DecodeAndValidate[testReq](w, r)
	if err == nil || !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected validation failed error for missing field, got %v", err)
	}
}
