package orderhandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandlerUpdateOrderStatus_Success verifies successful order status update.
func TestHandlerUpdateOrderStatus_Success(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	orderID := "order123"
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
			Values: []string{"order123"},
		},
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateOrderStatus(w, req, user)

	assert.Equal(t, http.StatusOK, w.Code)
	var response handlers.HandlerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order status updated successfully", response.Message)

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateOrderStatus_InvalidRequest checks that an invalid JSON request returns a bad request error.
func TestHandlerUpdateOrderStatus_InvalidRequest(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
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
			Values: []string{"order123"},
		},
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateOrderStatus(w, req, user)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertNotCalled(t, "UpdateOrderStatus")
}

// TestHandlerUpdateOrderStatus_MissingOrderID checks that missing order ID returns an error.
func TestHandlerUpdateOrderStatus_MissingOrderID(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
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
	assert.NoError(t, err)
	assert.Equal(t, "Order ID is required", response["error"])

	mockLogger.AssertExpectations(t)
	mockOrderService.AssertNotCalled(t, "UpdateOrderStatus")
}

// TestHandlerUpdateOrderStatus_OrderNotFound checks that order not found returns the correct error.
func TestHandlerUpdateOrderStatus_OrderNotFound(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
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
	mockOrderService.On("UpdateOrderStatus", mock.Anything, "order123", "shipped").Return(appError)
	mockLogger.On("LogHandlerError", mock.Anything, "update_order_status", "order_not_found", "Order not found", mock.Anything, mock.Anything, nil).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/orders/"+orderID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"order_id"},
			Values: []string{"order123"},
		},
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateOrderStatus(w, req, user)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order not found", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateOrderStatus_InvalidStatus checks that invalid status returns the correct error.
func TestHandlerUpdateOrderStatus_InvalidStatus(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	orderID := "order123"
	requestBody := UpdateOrderStatusRequest{
		Status: "invalid_status",
	}

	appError := &handlers.AppError{Code: "invalid_status", Message: "Invalid order status"}
	mockOrderService.On("UpdateOrderStatus", mock.Anything, orderID, "invalid_status").Return(appError)
	mockLogger.On("LogHandlerError", mock.Anything, "update_order_status", "invalid_status", "Invalid order status", mock.Anything, mock.Anything, nil).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/orders/"+orderID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"order_id"},
			Values: []string{"order123"},
		},
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateOrderStatus(w, req, user)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid order status", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateOrderStatus_UpdateFailed checks that update failure returns the correct error.
func TestHandlerUpdateOrderStatus_UpdateFailed(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	orderID := "order123"
	requestBody := UpdateOrderStatusRequest{
		Status: "shipped",
	}

	appError := &handlers.AppError{Code: "update_failed", Message: "Failed to update order status"}
	mockOrderService.On("UpdateOrderStatus", mock.Anything, orderID, "shipped").Return(appError)
	mockLogger.On("LogHandlerError", mock.Anything, "update_order_status", "update_failed", "Failed to update order status", mock.Anything, mock.Anything, nil).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/orders/"+orderID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"order_id"},
			Values: []string{"order123"},
		},
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateOrderStatus(w, req, user)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Something went wrong, please try again later", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateOrderStatus_UnknownError checks that unknown errors are handled properly.
func TestHandlerUpdateOrderStatus_UnknownError(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockLogger := new(mockHandlerLogger)

	cfg := &HandlersOrderConfig{
		HandlersConfig: &handlers.HandlersConfig{
			Logger: logrus.New(),
		},
		Logger:       mockLogger,
		orderService: mockOrderService,
	}

	user := database.User{ID: "user123"}
	orderID := "order123"
	requestBody := UpdateOrderStatusRequest{
		Status: "shipped",
	}

	unknownError := errors.New("unknown database error")
	mockOrderService.On("UpdateOrderStatus", mock.Anything, orderID, "shipped").Return(unknownError)
	mockLogger.On("LogHandlerError", mock.Anything, "update_order_status", "unknown_error", "Unknown error occurred", mock.Anything, mock.Anything, unknownError).Return()

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/orders/"+orderID+"/status", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"order_id"},
			Values: []string{"order123"},
		},
	})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	cfg.HandlerUpdateOrderStatus(w, req, user)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", response["error"])

	mockOrderService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// TestHandlerUpdateOrderStatus_ValidStatuses tests all valid order statuses.
func TestHandlerUpdateOrderStatus_ValidStatuses(t *testing.T) {
	validStatuses := []string{"pending", "paid", "shipped", "delivered", "cancelled"}

	for _, status := range validStatuses {
		t.Run("Status_"+status, func(t *testing.T) {
			mockOrderService := new(MockOrderService)
			mockLogger := new(mockHandlerLogger)

			cfg := &HandlersOrderConfig{
				HandlersConfig: &handlers.HandlersConfig{
					Logger: logrus.New(),
				},
				Logger:       mockLogger,
				orderService: mockOrderService,
			}

			user := database.User{ID: "user123"}
			requestBody := UpdateOrderStatusRequest{
				Status: status,
			}

			mockOrderService.On("UpdateOrderStatus", mock.Anything, "order123", status).Return(nil)
			mockLogger.On("LogHandlerSuccess", mock.Anything, "update_order_status", "Order status updated successfully", mock.Anything, mock.Anything).Return()

			jsonBody, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("PUT", "/orders/order123/status", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
				URLParams: chi.RouteParams{
					Keys:   []string{"order_id"},
					Values: []string{"order123"},
				},
			})
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			cfg.HandlerUpdateOrderStatus(w, req, user)

			assert.Equal(t, http.StatusOK, w.Code)
			var response handlers.HandlerResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Order status updated successfully", response.Message)

			mockOrderService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
