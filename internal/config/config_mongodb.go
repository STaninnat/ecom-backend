// Package config provides configuration management, validation, and provider logic for the ecom-backend project.
package config

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// config_mongodb.go: MongoDB connection helpers and legacy patterns.

// ConnectMongoDB establishes a connection to MongoDB using the legacy pattern.
// Uses log.Fatal on errors and is provided for backward compatibility.
// Prefer ConnectMongoDBWithError for better error handling.
func ConnectMongoDB(uri string) (*mongo.Client, *mongo.Database) {
	provider := NewMongoProvider(uri)
	client, db, err := provider.Connect(context.Background())
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	log.Println("Connected to MongoDB...")
	return client, db
}

// ConnectMongoDBWithError establishes a connection to MongoDB and returns the client, database, and error.
// Returns errors instead of calling log.Fatal, making it suitable for testing and graceful error handling.
func ConnectMongoDBWithError(ctx context.Context, uri string) (*mongo.Client, *mongo.Database, error) {
	provider := NewMongoProvider(uri)
	return provider.Connect(ctx)
}

// DisconnectMongoDB safely disconnects from MongoDB and releases associated resources.
// Handles nil clients gracefully and ensures proper cleanup of MongoDB connections.
func (cfg *APIConfig) DisconnectMongoDB(ctx context.Context) error {
	if cfg.MongoClient != nil {
		return cfg.MongoClient.Disconnect(ctx)
	}
	return nil
}
