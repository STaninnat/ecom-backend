// Package orderhandlers provides HTTP handlers and services for managing orders, including creation, retrieval, updating, deletion, with error handling and logging.
package orderhandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// handler_order_update_test.go: Tests for updating order status handler, covering success and failure cases with mocks.

// TestHandlerUpdateOrderStatus_Success verifies successful order status update.
func TestHandlerUpdateOrderStatus_Success(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	orderID := testOrderID
	requestBody := UpdateOrderStatusRequest{
		Status: "shipped",
	}

	mockOrderService.On("UpdateOrderStatus", mock.Anything, orderID, "shipped").Return(nil)
	mockLogger.On("LogHandlerSuccess", mock.Anything, "update_order_status", "Order status updated successfully", mock.Anything, mock.Anything).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/orders/"+orderID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"order_id"},
			Values: []string{testOrderID},
		},
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateOrderStatus(w, req, user)

	assert.Equal(t, http.StatusOK, w.Code)
	var response handlers.HandlerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Order status updated successfully", response.Message)

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateOrderStatus_InvalidRequest checks that an invalid JSON request returns a bad request error.
func TestHandlerUpdateOrderStatus_InvalidRequest(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	invalidJSON := `{"status": "shipped"`
	mockLogger.On("LogHandlerError", mock.Anything, "update_order_status", "invalid_request", "Failed to parse request body", mock.Anything, mock.Anything, mock.Anything).Return()

	req := httptest.NewRequest("PUT", "/orders/order123/status", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"order_id"},
			Values: []string{testOrderID},
		},
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateOrderStatus(w, req, user)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertNotCalled(t, "UpdateOrderStatus")
}

// TestHandlerUpdateOrderStatus_MissingOrderID checks that missing order ID returns an error.
func TestHandlerUpdateOrderStatus_MissingOrderID(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	requestBody := UpdateOrderStatusRequest{
		Status: "shipped",
	}

	mockLogger.On("LogHandlerError", mock.Anything, "update_order_status", "missing_order_id", "Order ID must be provided", mock.Anything, mock.Anything, nil).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/orders//status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"order_id"},
			Values: []string{""},
		},
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateOrderStatus(w, req, user)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Order ID is required", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertNotCalled(t, "UpdateOrderStatus")
}

// TestHandlerUpdateOrderStatus_OrderNotFound checks that order not found returns the correct error.
func TestHandlerUpdateOrderStatus_OrderNotFound(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		Config: &handlers.Config{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	orderID := "nonexistent"
	requestBody := UpdateOrderStatusRequest{
		Status: "shipped",
	}

	appError := &handlers.AppError{Code: "order_not_found", Message: "Order not found"}
	mockOrderService.On("UpdateOrderStatus", mock.Anything, testOrderID, "shipped").Return(appError)
	mockLogger.On("LogHandlerError", mock.Anything, "update_order_status", "order_not_found", "Order not found", mock.Anything, mock.Anything, nil).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/orders/"+orderID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"order_id"},
			Values: []string{testOrderID},
		},
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateOrderStatus(w, req, user)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Order not found", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateOrderStatus_ErrorScenarios tests all error scenarios for order status update.
func TestHandlerUpdateOrderStatus_ErrorScenarios(t *testing.T) {
	cases := []struct {
		name           string
		orderID        string
		status         string
		serviceErr     error
		loggerCall     func(*mockHandlerLogger)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:       "InvalidStatus",
			orderID:    testOrderID,
			status:     "invalid_status",
			serviceErr: &handlers.AppError{Code: "invalid_status", Message: "Invalid order status"},
			loggerCall: func(l *mockHandlerLogger) {
				l.On("LogHandlerError", mock.Anything, "update_order_status", "invalid_status", "Invalid order status", mock.Anything, mock.Anything, nil).Return()
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid order status",
		},
		{
			name:       "UpdateFailed",
			orderID:    testOrderID,
			status:     "shipped",
			serviceErr: &handlers.AppError{Code: "update_failed", Message: "Failed to update order status"},
			loggerCall: func(l *mockHandlerLogger) {
				l.On("LogHandlerError", mock.Anything, "update_order_status", "update_failed", "Failed to update order status", mock.Anything, mock.Anything, nil).Return()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Something went wrong, please try again later",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockOrderService := new(MockOrderService)
			mockLogger := new(mockHandlerLogger)
			cfg := &HandlersOrderConfig{
				Config:       &handlers.Config{Logger: logrus.New()},
				Logger:       mockLogger,
				orderService: mockOrderService,
			}
			user := database.User{ID: "user123"}
			mockOrderService.On("UpdateOrderStatus", mock.Anything, tc.orderID, tc.status).Return(tc.serviceErr)
			if tc.loggerCall != nil {
				tc.loggerCall(mockLogger)
			}
			requestBody := UpdateOrderStatusRequest{Status: tc.status}
			jsonBody, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("PUT", "/orders/"+tc.orderID+"/status", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
				URLParams: chi.RouteParams{
					Keys:   []string{"order_id"},
					Values: []string{tc.orderID},
				},
			})
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			cfg.HandlerUpdateOrderStatus(w, req, user)
			assert.Equal(t, tc.expectedStatus, w.Code)
			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedMsg, response["error"])
			mockOrderService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

// TestHandlerUpdateOrderStatus_ValidStatuses tests all valid order statuses.
func TestHandlerUpdateOrderStatus_ValidStatuses(t *testing.T) {
	validStatuses := []string{"pending", "paid", "shipped", "delivered", "cancelled"}

	for _, status := range validStatuses {
		t.Run("Status_"+status, func(t *testing.T) {
			mockOrderService := new(MockOrderService)
			mockLogger := new(mockHandlerLogger)

			cfg := &HandlersOrderConfig{
				Config: &handlers.Config{
					Logger: logrus.New(),
				},
				Logger:       mockLogger,
				orderService: mockOrderService,
			}

			user := database.User{ID: "user123"}
			requestBody := UpdateOrderStatusRequest{
				Status: status,
			}

			mockOrderService.On("UpdateOrderStatus", mock.Anything, testOrderID, status).Return(nil)
			mockLogger.On("LogHandlerSuccess", mock.Anything, "update_order_status", "Order status updated successfully", mock.Anything, mock.Anything).Return()

			jsonBody, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("PUT", "/orders/order123/status", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
				URLParams: chi.RouteParams{
					Keys:   []string{"order_id"},
					Values: []string{testOrderID},
				},
			})
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			cfg.HandlerUpdateOrderStatus(w, req, user)

			assert.Equal(t, http.StatusOK, w.Code)
			var response handlers.HandlerResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "Order status updated successfully", response.Message)

			mockOrderService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
