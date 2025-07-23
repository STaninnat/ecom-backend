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

// handler_payment_create_test.go: Tests for payment creation HTTP handler behavior and error handling.

// TestHandlerCreatePayment_Success tests the successful creation of a payment via the handler.
// It verifies that the handler returns HTTP 201, the correct response message, and client secret when the service succeeds.
func TestHandlerCreatePayment_Success(t *testing.T) {
	mockService := new(MockPaymentServiceForCreate)
	mockLog := new(MockLoggerForCreate)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	req := CreatePaymentIntentRequest{
		OrderID:  "order1",
		Currency: "USD",
	}
	jsonBody, _ := json.Marshal(req)

	expectedParams := CreatePaymentParams{
		OrderID:  "order1",
		UserID:   "u1",
		Currency: "USD",
	}
	expectedResult := &CreatePaymentResult{
		PaymentID:    "payment1",
		ClientSecret: "pi_test_secret",
	}

	mockService.On("CreatePayment", mock.Anything, expectedParams).Return(expectedResult, nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "create_payment", "Created payment successful", mock.Anything, mock.Anything).Return()

	httpReq := httptest.NewRequest("POST", "/payments", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerCreatePayment(w, httpReq, user)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp CreatePaymentIntentResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	assert.Equal(t, "pi_test_secret", resp.ClientSecret)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerCreatePayment_InvalidPayload tests the handler's response to an invalid JSON payload.
// It checks that the handler returns HTTP 400 and logs the appropriate error.
func TestHandlerCreatePayment_InvalidPayload(t *testing.T) {
	mockService := new(MockPaymentServiceForCreate)
	mockLog := new(MockLoggerForCreate)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	badBody := []byte(`{"bad":}`)
	mockLog.On("LogHandlerError", mock.Anything, "create_payment", "invalid_request", "Invalid request payload", mock.Anything, mock.Anything, mock.Anything).Return()

	httpReq := httptest.NewRequest("POST", "/payments", bytes.NewBuffer(badBody))
	w := httptest.NewRecorder()

	cfg.HandlerCreatePayment(w, httpReq, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLog.AssertExpectations(t)
}

// TestHandlerCreatePayment_ServiceError tests the handler's behavior when the payment service returns an error.
// It ensures the handler returns HTTP 500 and logs the service error correctly.
func TestHandlerCreatePayment_ServiceError(t *testing.T) {
	mockService := new(MockPaymentServiceForCreate)
	mockLog := new(MockLoggerForCreate)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	req := CreatePaymentIntentRequest{
		OrderID:  "order1",
		Currency: "USD",
	}
	jsonBody, _ := json.Marshal(req)

	expectedParams := CreatePaymentParams{
		OrderID:  "order1",
		UserID:   "u1",
		Currency: "USD",
	}

	err := &handlers.AppError{Code: "create_payment_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("CreatePayment", mock.Anything, expectedParams).Return(nil, err)
	mockLog.On("LogHandlerError", mock.Anything, "create_payment", "internal_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	httpReq := httptest.NewRequest("POST", "/payments", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerCreatePayment(w, httpReq, user)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerCreatePayment_ValidationError tests the handler's behavior when the service returns a validation error.
// It ensures the handler returns HTTP 400 for validation errors.
func TestHandlerCreatePayment_ValidationError(t *testing.T) {
	mockService := new(MockPaymentServiceForCreate)
	mockLog := new(MockLoggerForCreate)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	user := database.User{ID: "u1"}
	req := CreatePaymentIntentRequest{
		OrderID:  "order1",
		Currency: "USD",
	}
	jsonBody, _ := json.Marshal(req)

	expectedParams := CreatePaymentParams{
		OrderID:  "order1",
		UserID:   "u1",
		Currency: "USD",
	}

	err := &handlers.AppError{Code: "invalid_request", Message: "Invalid order ID", Err: errors.New("invalid order")}
	mockService.On("CreatePayment", mock.Anything, expectedParams).Return(nil, err)
	mockLog.On("LogHandlerError", mock.Anything, "create_payment", "invalid_request", "Invalid order ID", mock.Anything, mock.Anything, nil).Return()

	httpReq := httptest.NewRequest("POST", "/payments", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	cfg.HandlerCreatePayment(w, httpReq, user)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}
