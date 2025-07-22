// Package mongo provides MongoDB repositories and helpers for the ecom-backend project.
package intmongo

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// integration_test.go: Integration tests for MongoDB repositories and adapters.

// TestContainer holds MongoDB test container
type TestContainer struct {
	Container testcontainers.Container
	URI       string
	Client    *mongo.Client
	Database  *mongo.Database
}

// setupTestContainer creates a MongoDB test container for integration testing.
// Returns a TestContainer with connection details or skips the test if Docker is unavailable.
func setupTestContainer(t *testing.T) *TestContainer {
	ctx := context.Background()

	// Check if Docker is available
	if !isDockerAvailable() {
		t.Skip("Docker not available - skipping integration tests")
	}

	// Create MongoDB container
	container, err := mongodb.Run(ctx, "mongo:7.0",
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForListeningPort("27017/tcp"),
				wait.ForLog("Waiting for connections").WithOccurrence(1),
			).WithDeadline(60*time.Second),
		),
	)
	if err != nil {
		t.Skipf("Failed to create MongoDB container: %v - skipping integration tests", err)
	}

	// Get connection URI
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		err := container.Terminate(ctx)
		if err != nil {
			t.Errorf("container.Terminate failed: %v", err)
		}
		t.Skipf("Failed to get container URI: %v - skipping integration tests", err)
	}

	// Add a small delay to ensure MongoDB is fully ready
	time.Sleep(2 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		err := container.Terminate(ctx)
		if err != nil {
			t.Errorf("container.Terminate failed: %v", err)
		}
		t.Skipf("Failed to connect to MongoDB: %v - skipping integration tests", err)
	}

	// Ping to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		err := client.Disconnect(ctx)
		if err != nil {
			t.Errorf("client.Disconnect failed: %v", err)
		}
		err = container.Terminate(ctx)
		if err != nil {
			t.Errorf("container.Terminate failed: %v", err)
		}
		t.Skipf("Failed to ping MongoDB: %v - skipping integration tests", err)
	}

	// Get database
	database := client.Database("testdb")

	return &TestContainer{
		Container: container,
		URI:       uri,
		Client:    client,
		Database:  database,
	}
}

// isDockerAvailable checks if Docker is available on the system.
// Returns true if Docker is accessible, false otherwise.
func isDockerAvailable() bool {
	// Try to run a simple docker command
	cmd := exec.Command("docker", "ps")
	err := cmd.Run()
	return err == nil
}

// cleanupTestContainer cleans up the test container and disconnects from MongoDB.
// Ensures proper cleanup of resources after integration tests.
func cleanupTestContainer(t *testing.T, tc *TestContainer) {
	if tc != nil {
		if tc.Client != nil {
			err := tc.Client.Disconnect(context.Background())
			assert.NoError(t, err)
		}
		if tc.Container != nil {
			err := tc.Container.Terminate(context.Background())
			assert.NoError(t, err)
		}
	}
}

// TestNewCartMongo_Integration tests the CartMongo constructor with a real MongoDB connection.
// It verifies that the constructor creates a valid instance with a proper collection reference.
func TestNewCartMongo_Integration(t *testing.T) {
	tc := setupTestContainer(t)
	defer cleanupTestContainer(t, tc)

	// Test constructor
	cartMongo := NewCartMongo(tc.Database)
	assert.NotNil(t, cartMongo)
	assert.NotNil(t, cartMongo.Collection)
}

// TestNewReviewMongo_Integration tests the ReviewMongo constructor with a real MongoDB connection.
// It verifies that the constructor creates a valid instance with a proper collection reference.
func TestNewReviewMongo_Integration(t *testing.T) {
	tc := setupTestContainer(t)
	defer cleanupTestContainer(t, tc)

	// Test constructor
	reviewMongo := NewReviewMongo(tc.Database)
	assert.NotNil(t, reviewMongo)
	assert.NotNil(t, reviewMongo.Collection)
}

// TestDatabaseManager_Integration tests the DatabaseManager with a real MongoDB connection.
// It verifies connection establishment, database access, and proper cleanup.
func TestDatabaseManager_Integration(t *testing.T) {
	tc := setupTestContainer(t)
	defer cleanupTestContainer(t, tc)

	// Test DatabaseManager with real connection
	config := &DatabaseConfig{
		URI:            tc.URI,
		DatabaseName:   "testdb",
		ConnectTimeout: 10 * time.Second,
		MaxPoolSize:    10,
		MinPoolSize:    1,
	}

	manager, err := NewDatabaseManager(config)
	require.NoError(t, err)
	defer func() {
		err := manager.Close(context.Background())
		assert.NoError(t, err)
	}()

	// Test GetDatabase
	db := manager.GetDatabase()
	assert.NotNil(t, db)

	// Test GetClient
	client := manager.GetClient()
	assert.NotNil(t, client)

	// Test ping to ensure connection is working
	err = client.Ping(context.Background(), nil)
	assert.NoError(t, err)
}

// TestMongoCollectionAdapter_Integration tests the MongoCollectionAdapter with real MongoDB operations.
// It verifies CRUD operations (InsertOne, FindOne, UpdateOne, DeleteOne) work correctly.
func TestMongoCollectionAdapter_Integration(t *testing.T) {
	tc := setupTestContainer(t)
	defer cleanupTestContainer(t, tc)

	// Test adapter methods with real collection
	collection := tc.Database.Collection("test_collection")
	adapter := &MongoCollectionAdapter{Inner: collection}

	ctx := context.Background()

	// Test InsertOne
	doc := bson.M{"test": "value", "number": 42}
	result, err := adapter.InsertOne(ctx, doc)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test FindOne
	var found bson.M
	singleResult := adapter.FindOne(ctx, bson.M{"test": "value"})
	err = singleResult.Decode(&found)
	assert.NoError(t, err)
	assert.Equal(t, "value", found["test"])
	assert.Equal(t, int32(42), found["number"])

	// Test UpdateOne
	updateResult, err := adapter.UpdateOne(ctx, bson.M{"test": "value"}, bson.M{"$set": bson.M{"updated": true}})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), updateResult.MatchedCount)

	// Test DeleteOne
	deleteResult, err := adapter.DeleteOne(ctx, bson.M{"test": "value"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deleteResult.DeletedCount)
}

// TestMongoCursorAdapter_Integration tests the MongoCursorAdapter with real MongoDB cursor operations.
// It verifies cursor navigation, document decoding, and the All method functionality.
func TestMongoCursorAdapter_Integration(t *testing.T) {
	tc := setupTestContainer(t)
	defer cleanupTestContainer(t, tc)

	collection := tc.Database.Collection("test_cursor")
	ctx := context.Background()

	// Insert test documents
	docs := []any{
		bson.M{"name": "doc1", "value": 1},
		bson.M{"name": "doc2", "value": 2},
		bson.M{"name": "doc3", "value": 3},
	}
	_, err := collection.InsertMany(ctx, docs)
	require.NoError(t, err)

	// Test cursor adapter
	cursor, err := collection.Find(ctx, bson.M{})
	require.NoError(t, err)
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			t.Errorf("cursor.Close failed: %v", err)
		}
	}()

	adapter := &MongoCursorAdapter{Inner: cursor}

	// Test Next and Decode
	var results []bson.M
	for adapter.Next(ctx) {
		var doc bson.M
		err := adapter.Decode(&doc)
		assert.NoError(t, err)
		results = append(results, doc)
	}

	assert.Len(t, results, 3)
	assert.Equal(t, "doc1", results[0]["name"])
	assert.Equal(t, "doc2", results[1]["name"])
	assert.Equal(t, "doc3", results[2]["name"])

	// Test Err
	err = adapter.Err()
	assert.NoError(t, err)

	// Test All method
	cursor2, err := collection.Find(ctx, bson.M{})
	require.NoError(t, err)
	defer func() {
		if err := cursor2.Close(ctx); err != nil {
			t.Errorf("cursor2.Close failed: %v", err)
		}
	}()

	adapter2 := &MongoCursorAdapter{Inner: cursor2}
	var allResults []bson.M
	err = adapter2.All(ctx, &allResults)
	assert.NoError(t, err)
	assert.Len(t, allResults, 3)
}

// TestMongoSingleResultAdapter_Integration tests the MongoSingleResultAdapter with real MongoDB operations.
// It verifies single document retrieval and decoding functionality.
func TestMongoSingleResultAdapter_Integration(t *testing.T) {
	tc := setupTestContainer(t)
	defer cleanupTestContainer(t, tc)

	collection := tc.Database.Collection("test_single_result")
	ctx := context.Background()

	// Insert test document
	doc := bson.M{"name": "test", "value": 123}
	_, err := collection.InsertOne(ctx, doc)
	require.NoError(t, err)

	// Test single result adapter
	result := collection.FindOne(ctx, bson.M{"name": "test"})
	adapter := &MongoSingleResultAdapter{Inner: result}

	var found bson.M
	err = adapter.Decode(&found)
	assert.NoError(t, err)
	assert.Equal(t, "test", found["name"])
	assert.Equal(t, int32(123), found["value"])

	// Test Err
	err = adapter.Err()
	assert.NoError(t, err)
}

// TestCreateIndexes_Integration tests index creation functionality with a real MongoDB database.
// It verifies that cart and review collection indexes are created successfully.
func TestCreateIndexes_Integration(t *testing.T) {
	tc := setupTestContainer(t)
	defer cleanupTestContainer(t, tc)

	// Test index creation
	err := CreateIndexes(tc.Database)
	assert.NoError(t, err)

	// Verify indexes were created
	ctx := context.Background()

	// Check cart indexes
	cartIndexes, err := tc.Database.Collection("carts").Indexes().List(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, cartIndexes)

	// Check review indexes
	reviewIndexes, err := tc.Database.Collection("reviews").Indexes().List(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, reviewIndexes)
}

// TestCartMongo_Integration tests CartMongo operations with a real MongoDB database.
// It verifies cart CRUD operations including adding, updating, removing, and clearing items.
func TestCartMongo_Integration(t *testing.T) {
	tc := setupTestContainer(t)
	defer cleanupTestContainer(t, tc)

	cartMongo := NewCartMongo(tc.Database)
	ctx := context.Background()

	// Test GetCartByUserID with non-existent cart
	cart, err := cartMongo.GetCartByUserID(ctx, "user123")
	assert.NoError(t, err)
	assert.NotNil(t, cart)
	assert.Equal(t, "user123", cart.UserID)
	assert.Empty(t, cart.Items)

	// Test AddItemToCart
	item := models.CartItem{
		ProductID: "product123",
		Quantity:  2,
		Price:     29.99,
	}
	err = cartMongo.AddItemToCart(ctx, "user123", item)
	assert.NoError(t, err)

	// Test GetCartByUserID with existing cart
	cart, err = cartMongo.GetCartByUserID(ctx, "user123")
	assert.NoError(t, err)
	assert.NotNil(t, cart)
	assert.Len(t, cart.Items, 1)
	assert.Equal(t, "product123", cart.Items[0].ProductID)

	// Test UpdateItemQuantity
	err = cartMongo.UpdateItemQuantity(ctx, "user123", "product123", 5)
	assert.NoError(t, err)

	// Verify quantity was updated
	cart, err = cartMongo.GetCartByUserID(ctx, "user123")
	assert.NoError(t, err)
	assert.Equal(t, 5, cart.Items[0].Quantity)

	// Test RemoveItemFromCart
	err = cartMongo.RemoveItemFromCart(ctx, "user123", "product123")
	assert.NoError(t, err)

	// Verify item was removed
	cart, err = cartMongo.GetCartByUserID(ctx, "user123")
	assert.NoError(t, err)
	assert.Empty(t, cart.Items)

	// Test ClearCart
	err = cartMongo.AddItemToCart(ctx, "user123", item)
	assert.NoError(t, err)
	err = cartMongo.ClearCart(ctx, "user123")
	assert.NoError(t, err)

	cart, err = cartMongo.GetCartByUserID(ctx, "user123")
	assert.NoError(t, err)
	assert.Empty(t, cart.Items)

	// Test DeleteCart
	err = cartMongo.DeleteCart(ctx, "user123")
	assert.NoError(t, err)

	// Verify cart was deleted
	cart, err = cartMongo.GetCartByUserID(ctx, "user123")
	assert.NoError(t, err)
	assert.Empty(t, cart.Items) // Should return empty cart for non-existent user
}

// TestReviewMongo_Integration tests ReviewMongo operations with a real MongoDB database.
// It verifies review CRUD operations including creation, retrieval, updates, and statistics.
func TestReviewMongo_Integration(t *testing.T) {
	tc := setupTestContainer(t)
	defer cleanupTestContainer(t, tc)

	reviewMongo := NewReviewMongo(tc.Database)
	ctx := context.Background()

	// Test CreateReview
	review := &models.Review{
		UserID:    "user123",
		ProductID: "product123",
		Rating:    5,
		Comment:   "Great product!",
	}
	err := reviewMongo.CreateReview(ctx, review)
	assert.NoError(t, err)
	assert.NotEmpty(t, review.ID)

	// Test GetReviewByID
	foundReview, err := reviewMongo.GetReviewByID(ctx, review.ID)
	assert.NoError(t, err)
	assert.Equal(t, review.ID, foundReview.ID)
	assert.Equal(t, "user123", foundReview.UserID)
	assert.Equal(t, "product123", foundReview.ProductID)

	// Test GetReviewsByProductID
	reviews, err := reviewMongo.GetReviewsByProductID(ctx, "product123")
	assert.NoError(t, err)
	assert.Len(t, reviews, 1)
	assert.Equal(t, review.ID, reviews[0].ID)

	// Test GetReviewsByUserID
	userReviews, err := reviewMongo.GetReviewsByUserID(ctx, "user123")
	assert.NoError(t, err)
	assert.Len(t, userReviews, 1)
	assert.Equal(t, review.ID, userReviews[0].ID)

	// Test UpdateReviewByID
	updatedReview := &models.Review{
		Rating:  4,
		Comment: "Updated comment",
	}
	err = reviewMongo.UpdateReviewByID(ctx, review.ID, updatedReview)
	assert.NoError(t, err)

	// Verify update
	foundReview, err = reviewMongo.GetReviewByID(ctx, review.ID)
	assert.NoError(t, err)
	assert.Equal(t, 4, foundReview.Rating)
	assert.Equal(t, "Updated comment", foundReview.Comment)

	// Test GetProductRatingStats
	stats, err := reviewMongo.GetProductRatingStats(ctx, "product123")
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, float64(4), stats["averageRating"])
	assert.EqualValues(t, 1, stats["totalReviews"])

	// Test DeleteReviewByID
	err = reviewMongo.DeleteReviewByID(ctx, review.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = reviewMongo.GetReviewByID(ctx, review.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "review not found")
}

// TestPagination_Integration tests pagination functionality with a real MongoDB database.
// It verifies that paginated queries return correct results with proper metadata.
func TestPagination_Integration(t *testing.T) {
	tc := setupTestContainer(t)
	defer cleanupTestContainer(t, tc)

	reviewMongo := NewReviewMongo(tc.Database)
	ctx := context.Background()

	// Create multiple reviews
	for i := 1; i <= 15; i++ {
		review := &models.Review{
			UserID:    "user123",
			ProductID: "product123",
			Rating:    5,
			Comment:   fmt.Sprintf("Review %d", i),
		}
		err := reviewMongo.CreateReview(ctx, review)
		assert.NoError(t, err)
	}

	// Test pagination
	pagination := NewPaginationOptions(1, 10)
	result, err := reviewMongo.GetReviewsByProductIDPaginated(ctx, "product123", pagination)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 10)
	assert.Equal(t, int64(15), result.TotalCount)
	assert.Equal(t, int64(2), result.TotalPages)
	assert.True(t, result.HasNext)
	assert.False(t, result.HasPrev)

	// Test second page
	pagination.Page = 2
	result, err = reviewMongo.GetReviewsByProductIDPaginated(ctx, "product123", pagination)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 5)
	assert.False(t, result.HasNext)
	assert.True(t, result.HasPrev)
}
