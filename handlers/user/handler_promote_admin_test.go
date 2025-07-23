// Package userhandlers provides HTTP handlers and services for user-related operations, including user retrieval, updates, and admin role management, with proper error handling and logging.
package userhandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// handler_promote_admin_test.go: Tests for HandlerPromoteUserToAdmin covering success, authorization, input validation, error handling, and context user presence.

// TestHandlerPromoteUserToAdmin_Success tests that HandlerPromoteUserToAdmin successfully
// promotes a user to admin role when called by an authorized admin user
func TestHandlerPromoteUserToAdmin_Success(t *testing.T) {
	mockService := new(mockPromoteUserService)
	mockLogger := new(mockPromoteLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	admin := database.User{ID: "admin1", Role: "admin"}
	params := map[string]string{"user_id": "user2"}
	body, _ := json.Marshal(params)
	mockService.On("PromoteUserToAdmin", mock.Anything, admin, "user2").Return(nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "promote_admin", "User promoted to admin success", mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("POST", "/admin/user/promote", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, admin))

	cfg.HandlerPromoteUserToAdmin(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp handlers.HandlerResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "User promoted to admin", resp.Message)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerPromoteUserToAdmin_NonAdminForbidden tests that HandlerPromoteUserToAdmin
// returns a forbidden error when a non-admin user attempts to promote another user
func TestHandlerPromoteUserToAdmin_NonAdminForbidden(t *testing.T) {
	mockService := new(mockPromoteUserService)
	mockLogger := new(mockPromoteLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	nonAdmin := database.User{ID: "u1", Role: "user"}
	params := map[string]string{"user_id": "user2"}
	body, _ := json.Marshal(params)
	mockService.On("PromoteUserToAdmin", mock.Anything, nonAdmin, "user2").Return(&handlers.AppError{Code: "unauthorized_user", Message: "Admin privileges required"})
	mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("POST", "/admin/user/promote", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, nonAdmin))

	cfg.HandlerPromoteUserToAdmin(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerPromoteUserToAdmin_AlreadyAdmin tests that HandlerPromoteUserToAdmin
// returns an error when attempting to promote a user who is already an admin
func TestHandlerPromoteUserToAdmin_AlreadyAdmin(t *testing.T) {
	mockService := new(mockPromoteUserService)
	mockLogger := new(mockPromoteLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	admin := database.User{ID: "admin1", Role: "admin"}
	params := map[string]string{"user_id": "user2"}
	body, _ := json.Marshal(params)
	mockService.On("PromoteUserToAdmin", mock.Anything, admin, "user2").Return(&handlers.AppError{Code: "already_admin", Message: "User is already admin"})
	mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("POST", "/admin/user/promote", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, admin))

	cfg.HandlerPromoteUserToAdmin(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerPromoteUserToAdmin_TargetUserNotFound tests that HandlerPromoteUserToAdmin
// returns a not found error when the target user doesn't exist
func TestHandlerPromoteUserToAdmin_TargetUserNotFound(t *testing.T) {
	mockService := new(mockPromoteUserService)
	mockLogger := new(mockPromoteLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	admin := database.User{ID: "admin1", Role: "admin"}
	params := map[string]string{"user_id": "user2"}
	body, _ := json.Marshal(params)
	mockService.On("PromoteUserToAdmin", mock.Anything, admin, "user2").Return(&handlers.AppError{Code: "user_not_found", Message: "Target user not found"})
	mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("POST", "/admin/user/promote", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, admin))

	cfg.HandlerPromoteUserToAdmin(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerPromoteUserToAdmin_InvalidPayload tests that HandlerPromoteUserToAdmin
// returns a bad request error when the request payload contains malformed JSON
func TestHandlerPromoteUserToAdmin_InvalidPayload(t *testing.T) {
	mockService := new(mockPromoteUserService)
	mockLogger := new(mockPromoteLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	admin := database.User{ID: "admin1", Role: "admin"}
	invalidJSON := `{"user_id":}` // malformed
	mockLogger.On("LogHandlerError", mock.Anything, "promote_admin", "invalid_request", "Invalid request payload", mock.Anything, mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("POST", "/admin/user/promote", bytes.NewBufferString(invalidJSON))
	w := httptest.NewRecorder()
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, admin))

	cfg.HandlerPromoteUserToAdmin(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "PromoteUserToAdmin")
	mockLogger.AssertExpectations(t)
}

// TestHandlerPromoteUserToAdmin_UserNotFoundInContext tests that HandlerPromoteUserToAdmin
// returns an unauthorized error when no user is found in the request context
func TestHandlerPromoteUserToAdmin_UserNotFoundInContext(t *testing.T) {
	mockService := new(mockPromoteUserService)
	mockLogger := new(mockPromoteLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	mockLogger.On("LogHandlerError", mock.Anything, "promote_admin", "user_not_found", "User not found in context", mock.Anything, mock.Anything, mock.Anything).Return()

	params := map[string]string{"user_id": "user2"}
	body, _ := json.Marshal(params)
	r := httptest.NewRequest("POST", "/admin/user/promote", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Don't set user in context
	cfg.HandlerPromoteUserToAdmin(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertNotCalled(t, "PromoteUserToAdmin")
	mockLogger.AssertExpectations(t)
}
