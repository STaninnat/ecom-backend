package auth

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

type testReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

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

func TestDecodeAndValidate_InvalidJSON(t *testing.T) {
	body := `{"email":"test@example.com",` // malformed
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	_, err := DecodeAndValidate[testReq](w, r)
	if err == nil || err.Error() != "invalid request format" {
		t.Errorf("expected invalid request format error, got %v", err)
	}
}

func TestDecodeAndValidate_ValidationError(t *testing.T) {
	body := `{"email":"notanemail","password":"short"}`
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	_, err := DecodeAndValidate[testReq](w, r)
	if err == nil || err.Error() == "invalid request format" {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestDecodeAndValidate_UnknownField(t *testing.T) {
	body := `{"email":"test@example.com","password":"longenough","extra":"field"}`
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	_, err := DecodeAndValidate[testReq](w, r)
	if err == nil || !strings.Contains(err.Error(), "invalid request format") {
		t.Errorf("expected invalid request format error for unknown field, got %v", err)
	}
}

func TestDecodeAndValidate_MissingField(t *testing.T) {
	body := `{"email":"test@example.com"}` // missing password
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	_, err := DecodeAndValidate[testReq](w, r)
	if err == nil || !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected validation failed error for missing field, got %v", err)
	}
}
