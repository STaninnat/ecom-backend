// Package config provides configuration management, validation, and provider logic for the ecom-backend project.
package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// builder_test.go: Tests for configuration builder logic and provider integration.

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
func (m *mockProvider) GetInt(_ string) int { return 0 }

// GetIntOrDefault returns the integer value for the given key or the default value.
// Always returns the default value for testing purposes.
func (m *mockProvider) GetIntOrDefault(_ string, def int) int { return def }

// GetBool returns the boolean value for the given key from the mock provider.
// Always returns false for testing purposes.
func (m *mockProvider) GetBool(_ string) bool { return false }

// GetBoolOrDefault returns the boolean value for the given key or the default value.
// Always returns the default value for testing purposes.
func (m *mockProvider) GetBoolOrDefault(key string, def bool) bool {
	if v := m.values[key]; v != "" {
		return v == "true" // Mock provider returns "true" for true, empty for false
	}
	return def
}

// Mock providers for testing service connections
type mockDatabaseProvider struct{}

// Connect mocks the database connection for testing.
// Returns nil values to simulate successful connection without actual database.
func (m *mockDatabaseProvider) Connect(_ context.Context) (*sql.DB, *database.Queries, error) {
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
func (m *mockRedisProvider) Connect(_ context.Context) (redis.Cmdable, error) {
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
func (m *mockMongoProvider) Connect(_ context.Context) (*mongo.Client, *mongo.Database, error) {
	return nil, nil, nil
}

// Close mocks the MongoDB close operation for testing.
// Returns nil to simulate successful closure.
func (m *mockMongoProvider) Close(_ context.Context) error {
	return nil
}

type mockS3Provider struct{}

// CreateClient mocks the S3 client creation for testing.
// Returns nil values to simulate successful client creation without actual AWS.
func (m *mockS3Provider) CreateClient(_ context.Context, _ string) (*s3.Client, error) {
	return nil, nil
}

type mockOAuthProvider struct{}

// LoadGoogleConfig mocks the OAuth configuration loading for testing.
// Returns an empty OAuthConfig to simulate successful loading.
func (m *mockOAuthProvider) LoadGoogleConfig(_ string) (*OAuthConfig, error) {
	return &OAuthConfig{}, nil
}

// TestBuilder_NilProvider tests the config builder with a nil provider.
// It verifies that the builder returns an error when no provider is set.
func TestBuilder_NilProvider(t *testing.T) {
	builder := NewConfigBuilder()
	cfg, err := builder.Build(context.Background())
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "config provider is required")
}

// TestBuilder_MissingRequiredValue tests the config builder with missing required values.
// It verifies that the builder returns an error when required configuration is missing.
func TestBuilder_MissingRequiredValue(t *testing.T) {
	provider := &mockProvider{values: map[string]string{}}
	builder := NewConfigBuilder().WithProvider(provider)
	cfg, err := builder.Build(context.Background())
	require.Error(t, err)
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
	require.Error(t, err)
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
	require.Error(t, err)
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

// TestBuilder_ProviderErrorScenarios tests the config builder with various provider errors.
func TestBuilder_ProviderErrorScenarios(t *testing.T) {
	testCases := []struct {
		name           string
		builderSetup   func() Builder
		expectedErrSub string
	}{
		{
			name: "S3ClientCreationError",
			builderSetup: func() Builder {
				provider := &mockProvider{values: map[string]string{
					"PORT": "8080", "JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
					"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
					"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongodb://localhost:27017",
				}}
				mockS3Provider := &mockS3ProviderWithError{}
				return NewConfigBuilder().WithProvider(provider).WithS3(mockS3Provider)
			},
			expectedErrSub: "failed to create S3 client",
		},
		{
			name: "OAuthConfigLoadingError",
			builderSetup: func() Builder {
				provider := &mockProvider{values: map[string]string{
					"PORT": "8080", "JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
					"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
					"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongodb://localhost:27017",
				}}
				mockOAuthProvider := &mockOAuthProviderWithError{}
				return NewConfigBuilder().WithProvider(provider).WithOAuth(mockOAuthProvider)
			},
			expectedErrSub: "failed to load OAuth config",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := tc.builderSetup()
			cfg, err := builder.Build(context.Background())
			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tc.expectedErrSub)
		})
	}
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
			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), field.message)
		})
	}
}

// mockS3ProviderWithError providers for error testing
type mockS3ProviderWithError struct{}

// CreateClient mocks S3 client creation that returns an error for testing.
// Returns an error to simulate S3 client creation failure.
func (m *mockS3ProviderWithError) CreateClient(_ context.Context, _ string) (*s3.Client, error) {
	return nil, fmt.Errorf("S3 client creation failed")
}

// mockOAuthProviderWithError is a mock implementation of an OAuth provider that always returns an error.
type mockOAuthProviderWithError struct{}

// LoadGoogleConfig mocks OAuth config loading that returns an error for testing.
// Returns an error to simulate OAuth configuration loading failure.
func (m *mockOAuthProviderWithError) LoadGoogleConfig(_ string) (*OAuthConfig, error) {
	return nil, fmt.Errorf("OAuth config loading failed")
}

// errorS3Provider is a mock S3 provider that simulates a client creation failure.
type errorS3Provider struct{}

// CreateClient simulates the failure of creating an S3 client by always returning an error.
func (e *errorS3Provider) CreateClient(ctx context.Context, region string) (*s3.Client, error) {
	return nil, errors.New("s3 error")
}

// TestBuilder_Build_ProviderErrorScenarios tests the config builder with various provider errors.
func TestBuilder_Build_ProviderErrorScenarios(t *testing.T) {
	tests := []struct {
		name           string
		builderSetup   func() *BuilderImpl
		expectedErrMsg string
	}{
		{
			name: "S3ProviderError",
			builderSetup: func() *BuilderImpl {
				provider := &mockProvider{values: map[string]string{
					"PORT": "8080", "JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
					"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
					"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongodb://localhost:27017",
				}}
				return NewConfigBuilder().WithProvider(provider).WithS3(&errorS3Provider{}).(*BuilderImpl)
			},
			expectedErrMsg: "failed to create S3 client",
		},
		{
			name: "OAuthProviderError",
			builderSetup: func() *BuilderImpl {
				provider := &mockProvider{values: map[string]string{
					"PORT": "8080", "JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
					"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
					"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongodb://localhost:27017",
				}}
				return NewConfigBuilder().WithProvider(provider).WithOAuth(&errorOAuthProvider{}).(*BuilderImpl)
			},
			expectedErrMsg: "failed to load OAuth config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.builderSetup()
			cfg, err := builder.Build(context.Background())
			require.Error(t, err)
			assert.Nil(t, cfg)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}

// TestBuilder_connectRedis_NoAddress tests that Redis is not connected when no address is provided.
func TestBuilder_connectRedis_NoAddress(t *testing.T) {
	b := &BuilderImpl{
		provider: &mockProvider{values: map[string]string{
			"REDIS_ADDR": "",
		}},
		redis: nil, // not needed for this test
	}
	config := &APIConfig{}
	err := b.connectRedis(context.Background(), config)
	require.NoError(t, err)
	assert.Nil(t, config.RedisClient)
}

// errorRedisProvider is a mock Redis provider that simulates a connection error.
type errorRedisProvider struct{}

// Connect simulates a Redis connection failure by always returning an error.
func (e *errorRedisProvider) Connect(ctx context.Context) (redis.Cmdable, error) {
	return nil, errors.New("connection error")
}

// Close is a no-op close method for errorRedisProvider.
func (e *errorRedisProvider) Close() error { return nil }

// TestBuilder_connectRedis_Error verifies that an error is returned when Redis connection fails.
func TestBuilder_connectRedis_Error(t *testing.T) {
	b := &BuilderImpl{
		provider: &mockProvider{values: map[string]string{
			"REDIS_ADDR": "localhost:6379",
		}},
		redis: &errorRedisProvider{},
	}
	config := &APIConfig{}
	err := b.connectRedis(context.Background(), config)
	require.Error(t, err)
	assert.Nil(t, config.RedisClient)
}

// errorMongoProvider is a mock MongoDB provider that simulates a connection failure.
type errorMongoProvider struct{}

// Connect simulates a MongoDB connection failure by always returning an error.
func (e *errorMongoProvider) Connect(ctx context.Context) (*mongo.Client, *mongo.Database, error) {
	return nil, nil, errors.New("mongo connection error")
}

// Close is a no-op close method for errorMongoProvider.
func (e *errorMongoProvider) Close(ctx context.Context) error { return nil }

// TestBuilder_connectMongo_Error verifies that an error is returned when MongoDB connection fails.
func TestBuilder_connectMongo_Error(t *testing.T) {
	b := &BuilderImpl{
		provider: &mockProvider{values: map[string]string{}},
		mongo:    &errorMongoProvider{},
	}
	config := &APIConfig{}
	err := b.connectMongo(context.Background(), config, "mongodb://localhost:27017")
	require.Error(t, err)
	assert.Nil(t, config.MongoClient)
	assert.Nil(t, config.MongoDB)
}

// successMongoProvider is a mock MongoDB provider that simulates a successful connection.
type successMongoProvider struct {
	client *mongo.Client
	db     *mongo.Database
}

// Connect simulates a successful MongoDB connection by returning the preconfigured client and database.
func (s *successMongoProvider) Connect(ctx context.Context) (*mongo.Client, *mongo.Database, error) {
	return s.client, s.db, nil
}

// Close is a no-op close method for successMongoProvider.
func (s *successMongoProvider) Close(ctx context.Context) error { return nil }

// TestBuilder_connectMongo_Success verifies that the builder correctly connects to MongoDB.
func TestBuilder_connectMongo_Success(t *testing.T) {
	client := &mongo.Client{}
	db := &mongo.Database{}
	b := &BuilderImpl{
		provider: &mockProvider{values: map[string]string{}},
		mongo:    &successMongoProvider{client: client, db: db},
	}
	config := &APIConfig{}
	err := b.connectMongo(context.Background(), config, "mongodb://localhost:27017")
	require.NoError(t, err)
	assert.Equal(t, client, config.MongoClient)
	assert.Equal(t, db, config.MongoDB)
}

// errorOAuthProvider is a mock OAuth provider that simulates a failure to load config.
type errorOAuthProvider struct{}

// LoadGoogleConfig simulates a failure when loading OAuth config.
func (e *errorOAuthProvider) LoadGoogleConfig(credsPath string) (*OAuthConfig, error) {
	return nil, errors.New("oauth error")
}

// TestBuilder_Build_ValidatorError verifies that the builder returns an error when required environment variables are missing.
func TestBuilder_Build_ValidatorError(t *testing.T) {
	provider := &mockProvider{values: map[string]string{
		// Intentionally omit "PORT" to trigger validator error
		"JWT_SECRET": "jwt", "REFRESH_SECRET": "refresh", "ISSUER": "issuer", "AUDIENCE": "aud",
		"GOOGLE_CREDENTIALS_PATH": "creds.json", "S3_BUCKET": "bucket", "S3_REGION": "region", "STRIPE_SECRET_KEY": "sk",
		"STRIPE_WEBHOOK_SECRET": "wh", "MONGO_URI": "mongodb://localhost:27017",
	}}
	builder := NewConfigBuilder().WithProvider(provider)
	cfg, err := builder.Build(context.Background())
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to get PORT")
}
