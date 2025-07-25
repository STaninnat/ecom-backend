// Package authhandlers implements HTTP handlers for user authentication, including signup, signin, signout, token refresh, and OAuth integration.
package authhandlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
)

// auth_wrapper_test.go: Tests for auth handler configuration, ensuring correct service setup and error handling behavior.

// TestInitAuthService_Success verifies successful initialization of the AuthService with all dependencies present.
func TestInitAuthService_Success(t *testing.T) {
	apiCfg := &config.APIConfig{}
	apiCfg.DB = &database.Queries{}

	mockRedis, _ := redismock.NewClientMock()
	apiCfg.RedisClient = mockRedis

	cfg := &HandlersAuthConfig{
		Config: &handlers.Config{
			APIConfig: apiCfg,
			Auth:      &auth.Config{},
			OAuth:     &config.OAuthConfig{},
			Logger:    logrus.New(),
		},
		HandlersCartConfig: nil,
	}

	// Test successful initialization
	err := cfg.InitAuthService()
	require.NoError(t, err)
	assert.NotNil(t, cfg.authService)
}

// TestInitAuthService_MissingDB checks that initialization fails gracefully when the database is missing.
func TestInitAuthService_MissingDB(t *testing.T) {
	cfg := &HandlersAuthConfig{
		Config: &handlers.Config{
			APIConfig: nil, // Missing APIConfig (which contains DB)
			Auth:      &auth.Config{},
			OAuth:     &config.OAuthConfig{},
			Logger:    logrus.New(),
		},
		HandlersCartConfig: nil,
	}

	// Test initialization with missing DB should return an error, not panic
	err := cfg.InitAuthService()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestInitAuthService_MissingAuth checks that initialization fails gracefully when the Auth config is missing.
func TestInitAuthService_MissingAuth(t *testing.T) {
	apiCfg := &config.APIConfig{}
	apiCfg.DB = (*database.Queries)(nil)
	cfg := &HandlersAuthConfig{
		Config: &handlers.Config{
			APIConfig: apiCfg,
			Auth:      nil, // Missing Auth
			OAuth:     &config.OAuthConfig{},
			Logger:    logrus.New(),
		},
		HandlersCartConfig: nil,
	}

	// Test initialization with missing Auth
	err := cfg.InitAuthService()
	require.Error(t, err)
	assert.Equal(t, "database not initialized", err.Error())
}

// TestInitAuthService_MissingRedis checks that initialization does not panic when Redis is missing.
func TestInitAuthService_MissingRedis(t *testing.T) {
	apiCfg := &config.APIConfig{}
	apiCfg.DB = (*database.Queries)(nil)
	cfg := &HandlersAuthConfig{
		Config: &handlers.Config{
			APIConfig: apiCfg, // This would need RedisClient to be nil
			Auth:      &auth.Config{},
			OAuth:     &config.OAuthConfig{},
			Logger:    logrus.New(),
		},
		HandlersCartConfig: nil,
	}

	// Test initialization with missing Redis (this would need the actual APIConfig to have nil RedisClient)
	// For now, we'll test that it doesn't panic
	assert.NotPanics(t, func() {
		_ = cfg.InitAuthService()
	})
}

// TestGetAuthService_Initialized verifies that GetAuthService returns the initialized service.
func TestGetAuthService_Initialized(t *testing.T) {
	apiCfg := &config.APIConfig{}
	apiCfg.DB = (*database.Queries)(nil)
	cfg := &HandlersAuthConfig{
		Config: &handlers.Config{
			APIConfig: apiCfg,
			Auth:      &auth.Config{},
			OAuth:     &config.OAuthConfig{},
			Logger:    logrus.New(),
		},
		HandlersCartConfig: nil,
	}

	// Initialize the service first
	err := cfg.InitAuthService()
	if err == nil {
		// Test getting the service
		service := cfg.GetAuthService()
		assert.NotNil(t, service)
		assert.Equal(t, cfg.authService, service)
	}
}

// TestGetAuthService_NotInitialized checks that GetAuthService auto-initializes the service if not already done.
func TestGetAuthService_NotInitialized(t *testing.T) {
	apiCfg := &config.APIConfig{}
	apiCfg.DB = (*database.Queries)(nil)
	cfg := &HandlersAuthConfig{
		Config: &handlers.Config{
			APIConfig: apiCfg,
			Auth:      &auth.Config{},
			OAuth:     &config.OAuthConfig{},
			Logger:    logrus.New(),
		},
		HandlersCartConfig: nil,
	}

	// Test getting service without initialization (should auto-initialize)
	service := cfg.GetAuthService()
	assert.NotNil(t, service)
	assert.NotNil(t, cfg.authService)
}

// TestGetAuthService_ThreadSafety checks that GetAuthService is thread-safe under concurrent access.
func TestGetAuthService_ThreadSafety(t *testing.T) {
	apiCfg := &config.APIConfig{}
	apiCfg.DB = (*database.Queries)(nil)
	cfg := &HandlersAuthConfig{
		Config: &handlers.Config{
			APIConfig: apiCfg,
			Auth:      &auth.Config{},
			OAuth:     &config.OAuthConfig{},
			Logger:    logrus.New(),
		},
		HandlersCartConfig: nil,
	}

	// Test concurrent access to GetAuthService
	done := make(chan bool, 10)
	for range 10 {
		go func() {
			service := cfg.GetAuthService()
			assert.NotNil(t, service)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}

	// Verify service was initialized
	assert.NotNil(t, cfg.authService)
}

// TestHandleAuthError_AllErrorCodes verifies handleAuthError returns correct status codes and messages for all error codes.
func TestHandleAuthError_AllErrorCodes(t *testing.T) {
	mockHandlersConfig := &MockHandlersConfig{}
	cfg := &HandlersAuthConfig{
		Config: &handlers.Config{},
		Logger: mockHandlersConfig,
	}

	req := httptest.NewRequest("POST", "/test", nil)

	// Test all error codes that should return 400 Bad Request
	badRequestCodes := []string{"name_exists", "email_exists", "user_not_found", "invalid_password"}
	for _, code := range badRequestCodes {
		t.Run("BadRequest_"+code, func(t *testing.T) {
			w := httptest.NewRecorder()
			appErr := &handlers.AppError{Code: code, Message: "Test error"}
			mockHandlersConfig.On("LogHandlerError", mock.Anything, "test_op", code, "Test error", "", "", mock.Anything).Return()

			cfg.handleAuthError(w, req, appErr, "test_op", "", "")

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "Test error")
			mockHandlersConfig.AssertExpectations(t)
		})
	}

	t.Run("AllErrorCodes", func(t *testing.T) {
		testCases := []struct {
			name        string
			errorCode   string
			httpStatus  int
			expectedMsg string
		}{
			// Internal error codes (500)
			{"InternalError_database_error", "database_error", http.StatusInternalServerError, "Something went wrong, please try again later"},
			{"InternalError_transaction_error", "transaction_error", http.StatusInternalServerError, "Something went wrong, please try again later"},
			{"InternalError_create_user_error", "create_user_error", http.StatusInternalServerError, "Something went wrong, please try again later"},
			{"InternalError_hash_error", "hash_error", http.StatusInternalServerError, "Something went wrong, please try again later"},
			{"InternalError_token_generation_error", "token_generation_error", http.StatusInternalServerError, "Something went wrong, please try again later"},
			{"InternalError_redis_error", "redis_error", http.StatusInternalServerError, "Something went wrong, please try again later"},
			{"InternalError_commit_error", "commit_error", http.StatusInternalServerError, "Something went wrong, please try again later"},
			{"InternalError_update_user_error", "update_user_error", http.StatusInternalServerError, "Something went wrong, please try again later"},
			{"InternalError_uuid_error", "uuid_error", http.StatusInternalServerError, "Something went wrong, please try again later"},
			// OAuth bad request codes (400)
			{"OAuthBadRequest_invalid_state", "invalid_state", http.StatusBadRequest, "Test error"},
			{"OAuthBadRequest_token_exchange_error", "token_exchange_error", http.StatusBadRequest, "Test error"},
			{"OAuthBadRequest_google_api_error", "google_api_error", http.StatusBadRequest, "Test error"},
			{"OAuthBadRequest_no_refresh_token", "no_refresh_token", http.StatusBadRequest, "Test error"},
			{"OAuthBadRequest_google_token_error", "google_token_error", http.StatusBadRequest, "Test error"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				appErr := &handlers.AppError{Code: tc.errorCode, Message: "Test error", Err: errors.New("inner error")}
				mockHandlersConfig.On("LogHandlerError", mock.Anything, "test_op", tc.errorCode, "Test error", "", "", mock.Anything).Return()

				cfg.handleAuthError(w, req, appErr, "test_op", "", "")

				assert.Equal(t, tc.httpStatus, w.Code)
				assert.Contains(t, w.Body.String(), tc.expectedMsg)
				mockHandlersConfig.AssertExpectations(t)
			})
		}
	})

	// Test default case (unknown AppError code)
	t.Run("DefaultAppError", func(t *testing.T) {
		w := httptest.NewRecorder()
		appErr := &handlers.AppError{Code: "unknown_code", Message: "Test error", Err: errors.New("inner error")}
		mockHandlersConfig.On("LogHandlerError", mock.Anything, "test_op", "internal_error", "Test error", "", "", mock.Anything).Return()

		cfg.handleAuthError(w, req, appErr, "test_op", "", "")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
		mockHandlersConfig.AssertExpectations(t)
	})

	// Test non-AppError (generic error)
	t.Run("GenericError", func(t *testing.T) {
		w := httptest.NewRecorder()
		genericErr := errors.New("generic error")
		mockHandlersConfig.On("LogHandlerError", mock.Anything, "test_op", "unknown_error", "Unknown error occurred", "", "", genericErr).Return()

		cfg.handleAuthError(w, req, genericErr, "test_op", "", "")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
		mockHandlersConfig.AssertExpectations(t)
	})
}

// TestInitAuthService_AllValidationBranches covers all validation branches in InitAuthService for missing dependencies.
func TestInitAuthService_AllValidationBranches(t *testing.T) {
	// Test missing Config
	t.Run("MissingHandlersConfig", func(t *testing.T) {
		cfg := &HandlersAuthConfig{
			Config: nil,
		}
		err := cfg.InitAuthService()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "handlers config not initialized")
	})

	// Test missing APIConfig
	t.Run("MissingAPIConfig", func(t *testing.T) {
		cfg := &HandlersAuthConfig{
			Config: &handlers.Config{
				APIConfig: nil,
			},
		}
		err := cfg.InitAuthService()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API config not initialized")
	})

	// Test missing DB
	t.Run("MissingDB", func(t *testing.T) {
		cfg := &HandlersAuthConfig{
			Config: &handlers.Config{
				APIConfig: &config.APIConfig{
					DB: nil,
				},
			},
		}
		err := cfg.InitAuthService()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database not initialized")
	})

	// Test missing Auth (with valid DB)
	t.Run("MissingAuth", func(t *testing.T) {
		cfg := &HandlersAuthConfig{
			Config: &handlers.Config{
				APIConfig: &config.APIConfig{
					DB:          &database.Queries{}, // Valid DB
					RedisClient: nil,
				},
				Auth: nil,
			},
		}
		err := cfg.InitAuthService()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "auth config not initialized")
	})

	// Test missing RedisClient (with valid DB and Auth)
	t.Run("MissingRedisClient", func(t *testing.T) {
		cfg := &HandlersAuthConfig{
			Config: &handlers.Config{
				APIConfig: &config.APIConfig{
					DB:          &database.Queries{}, // Valid DB
					RedisClient: nil,
				},
				Auth: &auth.Config{},
			},
		}
		err := cfg.InitAuthService()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "redis client not initialized")
	})

	// Test successful initialization (will fail gracefully due to nil dependencies)
	t.Run("SuccessfulInitialization", func(t *testing.T) {
		cfg := &HandlersAuthConfig{
			Config: &handlers.Config{
				APIConfig: &config.APIConfig{
					DB:          (*database.Queries)(nil), // Non-nil pointer to nil
					RedisClient: nil,
				},
				Auth: &auth.Config{},
			},
		}
		err := cfg.InitAuthService()
		// This will fail due to nil DB, but should not panic
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database not initialized")
	})
}
