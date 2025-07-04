package intmongo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MockCollectionInterface for testing
type MockCollectionInterface struct {
	mock.Mock
}

func (m *MockCollectionInterface) InsertOne(ctx context.Context, document any) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCollectionInterface) InsertMany(ctx context.Context, documents []any) (*mongo.InsertManyResult, error) {
	args := m.Called(ctx, documents)
	return args.Get(0).(*mongo.InsertManyResult), args.Error(1)
}

func (m *MockCollectionInterface) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (CursorInterface, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(CursorInterface), args.Error(1)
}

func (m *MockCollectionInterface) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) SingleResultInterface {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(SingleResultInterface)
}

func (m *MockCollectionInterface) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockCollectionInterface) UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update, opts)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockCollectionInterface) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCollectionInterface) DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockCollectionInterface) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	args := m.Called(ctx, filter, opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCollectionInterface) Aggregate(ctx context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (CursorInterface, error) {
	args := m.Called(ctx, pipeline, opts)
	return args.Get(0).(CursorInterface), args.Error(1)
}

func (m *MockCollectionInterface) Indexes() mongo.IndexView {
	args := m.Called()
	return args.Get(0).(mongo.IndexView)
}

// MockCursorInterface for testing
type MockCursorInterface struct {
	mock.Mock
}

func (m *MockCursorInterface) Next(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockCursorInterface) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}

func (m *MockCursorInterface) All(ctx context.Context, results any) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

func (m *MockCursorInterface) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCursorInterface) Err() error {
	args := m.Called()
	return args.Error(0)
}

// MockSingleResultInterface for testing
type MockSingleResultInterface struct {
	mock.Mock
}

func (m *MockSingleResultInterface) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}

func (m *MockSingleResultInterface) Err() error {
	args := m.Called()
	return args.Error(0)
}

func TestDefaultDatabaseConfig(t *testing.T) {
	config := DefaultDatabaseConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "mongodb://localhost:27017", config.URI)
	assert.Equal(t, "ecom", config.DatabaseName)
	assert.Equal(t, 10*time.Second, config.ConnectTimeout)
	assert.Equal(t, uint64(100), config.MaxPoolSize)
	assert.Equal(t, uint64(5), config.MinPoolSize)
}

func TestNewDatabaseManager(t *testing.T) {
	tests := []struct {
		name        string
		config      *DatabaseConfig
		expectError bool
	}{
		{
			name:        "nil config should use defaults",
			config:      nil,
			expectError: false,
		},
		{
			name: "valid config",
			config: &DatabaseConfig{
				URI:            "mongodb://localhost:27017",
				DatabaseName:   "testdb",
				ConnectTimeout: 5 * time.Second,
				MaxPoolSize:    50,
				MinPoolSize:    2,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test requires a running MongoDB instance
			// In a real test environment, you'd use a test container or mock
			// For now, we'll skip if MongoDB is not available
			manager, err := NewDatabaseManager(tt.config)

			if err != nil {
				// If MongoDB is not available, skip the test
				t.Skipf("MongoDB not available: %v", err)
			}

			if !tt.expectError {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
				assert.NotNil(t, manager.GetDatabase())
				assert.NotNil(t, manager.GetClient())

				// Test close
				ctx := context.Background()
				err = manager.Close(ctx)
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestPaginationOptions(t *testing.T) {
	tests := []struct {
		name     string
		page     int64
		pageSize int64
		expected *PaginationOptions
	}{
		{
			name:     "valid pagination",
			page:     2,
			pageSize: 20,
			expected: &PaginationOptions{
				Page:     2,
				PageSize: 20,
				Sort:     map[string]any{"created_at": -1},
			},
		},
		{
			name:     "page less than 1 should default to 1",
			page:     0,
			pageSize: 10,
			expected: &PaginationOptions{
				Page:     1,
				PageSize: 10,
				Sort:     map[string]any{"created_at": -1},
			},
		},
		{
			name:     "pageSize less than 1 should default to 10",
			page:     1,
			pageSize: 0,
			expected: &PaginationOptions{
				Page:     1,
				PageSize: 10,
				Sort:     map[string]any{"created_at": -1},
			},
		},
		{
			name:     "pageSize greater than 100 should cap at 100",
			page:     1,
			pageSize: 150,
			expected: &PaginationOptions{
				Page:     1,
				PageSize: 100,
				Sort:     map[string]any{"created_at": -1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewPaginationOptions(tt.page, tt.pageSize)

			assert.Equal(t, tt.expected.Page, result.Page)
			assert.Equal(t, tt.expected.PageSize, result.PageSize)
			assert.Equal(t, tt.expected.Sort, result.Sort)
		})
	}
}

func TestPaginatedResult(t *testing.T) {
	// Test generic PaginatedResult
	result := &PaginatedResult[string]{
		Data:       []string{"item1", "item2"},
		TotalCount: 10,
		Page:       1,
		PageSize:   2,
		TotalPages: 5,
		HasNext:    true,
		HasPrev:    false,
	}

	assert.Equal(t, 2, len(result.Data))
	assert.Equal(t, int64(10), result.TotalCount)
	assert.Equal(t, int64(1), result.Page)
	assert.Equal(t, int64(2), result.PageSize)
	assert.Equal(t, int64(5), result.TotalPages)
	assert.True(t, result.HasNext)
	assert.False(t, result.HasPrev)
}

func TestFactoryFunctions(t *testing.T) {
	// Test factory functions with nil inputs
	cursorAdapter := NewCursorAdapter(nil)
	assert.NotNil(t, cursorAdapter)

	singleResultAdapter := NewSingleResultAdapter(nil)
	assert.NotNil(t, singleResultAdapter)
}

func TestCreateIndexes(t *testing.T) {
	// This test requires a running MongoDB instance
	// In a real test environment, you'd use a test container
	t.Skip("Skipping CreateIndexes test - requires MongoDB instance")

	// Test would look like:
	// config := DefaultDatabaseConfig()
	// manager, err := NewDatabaseManager(config)
	// if err != nil {
	//     t.Skipf("MongoDB not available: %v", err)
	// }
	//
	// err = CreateIndexes(manager.GetDatabase())
	// assert.NoError(t, err)
}

func TestCollectionInterfaceMethods(t *testing.T) {
	mockCollection := &MockCollectionInterface{}

	ctx := context.Background()
	document := bson.M{"test": "data"}
	filter := bson.M{"_id": "test"}
	update := bson.M{"$set": bson.M{"test": "updated"}}

	// Test InsertOne
	mockCollection.On("InsertOne", ctx, document).Return(&mongo.InsertOneResult{}, nil)
	result, err := mockCollection.InsertOne(ctx, document)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test InsertMany
	documents := []any{document}
	mockCollection.On("InsertMany", ctx, documents).Return(&mongo.InsertManyResult{}, nil)
	resultMany, err := mockCollection.InsertMany(ctx, documents)
	assert.NoError(t, err)
	assert.NotNil(t, resultMany)

	// Test Find
	mockCursor := &MockCursorInterface{}
	mockCollection.On("Find", ctx, filter, mock.Anything).Return(mockCursor, nil)
	cursor, err := mockCollection.Find(ctx, filter)
	assert.NoError(t, err)
	assert.NotNil(t, cursor)

	// Test FindOne
	mockSingleResult := &MockSingleResultInterface{}
	mockCollection.On("FindOne", ctx, filter, mock.Anything).Return(mockSingleResult)
	singleResult := mockCollection.FindOne(ctx, filter)
	assert.NotNil(t, singleResult)

	// Test UpdateOne
	mockCollection.On("UpdateOne", ctx, filter, update, mock.Anything).Return(&mongo.UpdateResult{}, nil)
	updateResult, err := mockCollection.UpdateOne(ctx, filter, update)
	assert.NoError(t, err)
	assert.NotNil(t, updateResult)

	// Test UpdateMany
	mockCollection.On("UpdateMany", ctx, filter, update, mock.Anything).Return(&mongo.UpdateResult{}, nil)
	updateManyResult, err := mockCollection.UpdateMany(ctx, filter, update)
	assert.NoError(t, err)
	assert.NotNil(t, updateManyResult)

	// Test DeleteOne
	mockCollection.On("DeleteOne", ctx, filter, mock.Anything).Return(&mongo.DeleteResult{}, nil)
	deleteResult, err := mockCollection.DeleteOne(ctx, filter)
	assert.NoError(t, err)
	assert.NotNil(t, deleteResult)

	// Test DeleteMany
	mockCollection.On("DeleteMany", ctx, filter, mock.Anything).Return(&mongo.DeleteResult{}, nil)
	deleteManyResult, err := mockCollection.DeleteMany(ctx, filter)
	assert.NoError(t, err)
	assert.NotNil(t, deleteManyResult)

	// Test CountDocuments
	mockCollection.On("CountDocuments", ctx, filter, mock.Anything).Return(int64(10), nil)
	count, err := mockCollection.CountDocuments(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), count)

	// Test Aggregate
	pipeline := []bson.M{{"$match": filter}}
	mockAggCursor := &MockCursorInterface{}
	mockCollection.On("Aggregate", ctx, pipeline, mock.Anything).Return(mockAggCursor, nil)
	aggCursor, err := mockCollection.Aggregate(ctx, pipeline)
	assert.NoError(t, err)
	assert.NotNil(t, aggCursor)

	// Test Indexes
	mockCollection.On("Indexes").Return(mongo.IndexView{})
	indexes := mockCollection.Indexes()
	assert.NotNil(t, indexes)

	mockCollection.AssertExpectations(t)
}

func TestCursorInterfaceMethods(t *testing.T) {
	mockCursor := &MockCursorInterface{}

	ctx := context.Background()
	val := bson.M{}
	results := []bson.M{}

	// Test Next
	mockCursor.On("Next", ctx).Return(true)
	hasNext := mockCursor.Next(ctx)
	assert.True(t, hasNext)

	// Test Decode
	mockCursor.On("Decode", val).Return(nil)
	err := mockCursor.Decode(val)
	assert.NoError(t, err)

	// Test All
	mockCursor.On("All", ctx, &results).Return(nil)
	err = mockCursor.All(ctx, &results)
	assert.NoError(t, err)

	// Test Close
	mockCursor.On("Close", ctx).Return(nil)
	err = mockCursor.Close(ctx)
	assert.NoError(t, err)

	// Test Err
	mockCursor.On("Err").Return(nil)
	err = mockCursor.Err()
	assert.NoError(t, err)

	mockCursor.AssertExpectations(t)
}

func TestSingleResultInterfaceMethods(t *testing.T) {
	mockResult := &MockSingleResultInterface{}

	val := bson.M{}

	// Test Decode
	mockResult.On("Decode", val).Return(nil)
	err := mockResult.Decode(val)
	assert.NoError(t, err)

	// Test Err
	mockResult.On("Err").Return(nil)
	err = mockResult.Err()
	assert.NoError(t, err)

	mockResult.AssertExpectations(t)
}

func TestNewPaginationOptions(t *testing.T) {
	// Normal values
	opt := NewPaginationOptions(2, 20)
	assert.Equal(t, int64(2), opt.Page)
	assert.Equal(t, int64(20), opt.PageSize)
	assert.Equal(t, map[string]any{"created_at": -1}, opt.Sort)

	// Page less than 1
	opt = NewPaginationOptions(0, 10)
	assert.Equal(t, int64(1), opt.Page)

	// PageSize less than 1
	opt = NewPaginationOptions(1, 0)
	assert.Equal(t, int64(10), opt.PageSize)

	// PageSize greater than 100
	opt = NewPaginationOptions(1, 200)
	assert.Equal(t, int64(100), opt.PageSize)
}
