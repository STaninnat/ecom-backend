// Package carthandlers implements HTTP handlers for cart operations including user and guest carts.
package carthandlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// handler_cart_get_test.go: Tests for user and guest cart retrieval handlers with various scenarios.

// TestHandlerGetUserCart tests the HandlerGetUserCart function for retrieving a user's cart.
// It covers scenarios such as successful retrieval, cart not found, and database errors.
// The test verifies the returned status and response body for each scenario.
func TestHandlerGetUserCart(t *testing.T) {
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
				expectedCart := &models.Cart{
					ID: "cart1",
					Items: []models.CartItem{
						{ProductID: "prod1", Quantity: 2},
					},
				}
				mockService.On("GetUserCart", mock.Anything, "user1").Return(expectedCart, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &models.Cart{
				ID: "cart1",
				Items: []models.CartItem{
					{ProductID: "prod1", Quantity: 2},
				},
			},
		},
		{
			name: "cart not found",
			user: database.User{ID: "user1"},
			setupMock: func(mockService *MockCartService) {
				mockService.On("GetUserCart", mock.Anything, "user1").Return(nil, &handlers.AppError{Code: "cart_not_found", Message: "Cart not found"})
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]any{
				"error": "Cart not found",
				"code":  "cart_not_found",
			},
		},
		{
			name: "database error",
			user: database.User{ID: "user1"},
			setupMock: func(mockService *MockCartService) {
				mockService.On("GetUserCart", mock.Anything, "user1").Return(nil, &handlers.AppError{Code: "database_error", Message: "Database error"})
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"error": "Database error",
				"code":  "database_error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := &MockCartService{}
			mockLogger := &MockLogger{}
			tt.setupMock(mockService)

			config := &HandlersCartConfig{
				CartService: mockService,
				Logger:      mockLogger,
			}

			req := httptest.NewRequest("GET", "/cart", nil)
			w := httptest.NewRecorder()

			// Set up logger expectations
			if tt.expectedStatus == http.StatusOK {
				mockLogger.On("LogHandlerSuccess", mock.Anything, "get_cart", "Got user cart successfully", mock.Anything, mock.Anything).Return()
			} else {
				mockLogger.On("LogHandlerError", mock.Anything, "get_cart", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
			}

			// Execute
			config.HandlerGetUserCart(w, req, tt.user)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.Cart
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(*models.Cart).ID, response.ID)
			} else {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(map[string]any)["error"], response["error"])
				assert.Equal(t, tt.expectedBody.(map[string]any)["code"], response["code"])
			}

			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

// TestHandlerGetGuestCart tests the HandlerGetGuestCart function for retrieving a guest cart (session-based).
// It covers scenarios such as successful retrieval, missing session ID, cart not found, and database errors.
// The test verifies the returned status and response body for each scenario.
func TestHandlerGetGuestCart(t *testing.T) {
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
				expectedCart := &models.Cart{
					ID: "cart1",
					Items: []models.CartItem{
						{ProductID: "prod1", Quantity: 1},
					},
				}
				mockService.On("GetGuestCart", mock.Anything, "sess1").Return(expectedCart, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &models.Cart{
				ID: "cart1",
				Items: []models.CartItem{
					{ProductID: "prod1", Quantity: 1},
				},
			},
		},
		{
			name:      "missing session ID",
			sessionID: "",
			setupMock: func(_ *MockCartService) {
				// No mock setup needed for this case
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]any{
				"error": "Missing session ID",
			},
		},
		{
			name:      "cart not found",
			sessionID: "sess1",
			setupMock: func(mockService *MockCartService) {
				mockService.On("GetGuestCart", mock.Anything, "sess1").Return(nil, &handlers.AppError{Code: "cart_not_found", Message: "Cart not found"})
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]any{
				"error": "Cart not found",
				"code":  "cart_not_found",
			},
		},
		{
			name:      "database error",
			sessionID: "sess1",
			setupMock: func(mockService *MockCartService) {
				mockService.On("GetGuestCart", mock.Anything, "sess1").Return(nil, &handlers.AppError{Code: "database_error", Message: "Database error"})
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"error": "Database error",
				"code":  "database_error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Patch getSessionIDFromRequest to return the test's sessionID
			orig := getSessionIDFromRequest
			getSessionIDFromRequest = func(_ *http.Request) string { return tt.sessionID }
			defer func() { getSessionIDFromRequest = orig }()

			// Setup
			mockService := &MockCartService{}
			mockLogger := &MockLogger{}
			tt.setupMock(mockService)

			config := &HandlersCartConfig{
				CartService: mockService,
				Logger:      mockLogger,
			}

			req := httptest.NewRequest("GET", "/cart", nil)
			w := httptest.NewRecorder()

			// Set up logger expectations
			switch {
			case tt.expectedStatus == http.StatusOK:
				mockLogger.On("LogHandlerSuccess", mock.Anything, "get_guest_cart", "Got guest cart successfully", mock.Anything, mock.Anything).Return()
			case tt.sessionID == "":
				mockLogger.On("LogHandlerError", mock.Anything, "get_guest_cart", "missing session ID", "Session ID not found in request", mock.Anything, mock.Anything, nil).Return()
			default:
				mockLogger.On("LogHandlerError", mock.Anything, "get_guest_cart", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
			}

			// Execute
			config.HandlerGetGuestCart(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.Cart
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(*models.Cart).ID, response.ID)
			} else {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(map[string]any)["error"], response["error"])
				if tt.expectedBody.(map[string]any)["code"] != nil {
					assert.Equal(t, tt.expectedBody.(map[string]any)["code"], response["code"])
				}
			}

			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
