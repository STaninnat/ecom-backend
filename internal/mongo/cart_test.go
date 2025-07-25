// Package mongo provides MongoDB repositories and helpers for the ecom-backend project.
package intmongo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/STaninnat/ecom-backend/models"
)

// cart_test.go: Tests for MongoDB cart repository and cart operations.

// MockCollectionInterface for testing
type MockCartCollectionInterface struct {
	mock.Mock
}

// InsertOne mocks the MongoDB InsertOne operation for testing.
// Returns the mocked InsertOneResult and error based on test expectations.
func (m *MockCartCollectionInterface) InsertOne(ctx context.Context, document any) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

// InsertMany mocks the MongoDB InsertMany operation for testing.
// Returns the mocked InsertManyResult and error based on test expectations.
func (m *MockCartCollectionInterface) InsertMany(ctx context.Context, documents []any) (*mongo.InsertManyResult, error) {
	args := m.Called(ctx, documents)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertManyResult), args.Error(1)
}

// Find mocks the MongoDB Find operation for testing.
// Returns a mocked cursor and error based on test expectations.
func (m *MockCartCollectionInterface) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (CursorInterface, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(CursorInterface), args.Error(1)
}

// FindOne mocks the MongoDB FindOne operation for testing.
// Returns a mocked SingleResultInterface for test expectations.
func (m *MockCartCollectionInterface) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) SingleResultInterface {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(SingleResultInterface)
}

// UpdateOne mocks the MongoDB UpdateOne operation for testing.
// Returns the mocked UpdateResult and error based on test expectations.
func (m *MockCartCollectionInterface) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

// UpdateMany mocks the MongoDB UpdateMany operation for testing.
// Returns the mocked UpdateResult and error based on test expectations.
func (m *MockCartCollectionInterface) UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

// DeleteOne mocks the MongoDB DeleteOne operation for testing.
// Returns the mocked DeleteResult and error based on test expectations.
func (m *MockCartCollectionInterface) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

// DeleteMany mocks the MongoDB DeleteMany operation for testing.
// Returns the mocked DeleteResult and error based on test expectations.
func (m *MockCartCollectionInterface) DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

// CountDocuments mocks the MongoDB CountDocuments operation for testing.
// Returns the mocked count and error based on test expectations.
func (m *MockCartCollectionInterface) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(int64), args.Error(1)
}

// Aggregate mocks the MongoDB Aggregate operation for testing.
// Returns a mocked cursor and error based on test expectations.
func (m *MockCartCollectionInterface) Aggregate(ctx context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (CursorInterface, error) {
	args := m.Called(ctx, pipeline, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(CursorInterface), args.Error(1)
}

// Indexes mocks the MongoDB Indexes operation for testing.
// Returns a mocked IndexView for test expectations.
func (m *MockCartCollectionInterface) Indexes() mongo.IndexView {
	args := m.Called()
	return args.Get(0).(mongo.IndexView)
}

// MockCursor for testing
type MockCartCursor struct {
	mock.Mock
}

// Next mocks the cursor Next operation for testing.
// Returns a boolean indicating if more documents are available.
func (m *MockCartCursor) Next(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

// Decode mocks the cursor Decode operation for testing.
// Decodes the current document into the provided value.
func (m *MockCartCursor) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}

// All mocks the cursor All operation for testing.
// Decodes all remaining documents into the provided results slice.
func (m *MockCartCursor) All(ctx context.Context, results any) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

// Close mocks the cursor Close operation for testing.
// Closes the cursor and returns any error that occurred.
func (m *MockCartCursor) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Err mocks the cursor Err operation for testing.
// Returns any error that occurred during cursor operations.
func (m *MockCartCursor) Err() error {
	args := m.Called()
	return args.Error(0)
}

// MockSingleResult for testing
type MockCartSingleResult struct {
	mock.Mock
}

// Decode mocks the SingleResult Decode operation for testing.
// Decodes the single document into the provided value.
func (m *MockCartSingleResult) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}

// Err mocks the SingleResult Err operation for testing.
// Returns any error that occurred during the find operation.
func (m *MockCartSingleResult) Err() error {
	args := m.Called()
	return args.Error(0)
}

// TestGetCartByUserID tests the GetCartByUserID function with various scenarios.
// It verifies successful cart retrieval, handling of non-existent carts, and database errors.
func TestGetCartByUserID(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name         string
		userID       string
		setupMock    func()
		expectError  bool
		expectedCart *models.Cart
	}{
		{
			name:   "existing cart should be returned",
			userID: "user123",
			setupMock: func() {
				mockResult := &MockCartSingleResult{}
				mockResult.On("Decode", mock.AnythingOfType("*models.Cart")).Run(func(args mock.Arguments) {
					cart := args.Get(0).(*models.Cart)
					cart.UserID = "user123"
					cart.Items = []models.CartItem{}
					cart.CreatedAt = time.Now().UTC()
					cart.UpdatedAt = time.Now().UTC()
				}).Return(nil)
				mockResult.On("Err").Return(nil)
				mockCollection.On("FindOne", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(mockResult).Once()
			},
			expectError:  false,
			expectedCart: &models.Cart{UserID: "user123"},
		},
		{
			name:   "non-existent cart should return empty cart",
			userID: "user123",
			setupMock: func() {
				mockResult := &MockCartSingleResult{}
				mockResult.On("Decode", mock.AnythingOfType("*models.Cart")).Return(mongo.ErrNoDocuments)
				mockResult.On("Err").Return(mongo.ErrNoDocuments)
				mockCollection.On("FindOne", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(mockResult).Once()
			},
			expectError: false,
			expectedCart: &models.Cart{
				UserID:    "user123",
				Items:     []models.CartItem{},
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
		},
		{
			name:   "database error should be returned",
			userID: "user123",
			setupMock: func() {
				mockResult := &MockCartSingleResult{}
				mockResult.On("Decode", mock.AnythingOfType("*models.Cart")).Return(assert.AnError)
				mockResult.On("Err").Return(assert.AnError)
				// Use more specific matching to distinguish from success case
				mockCollection.On("FindOne", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(mockResult).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			cart, err := cartMongo.GetCartByUserID(ctx, tt.userID)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, cart)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, cart)
				assert.Equal(t, tt.userID, cart.UserID)
				if tt.expectedCart.Items != nil {
					assert.Len(t, cart.Items, len(tt.expectedCart.Items))
				}
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestGetCartsByUserIDs tests the GetCartsByUserIDs function with multiple user IDs.
// It verifies successful retrieval of multiple carts and database error handling.
func TestGetCartsByUserIDs(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		userIDs     []string
		setupMock   func()
		expectError bool
		expectedLen int
	}{
		{
			name:        "empty user IDs should return empty slice",
			userIDs:     []string{},
			setupMock:   func() {},
			expectError: false,
			expectedLen: 0,
		},
		{
			name:    "valid user IDs should return carts",
			userIDs: []string{"user1", "user2"},
			setupMock: func() {
				mockCursor := &MockCartCursor{}
				mockCursor.On("Close", ctx).Return(nil)
				mockCursor.On("All", ctx, mock.AnythingOfType("*[]*models.Cart")).Run(func(args mock.Arguments) {
					carts := args.Get(1).(*[]*models.Cart)
					*carts = []*models.Cart{
						{UserID: "user1", Items: []models.CartItem{}},
						{UserID: "user2", Items: []models.CartItem{}},
					}
				}).Return(nil)
				mockCollection.On("Find", ctx, bson.M{"user_id": bson.M{"$in": []string{"user1", "user2"}}}, mock.Anything).Return(mockCursor, nil).Once()
			},
			expectError: false,
			expectedLen: 2,
		},
		{
			name:    "database error should be returned",
			userIDs: []string{"user1", "user2"},
			setupMock: func() {
				mockCollection.On("Find", ctx, bson.M{"user_id": bson.M{"$in": []string{"user1", "user2"}}}, mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			carts, err := cartMongo.GetCartsByUserIDs(ctx, tt.userIDs)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, carts)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, carts)
				assert.Len(t, carts, tt.expectedLen)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestGetCartsByUserIDs_EmptySlice tests GetCartsByUserIDs with an empty user IDs slice.
// It verifies that an empty slice returns an empty result without errors.
func TestGetCartsByUserIDs_EmptySlice(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	// Test with empty userIDs slice
	result, err := cartMongo.GetCartsByUserIDs(ctx, []string{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

// TestGetCartsByUserIDs_DatabaseError tests GetCartsByUserIDs when database operations fail.
// It verifies proper error handling when cursor operations return errors.
func TestGetCartsByUserIDs_DatabaseError(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCursor := &MockCartCursor{}
	mockCursor.On("Close", ctx).Return(nil)
	mockCursor.On("All", ctx, mock.AnythingOfType("*[]*models.Cart")).Return(assert.AnError)

	mockCollection.On("Find", ctx, bson.M{"user_id": bson.M{"$in": []string{"user1", "user2"}}}, mock.Anything).Return(mockCursor, nil)

	result, err := cartMongo.GetCartsByUserIDs(ctx, []string{"user1", "user2"})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to decode carts")
}

// TestGetCartsByUserIDs_FindError tests GetCartsByUserIDs when the Find operation fails.
// It verifies proper error handling when the initial database query fails.
func TestGetCartsByUserIDs_FindError(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("Find", ctx, bson.M{"user_id": bson.M{"$in": []string{"user1"}}}, mock.Anything).Return(nil, assert.AnError)

	result, err := cartMongo.GetCartsByUserIDs(ctx, []string{"user1"})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find carts")
}

// TestAddItemToCart tests the AddItemToCart function with various scenarios.
// It verifies successful item addition and database error handling.
func TestAddItemToCart(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	item := models.CartItem{
		ProductID: "product123",
		Quantity:  2,
		Price:     10.99,
		Name:      "Test Product",
	}

	tests := []struct {
		name        string
		userID      string
		item        models.CartItem
		setupMock   func()
		expectError bool
	}{
		{
			name:   "valid item should be added",
			userID: "user123",
			item:   item,
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{}, nil).Once()
			},
			expectError: false,
		},
		{
			name:   "database error should be returned",
			userID: "user123",
			item:   item,
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runCartItemTest(
				t,
				cartMongo.AddItemToCart,
				ctx,
				tt.userID,
				tt.item,
				tt.setupMock,
				tt.expectError,
			)
			mockCollection.AssertExpectations(t)
		})
	}
}

// TestAddItemsToCart tests the AddItemsToCart function with multiple items.
// It verifies successful addition of multiple items and database error handling.
func TestAddItemsToCart(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	items := []models.CartItem{
		{ProductID: "product1", Quantity: 1, Price: 10.99, Name: "Product 1"},
		{ProductID: "product2", Quantity: 2, Price: 20.99, Name: "Product 2"},
	}

	tests := []struct {
		name        string
		userID      string
		items       []models.CartItem
		setupMock   func()
		expectError bool
	}{
		{
			name:        "empty items should return error",
			userID:      "user123",
			items:       []models.CartItem{},
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:   "valid items should be added",
			userID: "user123",
			items:  items,
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{}, nil).Once()
			},
			expectError: false,
		},
		{
			name:   "database error should be returned",
			userID: "user123",
			items:  items,
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := cartMongo.AddItemsToCart(ctx, tt.userID, tt.items)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestRemoveItemFromCart tests the RemoveItemFromCart function with various scenarios.
// It verifies successful item removal and database error handling.
func TestRemoveItemFromCart(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      string
		productID   string
		setupMock   func()
		expectError bool
	}{
		{
			name:      "valid removal should succeed",
			userID:    "user123",
			productID: "product123",
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{}, nil).Once()
			},
			expectError: false,
		},
		{
			name:      "database error should be returned",
			userID:    "user123",
			productID: "product123",
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := cartMongo.RemoveItemFromCart(ctx, tt.userID, tt.productID)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestRemoveItemsFromCart tests the RemoveItemsFromCart function with multiple product IDs.
// It verifies successful removal of multiple items and database error handling.
func TestRemoveItemsFromCart(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      string
		productIDs  []string
		setupMock   func()
		expectError bool
	}{
		{
			name:        "empty product IDs should return error",
			userID:      "user123",
			productIDs:  []string{},
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:       "valid removal should succeed",
			userID:     "user123",
			productIDs: []string{"product1", "product2"},
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{}, nil).Once()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := cartMongo.RemoveItemsFromCart(ctx, tt.userID, tt.productIDs)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestRemoveItemsFromCart_EmptyProductIDs tests RemoveItemsFromCart with empty product IDs.
// It verifies that empty product IDs are handled gracefully without errors.
func TestRemoveItemsFromCart_EmptyProductIDs(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	err := cartMongo.RemoveItemsFromCart(ctx, "user123", []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "product IDs slice cannot be empty")
}

// TestRemoveItemsFromCart_DatabaseError tests RemoveItemsFromCart when database operations fail.
// It verifies proper error handling when the update operation fails.
func TestRemoveItemsFromCart_DatabaseError(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	err := cartMongo.RemoveItemsFromCart(ctx, "user123", []string{"product1", "product2"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove items from cart")
}

// TestClearCart tests the ClearCart function with various scenarios.
// It verifies successful cart clearing and database error handling.
func TestClearCart(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      string
		setupMock   func()
		expectError bool
	}{
		{
			name:   "valid clear should succeed",
			userID: "user123",
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{}, nil).Once()
			},
			expectError: false,
		},
		{
			name:   "database error should be returned",
			userID: "user123",
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := cartMongo.ClearCart(ctx, tt.userID)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestClearCarts tests the ClearCarts function with multiple user IDs.
// It verifies successful clearing of multiple carts and database error handling.
func TestClearCarts(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		userIDs     []string
		setupMock   func()
		expectError bool
	}{
		{
			name:        "empty user IDs should return error",
			userIDs:     []string{},
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:    "valid clear should succeed",
			userIDs: []string{"user1", "user2"},
			setupMock: func() {
				mockCollection.On("UpdateMany", ctx, bson.M{"user_id": bson.M{"$in": []string{"user1", "user2"}}}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{}, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := cartMongo.ClearCarts(ctx, tt.userIDs)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestClearCarts_EmptyUserIDs tests ClearCarts with empty user IDs slice.
// It verifies that empty user IDs are handled gracefully without errors.
func TestClearCarts_EmptyUserIDs(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	err := cartMongo.ClearCarts(ctx, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user IDs slice cannot be empty")
}

// TestClearCarts_DatabaseError tests ClearCarts when database operations fail.
// It verifies proper error handling when the update operation fails.
func TestClearCarts_DatabaseError(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("UpdateMany", ctx, bson.M{"user_id": bson.M{"$in": []string{"user1", "user2"}}}, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	err := cartMongo.ClearCarts(ctx, []string{"user1", "user2"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clear carts")
}

// TestUpdateItemQuantity tests the UpdateItemQuantity function with various scenarios.
// It verifies successful quantity updates and database error handling.
func TestUpdateItemQuantity(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      string
		productID   string
		quantity    int
		setupMock   func()
		expectError bool
	}{
		{
			name:      "positive quantity should update",
			userID:    "user123",
			productID: "product123",
			quantity:  5,
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123", "items.product_id": "product123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1}, nil).Once()
			},
			expectError: false,
		},
		{
			name:      "zero quantity should remove item",
			userID:    "user123",
			productID: "product123",
			quantity:  0,
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123", "items.product_id": "product123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1}, nil).Once()
			},
			expectError: false,
		},
		{
			name:      "negative quantity should remove item",
			userID:    "user123",
			productID: "product123",
			quantity:  -1,
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123", "items.product_id": "product123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1}, nil).Once()
			},
			expectError: false,
		},
		{
			name:      "item not found should return error",
			userID:    "user123",
			productID: "product123",
			quantity:  5,
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123", "items.product_id": "product123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 0}, nil).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := cartMongo.UpdateItemQuantity(ctx, tt.userID, tt.productID, tt.quantity)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestUpdateItemQuantity_ZeroQuantity tests UpdateItemQuantity with zero quantity.
// It verifies that zero quantity removes the item from the cart.
func TestUpdateItemQuantity_ZeroQuantity(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockResult := &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}
	mockCollection.On("UpdateOne", ctx, bson.M{
		"user_id":          "user123",
		"items.product_id": "product123",
	}, bson.M{
		"$pull": bson.M{
			"items": bson.M{"product_id": "product123"},
		},
	}, mock.Anything).Return(mockResult, nil)

	err := cartMongo.UpdateItemQuantity(ctx, "user123", "product123", 0)

	assert.NoError(t, err)
}

// TestUpdateItemQuantity_NegativeQuantity tests UpdateItemQuantity with negative quantity.
// It verifies that negative quantities are handled appropriately.
func TestUpdateItemQuantity_NegativeQuantity(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockResult := &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}
	mockCollection.On("UpdateOne", ctx, bson.M{
		"user_id":          "user123",
		"items.product_id": "product123",
	}, bson.M{
		"$pull": bson.M{
			"items": bson.M{"product_id": "product123"},
		},
	}, mock.Anything).Return(mockResult, nil)

	err := cartMongo.UpdateItemQuantity(ctx, "user123", "product123", -1)

	assert.NoError(t, err)
}

// TestUpdateItemQuantity_ItemNotFound tests UpdateItemQuantity when item is not found.
// It verifies proper handling when the item doesn't exist in the cart.
func TestUpdateItemQuantity_ItemNotFound(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockResult := &mongo.UpdateResult{MatchedCount: 0, ModifiedCount: 0}
	mockCollection.On("UpdateOne", ctx, bson.M{
		"user_id":          "user123",
		"items.product_id": "product123",
	}, mock.Anything, mock.Anything).Return(mockResult, nil)

	err := cartMongo.UpdateItemQuantity(ctx, "user123", "product123", 5)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "item not found in cart")
}

// TestUpdateItemQuantity_DatabaseError tests UpdateItemQuantity when database operations fail.
// It verifies proper error handling when the update operation fails.
func TestUpdateItemQuantity_DatabaseError(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("UpdateOne", ctx, bson.M{
		"user_id":          "user123",
		"items.product_id": "product123",
	}, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	err := cartMongo.UpdateItemQuantity(ctx, "user123", "product123", 5)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update item quantity")
}

// TestUpdateItemQuantities tests the UpdateItemQuantities function with multiple updates.
// It verifies successful batch quantity updates and database error handling.
func TestUpdateItemQuantities(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      string
		updates     map[string]int
		setupMock   func()
		expectError bool
	}{
		{
			name:        "empty updates should return error",
			userID:      "user123",
			updates:     map[string]int{},
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:   "valid updates should succeed",
			userID: "user123",
			updates: map[string]int{
				"product1": 5,
				"product2": 0,
				"product3": 3,
			},
			setupMock: func() {
				// Mock UpdateOne calls with flexible expectations
				mockCollection.On("UpdateOne", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1}, nil).Times(3)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := cartMongo.UpdateItemQuantities(ctx, tt.userID, tt.updates)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestUpdateItemQuantities_IndividualError tests UpdateItemQuantities when individual updates fail.
// It verifies proper error handling when some updates succeed and others fail.
func TestUpdateItemQuantities_IndividualError(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	// First call succeeds, second call fails
	mockResult1 := &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}
	mockResult2 := &mongo.UpdateResult{MatchedCount: 0, ModifiedCount: 0}

	mockCollection.On("UpdateOne", ctx, bson.M{
		"user_id":          "user123",
		"items.product_id": "product1",
	}, mock.Anything, mock.Anything).Return(mockResult1, nil)

	mockCollection.On("UpdateOne", ctx, bson.M{
		"user_id":          "user123",
		"items.product_id": "product2",
	}, mock.Anything, mock.Anything).Return(mockResult2, nil)

	updates := map[string]int{
		"product1": 5,
		"product2": 3,
	}

	err := cartMongo.UpdateItemQuantities(ctx, "user123", updates)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update item product2")
}

// TestUpdateItemQuantities_EmptyUpdates tests UpdateItemQuantities with empty updates slice.
// It verifies that empty updates are handled gracefully without errors.
func TestUpdateItemQuantities_EmptyUpdates(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	err := cartMongo.UpdateItemQuantities(ctx, "user123", map[string]int{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "updates map cannot be empty")
}

// TestCart_AddItemAndUpsertCart tests the AddItemToCart and UpsertCart functions with various scenarios.
// It verifies successful item addition and cart upsertion and database error handling.
func TestCart_AddItemAndUpsertCart(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	item := models.CartItem{ProductID: "product1", Quantity: 1, Price: 10.99, Name: "Product 1"}
	cart := models.Cart{UserID: "user123", Items: []models.CartItem{item}}

	tests := []struct {
		name        string
		userID      string
		item        *models.CartItem
		cart        *models.Cart
		callFunc    func(ctx context.Context, userID string, item *models.CartItem, cart *models.Cart) error
		setupMock   func()
		expectError bool
	}{
		{
			name:   "AddItemToCart: valid item should be added",
			userID: "user123",
			item:   &item,
			callFunc: func(ctx context.Context, userID string, item *models.CartItem, _ *models.Cart) error {
				return cartMongo.AddItemToCart(ctx, userID, *item)
			},
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{}, nil).Once()
			},
			expectError: false,
		},
		{
			name:   "AddItemToCart: database error should be returned",
			userID: "user123",
			item:   &item,
			callFunc: func(ctx context.Context, userID string, item *models.CartItem, _ *models.Cart) error {
				return cartMongo.AddItemToCart(ctx, userID, *item)
			},
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectError: true,
		},
		{
			name:   "UpsertCart: valid cart should be upserted",
			userID: "user123",
			cart:   &cart,
			callFunc: func(ctx context.Context, userID string, _ *models.CartItem, cart *models.Cart) error {
				return cartMongo.UpsertCart(ctx, userID, *cart)
			},
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{}, nil).Once()
			},
			expectError: false,
		},
		{
			name:   "UpsertCart: database error should be returned",
			userID: "user123",
			cart:   &cart,
			callFunc: func(ctx context.Context, userID string, _ *models.CartItem, cart *models.Cart) error {
				return cartMongo.UpsertCart(ctx, userID, *cart)
			},
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := tt.callFunc(ctx, tt.userID, tt.item, tt.cart)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mockCollection.AssertExpectations(t)
		})
	}
}

// runCartItemTest is a shared helper for testing AddItemToCart and UpsertCart.
func runCartItemTest(
	t *testing.T,
	testFunc func(ctx context.Context, userID string, item models.CartItem) error,
	ctx context.Context,
	userID string,
	item models.CartItem,
	setupMock func(),
	expectError bool,
) {
	setupMock()
	err := testFunc(ctx, userID, item)
	if expectError {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
	}
}

// TestGetCartStats tests the GetCartStats function with various scenarios.
// It verifies successful retrieval of cart statistics and database error handling.
func TestGetCartStats(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		setupMock   func()
		expectError bool
		expected    map[string]any
	}{
		{
			name: "valid stats should be returned",
			setupMock: func() {
				mockCursor := &MockCartCursor{}
				mockCursor.On("Close", ctx).Return(nil)
				mockCursor.On("All", ctx, mock.AnythingOfType("*[]map[string]interface {}")).Return(nil)
				mockCollection.On("Aggregate", ctx, mock.AnythingOfType("[]bson.M"), mock.Anything).Return(mockCursor, nil).Once()
			},
			expectError: false,
			expected: map[string]any{
				"totalCarts":      0,
				"totalItems":      0,
				"avgItemsPerCart": 0.0,
			},
		},
		{
			name: "database error should be returned",
			setupMock: func() {
				mockCollection.On("Aggregate", ctx, mock.AnythingOfType("[]bson.M"), mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			stats, err := cartMongo.GetCartStats(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, stats)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, stats)
				if tt.expected != nil {
					assert.Equal(t, tt.expected["totalCarts"], stats["totalCarts"])
					assert.Equal(t, tt.expected["totalItems"], stats["totalItems"])
					assert.Equal(t, tt.expected["avgItemsPerCart"], stats["avgItemsPerCart"])
				}
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestGetCartStats_DatabaseError tests GetCartStats when database operations fail.
// It verifies proper error handling when the aggregate operation fails.
func TestGetCartStats_DatabaseError(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("Aggregate", ctx, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	result, err := cartMongo.GetCartStats(ctx)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to aggregate cart stats")
}

// TestGetCartStats_DecodeError tests GetCartStats when document decoding fails.
// It verifies proper error handling when cursor decoding operations fail.
func TestGetCartStats_DecodeError(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCursor := &MockCartCursor{}
	mockCursor.On("Close", ctx).Return(nil)
	mockCursor.On("All", ctx, mock.AnythingOfType("*[]map[string]interface {}")).Return(assert.AnError)

	mockCollection.On("Aggregate", ctx, mock.Anything, mock.Anything).Return(mockCursor, nil)

	result, err := cartMongo.GetCartStats(ctx)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to decode aggregation results")
}

// TestGetCartStats_EmptyResults tests GetCartStats when no results are returned.
// It verifies proper handling when the aggregation returns no documents.
func TestGetCartStats_EmptyResults(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCursor := &MockCartCursor{}
	mockCursor.On("Close", ctx).Return(nil)
	mockCursor.On("All", ctx, mock.AnythingOfType("*[]map[string]interface {}")).Run(func(args mock.Arguments) {
		results := args.Get(1).(*[]map[string]any)
		*results = []map[string]any{} // Empty results
	}).Return(nil)

	mockCollection.On("Aggregate", ctx, mock.Anything, mock.Anything).Return(mockCursor, nil)

	result, err := cartMongo.GetCartStats(ctx)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.EqualValues(t, int64(0), result["totalCarts"])
	assert.EqualValues(t, int64(0), result["totalItems"])
	assert.InDelta(t, 0.0, result["avgItemsPerCart"], 0.000001)
}

// TestDeleteCart tests the DeleteCart function with various scenarios.
// It verifies successful cart deletion and database error handling.
func TestDeleteCart(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      string
		setupMock   func()
		expectError bool
	}{
		{
			name:   "valid delete should succeed",
			userID: "user123",
			setupMock: func() {
				mockCollection.On("DeleteOne", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(&mongo.DeleteResult{DeletedCount: 1}, nil).Once()
			},
			expectError: false,
		},
		{
			name:   "cart not found should return error",
			userID: "user123",
			setupMock: func() {
				mockCollection.On("DeleteOne", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(&mongo.DeleteResult{DeletedCount: 0}, nil).Once()
			},
			expectError: true,
		},
		{
			name:   "database error should be returned",
			userID: "user123",
			setupMock: func() {
				mockCollection.On("DeleteOne", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := cartMongo.DeleteCart(ctx, tt.userID)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestDeleteCart_NotFound tests DeleteCart when the cart is not found.
// It verifies proper handling when the cart doesn't exist in the database.
func TestDeleteCart_NotFound(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockResult := &mongo.DeleteResult{DeletedCount: 0}
	mockCollection.On("DeleteOne", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(mockResult, nil)

	err := cartMongo.DeleteCart(ctx, "user123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cart not found")
}

// TestDeleteCart_DatabaseError tests DeleteCart when database operations fail.
// It verifies proper error handling when the delete operation fails.
func TestDeleteCart_DatabaseError(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("DeleteOne", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(nil, assert.AnError)

	err := cartMongo.DeleteCart(ctx, "user123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete cart")
}

// TestNewCartMongo tests the NewCartMongo constructor function.
// It verifies that the CartMongo instance is created correctly with the provided collection.
func TestNewCartMongo(t *testing.T) {
	// This test is now covered by integration tests
	// Constructor logic is tested in TestNewCartMongo_Integration
	t.Skip("Covered by integration tests")
}
