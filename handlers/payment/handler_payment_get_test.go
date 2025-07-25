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

// handler_payment_get_test.go: Tests for payment-related HTTP handlers: get payment, payment history, and admin payments, covering success, validation errors, and service error handling.

// TestHandlerGetPayment_Success tests successful payment retrieval
func TestHandlerGetPayment_Success(t *testing.T) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
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
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	assert.Equal(t, "p1", resp.ID)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// runHandlerGetPaymentErrorTest is a shared helper for HandlerGetPayment error scenario tests.
func runHandlerGetPaymentErrorTest(
	t *testing.T,
	appErr *handlers.AppError,
	expectedLogCode string,
	expectedStatus int,
) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	mockService.On("GetPayment", mock.Anything, "order1", "u1").Return(nil, appErr)
	mockLog.On("LogHandlerError", mock.Anything, "get_payment", expectedLogCode, appErr.Message, mock.Anything, mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("GET", "/payments/order1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("order_id", "order1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	cfg.HandlerGetPayment(w, r, user)
	assert.Equal(t, expectedStatus, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetPayment_ServiceError tests internal error from service
func TestHandlerGetPayment_ServiceError(t *testing.T) {
	err := &handlers.AppError{Code: "get_payment_error", Message: "fail", Err: errors.New("fail")}
	runHandlerGetPaymentErrorTest(t, err, "internal_error", http.StatusInternalServerError)
}

// TestHandlerGetPayment_NotFound tests not found error from service
func TestHandlerGetPayment_NotFound(t *testing.T) {
	err := &handlers.AppError{Code: "payment_not_found", Message: "not found", Err: errors.New("not found")}
	runHandlerGetPaymentErrorTest(t, err, "payment_not_found", http.StatusNotFound)
}

// TestHandlerGetPayment_ValidationError tests validation error from service
func TestHandlerGetPayment_ValidationError(t *testing.T) {
	err := &handlers.AppError{Code: "invalid_request", Message: "Invalid order ID", Err: errors.New("invalid order")}
	runHandlerGetPaymentErrorTest(t, err, "invalid_request", http.StatusBadRequest)
}

// TestHandlerGetPayment_MissingOrderID tests missing order_id parameter
func TestHandlerGetPayment_MissingOrderID(t *testing.T) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
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
			c.HandlerGetPayment(w, r, u)
		},
		cfg,
		user,
		"GET",
		"/payments/",
		"get_payment",
		&mockLog.Mock,
	)
}

// runHandlerGetPaymentsSuccessTest is a shared helper for HandlerGetPaymentHistory/AdminGetPayments success scenario tests.
func runHandlerGetPaymentsSuccessTest(
	t *testing.T,
	handler func(cfg *HandlersPaymentConfig, w http.ResponseWriter, r *http.Request, user database.User),
	serviceMethod string,
	serviceReturn []PaymentHistoryItem,
	logMethod string,
	logMessage string,
	user database.User,
	url string,
	statusFilter string, // pass "" for no filter
	expectedLen int, // expected length of response
) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}

	switch serviceMethod {
	case "GetPaymentHistory":
		mockService.On("GetPaymentHistory", mock.Anything, user.ID).Return(serviceReturn, nil)
	case "GetAllPayments":
		mockService.On("GetAllPayments", mock.Anything, statusFilter).Return(serviceReturn, nil)
	}
	mockLog.On(logMethod, mock.Anything, mock.Anything, logMessage, mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("GET", url, nil)
	if statusFilter != "" {
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("status", statusFilter)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	}
	w := httptest.NewRecorder()

	handler(cfg, w, r, user)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp []PaymentHistoryItem
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	assert.Len(t, resp, expectedLen)
	if expectedLen > 0 {
		assert.Equal(t, serviceReturn[0].ID, resp[0].ID)
	}
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetPayments_Success tests successful payment history retrieval
func TestHandlerGetPayments_Success(t *testing.T) {
	tests := []struct {
		name          string
		handler       func(cfg *HandlersPaymentConfig, w http.ResponseWriter, r *http.Request, user database.User)
		serviceMethod string
		serviceReturn []PaymentHistoryItem
		logMethod     string
		logMessage    string
		user          database.User
		url           string
		statusFilter  string
		expectedLen   int
	}{
		{
			name: "GetPaymentHistory_Success",
			handler: func(cfg *HandlersPaymentConfig, w http.ResponseWriter, r *http.Request, user database.User) {
				cfg.HandlerGetPaymentHistory(w, r, user)
			},
			serviceMethod: "GetPaymentHistory",
			serviceReturn: []PaymentHistoryItem{
				{ID: "p1", OrderID: "order1", Amount: "100.00", Currency: "USD", Status: "succeeded", Provider: "stripe", ProviderPaymentID: "pi_123"},
				{ID: "p2", OrderID: "order2", Amount: "200.00", Currency: "USD", Status: "pending", Provider: "stripe", ProviderPaymentID: "pi_456"},
			},
			logMethod:    "LogHandlerSuccess",
			logMessage:   "Get Payment history success",
			user:         database.User{ID: "u1"},
			url:          "/payments/history",
			statusFilter: "",
			expectedLen:  2,
		},
		{
			name: "AdminGetPayments_Success",
			handler: func(cfg *HandlersPaymentConfig, w http.ResponseWriter, r *http.Request, user database.User) {
				cfg.HandlerAdminGetPayments(w, r, user)
			},
			serviceMethod: "GetAllPayments",
			serviceReturn: []PaymentHistoryItem{
				{ID: "p1", OrderID: "order1", Amount: "100.00", Currency: "USD", Status: "succeeded", Provider: "stripe", ProviderPaymentID: "pi_123"},
				{ID: "p2", OrderID: "order2", Amount: "200.00", Currency: "USD", Status: "pending", Provider: "stripe", ProviderPaymentID: "pi_456"},
			},
			logMethod:    "LogHandlerSuccess",
			logMessage:   "Get all payments success",
			user:         database.User{ID: "admin1"},
			url:          "/admin/payments",
			statusFilter: "",
			expectedLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runHandlerGetPaymentsSuccessTest(
				t,
				tt.handler,
				tt.serviceMethod,
				tt.serviceReturn,
				tt.logMethod,
				tt.logMessage,
				tt.user,
				tt.url,
				tt.statusFilter,
				tt.expectedLen,
			)
		})
	}
}

// runHandlerGetPaymentsErrorTest is a shared helper for HandlerGetPaymentHistory/AdminGetPayments error scenario tests.
func runHandlerGetPaymentsErrorTest(
	t *testing.T,
	handler func(cfg *HandlersPaymentConfig, w http.ResponseWriter, r *http.Request, user database.User),
	serviceMethod string,
	appErr *handlers.AppError,
	logMethod string,
	logCode string,
	logMessage string,
	user database.User,
	url string,
	expectedStatus int,
) {
	mockService := new(MockPaymentServiceForGet)
	mockLog := new(MockLoggerForGet)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}

	switch serviceMethod {
	case "GetPaymentHistory":
		mockService.On("GetPaymentHistory", mock.Anything, user.ID).Return(nil, appErr)
	case "GetAllPayments":
		mockService.On("GetAllPayments", mock.Anything, "").Return(nil, appErr)
	}
	mockLog.On(logMethod, mock.Anything, mock.Anything, logCode, logMessage, mock.Anything, mock.Anything, appErr.Err).Return()

	r := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	handler(cfg, w, r, user)
	assert.Equal(t, expectedStatus, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerGetPaymentHistory_ServiceError tests internal error from service
func TestHandlerGetPaymentHistory_ServiceError(t *testing.T) {
	err := &handlers.AppError{Code: "get_history_error", Message: "fail", Err: errors.New("fail")}
	user := database.User{ID: "u1"}
	runHandlerGetPaymentsErrorTest(
		t,
		func(cfg *HandlersPaymentConfig, w http.ResponseWriter, r *http.Request, user database.User) {
			cfg.HandlerGetPaymentHistory(w, r, user)
		},
		"GetPaymentHistory",
		err,
		"LogHandlerError",
		"internal_error",
		"fail",
		user,
		"/payments/history",
		http.StatusInternalServerError,
	)
}

// TestHandlerAdminGetPayments_Success tests successful admin payments retrieval
func TestHandlerAdminGetPayments_Success(t *testing.T) {
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
	user := database.User{ID: "admin1"}
	runHandlerGetPaymentsSuccessTest(
		t,
		func(cfg *HandlersPaymentConfig, w http.ResponseWriter, r *http.Request, user database.User) {
			cfg.HandlerAdminGetPayments(w, r, user)
		},
		"GetAllPayments",
		expectedPayments,
		"LogHandlerSuccess",
		"Get all payments success",
		user,
		"/admin/payments",
		"",
		2,
	)
}

// TestHandlerAdminGetPayments_WithStatusFilter tests admin payments retrieval with status filter
func TestHandlerAdminGetPayments_WithStatusFilter(t *testing.T) {
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
	user := database.User{ID: "admin1"}
	runHandlerGetPaymentsSuccessTest(
		t,
		func(cfg *HandlersPaymentConfig, w http.ResponseWriter, r *http.Request, user database.User) {
			cfg.HandlerAdminGetPayments(w, r, user)
		},
		"GetAllPayments",
		expectedPayments,
		"LogHandlerSuccess",
		"Get all payments success",
		user,
		"/admin/payments/succeeded",
		"succeeded",
		1,
	)
}

// TestHandlerAdminGetPayments_ServiceError tests internal error from service
func TestHandlerAdminGetPayments_ServiceError(t *testing.T) {
	err := &handlers.AppError{Code: "get_all_payments_error", Message: "fail", Err: errors.New("fail")}
	user := database.User{ID: "admin1"}
	runHandlerGetPaymentsErrorTest(
		t,
		func(cfg *HandlersPaymentConfig, w http.ResponseWriter, r *http.Request, user database.User) {
			cfg.HandlerAdminGetPayments(w, r, user)
		},
		"GetAllPayments",
		err,
		"LogHandlerError",
		"internal_error",
		"fail",
		user,
		"/admin/payments",
		http.StatusInternalServerError,
	)
}
