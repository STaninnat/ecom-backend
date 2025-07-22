// Package config provides configuration management, validation, and provider logic for the ecom-backend project.
package config

import (
	"context"
	"fmt"
)

// builder.go: Configuration builder pattern and construction logic.

// BuilderImpl implements the ConfigBuilder interface for constructing APIConfig instances with various providers and settings.
type BuilderImpl struct {
	provider Provider
	database DatabaseProvider
	redis    RedisProvider
	mongo    MongoProvider
	s3       S3Provider
	oauth    OAuthProvider
}

// NewConfigBuilder creates and returns a new instance of ConfigBuilderImpl.
// Initializes a new configuration builder for constructing APIConfig instances.
func NewConfigBuilder() *BuilderImpl {
	return &BuilderImpl{}
}

// WithProvider sets the configuration provider for the builder.
// The provider supplies configuration values from sources such as environment variables or config files.
func (b *BuilderImpl) WithProvider(provider Provider) Builder {
	b.provider = provider
	return b
}

// WithDatabase sets the database provider for the builder.
// Handles creation and management of database connections.
func (b *BuilderImpl) WithDatabase(provider DatabaseProvider) Builder {
	b.database = provider
	return b
}

// WithRedis sets the Redis provider for the builder.
// Manages Redis connections and caching functionality.
func (b *BuilderImpl) WithRedis(provider RedisProvider) Builder {
	b.redis = provider
	return b
}

// WithMongo sets the MongoDB provider for the builder.
// Handles NoSQL database connections and operations.
func (b *BuilderImpl) WithMongo(provider MongoProvider) Builder {
	b.mongo = provider
	return b
}

// WithS3 sets the S3 provider for the builder.
// Manages AWS S3 client creation and file storage operations.
func (b *BuilderImpl) WithS3(provider S3Provider) Builder {
	b.s3 = provider
	return b
}

// WithOAuth sets the OAuth provider for the builder.
// Handles authentication configuration, particularly for Google OAuth integration.
func (b *BuilderImpl) WithOAuth(provider OAuthProvider) Builder {
	b.oauth = provider
	return b
}

// Helper to load required config values
func (b *BuilderImpl) loadRequiredConfig() (map[string]string, error) {
	requiredKeys := []string{
		"PORT", "JWT_SECRET", "REFRESH_SECRET", "ISSUER", "AUDIENCE",
		"GOOGLE_CREDENTIALS_PATH", "S3_BUCKET", "S3_REGION",
		"STRIPE_SECRET_KEY", "STRIPE_WEBHOOK_SECRET", "MONGO_URI",
	}
	values := make(map[string]string)
	for _, key := range requiredKeys {
		val, err := b.provider.GetRequiredString(key)
		if err != nil {
			return nil, fmt.Errorf("failed to get %s: %w", key, err)
		}
		values[key] = val
	}
	return values, nil
}

func (b *BuilderImpl) getOptionalConfig() (uploadBackend, uploadPath string) {
	uploadBackend = b.provider.GetStringOrDefault("UPLOAD_BACKEND", "local")
	uploadPath = b.provider.GetStringOrDefault("UPLOAD_PATH", "./uploads")
	return
}

func (b *BuilderImpl) connectRedis(ctx context.Context, config *APIConfig) error {
	redisAddr := b.provider.GetString("REDIS_ADDR")
	redisUsername := b.provider.GetString("REDIS_USERNAME")
	redisPassword := b.provider.GetString("REDIS_PASSWORD")
	if redisAddr != "" {
		redisProvider := NewRedisProvider(redisAddr, redisUsername, redisPassword)
		redisClient, err := redisProvider.Connect(ctx)
		if err != nil {
			return fmt.Errorf("failed to connect to Redis: %w", err)
		}
		config.RedisClient = redisClient
	}
	return nil
}

func (b *BuilderImpl) connectMongo(ctx context.Context, config *APIConfig, mongoURI string) error {
	var mongoProvider MongoProvider
	if b.mongo != nil {
		mongoProvider = b.mongo
	} else {
		mongoProvider = NewMongoProvider(mongoURI)
	}
	mongoClient, mongoDB, err := mongoProvider.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	config.MongoClient = mongoClient
	config.MongoDB = mongoDB
	return nil
}

func (b *BuilderImpl) createS3Client(ctx context.Context, config *APIConfig, s3Region string) error {
	s3Client, err := b.s3.CreateClient(ctx, s3Region)
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %w", err)
	}
	config.S3Client = s3Client
	return nil
}

func (b *BuilderImpl) loadOAuthConfig(credsPath string) error {
	oauthConfig, err := b.oauth.LoadGoogleConfig(credsPath)
	if err != nil {
		return fmt.Errorf("failed to load OAuth config: %w", err)
	}
	_ = oauthConfig // Not stored in APIConfig
	return nil
}

// Build constructs the configuration using the provided options.
func (b *BuilderImpl) Build(ctx context.Context) (*APIConfig, error) {
	if b.provider == nil {
		return nil, fmt.Errorf("config provider is required")
	}

	required, err := b.loadRequiredConfig()
	if err != nil {
		return nil, err
	}
	uploadBackend, uploadPath := b.getOptionalConfig()

	config := &APIConfig{
		Port:                required["PORT"],
		JWTSecret:           required["JWT_SECRET"],
		RefreshSecret:       required["REFRESH_SECRET"],
		Issuer:              required["ISSUER"],
		Audience:            required["AUDIENCE"],
		CredsPath:           required["GOOGLE_CREDENTIALS_PATH"],
		S3Bucket:            required["S3_BUCKET"],
		S3Region:            required["S3_REGION"],
		StripeSecretKey:     required["STRIPE_SECRET_KEY"],
		StripeWebhookSecret: required["STRIPE_WEBHOOK_SECRET"],
		UploadBackend:       uploadBackend,
		UploadPath:          uploadPath,
	}

	if b.redis != nil {
		if err := b.connectRedis(ctx, config); err != nil {
			return nil, err
		}
	}
	if b.mongo != nil {
		if err := b.connectMongo(ctx, config, required["MONGO_URI"]); err != nil {
			return nil, err
		}
	}
	if b.s3 != nil {
		if err := b.createS3Client(ctx, config, required["S3_REGION"]); err != nil {
			return nil, err
		}
	}
	if b.oauth != nil {
		if err := b.loadOAuthConfig(required["GOOGLE_CREDENTIALS_PATH"]); err != nil {
			return nil, err
		}
	}
	return config, nil
}
