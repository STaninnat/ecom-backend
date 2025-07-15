package carthandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandlerUpdateItemQuantity tests the HandlerUpdateItemQuantity function for updating the quantity of an item in a user's cart.
// It covers scenarios such as successful update, invalid JSON, missing fields, and service errors.
// The test verifies the returned status and response body for each scenario.
func TestHandlerUpdateItemQuantity(t *testing.T) {
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
			body: CartUpdateRequest{ProductID: "prod1", Quantity: 2},
			setupMock: func(mockService *MockCartService) {
				mockService.On("UpdateItemQuantity", mock.Anything, "user1", "prod1", 2).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   handlers.HandlerResponse{Message: "Item quantity updated"},
		},
		{
			name:           "invalid json",
			user:           database.User{ID: "user1"},
			body:           "not json",
			setupMock:      func(mockService *MockCartService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Invalid request payload"},
		},
		{
			name:           "missing fields",
			user:           database.User{ID: "user1"},
			body:           CartUpdateRequest{ProductID: "", Quantity: 0},
			setupMock:      func(mockService *MockCartService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Product ID and quantity are required"},
		},
		{
			name: "service error",
			user: database.User{ID: "user1"},
			body: CartUpdateRequest{ProductID: "prod1", Quantity: 2},
			setupMock: func(mockService *MockCartService) {
				err := &handlers.AppError{Code: "product_not_found", Message: "Product not found"}
				mockService.On("UpdateItemQuantity", mock.Anything, "user1", "prod1", 2).Return(err)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]any{"error": "Product not found", "code": "product_not_found"},
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
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("PUT", "/cart/item", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			if tt.expectedStatus == http.StatusOK {
				mockLogger.On("LogHandlerSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
			} else if tt.name == "invalid json" {
				mockLogger.On("LogHandlerError", mock.Anything, "update_item_quantity", "invalid request body", "Failed to parse body", mock.Anything, mock.Anything, mock.Anything).Return()
			} else if tt.name == "missing fields" {
				mockLogger.On("LogHandlerError", mock.Anything, "update_item_quantity", "missing fields", "Required fields are missing", mock.Anything, mock.Anything, nil).Return()
			} else if tt.name == "service error" {
				mockLogger.On("LogHandlerError", mock.Anything, "update_item_quantity", "product_not_found", "Product not found", mock.Anything, mock.Anything, mock.Anything).Return()
			}

			config.HandlerUpdateItemQuantity(w, req, tt.user)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp handlers.HandlerResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(handlers.HandlerResponse).Message, resp.Message)
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

// TestHandlerUpdateGuestItemQuantity tests the HandlerUpdateGuestItemQuantity function for updating the quantity of an item in a guest cart (session-based).
// It covers scenarios such as successful update, missing session ID, invalid JSON, missing fields, and service errors.
// The test verifies the returned status and response body for each scenario.
func TestHandlerUpdateGuestItemQuantity(t *testing.T) {
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
			body:      CartUpdateRequest{ProductID: "prod1", Quantity: 2},
			setupMock: func(mockService *MockCartService) {
				mockService.On("UpdateGuestItemQuantity", mock.Anything, "sess1", "prod1", 2).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   handlers.HandlerResponse{Message: "Item quantity updated"},
		},
		{
			name:           "missing session id",
			sessionID:      "",
			body:           CartUpdateRequest{ProductID: "prod1", Quantity: 2},
			setupMock:      func(mockService *MockCartService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Missing session ID"},
		},
		{
			name:           "invalid json",
			sessionID:      "sess1",
			body:           "not json",
			setupMock:      func(mockService *MockCartService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Invalid request payload"},
		},
		{
			name:           "missing fields",
			sessionID:      "sess1",
			body:           CartUpdateRequest{ProductID: "", Quantity: 0},
			setupMock:      func(mockService *MockCartService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Product ID and quantity are required"},
		},
		{
			name:      "service error",
			sessionID: "sess1",
			body:      CartUpdateRequest{ProductID: "prod1", Quantity: 2},
			setupMock: func(mockService *MockCartService) {
				err := &handlers.AppError{Code: "cart_not_found", Message: "Cart not found"}
				mockService.On("UpdateGuestItemQuantity", mock.Anything, "sess1", "prod1", 2).Return(err)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]any{"error": "Cart not found", "code": "cart_not_found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Patch getSessionIDFromRequest to return the test's sessionID
			orig := getSessionIDFromRequest
			getSessionIDFromRequest = func(r *http.Request) string { return tt.sessionID }
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

			req := httptest.NewRequest("PUT", "/cart/guest/item", bytes.NewReader(bodyBytes))
			if tt.sessionID != "" {
				req = req.WithContext(context.WithValue(req.Context(), utils.ContextKey("session_id"), tt.sessionID))
			}
			w := httptest.NewRecorder()

			mockLogger.On("LogHandlerSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
			mockLogger.On("LogHandlerError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

			config.HandlerUpdateGuestItemQuantity(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp handlers.HandlerResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(handlers.HandlerResponse).Message, resp.Message)
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
