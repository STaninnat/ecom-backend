// Package config provides configuration management, validation, and provider logic for the ecom-backend project.
package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// config.go: Main API configuration struct, loading, and environment integration.

// APIConfig holds all configuration for the API, including server, JWT, database, Redis, MongoDB, S3, Stripe, upload, and OAuth settings.
type APIConfig struct {
	// Server configuration
	Port string

	// JWT configuration
	JWTSecret     string
	RefreshSecret string
	Issuer        string
	Audience      string

	// Database configuration
	DBConn *sql.DB
	DB     *database.Queries

	// Redis configuration
	RedisClient redis.Cmdable

	// MongoDB configuration
	MongoClient *mongo.Client
	MongoDB     *mongo.Database

	// S3 configuration
	S3Client *s3.Client
	S3Bucket string
	S3Region string

	// Stripe configuration
	StripeSecretKey     string
	StripeWebhookSecret string

	// Upload configuration
	UploadBackend string
	UploadPath    string

	// OAuth configuration
	CredsPath string
}

// LoadConfig loads configuration from environment variables and initializes services.
// Uses the legacy pattern and calls log.Fatal on errors. Prefer LoadConfigWithError for better error handling.
func LoadConfig() *APIConfig {
	config, err := LoadConfigWithError(context.Background())
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}
	return config
}

// LoadConfigWithError loads configuration from environment variables and initializes services.
// Returns errors instead of calling log.Fatal, making it suitable for testing and graceful error handling.
func LoadConfigWithError(ctx context.Context) (*APIConfig, error) {
	return LoadConfigWithProviders(
		ctx,
		NewEnvironmentProvider(),
		NewPostgresProvider(""),      // Will be set from environment
		NewRedisProvider("", "", ""), // Will be set from environment
		NewMongoProvider(""),         // Will be set from environment
		NewS3Provider(),
		NewOAuthProvider(),
	)
}

// LoadConfigWithProviders loads configuration using the provided providers and initializes services.
// Uses dependency injection to load configuration values and establish connections to external services.
// Validates the configuration and returns a fully configured APIConfig instance.
func LoadConfigWithProviders(
	ctx context.Context,
	provider Provider,
	dbProvider DatabaseProvider,
	redisProvider RedisProvider,
	mongoProvider MongoProvider,
	s3Provider S3Provider,
	oauthProvider OAuthProvider,
) (*APIConfig, error) {
	builder := NewConfigBuilder().
		WithProvider(provider).
		WithDatabase(dbProvider).
		WithRedis(redisProvider).
		WithMongo(mongoProvider).
		WithS3(s3Provider).
		WithOAuth(oauthProvider)

	config, err := builder.Build(ctx)
	if err != nil {
		return nil, err
	}

	// Connect to database if provider is available
	if dbProvider != nil {
		dbURL := provider.GetString("DATABASE_URL")
		if dbURL != "" {
			postgresProvider := NewPostgresProvider(dbURL)
			db, dbQueries, err := postgresProvider.Connect(ctx)
			if err != nil {
				return nil, err
			}
			config.DB = dbQueries
			config.DBConn = db
		}
	}

	// Validate configuration
	validator := NewConfigValidator()
	if err := validator.ValidatePartial(config); err != nil {
		return nil, err
	}

	return config, nil
}

// MockConfigProvider is a mock implementation of ConfigProvider for testing
type MockConfigProvider struct {
	values map[string]string
}

// GetString returns the string value for the given key from the mock provider.
func (m *MockConfigProvider) GetString(key string) string {
	return m.values[key]
}

// GetStringOrDefault returns the string value for the given key or the default value.
func (m *MockConfigProvider) GetStringOrDefault(key, defaultValue string) string {
	if value, exists := m.values[key]; exists && value != "" {
		return value
	}
	return defaultValue
}

// GetRequiredString returns the string value for the given key or an error if not found.
func (m *MockConfigProvider) GetRequiredString(key string) (string, error) {
	if value, exists := m.values[key]; exists && value != "" {
		return value, nil
	}
	return "", fmt.Errorf("required environment variable %s is not set", key)
}

// GetInt returns the integer value for the given key from the mock provider.
func (m *MockConfigProvider) GetInt(key string) int {
	value := m.values[key]
	if value == "" {
		return 0
	}
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return 0
}

// GetIntOrDefault returns the integer value for the given key or the default value.
func (m *MockConfigProvider) GetIntOrDefault(key string, defaultValue int) int {
	value := m.values[key]
	if value == "" {
		return defaultValue
	}
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}

// GetBool returns the boolean value for the given key from the mock provider.
func (m *MockConfigProvider) GetBool(key string) bool {
	value := strings.ToLower(m.values[key])
	return value == "true" || value == "1" || value == "yes"
}

// GetBoolOrDefault returns the boolean value for the given key or the default value.
func (m *MockConfigProvider) GetBoolOrDefault(key string, defaultValue bool) bool {
	value := m.values[key]
	if value == "" {
		return defaultValue
	}
	return m.GetBool(key)
}

// LoadConfigForTesting loads a minimal configuration suitable for testing purposes.
// Creates a configuration with default values and mock providers, ideal for unit tests that don't require real external service connections.
func LoadConfigForTesting(ctx context.Context) (*APIConfig, error) {
	// Create mock providers for testing
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

	return LoadConfigWithProviders(
		ctx,
		mockProvider,
		nil, // No database for testing
		nil, // No Redis for testing
		nil, // No MongoDB for testing
		nil, // No S3 for testing
		nil, // No OAuth for testing
	)
}
