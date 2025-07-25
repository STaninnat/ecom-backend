// Package orderhandlers provides HTTP handlers and services for managing orders, including creation, retrieval, updating, deletion, with error handling and logging.
package orderhandlers

import (
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

// handler_order_delete_test.go: Tests for HandlerDeleteOrder covering all typical and edge cases.

const (
	testOrderID     = "order123"
	testNonexistent = "nonexistent"
)

// Helper to set chi URL param in request context
func setChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	return r.WithContext(ctx)
}

func TestHandlerDeleteOrder_Scenarios(t *testing.T) {
	cases := []struct {
		name           string
		orderID        string
		setOrderID     bool
		serviceErr     error
		loggerCall     func(*mockHandlerLogger)
		expectedStatus int
		expectedMsg    string
		expectedField  string // "Message" or "error"
	}{
		{
			name:       "Success",
			orderID:    testOrderID,
			setOrderID: true,
			serviceErr: nil,
			loggerCall: func(l *mockHandlerLogger) {
				l.On("LogHandlerSuccess", mock.Anything, "delete_order", "Deleted order successful", mock.Anything, mock.Anything).Return()
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "Order deleted successfully",
			expectedField:  "Message",
		},
		{
			name:       "MissingOrderID",
			orderID:    "",
			setOrderID: false,
			serviceErr: nil,
			loggerCall: func(l *mockHandlerLogger) {
				l.On("LogHandlerError", mock.Anything, "delete_order", "missing_order_id", "Order ID not found in URL", mock.Anything, mock.Anything, nil).Return()
			},
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Missing order_id",
			expectedField:  "error",
		},
		{
			name:       "OrderNotFound",
			orderID:    testNonexistent,
			setOrderID: true,
			serviceErr: &handlers.AppError{Code: "order_not_found", Message: "Order not found"},
			loggerCall: func(l *mockHandlerLogger) {
				l.On("LogHandlerError", mock.Anything, "delete_order", "order_not_found", "Order not found", mock.Anything, mock.Anything, nil).Return()
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "Order not found",
			expectedField:  "error",
		},
		{
			name:       "DeleteFailed",
			orderID:    testOrderID,
			setOrderID: true,
			serviceErr: &handlers.AppError{Code: "delete_order_error", Message: "Failed to delete order"},
			loggerCall: func(l *mockHandlerLogger) {
				l.On("LogHandlerError", mock.Anything, "delete_order", "delete_order_error", "Failed to delete order", mock.Anything, mock.Anything, nil).Return()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Something went wrong, please try again later",
			expectedField:  "error",
		},
		{
			name:       "Unauthorized",
			orderID:    testOrderID,
			setOrderID: true,
			serviceErr: &handlers.AppError{Code: "unauthorized", Message: "User is not authorized to delete this order"},
			loggerCall: func(l *mockHandlerLogger) {
				l.On("LogHandlerError", mock.Anything, "delete_order", "unauthorized", "User is not authorized to delete this order", mock.Anything, mock.Anything, nil).Return()
			},
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "User is not authorized to delete this order",
			expectedField:  "error",
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
			if tc.setOrderID {
				mockOrderService.On("DeleteOrder", mock.Anything, tc.orderID).Return(tc.serviceErr)
			}
			if tc.loggerCall != nil {
				tc.loggerCall(mockLogger)
			}
			var req *http.Request
			if tc.setOrderID {
				req = httptest.NewRequest("DELETE", "/orders/"+tc.orderID, nil)
				req = setChiURLParam(req, "order_id", tc.orderID)
			} else {
				req = httptest.NewRequest("DELETE", "/orders/", nil)
			}
			w := httptest.NewRecorder()
			cfg.HandlerDeleteOrder(w, req, user)
			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedField == "Message" {
				var response handlers.HandlerResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedMsg, response.Message)
			} else {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedMsg, response[tc.expectedField])
			}
			mockOrderService.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
