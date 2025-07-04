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

func (m *MockReviewCollectionInterface) InsertOne(ctx context.Context, document any) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockReviewCollectionInterface) InsertMany(ctx context.Context, documents []any) (*mongo.InsertManyResult, error) {
	args := m.Called(ctx, documents)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertManyResult), args.Error(1)
}

func (m *MockReviewCollectionInterface) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (CursorInterface, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(CursorInterface), args.Error(1)
}

func (m *MockReviewCollectionInterface) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) SingleResultInterface {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(SingleResultInterface)
}

func (m *MockReviewCollectionInterface) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockReviewCollectionInterface) UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockReviewCollectionInterface) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockReviewCollectionInterface) DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockReviewCollectionInterface) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReviewCollectionInterface) Aggregate(ctx context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (CursorInterface, error) {
	args := m.Called(ctx, pipeline, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(CursorInterface), args.Error(1)
}

func (m *MockReviewCollectionInterface) Indexes() mongo.IndexView {
	args := m.Called()
	return args.Get(0).(mongo.IndexView)
}

// MockCursor for testing
type MockCursor struct {
	mock.Mock
}

func (m *MockCursor) Next(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockCursor) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}

func (m *MockCursor) All(ctx context.Context, results any) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

func (m *MockCursor) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCursor) Err() error {
	args := m.Called()
	return args.Error(0)
}

// MockSingleResult for testing
type MockSingleResult struct {
	mock.Mock
}

func (m *MockSingleResult) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}

func (m *MockSingleResult) Err() error {
	args := m.Called()
	return args.Error(0)
}

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

func TestUpdateReviewsByProductID_EmptyProductID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	err := reviewMongo.UpdateReviewsByProductID(ctx, "", bson.M{"$set": bson.M{"rating": 5}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product ID cannot be empty")
}

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

func TestDeleteReviewsByUserID_EmptyUserID(t *testing.T) {
	mockCollection := &MockReviewCollectionInterface{}
	reviewMongo := &ReviewMongo{Collection: mockCollection}
	ctx := context.Background()

	err := reviewMongo.DeleteReviewsByUserID(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID cannot be empty")
}

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
