// Package internal_testutil provides shared test utilities and mock implementations to support unit testing of internal handlers and services.
package internal_testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// auth_helpers.go: Provides shared test cases and mocks for testing authentication token error scenarios.

// MockHandlersConfig is a minimal exported mock for handler logging.
type MockHandlersConfig struct {
	mock.Mock
}

// MockAuthService is a minimal exported mock for auth service.
type MockAuthService struct {
	mock.Mock
}

// RunAuthTokenErrorScenarios is a shared helper for sign-out and refresh token error scenario test cases.
// It runs a table-driven test for the given operation and handler.
func RunAuthTokenErrorScenarios(t *testing.T, operation string, handlerFunc func(w http.ResponseWriter, r *http.Request, logger *MockHandlersConfig, authService *MockAuthService)) {
	testCases := []struct {
		name           string
		setupLogger    func(*MockHandlersConfig, *MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success_LocalProvider",
			setupLogger: func(logger *MockHandlersConfig, _ *MockAuthService) {
				logger.On("LogHandlerError", mock.Anything, operation, "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything).Return()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   assert.AnError.Error(),
		},
		{
			name: "Success_GoogleProvider",
			setupLogger: func(logger *MockHandlersConfig, _ *MockAuthService) {
				logger.On("LogHandlerError", mock.Anything, operation, "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything).Return()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   assert.AnError.Error(),
		},
		{
			name: "ServiceError",
			setupLogger: func(logger *MockHandlersConfig, _ *MockAuthService) {
				logger.On("LogHandlerError", mock.Anything, operation, "invalid_token", "Error validating authentication token", mock.Anything, mock.Anything, mock.Anything).Return()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   assert.AnError.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := &MockHandlersConfig{}
			authService := &MockAuthService{}
			req := httptest.NewRequest("POST", "/", nil)
			w := httptest.NewRecorder()
			tc.setupLogger(logger, authService)
			handlerFunc(w, req, logger, authService)
			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedBody)
			logger.AssertExpectations(t)
		})
	}
}
