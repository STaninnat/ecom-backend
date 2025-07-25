// Package paymenthandlers provides HTTP handlers and configurations for processing payments, including Stripe integration, error handling, and payment-related request and response management.
package paymenthandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	testutil "github.com/STaninnat/ecom-backend/internal/testutil"
)

// handler_payment_refund_test.go: Tests for refund payment HTTP handler including success, missing params, and service errors.

// TestHandlerRefundPayment_Success tests successful refund
func TestHandlerRefundPayment_Success(t *testing.T) {
	mockService := new(MockPaymentServiceForRefund)
	mockLog := new(MockLoggerForRefund)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	params := RefundPaymentParams{OrderID: "order1", UserID: "u1"}
	mockService.On("RefundPayment", mock.Anything, params).Return(nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "refund_payment", "Refund successful", mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("POST", "/payments/order1/refund", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("order_id", "order1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.HandlerRefundPayment(w, r, user)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	assert.Equal(t, "Refund processed", resp["message"])
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// Remove the duplicated TestHandlerRefundPayment_MissingOrderID and use the shared helper from handler_payment_get_test.go

func TestHandlerRefundPayment_MissingOrderID(t *testing.T) {
	mockService := new(MockPaymentServiceForRefund)
	mockLog := new(MockLoggerForRefund)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	testutil.RunHandlerMissingOrderIDTest(
		t,
		func(cfg any, w http.ResponseWriter, r *http.Request, user any) {
			c := cfg.(*HandlersPaymentConfig)
			u := user.(database.User)
			c.HandlerRefundPayment(w, r, u)
		},
		cfg,
		user,
		"POST",
		"/payments//refund",
		"refund_payment",
		&mockLog.Mock,
	)
}

// TestHandlerRefundPayment_ErrorScenarios tests error scenarios for refund payment.
func TestHandlerRefundPayment_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name           string
		err            *handlers.AppError
		logCode        string
		logMsg         string
		expectedStatus int
	}{
		{
			name:           "ServiceError",
			err:            &handlers.AppError{Code: "refund_payment_error", Message: "fail", Err: errors.New("fail")},
			logCode:        "internal_error",
			logMsg:         "fail",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "NotFound",
			err:            &handlers.AppError{Code: "payment_not_found", Message: "not found", Err: errors.New("not found")},
			logCode:        "payment_not_found",
			logMsg:         "not found",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPaymentServiceForRefund)
			mockLog := new(MockLoggerForRefund)
			cfg := &HandlersPaymentConfig{
				Config:         &handlers.Config{},
				Logger:         mockLog,
				paymentService: mockService,
			}
			user := database.User{ID: "u1"}
			params := RefundPaymentParams{OrderID: "order1", UserID: "u1"}
			mockService.On("RefundPayment", mock.Anything, params).Return(tt.err)
			mockLog.On("LogHandlerError", mock.Anything, "refund_payment", tt.logCode, tt.logMsg, mock.Anything, mock.Anything, tt.err.Err).Return()

			r := httptest.NewRequest("POST", "/payments/order1/refund", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("order_id", "order1")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			cfg.HandlerRefundPayment(w, r, user)
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
			mockLog.AssertExpectations(t)
		})
	}
}

// TestHandlerRefundPayment_ValidationError tests validation error from service
func TestHandlerRefundPayment_ValidationError(t *testing.T) {
	mockService := new(MockPaymentServiceForRefund)
	mockLog := new(MockLoggerForRefund)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	params := RefundPaymentParams{OrderID: "order1", UserID: "u1"}
	err := &handlers.AppError{Code: "invalid_request", Message: "Invalid order ID", Err: errors.New("invalid order")}
	mockService.On("RefundPayment", mock.Anything, params).Return(err)
	mockLog.On("LogHandlerError", mock.Anything, "refund_payment", "invalid_request", "Invalid order ID", mock.Anything, mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("POST", "/payments/order1/refund", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("order_id", "order1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.HandlerRefundPayment(w, r, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}
