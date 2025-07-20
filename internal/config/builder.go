package config

import (
	"context"
	"fmt"
)

// ConfigBuilderImpl implements the ConfigBuilder interface for constructing APIConfig instances with various providers and settings.
type ConfigBuilderImpl struct {
	provider ConfigProvider
	database DatabaseProvider
	redis    RedisProvider
	mongo    MongoProvider
	s3       S3Provider
	oauth    OAuthProvider
}

// NewConfigBuilder creates and returns a new instance of ConfigBuilderImpl.
// Initializes a new configuration builder for constructing APIConfig instances.
func NewConfigBuilder() *ConfigBuilderImpl {
	return &ConfigBuilderImpl{}
}

// WithProvider sets the configuration provider for the builder.
// The provider supplies configuration values from sources such as environment variables or config files.
func (b *ConfigBuilderImpl) WithProvider(provider ConfigProvider) ConfigBuilder {
	b.provider = provider
	return b
}

// WithDatabase sets the database provider for the builder.
// Handles creation and management of database connections.
func (b *ConfigBuilderImpl) WithDatabase(provider DatabaseProvider) ConfigBuilder {
	b.database = provider
	return b
}

// WithRedis sets the Redis provider for the builder.
// Manages Redis connections and caching functionality.
func (b *ConfigBuilderImpl) WithRedis(provider RedisProvider) ConfigBuilder {
	b.redis = provider
	return b
}

// WithMongo sets the MongoDB provider for the builder.
// Handles NoSQL database connections and operations.
func (b *ConfigBuilderImpl) WithMongo(provider MongoProvider) ConfigBuilder {
	b.mongo = provider
	return b
}

// WithS3 sets the S3 provider for the builder.
// Manages AWS S3 client creation and file storage operations.
func (b *ConfigBuilderImpl) WithS3(provider S3Provider) ConfigBuilder {
	b.s3 = provider
	return b
}

// WithOAuth sets the OAuth provider for the builder.
// Handles authentication configuration, particularly for Google OAuth integration.
func (b *ConfigBuilderImpl) WithOAuth(provider OAuthProvider) ConfigBuilder {
	b.oauth = provider
	return b
}

// Build constructs and returns a complete APIConfig instance based on the builder's configuration.
// Validates required configuration, establishes service connections, and returns a ready-to-use APIConfig.
// Returns an error if any required configuration is missing or service connections fail.
func (b *ConfigBuilderImpl) Build(ctx context.Context) (*APIConfig, error) {
	if b.provider == nil {
		return nil, fmt.Errorf("config provider is required")
	}

	// Load required configuration values
	port, err := b.provider.GetRequiredString("PORT")
	if err != nil {
		return nil, fmt.Errorf("failed to get PORT: %w", err)
	}

	jwtSecret, err := b.provider.GetRequiredString("JWT_SECRET")
	if err != nil {
		return nil, fmt.Errorf("failed to get JWT_SECRET: %w", err)
	}

	refreshSecret, err := b.provider.GetRequiredString("REFRESH_SECRET")
	if err != nil {
		return nil, fmt.Errorf("failed to get REFRESH_SECRET: %w", err)
	}

	issuer, err := b.provider.GetRequiredString("ISSUER")
	if err != nil {
		return nil, fmt.Errorf("failed to get ISSUER: %w", err)
	}

	audience, err := b.provider.GetRequiredString("AUDIENCE")
	if err != nil {
		return nil, fmt.Errorf("failed to get AUDIENCE: %w", err)
	}

	credsPath, err := b.provider.GetRequiredString("GOOGLE_CREDENTIALS_PATH")
	if err != nil {
		return nil, fmt.Errorf("failed to get GOOGLE_CREDENTIALS_PATH: %w", err)
	}

	s3Bucket, err := b.provider.GetRequiredString("S3_BUCKET")
	if err != nil {
		return nil, fmt.Errorf("failed to get S3_BUCKET: %w", err)
	}

	s3Region, err := b.provider.GetRequiredString("S3_REGION")
	if err != nil {
		return nil, fmt.Errorf("failed to get S3_REGION: %w", err)
	}

	stripeSecretKey, err := b.provider.GetRequiredString("STRIPE_SECRET_KEY")
	if err != nil {
		return nil, fmt.Errorf("failed to get STRIPE_SECRET_KEY: %w", err)
	}

	stripeWebhookSecret, err := b.provider.GetRequiredString("STRIPE_WEBHOOK_SECRET")
	if err != nil {
		return nil, fmt.Errorf("failed to get STRIPE_WEBHOOK_SECRET: %w", err)
	}

	mongoURI, err := b.provider.GetRequiredString("MONGO_URI")
	if err != nil {
		return nil, fmt.Errorf("failed to get MONGO_URI: %w", err)
	}

	// Get optional values with defaults
	uploadBackend := b.provider.GetStringOrDefault("UPLOAD_BACKEND", "local")
	uploadPath := b.provider.GetStringOrDefault("UPLOAD_PATH", "./uploads")

	// Initialize services
	config := &APIConfig{
		Port:                port,
		JWTSecret:           jwtSecret,
		RefreshSecret:       refreshSecret,
		Issuer:              issuer,
		Audience:            audience,
		CredsPath:           credsPath,
		S3Bucket:            s3Bucket,
		S3Region:            s3Region,
		StripeSecretKey:     stripeSecretKey,
		StripeWebhookSecret: stripeWebhookSecret,
		UploadBackend:       uploadBackend,
		UploadPath:          uploadPath,
	}

	// Connect to Redis
	if b.redis != nil {
		redisAddr := b.provider.GetString("REDIS_ADDR")
		redisUsername := b.provider.GetString("REDIS_USERNAME")
		redisPassword := b.provider.GetString("REDIS_PASSWORD")

		if redisAddr != "" {
			redisProvider := NewRedisProvider(redisAddr, redisUsername, redisPassword)
			redisClient, err := redisProvider.Connect(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to Redis: %w", err)
			}
			config.RedisClient = redisClient
		}
	}

	// Connect to MongoDB
	if b.mongo != nil {
		mongoProvider := NewMongoProvider(mongoURI)
		mongoClient, mongoDB, err := mongoProvider.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
		}
		config.MongoClient = mongoClient
		config.MongoDB = mongoDB
	}

	// Create S3 client
	if b.s3 != nil {
		s3Client, err := b.s3.CreateClient(ctx, s3Region)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 client: %w", err)
		}
		config.S3Client = s3Client
	}

	// Load OAuth configuration
	if b.oauth != nil {
		oauthConfig, err := b.oauth.LoadGoogleConfig(credsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load OAuth config: %w", err)
		}
		// Note: OAuth config is not stored in APIConfig, it's used separately
		_ = oauthConfig
	}

	return config, nil
}
