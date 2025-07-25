// Package paymenthandlers provides HTTP handlers and configurations for processing payments, including Stripe integration, error handling, and payment-related request and response management.
package paymenthandlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
)

// handler_payment_webhook_test.go: Tests for Stripe webhook HTTP handler including payload validation and service error handling.

const (
	testSignatureWebhook = "t=1234567890,v1=abc123"
)

// TestHandlerStripeWebhook_Success tests successful webhook processing
func TestHandlerStripeWebhook_Success(t *testing.T) {
	mockService := new(MockPaymentServiceForWebhook)
	mockLog := new(MockLoggerForWebhook)
	cfg := &HandlersPaymentConfig{
		Config: &handlers.Config{
			APIConfig: &config.APIConfig{
				StripeWebhookSecret: "whsec_test",
			},
		},
		Logger:         mockLog,
		paymentService: mockService,
	}
	payload := []byte(`{"type":"payment_intent.succeeded"}`)
	signature := testSignatureWebhook
	mockService.On("HandleWebhook", mock.Anything, payload, signature, "whsec_test").Return(nil)
	mockLog.On("LogHandlerSuccess", mock.Anything, "payment_webhook", "Webhook processed successfully", mock.Anything, mock.Anything).Return()

	r := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewBuffer(payload))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Stripe-Signature", signature)
	w := httptest.NewRecorder()

	cfg.HandlerStripeWebhook(w, r)
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerStripeWebhook_ReadBodyError tests error reading request body
func TestHandlerStripeWebhook_ReadBodyError(t *testing.T) {
	mockService := new(MockPaymentServiceForWebhook)
	mockLog := new(MockLoggerForWebhook)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	mockLog.On("LogHandlerError", mock.Anything, "payment_webhook", "read_failed", "Error reading request body", mock.Anything, mock.Anything, mock.Anything).Return()

	// Create a request with a body that will cause a read error
	r := httptest.NewRequest("POST", "/webhooks/stripe", &errorReader{})
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	cfg.HandlerStripeWebhook(w, r)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	mockLog.AssertExpectations(t)
}

// TestHandlerStripeWebhook_BadRequestScenarios tests bad request scenarios for webhook processing.
func TestHandlerStripeWebhook_BadRequestScenarios(t *testing.T) {
	tests := []struct {
		name            string
		contentType     string
		setSignature    bool
		expectedLogCode string
		expectedLogMsg  string
		expectedStatus  int
	}{
		{
			name:            "InvalidContentType",
			contentType:     "text/plain",
			setSignature:    true,
			expectedLogCode: "invalid_content_type",
			expectedLogMsg:  "Expected application/json",
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:            "MissingSignature",
			contentType:     "application/json",
			setSignature:    false,
			expectedLogCode: "missing_signature",
			expectedLogMsg:  "Missing Stripe signature header",
			expectedStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPaymentServiceForWebhook)
			mockLog := new(MockLoggerForWebhook)
			cfg := &HandlersPaymentConfig{
				Config:         &handlers.Config{},
				Logger:         mockLog,
				paymentService: mockService,
			}
			mockLog.On("LogHandlerError", mock.Anything, "payment_webhook", tt.expectedLogCode, tt.expectedLogMsg, mock.Anything, mock.Anything, nil).Return()

			r := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewBuffer([]byte(`{}`)))
			r.Header.Set("Content-Type", tt.contentType)
			if tt.setSignature {
				r.Header.Set("Stripe-Signature", "test-signature")
			}
			w := httptest.NewRecorder()

			cfg.HandlerStripeWebhook(w, r)
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockLog.AssertExpectations(t)
		})
	}
}

// TestHandlerStripeWebhook_ErrorScenarios tests error scenarios for webhook processing.
func TestHandlerStripeWebhook_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name           string
		err            *handlers.AppError
		logCode        string
		logMsg         string
		expectedStatus int
	}{
		{
			name:           "ServiceError",
			err:            &handlers.AppError{Code: "webhook_error", Message: "fail", Err: errors.New("fail")},
			logCode:        "webhook_error",
			logMsg:         "fail",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "InvalidSignature",
			err:            &handlers.AppError{Code: "invalid_signature", Message: "Invalid signature", Err: errors.New("invalid")},
			logCode:        "internal_error",
			logMsg:         "Invalid signature",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPaymentServiceForWebhook)
			mockLog := new(MockLoggerForWebhook)
			cfg := &HandlersPaymentConfig{
				Config: &handlers.Config{
					APIConfig: &config.APIConfig{
						StripeWebhookSecret: "whsec_test",
					},
				},
				Logger:         mockLog,
				paymentService: mockService,
			}
			payload := []byte(`{"type":"payment_intent.succeeded"}`)
			signature := testSignatureWebhook
			mockService.On("HandleWebhook", mock.Anything, payload, signature, "whsec_test").Return(tt.err)
			mockLog.On("LogHandlerError", mock.Anything, "payment_webhook", tt.logCode, tt.logMsg, mock.Anything, mock.Anything, tt.err.Err).Return()

			r := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewBuffer(payload))
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Stripe-Signature", signature)
			w := httptest.NewRecorder()

			cfg.HandlerStripeWebhook(w, r)
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
			mockLog.AssertExpectations(t)
		})
	}
}

// errorReader is a reader that always returns an error
type errorReader struct{}

func (e *errorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error")
}
