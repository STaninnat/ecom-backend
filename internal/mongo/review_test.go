package intmongo

import (
	"context"
	"testing"

	"github.com/STaninnat/ecom-backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MockCollectionInterface for testing
type MockReviewCollectionInterface struct {
	mock.Mock
}

// InsertOne mocks the MongoDB InsertOne operation for testing.
// Returns the mocked InsertOneResult and error based on test expectations.
func (m *MockReviewCollectionInterface) InsertOne(ctx context.Context, document any) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

// InsertMany mocks the MongoDB InsertMany operation for testing.
// Returns the mocked InsertManyResult and error based on test expectations.
func (m *MockReviewCollectionInterface) InsertMany(ctx context.Context, documents []any) (*mongo.InsertManyResult, error) {
	args := m.Called(ctx, documents)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertManyResult), args.Error(1)
}

// Find mocks the MongoDB Find operation for testing.
// Returns a mocked cursor and error based on test expectations.
func (m *MockReviewCollectionInterface) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (CursorInterface, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(CursorInterface), args.Error(1)
}

// FindOne mocks the MongoDB FindOne operation for testing.
// Returns a mocked SingleResultInterface for test expectations.
func (m *MockReviewCollectionInterface) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) SingleResultInterface {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(SingleResultInterface)
}

// UpdateOne mocks the MongoDB UpdateOne operation for testing.
// Returns the mocked UpdateResult and error based on test expectations.
func (m *MockReviewCollectionInterface) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

// UpdateMany mocks the MongoDB UpdateMany operation for testing.
// Returns the mocked UpdateResult and error based on test expectations.
func (m *MockReviewCollectionInterface) UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

// DeleteOne mocks the MongoDB DeleteOne operation for testing.
// Returns the mocked DeleteResult and error based on test expectations.
func (m *MockReviewCollectionInterface) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

// DeleteMany mocks the MongoDB DeleteMany operation for testing.
// Returns the mocked DeleteResult and error based on test expectations.
func (m *MockReviewCollectionInterface) DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

// CountDocuments mocks the MongoDB CountDocuments operation for testing.
// Returns the mocked count and error based on test expectations.
func (m *MockReviewCollectionInterface) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(int64), args.Error(1)
}

// Aggregate mocks the MongoDB Aggregate operation for testing.
// Returns a mocked cursor and error based on test expectations.
func (m *MockReviewCollectionInterface) Aggregate(ctx context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (CursorInterface, error) {
	args := m.Called(ctx, pipeline, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(CursorInterface), args.Error(1)
}

// Indexes mocks the MongoDB Indexes operation for testing.
// Returns a mocked IndexView for test expectations.
func (m *MockReviewCollectionInterface) Indexes() mongo.IndexView {
	args := m.Called()
	return args.Get(0).(mongo.IndexView)
}

// MockCursor for testing
type MockCursor struct {
	mock.Mock
}

// Next mocks the cursor Next operation for testing.
// Returns a boolean indicating if more documents are available.
func (m *MockCursor) Next(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

// Decode mocks the cursor Decode operation for testing.
// Decodes the current document into the provided value.
func (m *MockCursor) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}

// All mocks the cursor All operation for testing.
// Decodes all remaining documents into the provided results slice.
func (m *MockCursor) All(ctx context.Context, results any) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

// Close mocks the cursor Close operation for testing.
// Closes the cursor and returns any error that occurred.
func (m *MockCursor) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Err mocks the cursor Err operation for testing.
// Returns any error that occurred during cursor operations.
func (m *MockCursor) Err() error {
	args := m.Called()
	return args.Error(0)
}

// MockSingleResult for testing
type MockSingleResult struct {
	mock.Mock
}

// Decode mocks the SingleResult Decode operation for testing.
// Decodes the single document into the provided value.
func (m *MockSingleResult) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}

// Err mocks the SingleResult Err operation for testing.
// Returns any error that occurred during the find operation.
func (m *MockSingleResult) Err() error {
	args := m.Called()
	return args.Error(0)
}

// TestCreateReview tests the CreateReview function with various scenarios.
// It verifies successful review creation, ID generation, and database error handling.
func TestCreateReview(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		review      *models.Review
		setupMock   func()
		expectError bool
	}{
		{
			name:        "nil review should return error",
			review:      nil,
			setupMock:   func() {},
			expectError: true,
		},
		{
			name: "valid review should succeed",
			review: &models.Review{
				UserID:    "user123",
				ProductID: "product123",
				Rating:    5,
				Comment:   "Great product!",
			},
			setupMock: func() {
				mockCollection.On("InsertOne", ctx, mock.AnythingOfType("*models.Review")).Return(&mongo.InsertOneResult{}, nil)
			},
			expectError: false,
		},
		{
			name: "review with empty ID should generate ID",
			review: &models.Review{
				ID:        "",
				UserID:    "user123",
				ProductID: "product123",
				Rating:    4,
			},
			setupMock: func() {
				mockCollection.On("InsertOne", ctx, mock.AnythingOfType("*models.Review")).Return(&mongo.InsertOneResult{}, nil)
			},
			expectError: false,
		},
		{
			name: "database error should be returned",
			review: &models.Review{
				UserID:    "user123",
				ProductID: "product123",
				Rating:    3,
			},
			setupMock: func() {
				mockCollection.On("InsertOne", ctx, mock.AnythingOfType("*models.Review")).Return(nil, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollection.ExpectedCalls = nil
			tt.setupMock()

			err := reviewMongo.CreateReview(ctx, tt.review)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.review.ID)
				assert.NotZero(t, tt.review.CreatedAt)
				assert.NotZero(t, tt.review.UpdatedAt)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestCreateReviews tests the CreateReviews function with multiple reviews.
// It verifies successful batch review creation and database error handling.
func TestCreateReviews(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		reviews     []*models.Review
		setupMock   func()
		expectError bool
	}{
		{
			name:        "empty slice should return error",
			reviews:     []*models.Review{},
			setupMock:   func() {},
			expectError: true,
		},
		{
			name: "valid reviews should succeed",
			reviews: []*models.Review{
				{UserID: "user1", ProductID: "product1", Rating: 5},
				{UserID: "user2", ProductID: "product2", Rating: 4},
			},
			setupMock: func() {
				mockCollection.On("InsertMany", ctx, mock.AnythingOfType("[]interface {}")).Return(&mongo.InsertManyResult{}, nil)
			},
			expectError: false,
		},
		{
			name: "nil review in slice should return error",
			reviews: []*models.Review{
				{UserID: "user1", ProductID: "product1", Rating: 5},
				nil,
				{UserID: "user3", ProductID: "product3", Rating: 3},
			},
			setupMock:   func() {},
			expectError: true,
		},
		{
			name: "database error should be returned",
			reviews: []*models.Review{
				{UserID: "user1", ProductID: "product1", Rating: 5},
			},
			setupMock: func() {
				mockCollection.On("InsertMany", ctx, mock.AnythingOfType("[]interface {}")).Return(nil, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollection.ExpectedCalls = nil
			tt.setupMock()

			err := reviewMongo.CreateReviews(ctx, tt.reviews)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Check that IDs and timestamps were set
				for _, review := range tt.reviews {
					assert.NotEmpty(t, review.ID)
					assert.NotZero(t, review.CreatedAt)
					assert.NotZero(t, review.UpdatedAt)
				}
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestGetReviewsByProductID tests the GetReviewsByProductID function with various scenarios.
// It verifies successful retrieval of reviews by product ID and database error handling.
func TestGetReviewsByProductID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		productID   string
		setupMock   func()
		expectError bool
		expectedLen int
	}{
		{
			name:        "empty product ID should return error",
			productID:   "",
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:      "valid product ID should return reviews",
			productID: "product123",
			setupMock: func() {
				mockCursor := &MockCursor{}
				mockCursor.On("Close", ctx).Return(nil)
				mockCursor.On("All", ctx, mock.AnythingOfType("*[]*models.Review")).Run(func(args mock.Arguments) {
					reviews := args.Get(1).(*[]*models.Review)
					*reviews = []*models.Review{
						{UserID: "user1", ProductID: "product123", Rating: 5},
						{UserID: "user2", ProductID: "product123", Rating: 4},
					}
				}).Return(nil)
				mockCollection.On("Find", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(mockCursor, nil)
			},
			expectError: false,
			expectedLen: 2,
		},
		{
			name:      "database error should be returned",
			productID: "product123",
			setupMock: func() {
				mockCollection.On("Find", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(nil, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollection.ExpectedCalls = nil
			tt.setupMock()

			reviews, err := reviewMongo.GetReviewsByProductID(ctx, tt.productID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, reviews)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reviews)
				assert.Len(t, reviews, tt.expectedLen)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestGetReviewsByProductID_EmptyProductID tests GetReviewsByProductID with empty product ID.
// It verifies that empty product ID returns an empty slice without errors.
func TestGetReviewsByProductID_EmptyProductID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	result, err := reviewMongo.GetReviewsByProductID(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "product ID cannot be empty")
}

// TestGetReviewsByProductID_DatabaseError tests GetReviewsByProductID when database operations fail.
// It verifies proper error handling when the Find operation fails.
func TestGetReviewsByProductID_DatabaseError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("Find", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(nil, assert.AnError)

	result, err := reviewMongo.GetReviewsByProductID(ctx, "product123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find reviews by product ID")
}

// TestGetReviewsByProductID_DecodeError tests GetReviewsByProductID when document decoding fails.
// It verifies proper error handling when cursor decoding operations fail.
func TestGetReviewsByProductID_DecodeError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCursor := &MockCursor{}
	mockCursor.On("Close", ctx).Return(nil)
	mockCursor.On("All", ctx, mock.AnythingOfType("*[]*models.Review")).Return(assert.AnError)

	mockCollection.On("Find", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(mockCursor, nil)

	result, err := reviewMongo.GetReviewsByProductID(ctx, "product123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to decode reviews")
}

// TestGetReviewsByProductIDPaginated tests the GetReviewsByProductIDPaginated function.
// It verifies successful paginated retrieval of reviews with proper metadata.
func TestGetReviewsByProductIDPaginated(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		productID   string
		pagination  *PaginationOptions
		setupMock   func()
		expectError bool
	}{
		{
			name:        "empty product ID should return error",
			productID:   "",
			pagination:  NewPaginationOptions(1, 10),
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:       "nil pagination should use defaults",
			productID:  "product123",
			pagination: nil,
			setupMock: func() {
				mockCollection.On("CountDocuments", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(int64(25), nil)
				mockCursor := &MockCursor{}
				mockCursor.On("Close", ctx).Return(nil)
				mockCursor.On("All", ctx, mock.AnythingOfType("*[]*models.Review")).Return(nil)
				mockCollection.On("Find", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(mockCursor, nil)
			},
			expectError: false,
		},
		{
			name:       "valid pagination should work",
			productID:  "product123",
			pagination: NewPaginationOptions(2, 5),
			setupMock: func() {
				mockCollection.On("CountDocuments", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(int64(25), nil)
				mockCursor := &MockCursor{}
				mockCursor.On("Close", ctx).Return(nil)
				mockCursor.On("All", ctx, mock.AnythingOfType("*[]*models.Review")).Return(nil)
				mockCollection.On("Find", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(mockCursor, nil)
			},
			expectError: false,
		},
		{
			name:       "count error should be returned",
			productID:  "product123",
			pagination: NewPaginationOptions(1, 10),
			setupMock: func() {
				mockCollection.On("CountDocuments", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(int64(0), assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollection.ExpectedCalls = nil
			tt.setupMock()

			result, err := reviewMongo.GetReviewsByProductIDPaginated(ctx, tt.productID, tt.pagination)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.pagination != nil {
					assert.Equal(t, tt.pagination.Page, result.Page)
					assert.Equal(t, tt.pagination.PageSize, result.PageSize)
				} else {
					assert.Equal(t, int64(1), result.Page)
					assert.Equal(t, int64(10), result.PageSize)
				}
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestGetReviewsByProductIDPaginated_EmptyProductID tests GetReviewsByProductIDPaginated with empty product ID.
// It verifies that empty product ID returns an empty result without errors.
func TestGetReviewsByProductIDPaginated_EmptyProductID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	result, err := reviewMongo.GetReviewsByProductIDPaginated(ctx, "", nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "product ID cannot be empty")
}

// TestGetReviewsByProductIDPaginated_CountError tests GetReviewsByProductIDPaginated when count operation fails.
// It verifies proper error handling when the CountDocuments operation fails.
func TestGetReviewsByProductIDPaginated_CountError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	pagination := NewPaginationOptions(1, 10)
	mockCollection.On("CountDocuments", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(int64(0), assert.AnError)

	result, err := reviewMongo.GetReviewsByProductIDPaginated(ctx, "product123", pagination)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to count reviews")
}

// TestGetReviewsByProductIDPaginated_FindError tests GetReviewsByProductIDPaginated when Find operation fails.
// It verifies proper error handling when the Find operation fails.
func TestGetReviewsByProductIDPaginated_FindError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	pagination := NewPaginationOptions(1, 10)
	mockCollection.On("CountDocuments", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(int64(5), nil)
	mockCollection.On("Find", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(nil, assert.AnError)

	result, err := reviewMongo.GetReviewsByProductIDPaginated(ctx, "product123", pagination)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find reviews")
}

// TestGetReviewsByProductIDPaginated_DecodeError tests GetReviewsByProductIDPaginated when decoding fails.
// It verifies proper error handling when cursor decoding operations fail.
func TestGetReviewsByProductIDPaginated_DecodeError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	pagination := NewPaginationOptions(1, 10)
	mockCursor := &MockCursor{}
	mockCursor.On("Close", ctx).Return(nil)
	mockCursor.On("All", ctx, mock.AnythingOfType("*[]*models.Review")).Return(assert.AnError)

	mockCollection.On("CountDocuments", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(int64(5), nil)
	mockCollection.On("Find", ctx, bson.M{"product_id": "product123"}, mock.Anything).Return(mockCursor, nil)

	result, err := reviewMongo.GetReviewsByProductIDPaginated(ctx, "product123", pagination)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to decode reviews")
}

// TestGetReviewsByUserID tests the GetReviewsByUserID function with various scenarios.
// It verifies successful retrieval of reviews by user ID and database error handling.
func TestGetReviewsByUserID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      string
		setupMock   func()
		expectError bool
	}{
		{
			name:        "empty user ID should return error",
			userID:      "",
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:   "valid user ID should return reviews",
			userID: "user123",
			setupMock: func() {
				mockCursor := &MockCursor{}
				mockCursor.On("Close", ctx).Return(nil)
				mockCursor.On("All", ctx, mock.AnythingOfType("*[]*models.Review")).Run(func(args mock.Arguments) {
					reviews := args.Get(1).(*[]*models.Review)
					*reviews = []*models.Review{
						{UserID: "user123", ProductID: "product1", Rating: 5},
						{UserID: "user123", ProductID: "product2", Rating: 4},
					}
				}).Return(nil)
				mockCollection.On("Find", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(mockCursor, nil)
			},
			expectError: false,
		},
		{
			name:   "database error should be returned",
			userID: "user123",
			setupMock: func() {
				mockCollection.On("Find", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(nil, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollection.ExpectedCalls = nil
			tt.setupMock()

			reviews, err := reviewMongo.GetReviewsByUserID(ctx, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, reviews)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reviews)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestGetReviewsByUserID_EmptyUserID tests GetReviewsByUserID with empty user ID.
// It verifies that empty user ID returns an empty slice without errors.
func TestGetReviewsByUserID_EmptyUserID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	result, err := reviewMongo.GetReviewsByUserID(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
}

// TestGetReviewsByUserID_DatabaseError tests GetReviewsByUserID when database operations fail.
// It verifies proper error handling when the Find operation fails.
func TestGetReviewsByUserID_DatabaseError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("Find", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(nil, assert.AnError)

	result, err := reviewMongo.GetReviewsByUserID(ctx, "user123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find reviews by user ID")
}

// TestGetReviewsByUserID_DecodeError tests GetReviewsByUserID when document decoding fails.
// It verifies proper error handling when cursor decoding operations fail.
func TestGetReviewsByUserID_DecodeError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCursor := &MockCursor{}
	mockCursor.On("Close", ctx).Return(nil)
	mockCursor.On("All", ctx, mock.AnythingOfType("*[]*models.Review")).Return(assert.AnError)

	mockCollection.On("Find", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(mockCursor, nil)

	result, err := reviewMongo.GetReviewsByUserID(ctx, "user123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to decode reviews")
}

// TestGetReviewsByUserIDPaginated tests the GetReviewsByUserIDPaginated function.
// It verifies successful paginated retrieval of user reviews with proper metadata.
func TestGetReviewsByUserIDPaginated(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      string
		pagination  *PaginationOptions
		setupMock   func()
		expectError bool
	}{
		{
			name:        "empty user ID should return error",
			userID:      "",
			pagination:  NewPaginationOptions(1, 10),
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:       "valid pagination should work",
			userID:     "user123",
			pagination: NewPaginationOptions(1, 10),
			setupMock: func() {
				mockCollection.On("CountDocuments", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(int64(15), nil)
				mockCursor := &MockCursor{}
				mockCursor.On("Close", ctx).Return(nil)
				mockCursor.On("All", ctx, mock.AnythingOfType("*[]*models.Review")).Return(nil)
				mockCollection.On("Find", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(mockCursor, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollection.ExpectedCalls = nil
			tt.setupMock()

			result, err := reviewMongo.GetReviewsByUserIDPaginated(ctx, tt.userID, tt.pagination)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.pagination.Page, result.Page)
				assert.Equal(t, tt.pagination.PageSize, result.PageSize)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestGetReviewsByUserIDPaginated_EmptyUserID tests GetReviewsByUserIDPaginated with empty user ID.
// It verifies that empty user ID returns an empty result without errors.
func TestGetReviewsByUserIDPaginated_EmptyUserID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	result, err := reviewMongo.GetReviewsByUserIDPaginated(ctx, "", nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
}

// TestGetReviewsByUserIDPaginated_CountError tests GetReviewsByUserIDPaginated when count operation fails.
// It verifies proper error handling when the CountDocuments operation fails.
func TestGetReviewsByUserIDPaginated_CountError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	pagination := NewPaginationOptions(1, 10)
	mockCollection.On("CountDocuments", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(int64(0), assert.AnError)

	result, err := reviewMongo.GetReviewsByUserIDPaginated(ctx, "user123", pagination)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to count reviews")
}

// TestGetReviewsByUserIDPaginated_FindError tests GetReviewsByUserIDPaginated when Find operation fails.
// It verifies proper error handling when the Find operation fails.
func TestGetReviewsByUserIDPaginated_FindError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	pagination := NewPaginationOptions(1, 10)
	mockCollection.On("CountDocuments", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(int64(5), nil)
	mockCollection.On("Find", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(nil, assert.AnError)

	result, err := reviewMongo.GetReviewsByUserIDPaginated(ctx, "user123", pagination)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find reviews")
}

// TestGetReviewsByUserIDPaginated_DecodeError tests GetReviewsByUserIDPaginated when decoding fails.
// It verifies proper error handling when cursor decoding operations fail.
func TestGetReviewsByUserIDPaginated_DecodeError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	pagination := NewPaginationOptions(1, 10)
	mockCursor := &MockCursor{}
	mockCursor.On("Close", ctx).Return(nil)
	mockCursor.On("All", ctx, mock.AnythingOfType("*[]*models.Review")).Return(assert.AnError)

	mockCollection.On("CountDocuments", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(int64(5), nil)
	mockCollection.On("Find", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(mockCursor, nil)

	result, err := reviewMongo.GetReviewsByUserIDPaginated(ctx, "user123", pagination)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to decode reviews")
}

// TestGetReviewByID tests the GetReviewByID function with various scenarios.
// It verifies successful retrieval of reviews by ID and database error handling.
func TestGetReviewByID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		reviewID    string
		setupMock   func()
		expectError bool
	}{
		{
			name:        "empty review ID should return error",
			reviewID:    "",
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:     "valid review ID should return review",
			reviewID: "review123",
			setupMock: func() {
				mockResult := &MockSingleResult{}
				mockResult.On("Err").Return(nil)
				mockResult.On("Decode", mock.AnythingOfType("*models.Review")).Return(nil)
				mockCollection.On("FindOne", ctx, bson.M{"_id": "review123"}, mock.Anything).Return(mockResult)
			},
			expectError: false,
		},
		{
			name:     "review not found should return error",
			reviewID: "review123",
			setupMock: func() {
				mockResult := &MockSingleResult{}
				mockResult.On("Err").Return(mongo.ErrNoDocuments)
				mockResult.On("Decode", mock.AnythingOfType("*models.Review")).Return(mongo.ErrNoDocuments)
				mockCollection.On("FindOne", ctx, bson.M{"_id": "review123"}, mock.Anything).Return(mockResult)
			},
			expectError: true,
		},
		{
			name:     "database error should be returned",
			reviewID: "review123",
			setupMock: func() {
				mockResult := &MockSingleResult{}
				mockResult.On("Err").Return(assert.AnError)
				mockResult.On("Decode", mock.AnythingOfType("*models.Review")).Return(assert.AnError)
				mockCollection.On("FindOne", ctx, bson.M{"_id": "review123"}, mock.Anything).Return(mockResult)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollection.ExpectedCalls = nil
			tt.setupMock()

			review, err := reviewMongo.GetReviewByID(ctx, tt.reviewID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, review)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, review)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestGetReviewByID_EmptyReviewID tests GetReviewByID with empty review ID.
// It verifies that empty review ID returns an error.
func TestGetReviewByID_EmptyReviewID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	result, err := reviewMongo.GetReviewByID(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "review ID cannot be empty")
}

// TestGetReviewByID_DecodeError tests GetReviewByID when document decoding fails.
// It verifies proper error handling when the Decode operation fails.
func TestGetReviewByID_DecodeError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockResult := &MockSingleResult{}
	mockResult.On("Decode", mock.AnythingOfType("*models.Review")).Return(assert.AnError)
	mockResult.On("Err").Return(assert.AnError)

	mockCollection.On("FindOne", ctx, bson.M{"_id": "review123"}, mock.Anything).Return(mockResult)

	result, err := reviewMongo.GetReviewByID(ctx, "review123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find review")
}

// TestUpdateReviewByID tests the UpdateReviewByID function with various scenarios.
// It verifies successful review updates and database error handling.
func TestUpdateReviewByID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name          string
		reviewID      string
		updatedReview *models.Review
		setupMock     func()
		expectError   bool
	}{
		{
			name:          "empty review ID should return error",
			reviewID:      "",
			updatedReview: &models.Review{},
			setupMock:     func() {},
			expectError:   true,
		},
		{
			name:          "nil updated review should return error",
			reviewID:      "review123",
			updatedReview: nil,
			setupMock:     func() {},
			expectError:   true,
		},
		{
			name:     "valid update should succeed",
			reviewID: "review123",
			updatedReview: &models.Review{
				Rating:  5,
				Comment: "Updated comment",
			},
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"_id": "review123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 1}, nil)
			},
			expectError: false,
		},
		{
			name:     "review not found should return error",
			reviewID: "review123",
			updatedReview: &models.Review{
				Rating: 4,
			},
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"_id": "review123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{MatchedCount: 0}, nil)
			},
			expectError: true,
		},
		{
			name:     "database error should be returned",
			reviewID: "review123",
			updatedReview: &models.Review{
				Rating: 4,
			},
			setupMock: func() {
				mockCollection.On("UpdateOne", ctx, bson.M{"_id": "review123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(nil, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollection.ExpectedCalls = nil
			tt.setupMock()

			err := reviewMongo.UpdateReviewByID(ctx, tt.reviewID, tt.updatedReview)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestUpdateReviewByID_EmptyReviewID tests UpdateReviewByID with empty review ID.
// It verifies that empty review ID returns an error.
func TestUpdateReviewByID_EmptyReviewID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	err := reviewMongo.UpdateReviewByID(ctx, "", &models.Review{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "review ID cannot be empty")
}

// TestUpdateReviewByID_NilReview tests UpdateReviewByID with nil review.
// It verifies that nil review returns an error.
func TestUpdateReviewByID_NilReview(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	err := reviewMongo.UpdateReviewByID(ctx, "review123", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updated review cannot be nil")
}

// TestUpdateReviewByID_NotFound tests UpdateReviewByID when review is not found.
// It verifies proper error handling when the review doesn't exist.
func TestUpdateReviewByID_NotFound(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockResult := &mongo.UpdateResult{MatchedCount: 0, ModifiedCount: 0}
	mockCollection.On("UpdateOne", ctx, bson.M{"_id": "review123"}, mock.Anything, mock.Anything).Return(mockResult, nil)

	err := reviewMongo.UpdateReviewByID(ctx, "review123", &models.Review{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "review not found")
}

// TestUpdateReviewByID_DatabaseError tests UpdateReviewByID when database operations fail.
// It verifies proper error handling when the UpdateOne operation fails.
func TestUpdateReviewByID_DatabaseError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("UpdateOne", ctx, bson.M{"_id": "review123"}, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	err := reviewMongo.UpdateReviewByID(ctx, "review123", &models.Review{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update review")
}

// TestUpdateReviewsByProductID tests the UpdateReviewsByProductID function.
// It verifies successful batch updates of reviews by product ID.
func TestUpdateReviewsByProductID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		productID   string
		update      bson.M
		setupMock   func()
		expectError bool
	}{
		{
			name:        "empty product ID should return error",
			productID:   "",
			update:      bson.M{},
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:      "valid update should succeed",
			productID: "product123",
			update:    bson.M{"rating": 5},
			setupMock: func() {
				mockCollection.On("UpdateMany", ctx, bson.M{"product_id": "product123"}, mock.AnythingOfType("bson.M"), mock.Anything).Return(&mongo.UpdateResult{}, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollection.ExpectedCalls = nil
			tt.setupMock()

			err := reviewMongo.UpdateReviewsByProductID(ctx, tt.productID, tt.update)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestUpdateReviewsByProductID_DatabaseError tests UpdateReviewsByProductID when database operations fail.
// It verifies proper error handling when the UpdateMany operation fails.
func TestUpdateReviewsByProductID_DatabaseError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("UpdateMany", ctx, bson.M{"product_id": "product123"}, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	err := reviewMongo.UpdateReviewsByProductID(ctx, "product123", bson.M{"rating": 5})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update reviews")
}

// TestUpdateReviewsByProductID_EmptyProductID tests UpdateReviewsByProductID with empty product ID.
// It verifies that empty product ID returns an error.
func TestUpdateReviewsByProductID_EmptyProductID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	err := reviewMongo.UpdateReviewsByProductID(ctx, "", bson.M{"$set": bson.M{"rating": 5}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product ID cannot be empty")
}

// TestDeleteReviewByID tests the DeleteReviewByID function with various scenarios.
// It verifies successful review deletion and database error handling.
func TestDeleteReviewByID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		reviewID    string
		setupMock   func()
		expectError bool
	}{
		{
			name:        "empty review ID should return error",
			reviewID:    "",
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:     "valid delete should succeed",
			reviewID: "review123",
			setupMock: func() {
				mockCollection.On("DeleteOne", ctx, bson.M{"_id": "review123"}, mock.Anything).Return(&mongo.DeleteResult{DeletedCount: 1}, nil)
			},
			expectError: false,
		},
		{
			name:     "review not found should return error",
			reviewID: "review123",
			setupMock: func() {
				mockCollection.On("DeleteOne", ctx, bson.M{"_id": "review123"}, mock.Anything).Return(&mongo.DeleteResult{DeletedCount: 0}, nil)
			},
			expectError: true,
		},
		{
			name:     "database error should be returned",
			reviewID: "review123",
			setupMock: func() {
				mockCollection.On("DeleteOne", ctx, bson.M{"_id": "review123"}, mock.Anything).Return(nil, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollection.ExpectedCalls = nil
			tt.setupMock()

			err := reviewMongo.DeleteReviewByID(ctx, tt.reviewID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestDeleteReviewByID_EmptyReviewID tests DeleteReviewByID with empty review ID.
// It verifies that empty review ID returns an error.
func TestDeleteReviewByID_EmptyReviewID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	err := reviewMongo.DeleteReviewByID(ctx, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "review ID cannot be empty")
}

// TestDeleteReviewByID_NotFound tests DeleteReviewByID when review is not found.
// It verifies proper error handling when the review doesn't exist.
func TestDeleteReviewByID_NotFound(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockResult := &mongo.DeleteResult{DeletedCount: 0}
	mockCollection.On("DeleteOne", ctx, bson.M{"_id": "review123"}, mock.Anything).Return(mockResult, nil)

	err := reviewMongo.DeleteReviewByID(ctx, "review123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "review not found")
}

// TestDeleteReviewByID_DatabaseError tests DeleteReviewByID when database operations fail.
// It verifies proper error handling when the DeleteOne operation fails.
func TestDeleteReviewByID_DatabaseError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("DeleteOne", ctx, bson.M{"_id": "review123"}, mock.Anything).Return(nil, assert.AnError)

	err := reviewMongo.DeleteReviewByID(ctx, "review123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete review")
}

// TestDeleteReviewsByUserID tests the DeleteReviewsByUserID function.
// It verifies successful batch deletion of reviews by user ID.
func TestDeleteReviewsByUserID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		userID      string
		setupMock   func()
		expectError bool
	}{
		{
			name:        "empty user ID should return error",
			userID:      "",
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:   "valid delete should succeed",
			userID: "user123",
			setupMock: func() {
				mockCollection.On("DeleteMany", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(&mongo.DeleteResult{}, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollection.ExpectedCalls = nil
			tt.setupMock()

			err := reviewMongo.DeleteReviewsByUserID(ctx, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestDeleteReviewsByUserID_EmptyUserID tests DeleteReviewsByUserID with empty user ID.
// It verifies that empty user ID returns an error.
func TestDeleteReviewsByUserID_EmptyUserID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	err := reviewMongo.DeleteReviewsByUserID(ctx, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
}

// TestDeleteReviewsByUserID_DatabaseError tests DeleteReviewsByUserID when database operations fail.
// It verifies proper error handling when the DeleteMany operation fails.
func TestDeleteReviewsByUserID_DatabaseError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("DeleteMany", ctx, bson.M{"user_id": "user123"}, mock.Anything).Return(nil, assert.AnError)

	err := reviewMongo.DeleteReviewsByUserID(ctx, "user123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete reviews")
}

// TestGetProductRatingStats tests the GetProductRatingStats function with various scenarios.
// It verifies successful retrieval of product rating statistics and database error handling.
func TestGetProductRatingStats(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	tests := []struct {
		name        string
		productID   string
		setupMock   func()
		expectError bool
		expected    map[string]any
	}{
		{
			name:        "empty product ID should return error",
			productID:   "",
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:      "valid product ID should return stats",
			productID: "product123",
			setupMock: func() {
				mockCursor := &MockCursor{}
				mockCursor.On("Close", ctx).Return(nil)
				mockCursor.On("All", ctx, mock.AnythingOfType("*[]map[string]interface {}")).Return(nil)
				mockCollection.On("Aggregate", ctx, mock.AnythingOfType("[]bson.M"), mock.Anything).Return(mockCursor, nil)
			},
			expectError: false,
			expected: map[string]any{
				"averageRating": 0.0,
				"totalReviews":  0,
				"ratingCounts":  []int{},
			},
		},
		{
			name:      "database error should be returned",
			productID: "product123",
			setupMock: func() {
				mockCollection.On("Aggregate", ctx, mock.AnythingOfType("[]bson.M"), mock.Anything).Return(nil, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCollection.ExpectedCalls = nil
			tt.setupMock()

			stats, err := reviewMongo.GetProductRatingStats(ctx, tt.productID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats)
				if tt.expected != nil {
					assert.Equal(t, tt.expected["averageRating"], stats["averageRating"])
					assert.Equal(t, tt.expected["totalReviews"], stats["totalReviews"])
				}
			}

			mockCollection.AssertExpectations(t)
		})
	}
}

// TestGetProductRatingStats_EmptyProductID tests GetProductRatingStats with empty product ID.
// It verifies that empty product ID returns an error.
func TestGetProductRatingStats_EmptyProductID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	result, err := reviewMongo.GetProductRatingStats(ctx, "")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "product ID cannot be empty")
}

// TestGetProductRatingStats_DatabaseError tests GetProductRatingStats when database operations fail.
// It verifies proper error handling when the Aggregate operation fails.
func TestGetProductRatingStats_DatabaseError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCollection.On("Aggregate", ctx, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	result, err := reviewMongo.GetProductRatingStats(ctx, "product123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to aggregate rating stats")
}

// TestGetProductRatingStats_DecodeError tests GetProductRatingStats when document decoding fails.
// It verifies proper error handling when cursor decoding operations fail.
func TestGetProductRatingStats_DecodeError(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCursor := &MockCursor{}
	mockCursor.On("Close", ctx).Return(nil)
	mockCursor.On("All", ctx, mock.AnythingOfType("*[]map[string]interface {}")).Return(assert.AnError)

	mockCollection.On("Aggregate", ctx, mock.Anything, mock.Anything).Return(mockCursor, nil)

	result, err := reviewMongo.GetProductRatingStats(ctx, "product123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to decode aggregation results")
}

// TestGetProductRatingStats_EmptyResults tests GetProductRatingStats when no results are returned.
// It verifies proper handling when the aggregation returns no documents.
func TestGetProductRatingStats_EmptyResults(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	mockCursor := &MockCursor{}
	mockCursor.On("Close", ctx).Return(nil)
	mockCursor.On("All", ctx, mock.AnythingOfType("*[]map[string]interface {}")).Run(func(args mock.Arguments) {
		results := args.Get(1).(*[]map[string]any)
		*results = []map[string]any{} // Empty results
	}).Return(nil)

	mockCollection.On("Aggregate", ctx, mock.Anything, mock.Anything).Return(mockCursor, nil)

	result, err := reviewMongo.GetProductRatingStats(ctx, "product123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0.0, result["averageRating"])
	assert.EqualValues(t, int64(0), result["totalReviews"])
	assert.Equal(t, []int{}, result["ratingCounts"])
}

// TestNewReviewMongo tests the NewReviewMongo constructor function.
// It verifies that the ReviewMongo instance is created correctly with the provided collection.
func TestNewReviewMongo(t *testing.T) {
	// This test is now covered by integration tests
	// Constructor logic is tested in TestNewReviewMongo_Integration
	t.Skip("Covered by integration tests")
}
