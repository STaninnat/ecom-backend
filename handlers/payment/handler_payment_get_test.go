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

// TestHandlerGetPayment_Success tests successful payment retrieval
func TestHandlerGetPayment_Success(t *testing.T) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	expectedResult := &GetPaymentResult{
		ID:       "p1",
		OrderID:  "order1",
		UserID:   "u1",
		Amount:   100.0,
		Currency: "USD",
		Status:   "succeeded",
	}
	mockService.On("GetPayment", mock.Anything, "order1", "u1").Return(expectedResult, nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "get_payment", "Get Payment success", mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("GET", "/payments/order1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("order_id", "order1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.HandlerGetPayment(w, r, user)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp GetPaymentResult
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "p1", resp.ID)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetPayment_MissingOrderID tests missing order_id param
func TestHandlerGetPayment_MissingOrderID(t *testing.T) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	mockLog.On("LogHandlerError", mock.Anything, "get_payment", "missing_order_id", "Order ID not found in URL", mock.Anything, mock.Anything, nil).Return()

	r := httptest.NewRequest("GET", "/payments/", nil)
	rctx := chi.NewRouteContext() // no order_id param
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.HandlerGetPayment(w, r, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetPayment_ServiceError tests internal error from service
func TestHandlerGetPayment_ServiceError(t *testing.T) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	err := &handlers.AppError{Code: "get_payment_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("GetPayment", mock.Anything, "order1", "u1").Return(nil, err)
	mockLog.On("LogHandlerError", mock.Anything, "get_payment", "internal_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	r := httptest.NewRequest("GET", "/payments/order1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("order_id", "order1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.HandlerGetPayment(w, r, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetPayment_NotFound tests not found error from service
func TestHandlerGetPayment_NotFound(t *testing.T) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	err := &handlers.AppError{Code: "payment_not_found", Message: "not found", Err: errors.New("not found")}
	mockService.On("GetPayment", mock.Anything, "order1", "u1").Return(nil, err)
	mockLog.On("LogHandlerError", mock.Anything, "get_payment", "payment_not_found", "not found", mock.Anything, mock.Anything, err.Err).Return()

	r := httptest.NewRequest("GET", "/payments/order1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("order_id", "order1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.HandlerGetPayment(w, r, user)
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetPayment_ValidationError tests validation error from service
func TestHandlerGetPayment_ValidationError(t *testing.T) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	err := &handlers.AppError{Code: "invalid_request", Message: "Invalid order ID", Err: errors.New("invalid order")}
	mockService.On("GetPayment", mock.Anything, "order1", "u1").Return(nil, err)
	mockLog.On("LogHandlerError", mock.Anything, "get_payment", "invalid_request", "Invalid order ID", mock.Anything, mock.Anything, nil).Return()

	r := httptest.NewRequest("GET", "/payments/order1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("order_id", "order1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.HandlerGetPayment(w, r, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetPaymentHistory_Success tests successful payment history retrieval
func TestHandlerGetPaymentHistory_Success(t *testing.T) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	expectedHistory := []PaymentHistoryItem{
		{
			ID:                "p1",
			OrderID:           "order1",
			Amount:            "100.00",
			Currency:          "USD",
			Status:            "succeeded",
			Provider:          "stripe",
			ProviderPaymentID: "pi_123",
		},
		{
			ID:                "p2",
			OrderID:           "order2",
			Amount:            "200.00",
			Currency:          "USD",
			Status:            "pending",
			Provider:          "stripe",
			ProviderPaymentID: "pi_456",
		},
	}
	mockService.On("GetPaymentHistory", mock.Anything, "u1").Return(expectedHistory, nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "get_history_payment", "Get Payment history success", mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("GET", "/payments/history", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetPaymentHistory(w, r, user)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp []PaymentHistoryItem
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 2)
	assert.Equal(t, "p1", resp[0].ID)
	assert.Equal(t, "p2", resp[1].ID)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetPaymentHistory_ServiceError tests internal error from service
func TestHandlerGetPaymentHistory_ServiceError(t *testing.T) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	err := &handlers.AppError{Code: "get_history_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("GetPaymentHistory", mock.Anything, "u1").Return(nil, err)
	mockLog.On("LogHandlerError", mock.Anything, "get_history_payment", "internal_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	r := httptest.NewRequest("GET", "/payments/history", nil)
	w := httptest.NewRecorder()

	cfg.HandlerGetPaymentHistory(w, r, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerAdminGetPayments_Success tests successful admin payments retrieval
func TestHandlerAdminGetPayments_Success(t *testing.T) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "admin1"}
	expectedPayments := []PaymentHistoryItem{
		{
			ID:                "p1",
			OrderID:           "order1",
			Amount:            "100.00",
			Currency:          "USD",
			Status:            "succeeded",
			Provider:          "stripe",
			ProviderPaymentID: "pi_123",
		},
		{
			ID:                "p2",
			OrderID:           "order2",
			Amount:            "200.00",
			Currency:          "USD",
			Status:            "pending",
			Provider:          "stripe",
			ProviderPaymentID: "pi_456",
		},
	}
	mockService.On("GetAllPayments", mock.Anything, "").Return(expectedPayments, nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "admin_get_payments", "Get all payments success", mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("GET", "/admin/payments", nil)
	w := httptest.NewRecorder()

	cfg.HandlerAdminGetPayments(w, r, user)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp []PaymentHistoryItem
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 2)
	assert.Equal(t, "p1", resp[0].ID)
	assert.Equal(t, "p2", resp[1].ID)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerAdminGetPayments_WithStatusFilter tests admin payments retrieval with status filter
func TestHandlerAdminGetPayments_WithStatusFilter(t *testing.T) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "admin1"}
	expectedPayments := []PaymentHistoryItem{
		{
			ID:                "p1",
			OrderID:           "order1",
			Amount:            "100.00",
			Currency:          "USD",
			Status:            "succeeded",
			Provider:          "stripe",
			ProviderPaymentID: "pi_123",
		},
	}
	mockService.On("GetAllPayments", mock.Anything, "succeeded").Return(expectedPayments, nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "admin_get_payments", "Get all payments success", mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("GET", "/admin/payments/succeeded", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("status", "succeeded")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.HandlerAdminGetPayments(w, r, user)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp []PaymentHistoryItem
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, "p1", resp[0].ID)
	assert.Equal(t, "succeeded", resp[0].Status)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerAdminGetPayments_ServiceError tests internal error from service
func TestHandlerAdminGetPayments_ServiceError(t *testing.T) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "admin1"}
	err := &handlers.AppError{Code: "admin_get_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("GetAllPayments", mock.Anything, "").Return(nil, err)
	mockLog.On("LogHandlerError", mock.Anything, "admin_get_payments", "internal_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	r := httptest.NewRequest("GET", "/admin/payments", nil)
	w := httptest.NewRecorder()

	cfg.HandlerAdminGetPayments(w, r, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}
