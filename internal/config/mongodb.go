package config

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ConnectMongoDB(uri string) (*mongo.Client, *mongo.Database) {
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Error pinging MongoDB:", err)
	}
	log.Println("Connected to MongoDB...")

	db := client.Database("ecommerce_db")
	return client, db
}

func (cfg *APIConfig) DisconnectMongoDB(ctx context.Context) error {
	if cfg.MongoClient != nil {
		return cfg.MongoClient.Disconnect(ctx)
	}
	return nil
}
