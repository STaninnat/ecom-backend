// Package userhandlers provides HTTP handlers and services for user-related operations, including user retrieval, updates, and admin role management, with proper error handling and logging.
package userhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_get_user_test.go: Tests HandlerGetUser for success, service errors, missing user context, and empty user fields.

// TestHandlerGetUser_Success tests that HandlerGetUser successfully retrieves
// and returns user information when a valid user is in the context
func TestHandlerGetUser_Success(t *testing.T) {
	mockService := new(mockGetUserService)
	mockLogger := new(mockGetHandlerLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	user := database.User{ID: "u1"}
	resp := &UserResponse{ID: "u1", Name: "Test"}
	mockService.On("GetUser", mock.Anything, user).Return(resp, nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "get_user", "Get user info success", mock.Anything, mock.Anything).Return()
	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	cfg.HandlerGetUser(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	var got UserResponse
	err := json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Equal(t, *resp, got)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetUser_ServiceError tests that HandlerGetUser properly handles
// errors returned by the user service
func TestHandlerGetUser_ServiceError(t *testing.T) {
	mockService := new(mockGetUserService)
	mockLogger := new(mockGetHandlerLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	user := database.User{ID: "u1"}
	mockService.On("GetUser", mock.Anything, user).Return(nil, errors.New("fail"))
	mockLogger.On("LogHandlerError", mock.Anything, "get_user", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	cfg.HandlerGetUser(w, r)
	assert.NotEqual(t, http.StatusOK, w.Code)
	// Check that we get an error response
	var errorResp map[string]any
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp, "error")
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetUser_UserNotFoundInContext tests that HandlerGetUser returns
// an unauthorized error when no user is found in the request context
func TestHandlerGetUser_UserNotFoundInContext(t *testing.T) {
	mockService := new(mockGetUserService)
	mockLogger := new(mockGetHandlerLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}

	// Mock error logging for user not found
	mockLogger.On("LogHandlerError", mock.Anything, "get_user", "user_not_found", "User not found in context", mock.Anything, mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()

	// Don't set user in context - this should trigger the error path

	cfg.HandlerGetUser(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var errorResp map[string]any
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp, "error")
	mockLogger.AssertExpectations(t)
}

// TestHandlerGetUser_EmptyUserData tests that HandlerGetUser correctly handles
// users with empty/null fields and returns them properly
func TestHandlerGetUser_EmptyUserData(t *testing.T) {
	mockService := new(mockGetUserService)
	mockLogger := new(mockGetHandlerLogger)
	cfg := &HandlersUserConfig{
		Config:      &handlers.Config{Logger: logrus.New()},
		Logger:      mockLogger,
		userService: mockService,
	}
	user := database.User{ID: "u1"}
	// Return user with empty/null fields
	resp := &UserResponse{ID: "u1", Name: "", Email: "", Phone: "", Address: ""}
	mockService.On("GetUser", mock.Anything, user).Return(resp, nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "get_user", "Get user info success", mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("GET", "/user", nil)
	w := httptest.NewRecorder()

	// Set user in context
	r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))

	cfg.HandlerGetUser(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	var response UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "u1", response.ID)
	assert.Empty(t, response.Name)
	assert.Empty(t, response.Email)
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
