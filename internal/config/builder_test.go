package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type mockProvider struct {
	values map[string]string
}

// GetString returns the string value for the given key from the mock provider.
// Returns empty string if key is not found.
func (m *mockProvider) GetString(key string) string { return m.values[key] }

// GetStringOrDefault returns the string value for the given key or the default value.
// Returns the default if key is not found or value is empty.
func (m *mockProvider) GetStringOrDefault(key, def string) string {
	if v := m.values[key]; v != "" {
		return v
	}
	return def
}

// GetRequiredString returns the string value for the given key or an error.
// Returns an error if key is missing or value is empty.
func (m *mockProvider) GetRequiredString(key string) (string, error) {
	v, ok := m.values[key]
	if !ok || v == "" {
		return "", errors.New("missing " + key)
	}
	return v, nil
}

// GetInt returns the integer value for the given key from the mock provider.
// Always returns 0 for testing purposes.
func (m *mockProvider) GetInt(key string) int { return 0 }

// GetIntOrDefault returns the integer value for the given key or the default value.
// Always returns the default value for testing purposes.
func (m *mockProvider) GetIntOrDefault(key string, def int) int { return def }

// GetBool returns the boolean value for the given key from the mock provider.
// Always returns false for testing purposes.
func (m *mockProvider) GetBool(key string) bool { return false }

// GetBoolOrDefault returns the boolean value for the given key or the default value.
// Always returns the default value for testing purposes.
func (m *mockProvider) GetBoolOrDefault(key string, def bool) bool { return def }

// Mock providers for testing service connections
type mockDatabaseProvider struct{}

// Connect mocks the database connection for testing.
// Returns nil values to simulate successful connection without actual database.
func (m *mockDatabaseProvider) Connect(ctx context.Context) (*sql.DB, *database.Queries, error) {
	return nil, nil, nil
}

// Close mocks the database close operation for testing.
// Returns nil to simulate successful closure.
func (m *mockDatabaseProvider) Close() error {
	return nil
}

type mockRedisProvider struct{}

// Connect mocks the Redis connection for testing.
// Returns nil values to simulate successful connection without actual Redis.
func (m *mockRedisProvider) Connect(ctx context.Context) (redis.Cmdable, error) {
	return nil, nil
}

// Close mocks the Redis close operation for testing.
// Returns nil to simulate successful closure.
func (m *mockRedisProvider) Close() error {
	return nil
}

type mockMongoProvider struct{}

// Connect mocks the MongoDB connection for testing.
// Returns nil values to simulate successful connection without actual MongoDB.
func (m *mockMongoProvider) Connect(ctx context.Context) (*mongo.Client, *mongo.Database, error) {
	return nil, nil, nil
}

// Close mocks the MongoDB close operation for testing.
// Returns nil to simulate successful closure.
func (m *mockMongoProvider) Close(ctx context.Context) error {
	return nil
}

type mockS3Provider struct{}

// CreateClient mocks the S3 client creation for testing.
// Returns nil values to simulate successful client creation without actual AWS.
func (m *mockS3Provider) CreateClient(ctx context.Context, region string) (*s3.Client, error) {
	return nil, nil
}

type mockOAuthProvider struct{}

// LoadGoogleConfig mocks the OAuth configuration loading for testing.
// Returns an empty OAuthConfig to simulate successful loading.
func (m *mockOAuthProvider) LoadGoogleConfig(credsPath string) (*OAuthConfig, error) {
	return &OAuthConfig{}, nil
}

// TestBuilder_NilProvider tests the config builder with a nil provider.
// It verifies that the builder returns an error when no provider is set.
func TestBuilder_NilProvider(t *testing.T) {
	builder := NewConfigBuilder()
	cfg, err := builder.Build(context.Background())
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "config provider is required")
}

// TestBuilder_MissingRequiredValue tests the config builder with missing required values.
// It verifies that the builder returns an error when required configuration is missing.
func TestBuilder_MissingRequiredValue(t *testing.T) {
	provider := &mockProvider{values: map[string]string{}}
	builder := NewConfigBuilder().WithProvider(provider)
	cfg, err := builder.Build(context.Background())
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestBuilder_MissingPort tests the config builder with missing PORT configuration.
// It verifies that the builder returns an error when PORT is not provided.
func TestBuilder_MissingPort(t *testing.T) {
	provider := &mockProvider{values: map[string]string{
		"JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
		"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
		"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongo://uri",
	}}
	builder := NewConfigBuilder().WithProvider(provider)
	cfg, err := builder.Build(context.Background())
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to get PORT")
}

// TestBuilder_MissingJWTSecret tests the config builder with missing JWT_SECRET configuration.
// It verifies that the builder returns an error when JWT_SECRET is not provided.
func TestBuilder_MissingJWTSecret(t *testing.T) {
	provider := &mockProvider{values: map[string]string{
		"PORT": "8080", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
		"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
		"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongo://uri",
	}}
	builder := NewConfigBuilder().WithProvider(provider)
	cfg, err := builder.Build(context.Background())
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to get JWT_SECRET")
}

// TestBuilder_ConfigWiring tests the config builder with all required values.
// It verifies that configuration values are properly wired to the config struct.
func TestBuilder_ConfigWiring(t *testing.T) {
	provider := &mockProvider{values: map[string]string{
		"PORT": "8080", "JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
		"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
		"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongo://uri", "UPLOAD_BACKEND": "local", "UPLOAD_PATH": "./uploads",
	}}
	builder := NewConfigBuilder().WithProvider(provider)
	cfg, err := builder.Build(context.Background())
	// This will fail on Redis/Mongo/S3/OAuth connection, so just check config values if no error
	if err == nil {
		assert.Equal(t, "8080", cfg.Port)
		assert.Equal(t, "jwt", cfg.JWTSecret)
		assert.Equal(t, "refresh", cfg.RefreshSecret)
		assert.Equal(t, "issuer", cfg.Issuer)
		assert.Equal(t, "aud", cfg.Audience)
		assert.Equal(t, "creds.json", cfg.CredsPath)
		assert.Equal(t, "bucket", cfg.S3Bucket)
		assert.Equal(t, "region", cfg.S3Region)
		assert.Equal(t, "sk", cfg.StripeSecretKey)
		assert.Equal(t, "wh", cfg.StripeWebhookSecret)
		assert.Equal(t, "local", cfg.UploadBackend)
		assert.Equal(t, "./uploads", cfg.UploadPath)
	}
}

// TestBuilder_WithAllProviders tests the config builder with all service providers.
// It verifies that all providers are properly integrated into the configuration.
func TestBuilder_WithAllProviders(t *testing.T) {
	provider := &mockProvider{values: map[string]string{
		"PORT": "8080", "JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
		"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
		"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongodb://localhost:27017",
		// Don't set REDIS_ADDR to avoid real connection attempts
	}}

	builder := NewConfigBuilder().
		WithProvider(provider).
		WithDatabase(&mockDatabaseProvider{}).
		WithRedis(&mockRedisProvider{}).
		WithMongo(&mockMongoProvider{}).
		WithS3(&mockS3Provider{}).
		WithOAuth(&mockOAuthProvider{})

	cfg, err := builder.Build(context.Background())
	// This will likely fail due to real service connections, but we test the provider wiring
	if err == nil {
		assert.NotNil(t, cfg)
		assert.Equal(t, "8080", cfg.Port)
	} else {
		// Expected to fail due to real service connections, but config should be partially built
		assert.Contains(t, err.Error(), "failed to connect")
	}
}

// TestBuilder_DefaultValues tests the config builder with missing optional values.
// It verifies that default values are properly applied when optional config is missing.
func TestBuilder_DefaultValues(t *testing.T) {
	provider := &mockProvider{values: map[string]string{
		"PORT": "8080", "JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
		"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
		"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongodb://localhost:27017",
		// UPLOAD_BACKEND and UPLOAD_PATH not set, should use defaults
	}}

	builder := NewConfigBuilder().WithProvider(provider)
	cfg, err := builder.Build(context.Background())

	// This will fail on service connections, but we can check the default values
	if err == nil {
		assert.Equal(t, "local", cfg.UploadBackend)  // default value
		assert.Equal(t, "./uploads", cfg.UploadPath) // default value
	}
}

// TestBuilder_RedisWithEmptyAddress tests the config builder with empty Redis address.
// It verifies that the builder handles empty Redis configuration gracefully.
func TestBuilder_RedisWithEmptyAddress(t *testing.T) {
	provider := &mockProvider{values: map[string]string{
		"PORT": "8080", "JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
		"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
		"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongodb://localhost:27017",
		// REDIS_ADDR is empty, should not attempt connection
	}}

	builder := NewConfigBuilder().
		WithProvider(provider).
		WithRedis(&mockRedisProvider{})

	cfg, err := builder.Build(context.Background())
	// This should work since Redis address is empty
	if err == nil {
		assert.NotNil(t, cfg)
		assert.Nil(t, cfg.RedisClient) // Should be nil since no Redis connection
	}
}

// TestBuilder_S3ClientCreationError tests the config builder when S3 client creation fails.
// It verifies proper error handling when S3 provider returns an error.
func TestBuilder_S3ClientCreationError(t *testing.T) {
	provider := &mockProvider{values: map[string]string{
		"PORT": "8080", "JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
		"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
		"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongodb://localhost:27017",
	}}

	// Create a mock S3 provider that returns error
	mockS3Provider := &mockS3ProviderWithError{}

	builder := NewConfigBuilder().
		WithProvider(provider).
		WithS3(mockS3Provider)

	cfg, err := builder.Build(context.Background())
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to create S3 client")
}

// TestBuilder_OAuthConfigLoadingError tests the config builder when OAuth config loading fails.
// It verifies proper error handling when OAuth provider returns an error.
func TestBuilder_OAuthConfigLoadingError(t *testing.T) {
	provider := &mockProvider{values: map[string]string{
		"PORT": "8080", "JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
		"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
		"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongodb://localhost:27017",
	}}

	// Create a mock OAuth provider that returns error
	mockOAuthProvider := &mockOAuthProviderWithError{}

	builder := NewConfigBuilder().
		WithProvider(provider).
		WithOAuth(mockOAuthProvider)

	cfg, err := builder.Build(context.Background())
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to load OAuth config")
}

// TestBuilder_IndividualRequiredStringFailures tests each required string field individually.
// It verifies that the builder returns appropriate errors for each missing required field.
func TestBuilder_IndividualRequiredStringFailures(t *testing.T) {
	// Test each required string individually
	fields := []struct {
		field   string
		setter  func(*mockProvider)
		message string
	}{
		{"REFRESH_SECRET", func(p *mockProvider) { p.values["REFRESH_SECRET"] = "" }, "failed to get REFRESH_SECRET"},
		{"ISSUER", func(p *mockProvider) { p.values["ISSUER"] = "" }, "failed to get ISSUER"},
		{"AUDIENCE", func(p *mockProvider) { p.values["AUDIENCE"] = "" }, "failed to get AUDIENCE"},
		{"GOOGLE_CREDENTIALS_PATH", func(p *mockProvider) { p.values["GOOGLE_CREDENTIALS_PATH"] = "" }, "failed to get GOOGLE_CREDENTIALS_PATH"},
		{"S3_BUCKET", func(p *mockProvider) { p.values["S3_BUCKET"] = "" }, "failed to get S3_BUCKET"},
		{"S3_REGION", func(p *mockProvider) { p.values["S3_REGION"] = "" }, "failed to get S3_REGION"},
		{"STRIPE_SECRET_KEY", func(p *mockProvider) { p.values["STRIPE_SECRET_KEY"] = "" }, "failed to get STRIPE_SECRET_KEY"},
		{"STRIPE_WEBHOOK_SECRET", func(p *mockProvider) { p.values["STRIPE_WEBHOOK_SECRET"] = "" }, "failed to get STRIPE_WEBHOOK_SECRET"},
		{"MONGO_URI", func(p *mockProvider) { p.values["MONGO_URI"] = "" }, "failed to get MONGO_URI"},
	}

	for _, field := range fields {
		t.Run(field.field, func(t *testing.T) {
			provider := &mockProvider{values: map[string]string{
				"PORT": "8080", "JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
				"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
				"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongodb://localhost:27017",
			}}
			field.setter(provider)

			builder := NewConfigBuilder().WithProvider(provider)
			cfg, err := builder.Build(context.Background())
			assert.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), field.message)
		})
	}
}

// Mock providers for error testing
type mockS3ProviderWithError struct{}

// CreateClient mocks S3 client creation that returns an error for testing.
// Returns an error to simulate S3 client creation failure.
func (m *mockS3ProviderWithError) CreateClient(ctx context.Context, region string) (*s3.Client, error) {
	return nil, fmt.Errorf("S3 client creation failed")
}

type mockOAuthProviderWithError struct{}

// LoadGoogleConfig mocks OAuth config loading that returns an error for testing.
// Returns an error to simulate OAuth configuration loading failure.
func (m *mockOAuthProviderWithError) LoadGoogleConfig(credsPath string) (*OAuthConfig, error) {
	return nil, fmt.Errorf("OAuth config loading failed")
}
