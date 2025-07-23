// Package carthandlers implements HTTP handlers for cart operations including user and guest carts.
package carthandlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// handler_cart_checkout_test.go: Tests for user and guest cart checkout handlers.

// TestHandlerCheckoutUserCart tests the HandlerCheckoutUserCart function for user cart checkout scenarios.
// It verifies both successful checkout and service error handling, checking the returned status and response body.
func TestHandlerCheckoutUserCart(t *testing.T) {
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
				result := &CartCheckoutResult{
					OrderID: "order123",
					Message: "Order created successfully",
				}
				mockService.On("CheckoutUserCart", mock.Anything, "user1").Return(result, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: CartResponse{
				Message: "Order created successfully",
				OrderID: "order123",
			},
		},
		{
			name: "service error",
			user: database.User{ID: "user1"},
			setupMock: func(mockService *MockCartService) {
				err := &handlers.AppError{Code: "cart_empty", Message: "Cart is empty"}
				mockService.On("CheckoutUserCart", mock.Anything, "user1").Return(nil, err)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Cart is empty", "code": "cart_empty"},
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

			req := httptest.NewRequest("POST", "/cart/checkout", nil)
			w := httptest.NewRecorder()

			mockLogger.On("LogHandlerSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
			mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

			config.HandlerCheckoutUserCart(w, req, tt.user)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp CartResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(CartResponse).Message, resp.Message)
				assert.Equal(t, tt.expectedBody.(CartResponse).OrderID, resp.OrderID)
			} else {
				var resp map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				for k, v := range tt.expectedBody.(map[string]any) {
					assert.Equal(t, v, resp[k])
				}
			}
			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

// TestHandlerCheckoutGuestCart tests the HandlerCheckoutGuestCart function for guest cart checkout scenarios.
// It covers cases such as successful checkout, missing session ID, invalid JSON, missing user ID, and service errors.
// The test verifies the returned status and response body for each scenario.
func TestHandlerCheckoutGuestCart(t *testing.T) {
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
			body:      map[string]string{"user_id": "user1"},
			setupMock: func(mockService *MockCartService) {
				result := &CartCheckoutResult{
					OrderID: "order123",
					Message: "Guest order created successfully",
				}
				mockService.On("CheckoutGuestCart", mock.Anything, "sess1", "user1").Return(result, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: CartResponse{
				Message: "Guest order created successfully",
				OrderID: "order123",
			},
		},
		{
			name:           "missing session ID",
			sessionID:      "",
			body:           map[string]string{"user_id": "user1"},
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
			name:           "missing user ID",
			sessionID:      "sess1",
			body:           map[string]string{"user_id": ""},
			setupMock:      func(_ *MockCartService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "User ID is required for guest checkout"},
		},
		{
			name:      "service error",
			sessionID: "sess1",
			body:      map[string]string{"user_id": "user1"},
			setupMock: func(mockService *MockCartService) {
				err := &handlers.AppError{Code: "cart_empty", Message: "Cart is empty"}
				mockService.On("CheckoutGuestCart", mock.Anything, "sess1", "user1").Return(nil, err)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Cart is empty", "code": "cart_empty"},
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
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/cart/guest/checkout", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			mockLogger.On("LogHandlerSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
			mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

			config.HandlerCheckoutGuestCart(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp CartResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(CartResponse).Message, resp.Message)
				assert.Equal(t, tt.expectedBody.(CartResponse).OrderID, resp.OrderID)
			} else {
				var resp map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				for k, v := range tt.expectedBody.(map[string]any) {
					assert.Equal(t, v, resp[k])
				}
			}
			mockService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
