// Package paymenthandlers provides HTTP handlers and configurations for processing payments, including Stripe integration, error handling, and payment-related request and response management.
package paymenthandlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
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

	testInvalidPayload(t, cfg.HandlerConfirmPayment, mockLog, user, "POST", "/payments/confirm", "confirm_payment")
}

// TestHandlerConfirmPayment_ErrorCases tests internal error from service and validation error from service
func TestHandlerConfirmPayment_ErrorCases(t *testing.T) {
	cases := []struct {
		name           string
		err            *handlers.AppError
		expectedStatus int
		logArgs        []any
	}{
		{
			name:           "service error",
			err:            &handlers.AppError{Code: "confirm_payment_error", Message: "fail", Err: errors.New("fail")},
			expectedStatus: http.StatusInternalServerError,
			logArgs:        []any{mock.Anything, "confirm_payment", "internal_error", "fail", mock.Anything, mock.Anything, errors.New("fail")},
		},
		{
			name:           "validation error",
			err:            &handlers.AppError{Code: "invalid_request", Message: "Invalid order ID", Err: errors.New("invalid order")},
			expectedStatus: http.StatusBadRequest,
			logArgs:        []any{mock.Anything, "confirm_payment", "invalid_request", "Invalid order ID", mock.Anything, mock.Anything, mock.Anything},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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
			mockService.On("ConfirmPayment", mock.Anything, expectedParams).Return(nil, tc.err)
			mockLog.On("LogHandlerError", tc.logArgs...).Return()

			httpReq := httptest.NewRequest("POST", "/payments/confirm", bytes.NewBuffer(jsonBody))
			w := httptest.NewRecorder()

			cfg.HandlerConfirmPayment(w, httpReq, user)
			assert.Equal(t, tc.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
			mockLog.AssertExpectations(t)
		})
	}
}
