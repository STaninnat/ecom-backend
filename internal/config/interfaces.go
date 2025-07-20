package config

import (
	"context"
	"database/sql"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ConfigProvider defines the interface for configuration providers
type ConfigProvider interface {
	GetString(key string) string
	GetStringOrDefault(key, defaultValue string) string
	GetRequiredString(key string) (string, error)
	GetInt(key string) int
	GetIntOrDefault(key string, defaultValue int) int
	GetBool(key string) bool
	GetBoolOrDefault(key string, defaultValue bool) bool
}

// DatabaseProvider defines the interface for database connections
type DatabaseProvider interface {
	Connect(ctx context.Context) (*sql.DB, *database.Queries, error)
	Close() error
}

// RedisProvider defines the interface for Redis connections
type RedisProvider interface {
	Connect(ctx context.Context) (redis.Cmdable, error)
	Close() error
}

// MongoProvider defines the interface for MongoDB connections
type MongoProvider interface {
	Connect(ctx context.Context) (*mongo.Client, *mongo.Database, error)
	Close(ctx context.Context) error
}

// S3Provider defines the interface for S3 client creation
type S3Provider interface {
	CreateClient(ctx context.Context, region string) (*s3.Client, error)
}

// OAuthProvider defines the interface for OAuth configuration
type OAuthProvider interface {
	LoadGoogleConfig(credsPath string) (*OAuthConfig, error)
}

// ConfigValidator defines the interface for configuration validation
type ConfigValidator interface {
	Validate() error
}

// ConfigBuilder defines the interface for building configuration
type ConfigBuilder interface {
	WithProvider(provider ConfigProvider) ConfigBuilder
	WithDatabase(provider DatabaseProvider) ConfigBuilder
	WithRedis(provider RedisProvider) ConfigBuilder
	WithMongo(provider MongoProvider) ConfigBuilder
	WithS3(provider S3Provider) ConfigBuilder
	WithOAuth(provider OAuthProvider) ConfigBuilder
	Build(ctx context.Context) (*APIConfig, error)
}
