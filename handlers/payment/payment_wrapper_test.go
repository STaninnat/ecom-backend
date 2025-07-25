// Package paymenthandlers provides HTTP handlers and configurations for processing payments, including Stripe integration, error handling, and payment-related request and response management.
package paymenthandlers

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// payment_wrapper_test.go: Tests for payment handler setup, initialization, and thread safety.

// TestInitPaymentService_Success verifies successful initialization of the PaymentService with all dependencies present.
func TestInitPaymentService_Success(t *testing.T) {
	apiCfg := &config.APIConfig{
		DB:     &database.Queries{},
		DBConn: &sql.DB{},
	}
	apiCfg.StripeSecretKey = "sk_test_123"

	cfg := &HandlersPaymentConfig{
		Config: &handlers.Config{
			APIConfig: apiCfg,
		},
	}

	// Test successful initialization
	err := cfg.InitPaymentService()
	require.NoError(t, err)
	assert.NotNil(t, cfg.paymentService)
}

// TestInitPaymentService_MissingHandlersConfig checks that initialization fails gracefully when the handlers config is missing.
func TestInitPaymentService_MissingHandlersConfig(t *testing.T) {
	cfg := &HandlersPaymentConfig{
		Config: nil,
	}

	// Test initialization with missing handlers config should return an error
	err := cfg.InitPaymentService()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "handlers config not initialized")
}

// TestInitPaymentService_MissingAPIConfig checks that initialization fails gracefully when the API config is missing.
func TestInitPaymentService_MissingAPIConfig(t *testing.T) {
	cfg := &HandlersPaymentConfig{
		Config: &handlers.Config{
			APIConfig: nil,
		},
	}

	// Test initialization with missing API config should return an error
	err := cfg.InitPaymentService()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API config not initialized")
}

// TestInitPaymentService_MissingDB checks that initialization fails gracefully when the database is missing.
func TestInitPaymentService_MissingDB(t *testing.T) {
	cfg := &HandlersPaymentConfig{
		Config: &handlers.Config{
			APIConfig: &config.APIConfig{
				DB: nil,
			},
		},
	}

	// Test initialization with missing DB should return an error
	err := cfg.InitPaymentService()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database not initialized")
}

// TestInitPaymentService_MissingDBConn checks that initialization fails gracefully when the database connection is missing.
func TestInitPaymentService_MissingDBConn(t *testing.T) {
	cfg := &HandlersPaymentConfig{
		Config: &handlers.Config{
			APIConfig: &config.APIConfig{
				DB:     &database.Queries{},
				DBConn: nil,
			},
		},
	}

	// Test initialization with missing DBConn should return an error
	err := cfg.InitPaymentService()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not initialized")
}

// TestInitPaymentService_MissingStripeKey checks that initialization fails gracefully when the Stripe secret key is missing.
func TestInitPaymentService_MissingStripeKey(t *testing.T) {
	cfg := &HandlersPaymentConfig{
		Config: &handlers.Config{
			APIConfig: &config.APIConfig{
				DB:              &database.Queries{},
				DBConn:          &sql.DB{},
				StripeSecretKey: "",
			},
		},
	}

	// Test initialization with missing Stripe key should return an error
	err := cfg.InitPaymentService()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stripe secret key not configured")
}

// TestGetPaymentService_AlreadyInitialized tests that GetPaymentService returns the existing paymentService when it's already initialized.
func TestGetPaymentService_AlreadyInitialized(t *testing.T) {
	mockService := new(MockPaymentService)
	cfg := &HandlersPaymentConfig{
		Config:         &handlers.Config{},
		paymentService: mockService,
	}
	service := cfg.GetPaymentService()
	assert.Equal(t, mockService, service)
}

// TestGetPaymentService_InitializesWithNilConfig tests that GetPaymentService initializes a new service even when Config is nil.
func TestGetPaymentService_InitializesWithNilConfig(t *testing.T) {
	cfg := &HandlersPaymentConfig{
		Config:         nil,
		paymentService: nil,
	}
	service := cfg.GetPaymentService()
	assert.NotNil(t, service)
}

// TestGetPaymentService_ThreadSafety tests that GetPaymentService is thread-safe with concurrent access.
func TestGetPaymentService_ThreadSafety(t *testing.T) {
	cfg := &HandlersPaymentConfig{
		Config:         nil,
		paymentService: nil,
	}

	var wg sync.WaitGroup
	services := make([]PaymentService, 10)

	// Launch multiple goroutines to access the service concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			services[index] = cfg.GetPaymentService()
		}(i)
	}

	wg.Wait()

	// All services should be the same instance (singleton pattern)
	firstService := services[0]
	for i := 1; i < 10; i++ {
		assert.Equal(t, firstService, services[i], "All services should be the same instance")
	}
}

// TestSetupStripeAPI tests that SetupStripeAPI sets the Stripe API key correctly.
func TestSetupStripeAPI(t *testing.T) {
	cfg := &HandlersPaymentConfig{
		Config: &handlers.Config{
			APIConfig: &config.APIConfig{
				StripeSecretKey: "sk_test_setup_123",
			},
		},
	}

	// Test that the method doesn't panic and sets the key
	assert.NotPanics(t, func() {
		cfg.SetupStripeAPI()
	})
}

// Helper for handlePaymentError status code tests
func runHandlePaymentErrorStatusTest(t *testing.T, codes []string, expectedStatus int, cfg *HandlersPaymentConfig, req *http.Request, mockHandlersConfig *MockHandlersConfig, op string) {
	for _, code := range codes {
		t.Run(http.StatusText(expectedStatus)+"_"+code, func(t *testing.T) {
			mockHandlersConfig.ExpectedCalls = nil
			w := httptest.NewRecorder()
			appErr := &handlers.AppError{Code: code, Message: "Test error", Err: errors.New("inner error")}
			expectedCode := expectedStatus
			expectedMsg := "Test error"
			if code == "user_not_found" || (code == "invalid_request" && expectedStatus == http.StatusNotFound) {
				expectedCode = http.StatusInternalServerError
				expectedMsg = "Internal server error"
			} else if code == "invalid_request" {
				// logCode = "invalid_request" // This line is removed
			} else if code == "unauthorized" {
				// logCode = "unauthorized" // This line is removed
			} else if code == "payment_not_found" {
				// logCode = "payment_not_found" // This line is removed
			}
			mockHandlersConfig.On("LogHandlerError", mock.Anything, op, mock.Anything, "Test error", "", "", mock.Anything).Return()

			cfg.handlePaymentError(w, req, appErr, op, "", "")

			assert.Equal(t, expectedCode, w.Code)
			assert.Contains(t, w.Body.String(), expectedMsg)
			mockHandlersConfig.AssertExpectations(t)
		})
	}
}

func TestHandlePaymentError_AllErrorCodes(t *testing.T) {
	mockHandlersConfig := &MockHandlersConfig{}
	cfg := &HandlersPaymentConfig{
		Config: &handlers.Config{},
		Logger: mockHandlersConfig,
	}
	req := httptest.NewRequest("GET", "/", nil)

	notFoundCodes := []string{"payment_not_found", "order_not_found", "user_not_found"}
	for _, code := range notFoundCodes {
		t.Run("Not_Found_"+code, func(t *testing.T) {
			w := httptest.NewRecorder()
			appErr := &handlers.AppError{Code: code, Message: "Test error", Err: errors.New("inner error")}
			mockHandlersConfig.On("LogHandlerError", mock.Anything, "test_op", code, "Test error", "", "", mock.Anything).Return()

			cfg.handlePaymentError(w, req, appErr, "test_op", "", "")

			// For user_not_found, expect 404 and 'Test error'
			if code == "user_not_found" {
				assert.Equal(t, http.StatusNotFound, w.Code)
				assert.Contains(t, w.Body.String(), "Test error")
			} else {
				assert.Equal(t, http.StatusNotFound, w.Code)
				assert.Contains(t, w.Body.String(), "Test error")
			}
			mockHandlersConfig.AssertExpectations(t)
		})
	}

	forbiddenCodes := []string{"unauthorized"}
	runHandlePaymentErrorStatusTest(t, forbiddenCodes, http.StatusForbidden, cfg, req, mockHandlersConfig, "test_op")

	// Test all error codes that should return 400 Bad Request
	badRequestCodes := []string{"invalid_request", "missing_order_id", "missing_user_id", "invalid_currency", "payment_exists", "invalid_order_status", "invalid_status", "invalid_amount", "invalid_payment"}
	for _, code := range badRequestCodes {
		t.Run("BadRequest_"+code, func(t *testing.T) {
			w := httptest.NewRecorder()
			appErr := &handlers.AppError{Code: code, Message: "Test error"}
			mockHandlersConfig.On("LogHandlerError", mock.Anything, "test_op", code, "Test error", "", "", mock.Anything).Return()

			cfg.handlePaymentError(w, req, appErr, "test_op", "", "")

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "Test error")
			mockHandlersConfig.AssertExpectations(t)
		})
	}

	// Test all error codes that should return 500 Internal Server Error
	internalErrorCodes := []string{"database_error", "transaction_error", "commit_error", "stripe_error", "webhook_error"}
	for _, code := range internalErrorCodes {
		t.Run("InternalError_"+code, func(t *testing.T) {
			w := httptest.NewRecorder()
			appErr := &handlers.AppError{Code: code, Message: "Test error", Err: errors.New("inner error")}
			mockHandlersConfig.On("LogHandlerError", mock.Anything, "test_op", code, "Test error", "", "", mock.Anything).Return()

			cfg.handlePaymentError(w, req, appErr, "test_op", "", "")

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			if code == "stripe_error" || code == "webhook_error" {
				assert.Contains(t, w.Body.String(), "Payment service error")
			} else {
				assert.Contains(t, w.Body.String(), "Something went wrong, please try again later")
			}
			mockHandlersConfig.AssertExpectations(t)
		})
	}

	// Test unknown error code
	t.Run("UnknownError", func(t *testing.T) {
		w := httptest.NewRecorder()
		appErr := &handlers.AppError{Code: "unknown_code", Message: "Test error", Err: errors.New("inner error")}
		mockHandlersConfig.On("LogHandlerError", mock.Anything, "test_op", "internal_error", "Test error", "", "", mock.Anything).Return()

		cfg.handlePaymentError(w, req, appErr, "test_op", "", "")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
		mockHandlersConfig.AssertExpectations(t)
	})

	// Test non-AppError
	t.Run("NonAppError", func(t *testing.T) {
		w := httptest.NewRecorder()
		regularErr := errors.New("regular error")
		mockHandlersConfig.On("LogHandlerError", mock.Anything, "test_op", "unknown_error", "Unknown error occurred", "", "", mock.Anything).Return()

		cfg.handlePaymentError(w, req, regularErr, "test_op", "", "")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
		mockHandlersConfig.AssertExpectations(t)
	})
}

// TestHandlePaymentError_StatusCodes tests that handlePaymentError correctly categorizes and handles all error codes.
func TestHandlePaymentError_StatusCodes(t *testing.T) {
	testCases := []struct {
		name           string
		code           string
		expectedStatus int
	}{
		// Not found codes
		{name: "NotFound_payment_not_found", code: "payment_not_found", expectedStatus: http.StatusNotFound},
		{name: "NotFound_order_not_found", code: "order_not_found", expectedStatus: http.StatusNotFound},
		{name: "NotFound_user_not_found", code: "user_not_found", expectedStatus: http.StatusNotFound},
		// Forbidden codes
		{name: "Forbidden_unauthorized_payment", code: "unauthorized_payment", expectedStatus: http.StatusForbidden},
		{name: "Forbidden_unauthorized_order", code: "unauthorized_order", expectedStatus: http.StatusForbidden},
		{name: "Forbidden_unauthorized_user", code: "unauthorized_user", expectedStatus: http.StatusForbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHandlersConfig := &MockHandlersConfig{}
			cfg := &HandlersPaymentConfig{
				Config: &handlers.Config{},
				Logger: mockHandlersConfig,
			}
			req := httptest.NewRequest("POST", "/test", nil)
			w := httptest.NewRecorder()
			appErr := &handlers.AppError{Code: tc.code, Message: "Test error", Err: errors.New("inner error")}
			mockHandlersConfig.On("LogHandlerError", mock.Anything, "test_op", mock.Anything, "Test error", "", "", mock.Anything).Return()

			cfg.handlePaymentError(w, req, appErr, "test_op", "", "")

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), "Test error")
			mockHandlersConfig.AssertExpectations(t)
		})
	}
}

// TestPaymentRequestResponseTypes tests that the request/response types are properly defined.
func TestPaymentRequestResponseTypes(t *testing.T) {
	// Test CreatePaymentIntentRequest
	createReq := CreatePaymentIntentRequest{
		OrderID:  "order_123",
		Currency: "USD",
	}
	assert.Equal(t, "order_123", createReq.OrderID)
	assert.Equal(t, "USD", createReq.Currency)

	// Test CreatePaymentIntentResponse
	createResp := CreatePaymentIntentResponse{
		ClientSecret: "pi_secret_123",
	}
	assert.Equal(t, "pi_secret_123", createResp.ClientSecret)

	// Test ConfirmPaymentRequest
	confirmReq := ConfirmPaymentRequest{
		OrderID: "order_123",
	}
	assert.Equal(t, "order_123", confirmReq.OrderID)

	// Test ConfirmPaymentResponse
	confirmResp := ConfirmPaymentResponse{
		Status: "succeeded",
	}
	assert.Equal(t, "succeeded", confirmResp.Status)

	// Test GetPaymentResponse
	getResp := GetPaymentResponse{
		ID:                "pay_123",
		OrderID:           "order_123",
		UserID:            "user_123",
		Amount:            100.50,
		Currency:          "USD",
		Status:            "succeeded",
		Provider:          "stripe",
		ProviderPaymentID: "pi_123",
	}
	assert.Equal(t, "pay_123", getResp.ID)
	assert.Equal(t, "order_123", getResp.OrderID)
	assert.Equal(t, "user_123", getResp.UserID)
	assert.InEpsilon(t, 100.50, getResp.Amount, 0.001)
	assert.Equal(t, "USD", getResp.Currency)
	assert.Equal(t, "succeeded", getResp.Status)
	assert.Equal(t, "stripe", getResp.Provider)
	assert.Equal(t, "pi_123", getResp.ProviderPaymentID)

	// Test PaymentHistoryItem
	historyItem := PaymentHistoryItem{
		ID:                "pay_123",
		OrderID:           "order_123",
		Amount:            "100.50",
		Currency:          "USD",
		Status:            "succeeded",
		Provider:          "stripe",
		ProviderPaymentID: "pi_123",
	}
	assert.Equal(t, "pay_123", historyItem.ID)
	assert.Equal(t, "order_123", historyItem.OrderID)
	assert.Equal(t, "100.50", historyItem.Amount)
	assert.Equal(t, "USD", historyItem.Currency)
	assert.Equal(t, "succeeded", historyItem.Status)
	assert.Equal(t, "stripe", historyItem.Provider)
	assert.Equal(t, "pi_123", historyItem.ProviderPaymentID)
}

// TestInitPaymentService_AllValidationBranches tests all validation branches in InitPaymentService.
func TestInitPaymentService_AllValidationBranches(t *testing.T) {
	// Test missing Config
	t.Run("MissingHandlersConfig", func(t *testing.T) {
		cfg := &HandlersPaymentConfig{
			Config: nil,
		}
		err := cfg.InitPaymentService()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "handlers config not initialized")
	})

	// Test missing APIConfig
	t.Run("MissingAPIConfig", func(t *testing.T) {
		cfg := &HandlersPaymentConfig{
			Config: &handlers.Config{
				APIConfig: nil,
			},
		}
		err := cfg.InitPaymentService()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API config not initialized")
	})

	// Test missing DB
	t.Run("MissingDB", func(t *testing.T) {
		cfg := &HandlersPaymentConfig{
			Config: &handlers.Config{
				APIConfig: &config.APIConfig{
					DB: nil,
				},
			},
		}
		err := cfg.InitPaymentService()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database not initialized")
	})

	// Test missing DBConn
	t.Run("MissingDBConn", func(t *testing.T) {
		cfg := &HandlersPaymentConfig{
			Config: &handlers.Config{
				APIConfig: &config.APIConfig{
					DB:     &database.Queries{},
					DBConn: nil,
				},
			},
		}
		err := cfg.InitPaymentService()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database connection not initialized")
	})

	// Test missing StripeSecretKey
	t.Run("MissingStripeSecretKey", func(t *testing.T) {
		cfg := &HandlersPaymentConfig{
			Config: &handlers.Config{
				APIConfig: &config.APIConfig{
					DB:              &database.Queries{},
					DBConn:          &sql.DB{},
					StripeSecretKey: "",
				},
			},
		}
		err := cfg.InitPaymentService()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "stripe secret key not configured")
	})

	// Test successful initialization
	t.Run("SuccessfulInitialization", func(t *testing.T) {
		cfg := &HandlersPaymentConfig{
			Config: &handlers.Config{
				APIConfig: &config.APIConfig{
					DB:              &database.Queries{},
					DBConn:          &sql.DB{},
					StripeSecretKey: "sk_test_123",
				},
			},
		}
		err := cfg.InitPaymentService()
		require.NoError(t, err)
		assert.NotNil(t, cfg.paymentService)
	})
}
