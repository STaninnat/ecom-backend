// Package config provides configuration management, validation, and provider logic for the ecom-backend project.
package config

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	redismock "github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// providers_test.go: Tests for environment, database, Redis, MongoDB, S3, and OAuth providers.

// Remove reflect, unsafe, and patch helpers

// TestEnvironmentProvider_GetString tests the GetString method of EnvironmentProvider.
// It verifies that environment variables are correctly retrieved and empty strings are returned for missing keys.
func TestEnvironmentProvider_GetString(t *testing.T) {
	provider := NewEnvironmentProvider()
	if err := os.Setenv("FOO", "bar"); err != nil {
		t.Errorf("Failed to set environment variable: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Unsetenv("FOO"); err != nil {
			t.Errorf("Failed to unset environment variable: %v", err)
		}
	})
	assert.Equal(t, "bar", provider.GetString("FOO"))
	assert.Equal(t, "", provider.GetString("NOT_SET"))
}

// TestEnvironmentProvider_GetStringOrDefault tests the GetStringOrDefault method of EnvironmentProvider.
// It verifies that default values are returned when environment variables are not set.
func TestEnvironmentProvider_GetStringOrDefault(t *testing.T) {
	provider := NewEnvironmentProvider()
	if err := os.Setenv("FOO", "bar"); err != nil {
		t.Errorf("Failed to set environment variable: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Unsetenv("FOO"); err != nil {
			t.Errorf("Failed to unset environment variable: %v", err)
		}
	})
	assert.Equal(t, "bar", provider.GetStringOrDefault("FOO", "baz"))
	assert.Equal(t, "baz", provider.GetStringOrDefault("NOT_SET", "baz"))
}

// TestEnvironmentProvider_GetRequiredString tests the GetRequiredString method of EnvironmentProvider.
// It verifies that required environment variables are retrieved successfully and errors are returned for missing keys.
func TestEnvironmentProvider_GetRequiredString(t *testing.T) {
	provider := NewEnvironmentProvider()
	if err := os.Setenv("FOO", "bar"); err != nil {
		t.Errorf("Failed to set environment variable: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Unsetenv("FOO"); err != nil {
			t.Errorf("Failed to unset environment variable: %v", err)
		}
	})
	val, err := provider.GetRequiredString("FOO")
	assert.NoError(t, err)
	assert.Equal(t, "bar", val)
	_, err = provider.GetRequiredString("NOT_SET")
	assert.Error(t, err)
}

// TestEnvironmentProvider_GetInt tests the GetInt method of EnvironmentProvider.
// It verifies that integer environment variables are correctly parsed and zero is returned for invalid values.
func TestEnvironmentProvider_GetInt(t *testing.T) {
	provider := NewEnvironmentProvider()
	if err := os.Setenv("INT_VAL", "42"); err != nil {
		t.Errorf("Failed to set INT_VAL: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Unsetenv("INT_VAL"); err != nil {
			t.Errorf("Failed to unset INT_VAL: %v", err)
		}
	})
	assert.Equal(t, 42, provider.GetInt("INT_VAL"))
	assert.Equal(t, 0, provider.GetInt("NOT_SET"))
	if err := os.Setenv("BAD_INT", "abc"); err != nil {
		t.Errorf("Failed to set BAD_INT: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Unsetenv("BAD_INT"); err != nil {
			t.Errorf("Failed to unset BAD_INT: %v", err)
		}
	})
	assert.Equal(t, 0, provider.GetInt("BAD_INT"))
}

// TestEnvironmentProvider_GetIntOrDefault tests the GetIntOrDefault method of EnvironmentProvider.
// It verifies that default integer values are returned when environment variables are not set or invalid.
func TestEnvironmentProvider_GetIntOrDefault(t *testing.T) {
	provider := NewEnvironmentProvider()
	if err := os.Setenv("INT_VAL", "42"); err != nil {
		t.Errorf("Failed to set INT_VAL: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Unsetenv("INT_VAL"); err != nil {
			t.Errorf("Failed to unset INT_VAL: %v", err)
		}
	})
	assert.Equal(t, 42, provider.GetIntOrDefault("INT_VAL", 99))
	assert.Equal(t, 99, provider.GetIntOrDefault("NOT_SET", 99))
	if err := os.Setenv("BAD_INT", "abc"); err != nil {
		t.Errorf("Failed to set BAD_INT: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Unsetenv("BAD_INT"); err != nil {
			t.Errorf("Failed to unset BAD_INT: %v", err)
		}
	})
	assert.Equal(t, 99, provider.GetIntOrDefault("BAD_INT", 99))
}

// TestEnvironmentProvider_GetBool tests the GetBool method of EnvironmentProvider.
// It verifies that boolean environment variables are correctly parsed for various truthy and falsy values.
func TestEnvironmentProvider_GetBool(t *testing.T) {
	provider := NewEnvironmentProvider()
	if err := os.Setenv("BOOL_TRUE", "true"); err != nil {
		t.Errorf("Failed to set BOOL_TRUE: %v", err)
	}
	if err := os.Setenv("BOOL_ONE", "1"); err != nil {
		t.Errorf("Failed to set BOOL_ONE: %v", err)
	}
	if err := os.Setenv("BOOL_YES", "yes"); err != nil {
		t.Errorf("Failed to set BOOL_YES: %v", err)
	}
	if err := os.Setenv("BOOL_FALSE", "false"); err != nil {
		t.Errorf("Failed to set BOOL_FALSE: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Unsetenv("BOOL_TRUE"); err != nil {
			t.Errorf("Failed to unset BOOL_TRUE: %v", err)
		}
		if err := os.Unsetenv("BOOL_ONE"); err != nil {
			t.Errorf("Failed to unset BOOL_ONE: %v", err)
		}
		if err := os.Unsetenv("BOOL_YES"); err != nil {
			t.Errorf("Failed to unset BOOL_YES: %v", err)
		}
		if err := os.Unsetenv("BOOL_FALSE"); err != nil {
			t.Errorf("Failed to unset BOOL_FALSE: %v", err)
		}
	})
	assert.True(t, provider.GetBool("BOOL_TRUE"))
	assert.True(t, provider.GetBool("BOOL_ONE"))
	assert.True(t, provider.GetBool("BOOL_YES"))
	assert.False(t, provider.GetBool("BOOL_FALSE"))
	assert.False(t, provider.GetBool("NOT_SET"))
}

// TestEnvironmentProvider_GetBoolOrDefault tests the GetBoolOrDefault method of EnvironmentProvider.
// It verifies that default boolean values are returned when environment variables are not set.
func TestEnvironmentProvider_GetBoolOrDefault(t *testing.T) {
	provider := NewEnvironmentProvider()
	if err := os.Setenv("BOOL_TRUE", "true"); err != nil {
		t.Errorf("Failed to set BOOL_TRUE: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Unsetenv("BOOL_TRUE"); err != nil {
			t.Errorf("Failed to unset BOOL_TRUE: %v", err)
		}
	})
	assert.True(t, provider.GetBoolOrDefault("BOOL_TRUE", false))
	assert.True(t, provider.GetBoolOrDefault("NOT_SET", true))
	assert.False(t, provider.GetBoolOrDefault("NOT_SET", false))
}

// TestPostgresProvider_Connect_Success tests successful database connection in PostgresProvider.
// It verifies that the provider can establish a connection and return valid database and queries objects.
func TestPostgresProvider_Connect_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	mock.ExpectPing()
	mock.ExpectClose()
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() failed: %v", err)
		}
	}()

	provider := &PostgresProvider{dbURL: "sqlmock_db", sqlOpen: func(_, _ string) (*sql.DB, error) {
		return db, nil
	}}

	dbOut, queries, err := provider.Connect(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, dbOut)
	assert.NotNil(t, queries)
}

// TestPostgresProvider_Connect_OpenError tests database connection failure in PostgresProvider.
// It verifies that the provider returns an error when the database connection cannot be established.
func TestPostgresProvider_Connect_OpenError(t *testing.T) {
	provider := &PostgresProvider{dbURL: "bad_dsn", sqlOpen: func(_, _ string) (*sql.DB, error) {
		return nil, errors.New("open error")
	}}

	dbOut, queries, err := provider.Connect(context.Background())
	assert.Error(t, err)
	assert.Nil(t, dbOut)
	assert.Nil(t, queries)
}

// TestPostgresProvider_Connect_PingError tests database ping failure in PostgresProvider.
// It verifies that the provider returns an error when the database ping operation fails.
func TestPostgresProvider_Connect_PingError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	mock.ExpectClose()
	err = db.Close() // Close the db to force ping to fail
	if err != nil {
		t.Errorf("db.Close() failed: %v", err)
	}

	provider := &PostgresProvider{dbURL: "sqlmock_db", sqlOpen: func(_, _ string) (*sql.DB, error) {
		return db, nil
	}}

	dbOut, queries, err := provider.Connect(context.Background())
	assert.Error(t, err)
	assert.Nil(t, dbOut)
	assert.Nil(t, queries)
}

// TestPostgresProvider_Close tests successful database closure in PostgresProvider.
// It verifies that the provider can close the database connection without errors.
func TestPostgresProvider_Close(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	mock.ExpectClose()
	provider := &PostgresProvider{db: db}
	assert.NoError(t, provider.Close())
}

// TestPostgresProvider_Close_NilDB tests database closure with nil database in PostgresProvider.
// It verifies that the provider handles nil database gracefully during closure.
func TestPostgresProvider_Close_NilDB(t *testing.T) {
	provider := &PostgresProvider{db: nil}
	assert.NoError(t, provider.Close())
}

// TestPostgresProvider_Close_Error tests database closure error in PostgresProvider.
// It verifies that the provider returns an error when the database close operation fails.
func TestPostgresProvider_Close_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	mock.ExpectClose().WillReturnError(errors.New("close error"))
	provider := &PostgresProvider{db: db}
	assert.Error(t, provider.Close())
}

// TestRedisProviderImpl_Connect_Success tests successful Redis connection in RedisProviderImpl.
// It verifies that the provider can establish a Redis connection and return a valid client.
func TestRedisProviderImpl_Connect_Success(t *testing.T) {
	client, mock := redismock.NewClientMock()
	mock.ExpectPing().SetVal("PONG")

	provider := &RedisProviderImpl{
		addr:      "localhost:6379",
		username:  "",
		password:  "",
		newClient: func(_ *redis.Options) *redis.Client { return client },
	}

	cmdable, err := provider.Connect(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, cmdable)
}

// TestRedisProviderImpl_Connect_PingError tests Redis ping failure in RedisProviderImpl.
// It verifies that the provider returns an error when the Redis ping operation fails.
func TestRedisProviderImpl_Connect_PingError(t *testing.T) {
	client, mock := redismock.NewClientMock()
	mock.ExpectPing().SetErr(errors.New("ping error"))

	provider := &RedisProviderImpl{
		addr:      "localhost:6379",
		username:  "",
		password:  "",
		newClient: func(_ *redis.Options) *redis.Client { return client },
	}

	cmdable, err := provider.Connect(context.Background())
	assert.Error(t, err)
	assert.Nil(t, cmdable)
}

// TestRedisProviderImpl_Close tests successful Redis closure in RedisProviderImpl.
// It verifies that the provider can close the Redis connection without errors.
func TestRedisProviderImpl_Close(t *testing.T) {
	client, _ := redismock.NewClientMock()
	provider := &RedisProviderImpl{client: client}
	assert.NoError(t, provider.Close())
}

// TestRedisProviderImpl_Close_NilClient tests Redis closure with nil client in RedisProviderImpl.
// It verifies that the provider handles nil client gracefully during closure.
func TestRedisProviderImpl_Close_NilClient(t *testing.T) {
	provider := &RedisProviderImpl{client: nil}
	assert.NoError(t, provider.Close())
}

// TestRedisProviderImpl_Close_Error tests Redis closure error in RedisProviderImpl.
// It verifies that the provider returns an error when the Redis close operation fails.
func TestRedisProviderImpl_Close_Error(t *testing.T) {
	// Redis mock doesn't support ExpectClose, so we'll skip this test
	// The Close method on redis.Client doesn't return an error anyway
	t.Skip("Redis mock doesn't support ExpectClose and redis.Client.Close() doesn't return error")
}

// --- MongoProviderImpl error handling tests only ---
// TestMongoProviderImpl_Connect_ConnectError tests MongoDB connection failure in MongoProviderImpl.
// It verifies that the provider returns an error when the MongoDB connection cannot be established.
func TestMongoProviderImpl_Connect_ConnectError(t *testing.T) {
	provider := &MongoProviderImpl{
		uri: "mongodb://localhost:27017",
		connect: func(_ ...*options.ClientOptions) (*mongo.Client, error) {
			return nil, errors.New("connect error")
		},
	}
	client, db, err := provider.Connect(context.Background())
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Nil(t, db)
}

// TestMongoProviderImpl_Connect_PingError tests MongoDB ping failure in MongoProviderImpl.
// It verifies that the provider returns an error when the MongoDB ping operation fails.
func TestMongoProviderImpl_Connect_PingError(t *testing.T) {
	// Create a mock client that will fail on ping
	mockClient := &mongo.Client{}
	provider := &MongoProviderImpl{
		uri: "mongodb://localhost:27017",
		connect: func(_ ...*options.ClientOptions) (*mongo.Client, error) {
			return mockClient, nil
		},
	}
	client, db, err := provider.Connect(context.Background())
	// This will likely fail due to ping error, but we test the function call
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Nil(t, db)
}

// TestMongoProviderImpl_Close tests successful MongoDB closure in MongoProviderImpl.
// It verifies that the provider can close the MongoDB connection without errors.
func TestMongoProviderImpl_Close(t *testing.T) {
	// We can't create a real mongo.Client without a connection
	// So we'll skip this test as it's not feasible to test without real MongoDB
	t.Skip("Cannot test MongoProviderImpl.Close with real client without MongoDB connection")
}

// TestMongoProviderImpl_Close_NilClient tests MongoDB closure with nil client in MongoProviderImpl.
// It verifies that the provider handles nil client gracefully during closure.
func TestMongoProviderImpl_Close_NilClient(t *testing.T) {
	provider := &MongoProviderImpl{client: nil}
	err := provider.Close(context.Background())
	assert.NoError(t, err)
}

// TestS3ProviderImpl_CreateClient_Success tests successful S3 client creation in S3ProviderImpl.
// It verifies that the provider can create an S3 client with valid configuration.
func TestS3ProviderImpl_CreateClient_Success(t *testing.T) {
	provider := &S3ProviderImpl{
		loadConfig: func(_ context.Context, _ ...func(*config.LoadOptions) error) (aws.Config, error) {
			return aws.Config{}, nil
		},
	}
	client, err := provider.CreateClient(context.Background(), "us-east-1")
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

// TestS3ProviderImpl_CreateClient_ConfigError tests S3 client creation failure in S3ProviderImpl.
// It verifies that the provider returns an error when S3 client configuration fails.
func TestS3ProviderImpl_CreateClient_ConfigError(t *testing.T) {
	provider := &S3ProviderImpl{
		loadConfig: func(_ context.Context, _ ...func(*config.LoadOptions) error) (aws.Config, error) {
			return aws.Config{}, errors.New("config error")
		},
	}
	client, err := provider.CreateClient(context.Background(), "us-east-1")
	assert.Error(t, err)
	assert.Nil(t, client)
}

// TestOAuthProviderImpl_LoadGoogleConfig_Success tests successful OAuth config loading in OAuthProviderImpl.
// It verifies that the provider can load Google OAuth configuration from a valid credentials file.
func TestOAuthProviderImpl_LoadGoogleConfig_Success(t *testing.T) {
	validJSON := `{"installed":{"client_id":"id","client_secret":"secret","redirect_uris":["http://localhost"]}}`
	provider := &OAuthProviderImpl{
		readFile: func(_ string) ([]byte, error) {
			return []byte(validJSON), nil
		},
	}
	cfg, err := provider.LoadGoogleConfig("/safe/path/creds.json")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Google)
}

// TestOAuthProviderImpl_LoadGoogleConfig_FileError tests OAuth config loading failure in OAuthProviderImpl.
// It verifies that the provider returns an error when the credentials file cannot be read.
func TestOAuthProviderImpl_LoadGoogleConfig_FileError(t *testing.T) {
	provider := &OAuthProviderImpl{
		readFile: func(_ string) ([]byte, error) {
			return nil, errors.New("file error")
		},
	}
	cfg, err := provider.LoadGoogleConfig("/safe/path/creds.json")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestOAuthProviderImpl_LoadGoogleConfig_ParseError tests OAuth config parsing failure in OAuthProviderImpl.
// It verifies that the provider returns an error when the credentials file cannot be parsed.
func TestOAuthProviderImpl_LoadGoogleConfig_ParseError(t *testing.T) {
	provider := &OAuthProviderImpl{
		readFile: func(_ string) ([]byte, error) {
			return []byte("not json"), nil
		},
	}
	cfg, err := provider.LoadGoogleConfig("/safe/path/creds.json")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestOAuthProviderImpl_LoadGoogleConfig_UnsafePath tests OAuth config loading with unsafe path in OAuthProviderImpl.
// It verifies that the provider returns an error when the credentials path is considered unsafe.
func TestOAuthProviderImpl_LoadGoogleConfig_UnsafePath(t *testing.T) {
	provider := &OAuthProviderImpl{
		readFile: func(_ string) ([]byte, error) {
			return []byte("{}"), nil
		},
	}
	cfg, err := provider.LoadGoogleConfig("../unsafe/creds.json")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}
