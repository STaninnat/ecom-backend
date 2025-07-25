// Package carthandlers implements HTTP handlers for cart operations including user and guest carts.
package carthandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_cart_delete_test.go: Tests for cart handlers covering item removal and cart clearing for users and guests.

// TestHandlerRemoveItemFromUserCart tests the HandlerRemoveItemFromUserCart function for removing an item from a user's cart.
// It covers scenarios such as successful removal, invalid JSON, missing product ID, and service errors.
// The test verifies the returned status and response body for each scenario.
func TestHandlerRemoveItemFromUserCart(t *testing.T) {
	tests := []struct {
		name           string
		user           database.User
		body           any
		setupMock      func(*MockCartService)
		expectedStatus int
		expectedBody   any
	}{
		{
			name: "success",
			user: database.User{ID: "user1"},
			body: DeleteItemRequest{ProductID: "prod1"},
			setupMock: func(mockService *MockCartService) {
				mockService.On("RemoveItem", mock.Anything, "user1", "prod1").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   handlers.HandlerResponse{Message: "Item removed from cart"},
		},
		{
			name:           "invalid json",
			user:           database.User{ID: "user1"},
			body:           "not json",
			setupMock:      func(_ *MockCartService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Invalid request payload"},
		},
		{
			name:           "missing product ID",
			user:           database.User{ID: "user1"},
			body:           DeleteItemRequest{ProductID: ""},
			setupMock:      func(_ *MockCartService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Product ID is required"},
		},
		{
			name: "service error",
			user: database.User{ID: "user1"},
			body: DeleteItemRequest{ProductID: "prod1"},
			setupMock: func(mockService *MockCartService) {
				err := &handlers.AppError{Code: "item_not_found", Message: "Item not found"}
				mockService.On("RemoveItem", mock.Anything, "user1", "prod1").Return(err)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]any{"error": "Item not found", "code": "item_not_found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockCartService{}
			mockLogger := &MockLogger{}
			tt.setupMock(mockService)

			config := &HandlersCartConfig{
				CartService: mockService,
				Logger:      mockLogger,
			}

			var bodyBytes []byte
			var err error
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, err = json.Marshal(v)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("DELETE", "/cart/item", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			mockLogger.On("LogHandlerSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
			mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

			config.HandlerRemoveItemFromUserCart(w, req, tt.user)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assertHTTPResponse(t, w, tt.expectedBody, tt.expectedStatus)
			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

// TestHandlerClearUserCart tests the HandlerClearUserCart function for clearing a user's cart.
// It covers scenarios such as successful cart clearing and service errors.
// The test verifies the returned status and response body for each scenario.
func TestHandlerClearUserCart(t *testing.T) {
	tests := []struct {
		name           string
		user           database.User
		setupMock      func(*MockCartService)
		expectedStatus int
		expectedBody   any
	}{
		{
			name: "success",
			user: database.User{ID: "user1"},
			setupMock: func(mockService *MockCartService) {
				mockService.On("DeleteUserCart", mock.Anything, "user1").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   handlers.HandlerResponse{Message: "Cart cleared"},
		},
		{
			name: "service error",
			user: database.User{ID: "user1"},
			setupMock: func(mockService *MockCartService) {
				err := &handlers.AppError{Code: "cart_not_found", Message: "Cart not found"}
				mockService.On("DeleteUserCart", mock.Anything, "user1").Return(err)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]any{"error": "Cart not found", "code": "cart_not_found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockCartService{}
			mockLogger := &MockLogger{}
			tt.setupMock(mockService)

			config := &HandlersCartConfig{
				CartService: mockService,
				Logger:      mockLogger,
			}

			req := httptest.NewRequest("DELETE", "/cart", nil)
			w := httptest.NewRecorder()

			mockLogger.On("LogHandlerSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
			mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

			config.HandlerClearUserCart(w, req, tt.user)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assertHTTPResponse(t, w, tt.expectedBody, tt.expectedStatus)
			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

// TestHandlerRemoveItemFromGuestCart tests the HandlerRemoveItemFromGuestCart function for removing an item from a guest cart.
// It covers scenarios such as successful removal, missing session ID, invalid JSON, missing product ID, and service errors.
// The test verifies the returned status and response body for each scenario.
func TestHandlerRemoveItemFromGuestCart(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		body           any
		setupMock      func(*MockCartService)
		expectedStatus int
		expectedBody   any
	}{
		{
			name:      "success",
			sessionID: "sess1",
			body:      DeleteItemRequest{ProductID: "prod1"},
			setupMock: func(mockService *MockCartService) {
				mockService.On("RemoveGuestItem", mock.Anything, "sess1", "prod1").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   handlers.HandlerResponse{Message: "Item removed from cart"},
		},
		{
			name:           "missing session ID",
			sessionID:      "",
			body:           DeleteItemRequest{ProductID: "prod1"},
			setupMock:      func(_ *MockCartService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Missing session ID"},
		},
		{
			name:           "invalid json",
			sessionID:      "sess1",
			body:           "not json",
			setupMock:      func(_ *MockCartService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Invalid request payload"},
		},
		{
			name:           "missing product ID",
			sessionID:      "sess1",
			body:           DeleteItemRequest{ProductID: ""},
			setupMock:      func(_ *MockCartService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Product ID is required"},
		},
		{
			name:      "service error",
			sessionID: "sess1",
			body:      DeleteItemRequest{ProductID: "prod1"},
			setupMock: func(mockService *MockCartService) {
				err := &handlers.AppError{Code: "item_not_found", Message: "Item not found"}
				mockService.On("RemoveGuestItem", mock.Anything, "sess1", "prod1").Return(err)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]any{"error": "Item not found", "code": "item_not_found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Patch getSessionIDFromRequest to return the test's sessionID
			orig := getSessionIDFromRequest
			getSessionIDFromRequest = func(_ *http.Request) string { return tt.sessionID }
			defer func() { getSessionIDFromRequest = orig }()

			mockService := &MockCartService{}
			mockLogger := &MockLogger{}
			tt.setupMock(mockService)

			config := &HandlersCartConfig{
				CartService: mockService,
				Logger:      mockLogger,
			}

			var bodyBytes []byte
			var err error
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, err = json.Marshal(v)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("DELETE", "/cart/guest/item", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			mockLogger.On("LogHandlerSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
			mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

			config.HandlerRemoveItemFromGuestCart(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assertHTTPResponse(t, w, tt.expectedBody, tt.expectedStatus)
			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

// TestHandlerClearGuestCart tests the HandlerClearGuestCart function for clearing a guest cart.
// It covers scenarios such as successful cart clearing, missing session ID, and service errors.
// The test verifies the returned status and response body for each scenario.
func TestHandlerClearGuestCart(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		setupMock      func(*MockCartService)
		expectedStatus int
		expectedBody   any
	}{
		{
			name:      "success",
			sessionID: "sess1",
			setupMock: func(mockService *MockCartService) {
				mockService.On("DeleteGuestCart", mock.Anything, "sess1").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   handlers.HandlerResponse{Message: "Guest cart cleared"},
		},
		{
			name:           "missing session ID",
			sessionID:      "",
			setupMock:      func(_ *MockCartService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Missing session ID"},
		},
		{
			name:      "service error",
			sessionID: "sess1",
			setupMock: func(mockService *MockCartService) {
				err := &handlers.AppError{Code: "cart_not_found", Message: "Cart not found"}
				mockService.On("DeleteGuestCart", mock.Anything, "sess1").Return(err)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]any{"error": "Cart not found", "code": "cart_not_found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Patch getSessionIDFromRequest to return the test's sessionID
			orig := getSessionIDFromRequest
			getSessionIDFromRequest = func(_ *http.Request) string { return tt.sessionID }
			defer func() { getSessionIDFromRequest = orig }()

			mockService := &MockCartService{}
			mockLogger := &MockLogger{}
			tt.setupMock(mockService)

			config := &HandlersCartConfig{
				CartService: mockService,
				Logger:      mockLogger,
			}

			req := httptest.NewRequest("DELETE", "/cart/guest", nil)
			w := httptest.NewRecorder()

			mockLogger.On("LogHandlerSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
			mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

			config.HandlerClearGuestCart(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assertHTTPResponse(t, w, tt.expectedBody, tt.expectedStatus)
			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
