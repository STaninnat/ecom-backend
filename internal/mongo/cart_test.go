package intmongo

import (
	"context"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MockCollectionInterface for testing
type MockCartCollectionInterface struct {
	mock.Mock
}

func (m *MockCartCollectionInterface) InsertOne(ctx context.Context, document any) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCartCollectionInterface) InsertMany(ctx context.Context, documents []any) (*mongo.InsertManyResult, error) {
	args := m.Called(ctx, documents)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertManyResult), args.Error(1)
}

func (m *MockCartCollectionInterface) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (CursorInterface, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(CursorInterface), args.Error(1)
}

func (m *MockCartCollectionInterface) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) SingleResultInterface {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(SingleResultInterface)
}

func (m *MockCartCollectionInterface) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockCartCollectionInterface) UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockCartCollectionInterface) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCartCollectionInterface) DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCartCollectionInterface) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCartCollectionInterface) Aggregate(ctx context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (CursorInterface, error) {
	args := m.Called(ctx, pipeline, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(CursorInterface), args.Error(1)
}

func (m *MockCartCollectionInterface) Indexes() mongo.IndexView {
	args := m.Called()
	return args.Get(0).(mongo.IndexView)
}

// MockCursor for testing
type MockCartCursor struct {
	mock.Mock
}

func (m *MockCartCursor) Next(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockCartCursor) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}

func (m *MockCartCursor) All(ctx context.Context, results any) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

func (m *MockCartCursor) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCartCursor) Err() error {
	args := m.Called()
	return args.Error(0)
}

// MockSingleResult for testing
type MockCartSingleResult struct {
	mock.Mock
}

func (m *MockCartSingleResult) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}

func (m *MockCartSingleResult) Err() error {
	args := m.Called()
	return args.Error(0)
}

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
				assert.Error(t, err)
				assert.Nil(t, cart)
			} else {
				assert.NoError(t, err)
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
				assert.Error(t, err)
				assert.Nil(t, carts)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, carts)
				assert.Len(t, carts, tt.expectedLen)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

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
			tt.setupMock()

			err := cartMongo.AddItemToCart(ctx, tt.userID, tt.item)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

func TestRemoveItemsFromCart_EmptyProductIDs(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	err := cartMongo.RemoveItemsFromCart(ctx, "user123", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product IDs slice cannot be empty")
}

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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

func TestClearCarts_EmptyUserIDs(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	err := cartMongo.ClearCarts(ctx, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user IDs slice cannot be empty")
}

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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

func TestUpdateItemQuantities_EmptyUpdates(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	err := cartMongo.UpdateItemQuantities(ctx, "user123", map[string]int{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updates map cannot be empty")
}

func TestUpsertCart(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	cart := models.Cart{
		UserID: "user123",
		Items: []models.CartItem{
			{ProductID: "product1", Quantity: 1, Price: 10.99, Name: "Product 1"},
		},
	}

	tests := []struct {
		name        string
		userID      string
		cart        models.Cart
		setupMock   func()
		expectError bool
	}{
		{
			name:   "valid cart should be upserted",
			userID: "user123",
			cart:   cart,
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{}, nil).Once()
			},
			expectError: false,
		},
		{
			name:   "database error should be returned",
			userID: "user123",
			cart:   cart,
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := cartMongo.UpsertCart(ctx, tt.userID, tt.cart)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

func TestMergeGuestCartToUser(t *testing.T) {
	mockCollection := &MockCartCollectionInterface{}
	cartMongo := &CartMongo{Collection: mockCollection}
	ctx := context.Background()

	guestItems := []models.CartItem{
		{ProductID: "product1", Quantity: 2, Price: 10.99, Name: "Product 1"},
		{ProductID: "product2", Quantity: 1, Price: 20.99, Name: "Product 2"},
	}

	tests := []struct {
		name        string
		userID      string
		items       []models.CartItem
		setupMock   func()
		expectError bool
	}{
		{
			name:   "valid merge should succeed",
			userID: "user123",
			items:  guestItems,
			setupMock: func() {
				// Mock GetCartByUserID
				mockResult := &MockCartSingleResult{}
				mockResult.On("Decode", mock.AnythingOfType("*models.Cart")).Return(nil)
				mockResult.On("Err").Return(nil)
				mockCollection.On("FindOne", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(mockResult).Once()
				// Mock UpsertCart
				mockCollection.On("UpdateOne", ctx, bson.M{"user_id": "user123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{}, nil).Once()
			},
			expectError: false,
		},
		{
			name:   "database error in GetCartByUserID should be returned",
			userID: "user123",
			items:  guestItems,
			setupMock: func() {
				mockResult := &MockCartSingleResult{}
				mockResult.On("Decode", mock.AnythingOfType("*models.Cart")).Return(assert.AnError)
				mockResult.On("Err").Return(assert.AnError)
				mockCollection.On("FindOne", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(mockResult).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := cartMongo.MergeGuestCartToUser(ctx, tt.userID, tt.items)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

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
				assert.Error(t, err)
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}
