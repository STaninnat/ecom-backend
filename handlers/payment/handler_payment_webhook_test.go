package paymenthandlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestHandlerStripeWebhook_Success tests successful webhook processing
func TestHandlerStripeWebhook_Success(t *testing.T) {
	mockService := new(MockPaymentServiceForWebhook)
	mockLog := new(MockLoggerForWebhook)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{
				StripeWebhookSecret: "whsec_test",
			},
		},
		Logger:         mockLog,
		paymentService: mockService,
	}
	payload := []byte(`{"type":"payment_intent.succeeded"}`)
	signature := "t=1234567890,v1=abc123"
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
		HandlersConfig: &handlers.HandlersConfig{},
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

// TestHandlerStripeWebhook_InvalidContentType tests invalid content type
func TestHandlerStripeWebhook_InvalidContentType(t *testing.T) {
	mockService := new(MockPaymentServiceForWebhook)
	mockLog := new(MockLoggerForWebhook)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	mockLog.On("LogHandlerError", mock.Anything, "payment_webhook", "invalid_content_type", "Expected application/json", mock.Anything, mock.Anything, nil).Return()

	r := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewBuffer([]byte(`{}`)))
	r.Header.Set("Content-Type", "text/plain") // wrong content type
	w := httptest.NewRecorder()

	cfg.HandlerStripeWebhook(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLog.AssertExpectations(t)
}

// TestHandlerStripeWebhook_MissingSignature tests missing signature header
func TestHandlerStripeWebhook_MissingSignature(t *testing.T) {
	mockService := new(MockPaymentServiceForWebhook)
	mockLog := new(MockLoggerForWebhook)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{},
		Logger:         mockLog,
		paymentService: mockService,
	}
	mockLog.On("LogHandlerError", mock.Anything, "payment_webhook", "missing_signature", "Missing Stripe signature header", mock.Anything, mock.Anything, nil).Return()

	r := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewBuffer([]byte(`{}`)))
	r.Header.Set("Content-Type", "application/json")
	// no Stripe-Signature header
	w := httptest.NewRecorder()

	cfg.HandlerStripeWebhook(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockLog.AssertExpectations(t)
}

// TestHandlerStripeWebhook_ServiceError tests internal error from service
func TestHandlerStripeWebhook_ServiceError(t *testing.T) {
	mockService := new(MockPaymentServiceForWebhook)
	mockLog := new(MockLoggerForWebhook)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{
				StripeWebhookSecret: "whsec_test",
			},
		},
		Logger:         mockLog,
		paymentService: mockService,
	}
	payload := []byte(`{"type":"payment_intent.succeeded"}`)
	signature := "t=1234567890,v1=abc123"
	err := &handlers.AppError{Code: "webhook_error", Message: "fail", Err: errors.New("fail")}
	mockService.On("HandleWebhook", mock.Anything, payload, signature, "whsec_test").Return(err)
	mockLog.On("LogHandlerError", mock.Anything, "payment_webhook", "webhook_error", "fail", mock.Anything, mock.Anything, err.Err).Return()

	r := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewBuffer(payload))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Stripe-Signature", signature)
	w := httptest.NewRecorder()

	cfg.HandlerStripeWebhook(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// TestHandlerStripeWebhook_InvalidSignature tests invalid signature error
func TestHandlerStripeWebhook_InvalidSignature(t *testing.T) {
	mockService := new(MockPaymentServiceForWebhook)
	mockLog := new(MockLoggerForWebhook)
	cfg := &HandlersPaymentConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: &config.APIConfig{
				StripeWebhookSecret: "whsec_test",
			},
		},
		Logger:         mockLog,
		paymentService: mockService,
	}
	payload := []byte(`{"type":"payment_intent.succeeded"}`)
	signature := "t=1234567890,v1=abc123"
	err := &handlers.AppError{Code: "invalid_signature", Message: "Invalid signature", Err: errors.New("invalid signature")}
	mockService.On("HandleWebhook", mock.Anything, payload, signature, "whsec_test").Return(err)
	mockLog.On("LogHandlerError", mock.Anything, "payment_webhook", "internal_error", "Invalid signature", mock.Anything, mock.Anything, err.Err).Return()

	r := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewBuffer(payload))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Stripe-Signature", signature)
	w := httptest.NewRecorder()

	cfg.HandlerStripeWebhook(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
	mockLog.AssertExpectations(t)
}

// errorReader is a reader that always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
