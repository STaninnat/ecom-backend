package intmongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// =====================
// Database Configuration
// =====================

// DatabaseConfig holds MongoDB configuration settings.
type DatabaseConfig struct {
	URI            string
	DatabaseName   string
	ConnectTimeout time.Duration
	MaxPoolSize    uint64
	MinPoolSize    uint64
}

// DefaultDatabaseConfig returns default MongoDB configuration.
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		URI:            "mongodb://localhost:27017",
		DatabaseName:   "ecom",
		ConnectTimeout: 10 * time.Second,
		MaxPoolSize:    100,
		MinPoolSize:    5,
	}
}

// DatabaseManager manages MongoDB connections and provides database access.
type DatabaseManager struct {
	client   *mongo.Client
	database *mongo.Database
	config   *DatabaseConfig
}

// NewDatabaseManager creates a new database manager with the given configuration.
func NewDatabaseManager(config *DatabaseConfig) (*DatabaseManager, error) {
	if config == nil {
		config = DefaultDatabaseConfig()
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(config.URI).
		SetMaxPoolSize(config.MaxPoolSize).
		SetMinPoolSize(config.MinPoolSize).
		SetMaxConnIdleTime(30 * time.Minute)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("mongo connect error: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongo ping error: %w", err)
	}

	database := client.Database(config.DatabaseName)

	return &DatabaseManager{
		client:   client,
		database: database,
		config:   config,
	}, nil
}

// GetDatabase returns the MongoDB database instance.
func (dm *DatabaseManager) GetDatabase() *mongo.Database {
	return dm.database
}

// GetClient returns the MongoDB client instance.
func (dm *DatabaseManager) GetClient() *mongo.Client {
	return dm.client
}

// Close closes the database connection.
func (dm *DatabaseManager) Close(ctx context.Context) error {
	return dm.client.Disconnect(ctx)
}

// =====================
// Pagination Support
// =====================

// PaginationOptions holds pagination parameters for MongoDB queries.
type PaginationOptions struct {
	Page     int64
	PageSize int64
	Sort     map[string]any
	Filter   map[string]any // Added for flexible filtering
}

// NewPaginationOptions creates pagination options with defaults.
func NewPaginationOptions(page, pageSize int64) *PaginationOptions {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return &PaginationOptions{
		Page:     page,
		PageSize: pageSize,
		Sort:     map[string]any{"created_at": -1}, // Default sort by creation date desc
	}
}

// PaginatedResult holds paginated query results.
type PaginatedResult[T any] struct {
	Data       []T
	TotalCount int64
	Page       int64
	PageSize   int64
	TotalPages int64
	HasNext    bool
	HasPrev    bool
}

// =====================
// MongoDB Abstractions
// =====================

// CollectionInterface abstracts mongo.Collection for testability.
type CollectionInterface interface {
	InsertOne(ctx context.Context, document any) (*mongo.InsertOneResult, error)
	InsertMany(ctx context.Context, documents []any) (*mongo.InsertManyResult, error)
	Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (CursorInterface, error)
	FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) SingleResultInterface
	UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error)
	UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error)
	DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error)
	CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error)
	Aggregate(ctx context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (CursorInterface, error)
	Indexes() mongo.IndexView
}

// CursorInterface abstracts mongo.Cursor for testability.
type CursorInterface interface {
	Next(ctx context.Context) bool
	Decode(val any) error
	All(ctx context.Context, results any) error
	Close(ctx context.Context) error
	Err() error
}

// SingleResultInterface abstracts mongo.SingleResult for testability.
type SingleResultInterface interface {
	Decode(val any) error
	Err() error
}

// =====================
// Adapters for MongoDB
// =====================

// MongoCollectionAdapter adapts mongo.Collection to CollectionInterface.
type MongoCollectionAdapter struct {
	Inner *mongo.Collection
}

func (m *MongoCollectionAdapter) InsertOne(ctx context.Context, doc any) (*mongo.InsertOneResult, error) {
	return m.Inner.InsertOne(ctx, doc)
}

func (m *MongoCollectionAdapter) InsertMany(ctx context.Context, docs []any) (*mongo.InsertManyResult, error) {
	return m.Inner.InsertMany(ctx, docs)
}

func (m *MongoCollectionAdapter) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (CursorInterface, error) {
	cursor, err := m.Inner.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return &MongoCursorAdapter{Inner: cursor}, nil
}

func (m *MongoCollectionAdapter) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) SingleResultInterface {
	result := m.Inner.FindOne(ctx, filter, opts...)
	return &MongoSingleResultAdapter{Inner: result}
}

func (m *MongoCollectionAdapter) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	return m.Inner.UpdateOne(ctx, filter, update, opts...)
}

func (m *MongoCollectionAdapter) UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	return m.Inner.UpdateMany(ctx, filter, update, opts...)
}

func (m *MongoCollectionAdapter) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	return m.Inner.DeleteOne(ctx, filter, opts...)
}

func (m *MongoCollectionAdapter) DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	return m.Inner.DeleteMany(ctx, filter, opts...)
}

func (m *MongoCollectionAdapter) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	return m.Inner.CountDocuments(ctx, filter, opts...)
}

func (m *MongoCollectionAdapter) Aggregate(ctx context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (CursorInterface, error) {
	cursor, err := m.Inner.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return nil, err
	}
	return &MongoCursorAdapter{Inner: cursor}, nil
}

func (m *MongoCollectionAdapter) Indexes() mongo.IndexView {
	return m.Inner.Indexes()
}

// MongoCursorAdapter adapts mongo.Cursor to CursorInterface.
type MongoCursorAdapter struct {
	Inner *mongo.Cursor
}

func (c *MongoCursorAdapter) Next(ctx context.Context) bool { return c.Inner.Next(ctx) }
func (c *MongoCursorAdapter) Decode(val any) error          { return c.Inner.Decode(val) }
func (c *MongoCursorAdapter) All(ctx context.Context, results any) error {
	return c.Inner.All(ctx, results)
}
func (c *MongoCursorAdapter) Close(ctx context.Context) error { return c.Inner.Close(ctx) }
func (c *MongoCursorAdapter) Err() error                      { return c.Inner.Err() }

// MongoSingleResultAdapter adapts mongo.SingleResult to SingleResultInterface.
type MongoSingleResultAdapter struct {
	Inner *mongo.SingleResult
}

func (r *MongoSingleResultAdapter) Decode(val any) error { return r.Inner.Decode(val) }
func (r *MongoSingleResultAdapter) Err() error           { return r.Inner.Err() }

// Factory functions for easy mocking
var (
	NewCursorAdapter       = func(cur *mongo.Cursor) CursorInterface { return &MongoCursorAdapter{Inner: cur} }
	NewSingleResultAdapter = func(res *mongo.SingleResult) SingleResultInterface { return &MongoSingleResultAdapter{Inner: res} }
)

// =====================
// Index Management
// =====================

// CreateIndexes creates necessary indexes for optimal performance
func CreateIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Cart indexes
	cartCollection := db.Collection("carts")
	_, err := cartCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("cart index error: %w", err)
	}

	// Review indexes
	reviewCollection := db.Collection("reviews")
	_, err = reviewCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "product_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
	})
	if err != nil {
		return fmt.Errorf("review index error: %w", err)
	}

	_, err = reviewCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
		},
	})
	if err != nil {
		return fmt.Errorf("review index error: %w", err)
	}

	return nil
}
