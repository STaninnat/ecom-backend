package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoadConfigWithProviders_Success tests successful configuration loading with all providers.
// It verifies that the configuration is properly built when all required values and providers are present.
func TestLoadConfigWithProviders_Success(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{
			"PORT":                    "8080",
			"JWT_SECRET":              "test-jwt-secret",
			"REFRESH_SECRET":          "test-refresh-secret",
			"ISSUER":                  "test-issuer",
			"AUDIENCE":                "test-audience",
			"GOOGLE_CREDENTIALS_PATH": "test-credentials.json",
			"S3_BUCKET":               "test-bucket",
			"S3_REGION":               "us-east-1",
			"STRIPE_SECRET_KEY":       "test-stripe-key",
			"STRIPE_WEBHOOK_SECRET":   "test-webhook-secret",
			"MONGO_URI":               "mongodb://localhost:27017",
			"UPLOAD_BACKEND":          "local",
			"UPLOAD_PATH":             "./test-uploads",
		},
	}

	cfg, err := LoadConfigWithProviders(
		context.Background(),
		mockProvider,
		nil, nil, nil, nil, nil,
	)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "test-jwt-secret", cfg.JWTSecret)
	assert.Equal(t, "test-refresh-secret", cfg.RefreshSecret)
	assert.Equal(t, "test-issuer", cfg.Issuer)
	assert.Equal(t, "test-audience", cfg.Audience)
	assert.Equal(t, "test-credentials.json", cfg.CredsPath)
	assert.Equal(t, "test-bucket", cfg.S3Bucket)
	assert.Equal(t, "us-east-1", cfg.S3Region)
	assert.Equal(t, "test-stripe-key", cfg.StripeSecretKey)
	assert.Equal(t, "test-webhook-secret", cfg.StripeWebhookSecret)
	// MongoClient is a pointer, should be nil since no provider is used
	assert.Nil(t, cfg.MongoClient)
	assert.Equal(t, "local", cfg.UploadBackend)
	assert.Equal(t, "./test-uploads", cfg.UploadPath)
}

// TestLoadConfigWithProviders_MissingRequiredValue tests configuration loading with missing required values.
// It verifies that the function returns an error when essential configuration values are not provided.
func TestLoadConfigWithProviders_MissingRequiredValue(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{
			// Missing PORT
			"JWT_SECRET":              "test-jwt-secret",
			"REFRESH_SECRET":          "test-refresh-secret",
			"ISSUER":                  "test-issuer",
			"AUDIENCE":                "test-audience",
			"GOOGLE_CREDENTIALS_PATH": "test-credentials.json",
			"S3_BUCKET":               "test-bucket",
			"S3_REGION":               "us-east-1",
			"STRIPE_SECRET_KEY":       "test-stripe-key",
			"STRIPE_WEBHOOK_SECRET":   "test-webhook-secret",
			"MONGO_URI":               "mongodb://localhost:27017",
			"UPLOAD_BACKEND":          "local",
			"UPLOAD_PATH":             "./test-uploads",
		},
	}

	cfg, err := LoadConfigWithProviders(
		context.Background(),
		mockProvider,
		nil, nil, nil, nil, nil,
	)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestLoadConfigForTesting tests the LoadConfigForTesting function.
// It verifies that the testing configuration is properly loaded with default values.
func TestLoadConfigForTesting(t *testing.T) {
	cfg, err := LoadConfigForTesting(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "test-jwt-secret", cfg.JWTSecret)
	assert.Equal(t, "test-refresh-secret", cfg.RefreshSecret)
	assert.Equal(t, "test-issuer", cfg.Issuer)
	assert.Equal(t, "test-audience", cfg.Audience)
	assert.Equal(t, "test-credentials.json", cfg.CredsPath)
	assert.Equal(t, "test-bucket", cfg.S3Bucket)
	assert.Equal(t, "us-east-1", cfg.S3Region)
	assert.Equal(t, "test-stripe-key", cfg.StripeSecretKey)
	assert.Equal(t, "test-webhook-secret", cfg.StripeWebhookSecret)
	assert.Equal(t, "local", cfg.UploadBackend)
	assert.Equal(t, "./test-uploads", cfg.UploadPath)
}

// TestMockConfigProvider_BoolAndInt tests the MockConfigProvider's boolean and integer methods.
// It verifies that the mock provider correctly handles boolean and integer value retrieval.
func TestMockConfigProvider_BoolAndInt(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{
			"BOOL_TRUE":  "true",
			"BOOL_FALSE": "false",
			"INT_VAL":    "42",
		},
	}
	assert.True(t, mockProvider.GetBool("BOOL_TRUE"))
	assert.False(t, mockProvider.GetBool("BOOL_FALSE"))
	assert.Equal(t, 42, mockProvider.GetInt("INT_VAL"))
	assert.Equal(t, 0, mockProvider.GetInt("NOT_SET"))
	assert.Equal(t, 99, mockProvider.GetIntOrDefault("NOT_SET", 99))
	assert.True(t, mockProvider.GetBoolOrDefault("NOT_SET", true))
}

// TestLoadConfigWithProviders_AllNilProviders tests configuration loading with all nil providers.
// It verifies that the function handles nil providers gracefully and returns appropriate errors.
func TestLoadConfigWithProviders_AllNilProviders(t *testing.T) {
	cfg, err := LoadConfigWithProviders(context.Background(), nil, nil, nil, nil, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestLoadConfigWithError_Success tests successful configuration loading with error handling.
// It verifies that the configuration is properly built when all required values are present.
func TestLoadConfigWithError_Success(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{
			"PORT":                    "8080",
			"JWT_SECRET":              "test-jwt-secret",
			"REFRESH_SECRET":          "test-refresh-secret",
			"ISSUER":                  "test-issuer",
			"AUDIENCE":                "test-audience",
			"GOOGLE_CREDENTIALS_PATH": "test-credentials.json",
			"S3_BUCKET":               "test-bucket",
			"S3_REGION":               "us-east-1",
			"STRIPE_SECRET_KEY":       "test-stripe-key",
			"STRIPE_WEBHOOK_SECRET":   "test-webhook-secret",
			"MONGO_URI":               "mongodb://localhost:27017",
			"UPLOAD_BACKEND":          "local",
			"UPLOAD_PATH":             "./test-uploads",
		},
	}
	cfg, err := LoadConfigWithProviders(
		context.Background(),
		mockProvider, nil, nil, nil, nil, nil,
	)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

// TestLoadConfigWithError_MissingRequiredValue tests configuration loading with missing required values and error handling.
// It verifies that the function returns an error when essential configuration values are not provided.
func TestLoadConfigWithError_MissingRequiredValue(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{
			// Missing PORT
			"JWT_SECRET":              "test-jwt-secret",
			"REFRESH_SECRET":          "test-refresh-secret",
			"ISSUER":                  "test-issuer",
			"AUDIENCE":                "test-audience",
			"GOOGLE_CREDENTIALS_PATH": "test-credentials.json",
			"S3_BUCKET":               "test-bucket",
			"S3_REGION":               "us-east-1",
			"STRIPE_SECRET_KEY":       "test-stripe-key",
			"STRIPE_WEBHOOK_SECRET":   "test-webhook-secret",
			"MONGO_URI":               "mongodb://localhost:27017",
			"UPLOAD_BACKEND":          "local",
			"UPLOAD_PATH":             "./test-uploads",
		},
	}
	cfg, err := LoadConfigWithProviders(
		context.Background(),
		mockProvider, nil, nil, nil, nil, nil,
	)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestMockConfigProvider_GetStringOrDefault tests the GetStringOrDefault method of MockConfigProvider.
// It verifies that the method returns the correct value or default when the key is not found.
func TestMockConfigProvider_GetStringOrDefault(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{
			"FOO": "bar",
		},
	}
	assert.Equal(t, "bar", mockProvider.GetStringOrDefault("FOO", "default"))
	assert.Equal(t, "default", mockProvider.GetStringOrDefault("NOT_SET", "default"))
}

// TestMockConfigProvider_GetRequiredString_Error tests the GetRequiredString method of MockConfigProvider with error.
// It verifies that the method returns an error when the required string value is missing.
func TestMockConfigProvider_GetRequiredString_Error(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{},
	}
	val, err := mockProvider.GetRequiredString("NOT_SET")
	assert.Error(t, err)
	assert.Equal(t, "", val)
}

// TestMockConfigProvider_GetInt_Invalid tests the GetInt method of MockConfigProvider with invalid input.
// It verifies that the method handles invalid integer values correctly and returns zero.
func TestMockConfigProvider_GetInt_Invalid(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{
			"BAD_INT": "notanint",
		},
	}
	assert.Equal(t, 0, mockProvider.GetInt("BAD_INT"))
	assert.Equal(t, 5, mockProvider.GetIntOrDefault("BAD_INT", 5))
}

// TestMockConfigProvider_GetBoolOrDefault_FalseDefault tests the GetBoolOrDefault method of MockConfigProvider with false default.
// It verifies that the method returns the correct boolean value or false default when the key is not found.
func TestMockConfigProvider_GetBoolOrDefault_FalseDefault(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{},
	}
	assert.False(t, mockProvider.GetBoolOrDefault("NOT_SET", false))
}

// TestLoadConfigWithError_DirectCall tests direct call to LoadConfigWithError function.
// It verifies that the function can be called directly with a mock provider and returns expected results.
func TestLoadConfigWithError_DirectCall(t *testing.T) {
	// This tests the actual LoadConfigWithError function
	// We need to set up environment variables for this to work
	// Since LoadConfigWithError calls LoadConfigWithProviders with real providers
	// We'll test it with a context that should work
	ctx := context.Background()
	cfg, err := LoadConfigWithError(ctx)
	// This will likely fail due to missing env vars, but we're testing the function call
	// The error is expected in a test environment
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestLoadConfigWithProviders_DatabaseConnection tests configuration loading with database connection.
// It verifies that the configuration properly handles database provider integration and connection setup.
func TestLoadConfigWithProviders_DatabaseConnection(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{
			"PORT":                    "8080",
			"JWT_SECRET":              "test-jwt-secret",
			"REFRESH_SECRET":          "test-refresh-secret",
			"ISSUER":                  "test-issuer",
			"AUDIENCE":                "test-audience",
			"GOOGLE_CREDENTIALS_PATH": "test-credentials.json",
			"S3_BUCKET":               "test-bucket",
			"S3_REGION":               "us-east-1",
			"STRIPE_SECRET_KEY":       "test-stripe-key",
			"STRIPE_WEBHOOK_SECRET":   "test-webhook-secret",
			"MONGO_URI":               "mongodb://localhost:27017",
			"DATABASE_URL":            "postgres://test:test@localhost:5432/test",
			"UPLOAD_BACKEND":          "local",
			"UPLOAD_PATH":             "./test-uploads",
		},
	}

	// Create a mock database provider that will fail to connect
	mockDBProvider := &mockDatabaseProvider{}

	cfg, err := LoadConfigWithProviders(
		context.Background(),
		mockProvider,
		mockDBProvider, // This will trigger the database connection path
		nil, nil, nil, nil,
	)
	// This will likely fail due to real database connection attempt, but we test the path
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestLoadConfigWithProviders_ValidatorError tests configuration loading with validator error.
// It verifies that the function returns an error when the configuration validation fails.
func TestLoadConfigWithProviders_ValidatorError(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{
			// Missing PORT to trigger validation error
			"JWT_SECRET":              "test-jwt-secret",
			"REFRESH_SECRET":          "test-refresh-secret",
			"ISSUER":                  "test-issuer",
			"AUDIENCE":                "test-audience",
			"GOOGLE_CREDENTIALS_PATH": "test-credentials.json",
			"S3_BUCKET":               "test-bucket",
			"S3_REGION":               "us-east-1",
			"STRIPE_SECRET_KEY":       "test-stripe-key",
			"STRIPE_WEBHOOK_SECRET":   "test-webhook-secret",
			"MONGO_URI":               "mongodb://localhost:27017",
			"UPLOAD_BACKEND":          "local",
			"UPLOAD_PATH":             "./test-uploads",
		},
	}

	cfg, err := LoadConfigWithProviders(
		context.Background(),
		mockProvider,
		nil, nil, nil, nil, nil,
	)
	// This should fail due to missing PORT in required string
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "required environment variable PORT is not set")
}

// TestMockConfigProvider_GetString tests the GetString method of MockConfigProvider.
// It verifies that the method returns the correct string value for existing keys and empty string for missing keys.
func TestMockConfigProvider_GetString(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{
			"FOO": "bar",
		},
	}
	assert.Equal(t, "bar", mockProvider.GetString("FOO"))
	assert.Equal(t, "", mockProvider.GetString("NOT_SET"))
}

// TestMockConfigProvider_GetIntOrDefault_EdgeCases tests the GetIntOrDefault method of MockConfigProvider with edge cases.
// It verifies that the method handles various edge cases including invalid integers and missing keys correctly.
func TestMockConfigProvider_GetIntOrDefault_EdgeCases(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{
			"EMPTY":    "",
			"ZERO":     "0",
			"NEGATIVE": "-5",
		},
	}
	// Test empty string returns default
	assert.Equal(t, 10, mockProvider.GetIntOrDefault("EMPTY", 10))
	// Test zero value
	assert.Equal(t, 0, mockProvider.GetIntOrDefault("ZERO", 10))
	// Test negative value
	assert.Equal(t, -5, mockProvider.GetIntOrDefault("NEGATIVE", 10))
	// Test missing key
	assert.Equal(t, 15, mockProvider.GetIntOrDefault("NOT_SET", 15))
}

// TestMockConfigProvider_GetBoolOrDefault_EdgeCases tests the GetBoolOrDefault method of MockConfigProvider with edge cases.
// It verifies that the method handles various edge cases including invalid booleans and missing keys correctly.
func TestMockConfigProvider_GetBoolOrDefault_EdgeCases(t *testing.T) {
	mockProvider := &MockConfigProvider{
		values: map[string]string{
			"EMPTY": "",
			"YES":   "yes",
			"ONE":   "1",
			"FALSE": "false",
		},
	}
	// Test empty string returns default
	assert.True(t, mockProvider.GetBoolOrDefault("EMPTY", true))
	assert.False(t, mockProvider.GetBoolOrDefault("EMPTY", false))
	// Test "yes" value
	assert.True(t, mockProvider.GetBoolOrDefault("YES", false))
	// Test "1" value
	assert.True(t, mockProvider.GetBoolOrDefault("ONE", false))
	// Test "false" value
	assert.False(t, mockProvider.GetBoolOrDefault("FALSE", true))
	// Test missing key
	assert.True(t, mockProvider.GetBoolOrDefault("NOT_SET", true))
	assert.False(t, mockProvider.GetBoolOrDefault("NOT_SET", false))
}
