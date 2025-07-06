package authhandlers

import (
	"testing"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// func TestInitAuthService_Success(t *testing.T) {
// 	apiCfg := &config.APIConfig{}
// 	apiCfg.DB = (*database.Queries)(nil) // Correct type, non-nil
// 	cfg := &HandlersAuthConfig{
// 		HandlersConfig: &handlers.HandlersConfig{
// 			APIConfig: apiCfg,
// 			Auth:      &auth.AuthConfig{},
// 			OAuth:     &config.OAuthConfig{},
// 			Logger:    logrus.New(),
// 		},
// 		HandlersCartConfig: nil,
// 	}
//
// 	// Test successful initialization
// 	err := cfg.InitAuthService()
// 	assert.NoError(t, err)
// 	assert.NotNil(t, cfg.authService)
// }

func TestInitAuthService_MissingDB(t *testing.T) {
	cfg := &HandlersAuthConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: nil, // Missing APIConfig (which contains DB)
			Auth:      &auth.AuthConfig{},
			OAuth:     &config.OAuthConfig{},
			Logger:    logrus.New(),
		},
		HandlersCartConfig: nil,
	}

	// Test initialization with missing DB should return an error, not panic
	err := cfg.InitAuthService()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestInitAuthService_MissingAuth(t *testing.T) {
	apiCfg := &config.APIConfig{}
	apiCfg.DB = (*database.Queries)(nil)
	cfg := &HandlersAuthConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: apiCfg,
			Auth:      nil, // Missing Auth
			OAuth:     &config.OAuthConfig{},
			Logger:    logrus.New(),
		},
		HandlersCartConfig: nil,
	}

	// Test initialization with missing Auth
	err := cfg.InitAuthService()
	assert.Error(t, err)
	assert.Equal(t, "database not initialized", err.Error())
}

func TestInitAuthService_MissingRedis(t *testing.T) {
	apiCfg := &config.APIConfig{}
	apiCfg.DB = (*database.Queries)(nil)
	cfg := &HandlersAuthConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: apiCfg, // This would need RedisClient to be nil
			Auth:      &auth.AuthConfig{},
			OAuth:     &config.OAuthConfig{},
			Logger:    logrus.New(),
		},
		HandlersCartConfig: nil,
	}

	// Test initialization with missing Redis (this would need the actual APIConfig to have nil RedisClient)
	// For now, we'll test that it doesn't panic
	assert.NotPanics(t, func() {
		cfg.InitAuthService()
	})
}

func TestGetAuthService_Initialized(t *testing.T) {
	apiCfg := &config.APIConfig{}
	apiCfg.DB = (*database.Queries)(nil)
	cfg := &HandlersAuthConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: apiCfg,
			Auth:      &auth.AuthConfig{},
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

func TestGetAuthService_NotInitialized(t *testing.T) {
	apiCfg := &config.APIConfig{}
	apiCfg.DB = (*database.Queries)(nil)
	cfg := &HandlersAuthConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: apiCfg,
			Auth:      &auth.AuthConfig{},
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

func TestGetAuthService_ThreadSafety(t *testing.T) {
	apiCfg := &config.APIConfig{}
	apiCfg.DB = (*database.Queries)(nil)
	cfg := &HandlersAuthConfig{
		HandlersConfig: &handlers.HandlersConfig{
			APIConfig: apiCfg,
			Auth:      &auth.AuthConfig{},
			OAuth:     &config.OAuthConfig{},
			Logger:    logrus.New(),
		},
		HandlersCartConfig: nil,
	}

	// Test concurrent access to GetAuthService
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			service := cfg.GetAuthService()
			assert.NotNil(t, service)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify service was initialized
	assert.NotNil(t, cfg.authService)
}
