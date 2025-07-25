// Package config provides configuration management, validation, and provider logic for the ecom-backend project.
package config

import (
	"fmt"
	"strings"
)

// validator.go: Configuration validation logic and helpers.

const (
	uploadBackendLocal = "local"
	uploadBackendS3    = "s3"
)

// ValidatorImpl implements the Validator interface for configuration validation.
type ValidatorImpl struct{}

// NewConfigValidator creates and returns a new ConfigValidatorImpl instance.
// Ensures all required configuration values are present and valid before the application starts.
func NewConfigValidator() *ValidatorImpl {
	return &ValidatorImpl{}
}

// Validate performs comprehensive validation of the entire APIConfig.
// Checks all required fields, validates upload backend configuration, and ensures all necessary service clients are configured.
// Returns an error if any validation fails, with detailed error messages.
func (v *ValidatorImpl) Validate(config *APIConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	var errors []string

	// Validate required string fields
	if config.Port == "" {
		errors = append(errors, "PORT is required")
	}

	if config.JWTSecret == "" {
		errors = append(errors, "JWT_SECRET is required")
	}

	if config.RefreshSecret == "" {
		errors = append(errors, "REFRESH_SECRET is required")
	}

	if config.Issuer == "" {
		errors = append(errors, "ISSUER is required")
	}

	if config.Audience == "" {
		errors = append(errors, "AUDIENCE is required")
	}

	if config.CredsPath == "" {
		errors = append(errors, "GOOGLE_CREDENTIALS_PATH is required")
	}

	if config.S3Bucket == "" {
		errors = append(errors, "S3_BUCKET is required")
	}

	if config.S3Region == "" {
		errors = append(errors, "S3_REGION is required")
	}

	if config.StripeSecretKey == "" {
		errors = append(errors, "STRIPE_SECRET_KEY is required")
	}

	if config.StripeWebhookSecret == "" {
		errors = append(errors, "STRIPE_WEBHOOK_SECRET is required")
	}

	// Validate upload backend
	if config.UploadBackend != "" && config.UploadBackend != uploadBackendS3 && config.UploadBackend != uploadBackendLocal {
		errors = append(errors, "UPLOAD_BACKEND must be either 's3' or 'local'")
	}

	// Validate upload path for local backend
	if config.UploadBackend == uploadBackendLocal && config.UploadPath == "" {
		errors = append(errors, "UPLOAD_PATH is required when using local upload backend")
	}

	// Validate S3 client when using S3 backend
	if config.UploadBackend == uploadBackendS3 && config.S3Client == nil {
		errors = append(errors, "S3_CLIENT is required when using S3 upload backend")
	}

	// Validate Redis client
	if config.RedisClient == nil {
		errors = append(errors, "Redis client is required")
	}

	// Validate MongoDB client
	if config.MongoClient == nil {
		errors = append(errors, "MongoDB client is required")
	}

	if config.MongoDB == nil {
		errors = append(errors, "MongoDB database is required")
	}

	// Return combined error if any validation failed
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ValidatePartial performs validation on specific configuration fields without requiring all fields.
// Useful for testing individual components or when only certain configuration sections need to be validated.
// Provides more granular error reporting for specific validation failures.
func (v *ValidatorImpl) ValidatePartial(config *APIConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	var errors []string

	// Validate required string fields
	if config.Port == "" {
		errors = append(errors, "PORT is required")
	}

	if config.JWTSecret == "" {
		errors = append(errors, "JWT_SECRET is required")
	}

	if config.RefreshSecret == "" {
		errors = append(errors, "REFRESH_SECRET is required")
	}

	if config.Issuer == "" {
		errors = append(errors, "ISSUER is required")
	}

	if config.Audience == "" {
		errors = append(errors, "AUDIENCE is required")
	}

	if config.CredsPath == "" {
		errors = append(errors, "GOOGLE_CREDENTIALS_PATH is required")
	}

	if config.S3Bucket == "" {
		errors = append(errors, "S3_BUCKET is required")
	}

	if config.S3Region == "" {
		errors = append(errors, "S3_REGION is required")
	}

	if config.StripeSecretKey == "" {
		errors = append(errors, "STRIPE_SECRET_KEY is required")
	}

	if config.StripeWebhookSecret == "" {
		errors = append(errors, "STRIPE_WEBHOOK_SECRET is required")
	}

	// Validate upload backend
	if config.UploadBackend != "" && config.UploadBackend != uploadBackendS3 && config.UploadBackend != uploadBackendLocal {
		errors = append(errors, "UPLOAD_BACKEND must be either 's3' or 'local'")
	}

	// Return combined error if any validation failed
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
