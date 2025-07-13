package paymenthandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandlerRefundPayment_Success tests successful refund
func TestHandlerRefundPayment_Success(t *testing.T) {
	mockService := new(MockPaymentServiceForRefund)
	mockLog := new(MockLoggerForRefund)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
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
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "Refund processed", resp["message"])
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerRefundPayment_MissingOrderID tests missing order_id param
func TestHandlerRefundPayment_MissingOrderID(t *testing.T) {
	mockService := new(MockPaymentServiceForRefund)
	mockLog := new(MockLoggerForRefund)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	mockLog.On("LogHandlerError", mock.Anything, "refund_payment", "missing_order_id", "Order ID not found in URL", mock.Anything, mock.Anything, nil).Return()

	r := httptest.NewRequest("POST", "/payments//refund", nil)
	rctx := chi.NewRouteContext() // no order_id param
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.HandlerRefundPayment(w, r, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLog.AssertExpectations(t)
}

// TestHandlerRefundPayment_ServiceError tests internal error from service
func TestHandlerRefundPayment_ServiceError(t *testing.T) {
	mockService := new(MockPaymentServiceForRefund)
	mockLog := new(MockLoggerForRefund)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	params := RefundPaymentParams{OrderID: "order1", UserID: "u1"}
	err := &handlers.AppError{Code: "refund_payment_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("RefundPayment", mock.Anything, params).Return(err)
	mockLog.On("LogHandlerError", mock.Anything, "refund_payment", "internal_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	r := httptest.NewRequest("POST", "/payments/order1/refund", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("order_id", "order1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.HandlerRefundPayment(w, r, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerRefundPayment_NotFound tests not found error from service
func TestHandlerRefundPayment_NotFound(t *testing.T) {
	mockService := new(MockPaymentServiceForRefund)
	mockLog := new(MockLoggerForRefund)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	params := RefundPaymentParams{OrderID: "order1", UserID: "u1"}
	err := &handlers.AppError{Code: "payment_not_found", Message: "not found", Err: errors.New("not found")}
	mockService.On("RefundPayment", mock.Anything, params).Return(err)
	mockLog.On("LogHandlerError", mock.Anything, "refund_payment", "payment_not_found", "not found", mock.Anything, mock.Anything, err.Err).Return()

	r := httptest.NewRequest("POST", "/payments/order1/refund", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("order_id", "order1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.HandlerRefundPayment(w, r, user)
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerRefundPayment_ValidationError tests validation error from service
func TestHandlerRefundPayment_ValidationError(t *testing.T) {
	mockService := new(MockPaymentServiceForRefund)
	mockLog := new(MockLoggerForRefund)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	params := RefundPaymentParams{OrderID: "order1", UserID: "u1"}
	err := &handlers.AppError{Code: "invalid_request", Message: "Invalid order ID", Err: errors.New("invalid order")}
	mockService.On("RefundPayment", mock.Anything, params).Return(err)
	mockLog.On("LogHandlerError", mock.Anything, "refund_payment", "invalid_request", "Invalid order ID", mock.Anything, mock.Anything, nil).Return()

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
