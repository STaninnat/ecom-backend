// Package userhandlers provides HTTP handlers and services for user-related operations, including user retrieval, updates, and admin role management, with proper error handling and logging.
package userhandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_update_user_test.go: Tests for HandlerUpdateUser covering success, validation errors, service errors, and missing user context.

// TestHandlerUpdateUser_Success tests that HandlerUpdateUser successfully updates
// user information when valid parameters are provided
func TestHandlerUpdateUser_Success(t *testing.T) {
	mockService := new(mockUpdateUserService)
	mockLogger := new(mockUpdateHandlerLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "A", Email: "a@b.com", Phone: "123", Address: "Addr"}
	mockService.On("UpdateUser", mock.Anything, user, params).Return(nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "update_user", "User info updated", mock.Anything, mock.Anything).Return()

	body, _ := json.Marshal(params)
	r := httptest.NewRequest("PUT", "/user", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	cfg.HandlerUpdateUser(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp handlers.HandlerResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	assert.Equal(t, "Updated user info successful", resp.Message)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateUser_ValidationError tests that HandlerUpdateUser returns a validation
// error when the request payload contains invalid data types
func TestHandlerUpdateUser_ValidationError(t *testing.T) {
	mockService := new(mockUpdateUserService)
	mockLogger := new(mockUpdateHandlerLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	user := database.User{ID: "u1"}
	// Send invalid JSON that will fail auth.DecodeAndValidate
	invalidJSON := `{"name": "A", "email": "a@b.com", "phone": 123}` // phone should be string, not number
	body := []byte(invalidJSON)
	r := httptest.NewRequest("PUT", "/user", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	// Expect error logging for invalid JSON
	mockLogger.On("LogHandlerError", mock.Anything, "update_user", "invalid_request", "Invalid update payload", mock.Anything, mock.Anything, mock.Anything).Return()

	cfg.HandlerUpdateUser(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotEmpty(t, w.Body.String())
	mockService.AssertNotCalled(t, "UpdateUser")
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateUser_InvalidJSON tests that HandlerUpdateUser returns an error
// when the request body contains malformed JSON
func TestHandlerUpdateUser_InvalidJSON(t *testing.T) {
	mockService := new(mockUpdateUserService)
	mockLogger := new(mockUpdateHandlerLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	user := database.User{ID: "u1"}
	invalidJSON := `{"name": "A", "email": "a@b.com"` // missing closing brace

	// Expect error logging for invalid JSON
	mockLogger.On(
		"LogHandlerError",
		mock.Anything, // ctx
		"update_user",
		"invalid_request",
		"Invalid update payload",
		mock.Anything, // ip
		mock.Anything, // userAgent
		mock.Anything, // error (should be an error like unexpected EOF)
	).Return()

	r := httptest.NewRequest("PUT", "/user", bytes.NewBufferString(invalidJSON))
	w := httptest.NewRecorder()

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	cfg.HandlerUpdateUser(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotEmpty(t, w.Body.String())
	mockService.AssertNotCalled(t, "UpdateUser")
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateUser_ServiceError tests that HandlerUpdateUser properly handles
// errors returned by the user service
func TestHandlerUpdateUser_ServiceError(t *testing.T) {
	mockService := new(mockUpdateUserService)
	mockLogger := new(mockUpdateHandlerLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	user := database.User{ID: "u1"}
	params := UpdateUserParams{Name: "A", Email: "a@b.com", Phone: "123", Address: "Addr"}
	mockService.On("UpdateUser", mock.Anything, user, params).Return(errors.New("fail"))
	mockLogger.On("LogHandlerError", mock.Anything, "update_user", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	body, _ := json.Marshal(params)
	r := httptest.NewRequest("PUT", "/user", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	cfg.HandlerUpdateUser(w, r)
	assert.NotEqual(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateUser_UserNotFoundInContext tests that HandlerUpdateUser returns
// an error when no user is found in the request context
func TestHandlerUpdateUser_UserNotFoundInContext(t *testing.T) {
	mockService := new(mockUpdateUserService)
	mockLogger := new(mockUpdateHandlerLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}

	// Mock error logging for user not found
	mockLogger.On("LogHandlerError", mock.Anything, "update_user", "user_not_found", "User not found in context", mock.Anything, mock.Anything, mock.Anything).Return()

	params := UpdateUserParams{Name: "A", Email: "a@b.com", Phone: "123", Address: "Addr"}
	body, _ := json.Marshal(params)
	r := httptest.NewRequest("PUT", "/user", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Don't set user in context - this should trigger the error path

	cfg.HandlerUpdateUser(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var errorResp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&errorResp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	assert.Contains(t, errorResp, "error")
	mockLogger.AssertExpectations(t)
}
