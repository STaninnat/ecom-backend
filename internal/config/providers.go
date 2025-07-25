// Package config provides configuration management, validation, and provider logic for the ecom-backend project.
package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/oauth2/google"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// providers.go: Environment, database, Redis, MongoDB, S3, and OAuth provider implementations.

const strTrue = "true"

// EnvironmentProvider implements ConfigProvider using environment variables
type EnvironmentProvider struct{}

// NewEnvironmentProvider creates and returns a new EnvironmentProvider instance.
// This provider reads configuration values from environment variables,
// making it suitable for containerized deployments and cloud environments.
func NewEnvironmentProvider() *EnvironmentProvider {
	return &EnvironmentProvider{}
}

// GetString retrieves a string value from environment variables.
// Returns the environment variable value if set, or an empty string if not found.
func (e *EnvironmentProvider) GetString(key string) string {
	return os.Getenv(key)
}

// GetStringOrDefault retrieves a string value from environment variables with a default fallback.
// Returns the environment variable value if set and non-empty, otherwise returns the default value.
// This is useful for optional configuration values that should have sensible defaults.
func (e *EnvironmentProvider) GetStringOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetRequiredString retrieves a required string value from environment variables.
// Returns the environment variable value if set, or an error if the value is missing.
// This method is used for essential configuration values that must be present for the application to function.
func (e *EnvironmentProvider) GetRequiredString(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return value, nil
}

// GetInt retrieves an integer value from environment variables.
// Attempts to parse the environment variable as an integer and returns the parsed value.
// Returns 0 if the environment variable is not set or cannot be parsed as an integer.
func (e *EnvironmentProvider) GetInt(key string) int {
	value := os.Getenv(key)
	if value == "" {
		return 0
	}
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return 0
}

// GetIntOrDefault retrieves an integer value from environment variables with a default fallback.
// Attempts to parse the environment variable as an integer and returns the parsed value.
// Returns the default value if the environment variable is not set or cannot be parsed as an integer.
func (e *EnvironmentProvider) GetIntOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}

// GetBool retrieves a boolean value from environment variables.
// Parses various truthy values including "true", "1", and "yes" as true.
// Returns false for any other value or if the environment variable is not set.
func (e *EnvironmentProvider) GetBool(key string) bool {
	value := strings.ToLower(os.Getenv(key))
	return value == strTrue || value == "1" || value == "yes"
}

// GetBoolOrDefault retrieves a boolean value from environment variables with a default fallback.
// Parses various truthy values including "true", "1", and "yes" as true.
// Returns the default value if the environment variable is not set or cannot be parsed as a boolean.
func (e *EnvironmentProvider) GetBoolOrDefault(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return e.GetBool(key)
}

// PostgresProvider implements DatabaseProvider for PostgreSQL
type PostgresProvider struct {
	dbURL   string
	db      *sql.DB
	sqlOpen func(driverName, dataSourceName string) (*sql.DB, error)
}

// NewPostgresProvider creates and returns a new PostgresProvider instance.
// This provider manages PostgreSQL database connections and provides access to the database
// and generated SQL queries for the application's data layer.
func NewPostgresProvider(dbURL string) *PostgresProvider {
	return &PostgresProvider{dbURL: dbURL, sqlOpen: sql.Open}
}

// Connect establishes a connection to the PostgreSQL database and initializes the queries object.
// This method opens a database connection, verifies connectivity with a ping operation,
// and returns both the database connection and a queries object for executing SQL operations.
// Returns an error if the connection cannot be established or if the ping operation fails.
func (p *PostgresProvider) Connect(ctx context.Context) (*sql.DB, *database.Queries, error) {
	sqlOpen := p.sqlOpen
	if sqlOpen == nil {
		sqlOpen = sql.Open
	}
	db, err := sqlOpen("postgres", p.dbURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	p.db = db
	dbQueries := database.New(db)
	return db, dbQueries, nil
}

// Close terminates the database connection and releases associated resources.
// This method safely closes the PostgreSQL connection and sets the internal database
// reference to nil to prevent further usage of the closed connection.
func (p *PostgresProvider) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// RedisProviderImpl implements RedisProvider
type RedisProviderImpl struct {
	addr      string
	username  string
	password  string
	client    *redis.Client
	newClient func(opt *redis.Options) *redis.Client
}

// NewRedisProvider creates and returns a new RedisProviderImpl instance.
// This provider manages Redis connections for caching, session storage, and rate limiting.
// It supports authentication and can be configured with custom connection parameters.
func NewRedisProvider(addr, username, password string) *RedisProviderImpl {
	return &RedisProviderImpl{
		addr:      addr,
		username:  username,
		password:  password,
		newClient: redis.NewClient,
	}
}

// Connect establishes a connection to the Redis server and verifies connectivity.
// This method creates a Redis client with the configured connection parameters,
// performs a ping operation to verify the connection is working, and returns
// the Redis client for use in the application.
func (r *RedisProviderImpl) Connect(ctx context.Context) (redis.Cmdable, error) {
	newClient := r.newClient
	if newClient == nil {
		newClient = redis.NewClient
	}
	client := newClient(&redis.Options{
		Addr:     r.addr,
		Username: r.username,
		Password: r.password,
		DB:       0,
	})

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	r.client = client
	return client, nil
}

// Close terminates the Redis connection and releases associated resources.
// This method safely closes the Redis client connection and sets the internal
// client reference to nil to prevent further usage of the closed connection.
func (r *RedisProviderImpl) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// MongoProviderImpl implements MongoProvider
type MongoProviderImpl struct {
	uri     string
	client  *mongo.Client
	connect func(opts ...*options.ClientOptions) (*mongo.Client, error)
}

// NewMongoProvider creates and returns a new MongoProviderImpl instance.
// This provider manages MongoDB connections for NoSQL database operations,
// particularly for features like cart management and user reviews.
func NewMongoProvider(uri string) *MongoProviderImpl {
	return &MongoProviderImpl{uri: uri, connect: mongo.Connect}
}

// Connect establishes a connection to the MongoDB server and returns the client and database.
// This method creates a MongoDB client with the configured connection URI,
// performs a ping operation to verify connectivity, and returns both the client
// and a database instance for performing MongoDB operations.
func (m *MongoProviderImpl) Connect(ctx context.Context) (*mongo.Client, *mongo.Database, error) {
	clientOptions := options.Client().ApplyURI(m.uri)

	connect := m.connect
	if connect == nil {
		connect = mongo.Connect
	}
	client, err := connect(clientOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	db := client.Database("ecommerce_db")
	return client, db, nil
}

// Close terminates the MongoDB connection and releases associated resources.
// This method safely disconnects the MongoDB client and sets the internal
// client reference to nil to prevent further usage of the closed connection.
func (m *MongoProviderImpl) Close(ctx context.Context) error {
	if m.client != nil {
		return m.client.Disconnect(ctx)
	}
	return nil
}

// S3ProviderImpl implements S3Provider
type S3ProviderImpl struct {
	loadConfig func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error)
}

// NewS3Provider creates and returns a new S3ProviderImpl instance.
// This provider manages AWS S3 client creation for cloud-based file storage operations,
// enabling the application to upload and manage files in the cloud.
func NewS3Provider() *S3ProviderImpl {
	return &S3ProviderImpl{loadConfig: config.LoadDefaultConfig}
}

// CreateClient creates an AWS S3 client with the specified region configuration.
// This method initializes AWS configuration and creates an S3 client that can be used
// for file upload, download, and management operations in the specified AWS region.
func (s *S3ProviderImpl) CreateClient(ctx context.Context, region string) (*s3.Client, error) {
	loadConfig := s.loadConfig
	if loadConfig == nil {
		loadConfig = config.LoadDefaultConfig
	}
	awsCfg, err := loadConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	return client, nil
}

// OAuthProviderImpl implements OAuthProvider
type OAuthProviderImpl struct {
	readFile func(filename string) ([]byte, error)
}

// NewOAuthProvider creates and returns a new OAuthProviderImpl instance.
// This provider manages Google OAuth configuration loading and validation,
// enabling the application to authenticate users through Google's OAuth service.
func NewOAuthProvider() *OAuthProviderImpl {
	return &OAuthProviderImpl{readFile: os.ReadFile}
}

// LoadGoogleConfig loads and validates Google OAuth configuration from a credentials file.
// This method reads the Google credentials JSON file, validates the file path for security,
// parses the configuration, and returns a properly configured OAuth configuration object.
// The credentials file must contain valid Google OAuth client information.
func (o *OAuthProviderImpl) LoadGoogleConfig(credsPath string) (*OAuthConfig, error) {
	safePath := cleanPath(credsPath)
	if !isSafePath(safePath) {
		return nil, fmt.Errorf("unsafe file path")
	}

	readFile := o.readFile
	if readFile == nil {
		readFile = os.ReadFile
	}

	data, err := readFile(safePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	googleConfig, err := google.ConfigFromJSON(data,
		"https://www.googleapis.com/auth/userinfo.profile",
		"https://www.googleapis.com/auth/userinfo.email",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Google config: %w", err)
	}

	return &OAuthConfig{Google: googleConfig}, nil
}

// Helper functions
// cleanPath normalizes and cleans a file path for security validation.
func cleanPath(path string) string {
	return filepath.Clean(path)
}
