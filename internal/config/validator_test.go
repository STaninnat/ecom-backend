// Package config provides configuration management, validation, and provider logic for the ecom-backend project.
package config

import (
	"testing"

	redismock "github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// validator_test.go: Tests for configuration validation logic and helpers.

// validAPIConfig creates a valid APIConfig for testing purposes.
func validAPIConfig() *APIConfig {
	client, _ := redismock.NewClientMock()
	return &APIConfig{
		Port:                "8080",
		JWTSecret:           "jwt",
		RefreshSecret:       "refresh",
		Issuer:              "issuer",
		Audience:            "aud",
		CredsPath:           "creds.json",
		S3Bucket:            "bucket",
		S3Region:            "region",
		StripeSecretKey:     "sk",
		StripeWebhookSecret: "wh",
		UploadBackend:       "local",
		UploadPath:          "./uploads",
		RedisClient:         client,
		MongoClient:         new(mongo.Client),   // dummy non-nil
		MongoDB:             new(mongo.Database), // dummy non-nil
	}
}

// TestValidator_AllRequiredFieldsPresent tests the validator with all required fields present.
// It verifies that the validator passes when all essential configuration fields are provided.
func TestValidator_AllRequiredFieldsPresent(t *testing.T) {
	v := NewConfigValidator()
	cfg := validAPIConfig()
	err := v.Validate(cfg)
	assert.NoError(t, err)
}

// TestValidator_MissingRequiredFields tests the validator with missing required fields.
// It verifies that the validator returns errors when essential configuration fields are missing.
func TestValidator_MissingRequiredFields(t *testing.T) {
	v := NewConfigValidator()
	cfg := &APIConfig{}
	err := v.Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PORT is required")
	assert.Contains(t, err.Error(), "JWT_SECRET is required")
	assert.Contains(t, err.Error(), "GOOGLE_CREDENTIALS_PATH is required")
}

// TestValidator_InvalidUploadBackend tests the validator with invalid upload backend.
// It verifies that the validator returns an error when an unsupported upload backend is specified.
func TestValidator_InvalidUploadBackend(t *testing.T) {
	v := NewConfigValidator()
	cfg := validAPIConfig()
	cfg.UploadBackend = "ftp"
	err := v.Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UPLOAD_BACKEND must be either 's3' or 'local'")
}

// TestValidator_MissingUploadPathForLocal tests the validator with missing upload path for local backend.
// It verifies that the validator returns an error when local backend is used without upload path.
func TestValidator_MissingUploadPathForLocal(t *testing.T) {
	v := NewConfigValidator()
	cfg := validAPIConfig()
	cfg.UploadBackend = "local"
	cfg.UploadPath = ""
	err := v.Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UPLOAD_PATH is required when using local upload backend")
}

// TestValidator_MissingS3ClientForS3Backend tests the validator with missing S3 client for S3 backend.
// It verifies that the validator returns an error when S3 backend is used without S3 client.
func TestValidator_MissingS3ClientForS3Backend(t *testing.T) {
	v := NewConfigValidator()
	cfg := validAPIConfig()
	cfg.UploadBackend = "s3"
	cfg.S3Client = nil
	err := v.Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "S3_CLIENT is required when using S3 upload backend")
}

// TestValidator_MissingRedisClient tests the validator with missing Redis client.
// It verifies that the validator returns an error when Redis client is not provided.
func TestValidator_MissingRedisClient(t *testing.T) {
	v := NewConfigValidator()
	cfg := validAPIConfig()
	cfg.RedisClient = nil
	err := v.Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Redis client is required")
}

// TestValidator_MissingMongoClients tests the validator with missing MongoDB clients.
// It verifies that the validator returns an error when MongoDB client and database are not provided.
func TestValidator_MissingMongoClients(t *testing.T) {
	v := NewConfigValidator()
	cfg := validAPIConfig()
	cfg.MongoClient = nil
	cfg.MongoDB = nil
	err := v.Validate(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MongoDB client is required")
	assert.Contains(t, err.Error(), "MongoDB database is required")
}

// TestValidator_ValidatePartial tests the partial validation functionality.
// It verifies that the validator can validate specific fields without requiring all fields to be present.
func TestValidator_ValidatePartial(t *testing.T) {
	v := NewConfigValidator()
	cfg := validAPIConfig()
	cfg.RedisClient = nil
	cfg.MongoClient = nil
	cfg.MongoDB = nil
	cfg.S3Client = nil
	err := v.ValidatePartial(cfg)
	require.NoError(t, err)

	cfg.Port = ""
	err = v.ValidatePartial(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PORT is required")
}

// TestValidator_ValidatePartial_NilConfig tests partial validation with nil configuration.
// It verifies that the validator handles nil configuration gracefully during partial validation.
func TestValidator_ValidatePartial_NilConfig(t *testing.T) {
	v := NewConfigValidator()
	err := v.ValidatePartial(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
}

// TestValidator_ValidatePartial_IndividualFieldErrors tests individual field validation errors.
// It verifies that the validator correctly identifies and reports errors for specific invalid fields.
func TestValidator_ValidatePartial_IndividualFieldErrors(t *testing.T) {
	v := NewConfigValidator()
	cfg := validAPIConfig()
	cfg.RedisClient = nil
	cfg.MongoClient = nil
	cfg.MongoDB = nil
	cfg.S3Client = nil

	// Test each required field individually
	fields := []struct {
		field   string
		setter  func(*APIConfig)
		message string
	}{
		{"JWT_SECRET", func(c *APIConfig) { c.JWTSecret = "" }, "JWT_SECRET is required"},
		{"REFRESH_SECRET", func(c *APIConfig) { c.RefreshSecret = "" }, "REFRESH_SECRET is required"},
		{"ISSUER", func(c *APIConfig) { c.Issuer = "" }, "ISSUER is required"},
		{"AUDIENCE", func(c *APIConfig) { c.Audience = "" }, "AUDIENCE is required"},
		{"GOOGLE_CREDENTIALS_PATH", func(c *APIConfig) { c.CredsPath = "" }, "GOOGLE_CREDENTIALS_PATH is required"},
		{"S3_BUCKET", func(c *APIConfig) { c.S3Bucket = "" }, "S3_BUCKET is required"},
		{"S3_REGION", func(c *APIConfig) { c.S3Region = "" }, "S3_REGION is required"},
		{"STRIPE_SECRET_KEY", func(c *APIConfig) { c.StripeSecretKey = "" }, "STRIPE_SECRET_KEY is required"},
		{"STRIPE_WEBHOOK_SECRET", func(c *APIConfig) { c.StripeWebhookSecret = "" }, "STRIPE_WEBHOOK_SECRET is required"},
	}

	for _, field := range fields {
		t.Run(field.field, func(t *testing.T) {
			cfgCopy := *cfg
			field.setter(&cfgCopy)
			err := v.ValidatePartial(&cfgCopy)
			require.Error(t, err)
			assert.Contains(t, err.Error(), field.message)
		})
	}
}

// TestValidator_ValidatePartial_UploadBackendEdgeCases tests upload backend validation edge cases.
// It verifies that the validator handles various edge cases in upload backend configuration correctly.
func TestValidator_ValidatePartial_UploadBackendEdgeCases(t *testing.T) {
	v := NewConfigValidator()
	cfg := validAPIConfig()
	cfg.RedisClient = nil
	cfg.MongoClient = nil
	cfg.MongoDB = nil
	cfg.S3Client = nil

	// Test empty upload backend (should be valid)
	cfg.UploadBackend = ""
	err := v.ValidatePartial(cfg)
	require.NoError(t, err)

	// Test "s3" upload backend (should be valid)
	cfg.UploadBackend = "s3"
	err = v.ValidatePartial(cfg)
	require.NoError(t, err)

	// Test "local" upload backend (should be valid)
	cfg.UploadBackend = "local"
	err = v.ValidatePartial(cfg)
	require.NoError(t, err)

	// Test invalid upload backend
	cfg.UploadBackend = "invalid"
	err = v.ValidatePartial(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UPLOAD_BACKEND must be either 's3' or 'local'")
}
