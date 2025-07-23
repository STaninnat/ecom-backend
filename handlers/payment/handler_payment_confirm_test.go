// Package paymenthandlers provides HTTP handlers and configurations for processing payments, including Stripe integration, error handling, and payment-related request and response management.
package paymenthandlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// handler_payment_confirm_test.go: Tests for payment confirmation HTTP handler covering success, validation, and error cases.

// TestHandlerConfirmPayment_Success tests successful payment confirmation
func TestHandlerConfirmPayment_Success(t *testing.T) {
	mockService := new(MockPaymentServiceForConfirm)
	mockLog := new(MockLoggerForConfirm)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	req := ConfirmPaymentRequest{OrderID: "order1"}
	jsonBody, _ := json.Marshal(req)
	expectedParams := ConfirmPaymentParams{OrderID: "order1", UserID: "u1"}
	expectedResult := &ConfirmPaymentResult{Status: "succeeded"}
	mockService.On("ConfirmPayment", mock.Anything, expectedParams).Return(expectedResult, nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "confirm_payment", "Payment confirmation success", mock.Anything, mock.Anything).Return()

	httpReq := httptest.NewRequest("POST", "/payments/confirm", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerConfirmPayment(w, httpReq, user)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp ConfirmPaymentResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	assert.Equal(t, "succeeded", resp.Status)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerConfirmPayment_InvalidPayload tests invalid JSON payload
func TestHandlerConfirmPayment_InvalidPayload(t *testing.T) {
	mockService := new(MockPaymentServiceForConfirm)
	mockLog := new(MockLoggerForConfirm)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	badBody := []byte(`{"bad":}`)
	mockLog.On("LogHandlerError", mock.Anything, "confirm_payment", "invalid_request", "Invalid request payload", mock.Anything, mock.Anything, mock.Anything).Return()

	httpReq := httptest.NewRequest("POST", "/payments/confirm", bytes.NewBuffer(badBody))
	w := httptest.NewRecorder()

	cfg.HandlerConfirmPayment(w, httpReq, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLog.AssertExpectations(t)
}

// TestHandlerConfirmPayment_ServiceError tests internal error from service
func TestHandlerConfirmPayment_ServiceError(t *testing.T) {
	mockService := new(MockPaymentServiceForConfirm)
	mockLog := new(MockLoggerForConfirm)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	req := ConfirmPaymentRequest{OrderID: "order1"}
	jsonBody, _ := json.Marshal(req)
	expectedParams := ConfirmPaymentParams{OrderID: "order1", UserID: "u1"}
	err := &handlers.AppError{Code: "confirm_payment_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("ConfirmPayment", mock.Anything, expectedParams).Return(nil, err)
	mockLog.On("LogHandlerError", mock.Anything, "confirm_payment", "internal_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	httpReq := httptest.NewRequest("POST", "/payments/confirm", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerConfirmPayment(w, httpReq, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerConfirmPayment_ValidationError tests validation error from service
func TestHandlerConfirmPayment_ValidationError(t *testing.T) {
	mockService := new(MockPaymentServiceForConfirm)
	mockLog := new(MockLoggerForConfirm)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	req := ConfirmPaymentRequest{OrderID: "order1"}
	jsonBody, _ := json.Marshal(req)
	expectedParams := ConfirmPaymentParams{OrderID: "order1", UserID: "u1"}
	err := &handlers.AppError{Code: "invalid_request", Message: "Invalid order ID", Err: errors.New("invalid order")}
	mockService.On("ConfirmPayment", mock.Anything, expectedParams).Return(nil, err)
	mockLog.On("LogHandlerError", mock.Anything, "confirm_payment", "invalid_request", "Invalid order ID", mock.Anything, mock.Anything, nil).Return()

	httpReq := httptest.NewRequest("POST", "/payments/confirm", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerConfirmPayment(w, httpReq, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}
