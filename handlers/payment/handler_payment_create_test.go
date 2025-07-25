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

	testInvalidPayload(t, cfg.HandlerCreatePayment, mockLog, user, "POST", "/payments", "create_payment")
}

// TestHandlerCreatePayment_ErrorCases tests the handler's behavior for different error cases.
func TestHandlerCreatePayment_ErrorCases(t *testing.T) {
	cases := []struct {
		name           string
		err            *handlers.AppError
		expectedStatus int
		logArgs        []any
	}{
		{
			name:           "service error",
			err:            &handlers.AppError{Code: "create_payment_error", Message: "fail", Err: errors.New("fail")},
			expectedStatus: http.StatusInternalServerError,
			logArgs:        []any{mock.Anything, "create_payment", "internal_error", "fail", mock.Anything, mock.Anything, errors.New("fail")},
		},
		{
			name:           "validation error",
			err:            &handlers.AppError{Code: "invalid_request", Message: "Invalid order ID", Err: errors.New("invalid order")},
			expectedStatus: http.StatusBadRequest,
			logArgs:        []any{mock.Anything, "create_payment", "invalid_request", "Invalid order ID", mock.Anything, mock.Anything, mock.Anything},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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
			mockService.On("CreatePayment", mock.Anything, expectedParams).Return(nil, tc.err)
			mockLog.On("LogHandlerError", tc.logArgs...).Return()

			httpReq := httptest.NewRequest("POST", "/payments", bytes.NewBuffer(jsonBody))
			w := httptest.NewRecorder()

			cfg.HandlerCreatePayment(w, httpReq, user)
			assert.Equal(t, tc.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
			mockLog.AssertExpectations(t)
		})
	}
}
