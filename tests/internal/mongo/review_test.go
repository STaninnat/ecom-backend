package intmongo_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// --- Mock Collection ---
type mockCollection struct {
	mock.Mock
}

func (m *mockCollection) InsertOne(ctx context.Context, document any) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}
func (m *mockCollection) Find(ctx context.Context, filter any) (*mongo.Cursor, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}
func (m *mockCollection) FindOne(ctx context.Context, filter any) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}
func (m *mockCollection) UpdateOne(ctx context.Context, filter any, update any) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}
func (m *mockCollection) DeleteOne(ctx context.Context, filter any) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

type mockCursor struct {
	mock.Mock
}

func (m *mockCursor) Next(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}
func (m *mockCursor) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}
func (m *mockCursor) All(ctx context.Context, results any) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}
func (m *mockCursor) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type mockSingleResult struct {
	mock.Mock
}

func (m *mockSingleResult) Decode(val any) error {
	args := m.Called(val)
	return args.Error(0)
}

// --- TEST CASES ---
func TestCreateReview_Success(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCollection)
	store := &intmongo.ReviewMongo{
		Collection: mockCol,
	}

	review := &models.Review{UserID: "u1", ProductID: "p1", Rating: 5}

	mockCol.On("InsertOne", ctx, mock.AnythingOfType("*models.Review")).Return(&mongo.InsertOneResult{}, nil)

	err := store.CreateReview(ctx, review)

	assert.NoError(t, err)
	mockCol.AssertExpectations(t)
}

func TestCreateReview_Error(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCollection)
	store := &intmongo.ReviewMongo{
		Collection: mockCol,
	}

	review := &models.Review{UserID: "u1", ProductID: "p1", Rating: 5}

	mockCol.On("InsertOne", ctx, mock.Anything).Return((*mongo.InsertOneResult)(nil), errors.New("insert error"))

	err := store.CreateReview(ctx, review)

	assert.Error(t, err)
}

func TestUpdateReviewByID_Success(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCollection)
	store := &intmongo.ReviewMongo{
		Collection: mockCol,
	}

	updated := &models.Review{
		Rating:    4,
		Comment:   "updated",
		MediaURLs: []string{},
	}

	mockCol.On("UpdateOne",
		ctx,
		bson.M{"_id": "rev123"},
		mock.MatchedBy(func(update any) bool {
			updMap, ok := update.(bson.M)
			if !ok {
				return false
			}
			setMap, ok := updMap["$set"].(bson.M)
			if !ok {
				return false
			}

			return setMap["rating"] == 4 &&
				setMap["comment"] == "updated" &&
				reflect.DeepEqual(setMap["media_urls"], []string{})
		}),
	).Return(&mongo.UpdateResult{}, nil)

	err := store.UpdateReviewByID(ctx, "rev123", updated)
	assert.NoError(t, err)
}

func TestUpdateReviewByID_Error(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCollection)
	store := &intmongo.ReviewMongo{
		Collection: mockCol,
	}

	review := &models.Review{Rating: 2}

	mockCol.On("UpdateOne", ctx, mock.Anything, mock.Anything).
		Return((*mongo.UpdateResult)(nil), errors.New("update error"))

	err := store.UpdateReviewByID(ctx, "bad_id", review)

	assert.Error(t, err)
}

func TestDeleteReviewByID_Success(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCollection)
	store := &intmongo.ReviewMongo{
		Collection: mockCol,
	}

	mockCol.On("DeleteOne", ctx, bson.M{"_id": "del1"}).Return(&mongo.DeleteResult{}, nil)

	err := store.DeleteReviewByID(ctx, "del1")
	assert.NoError(t, err)
}

func TestDeleteReviewByID_Error(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCollection)
	store := &intmongo.ReviewMongo{
		Collection: mockCol,
	}

	mockCol.On("DeleteOne", ctx, mock.Anything).
		Return((*mongo.DeleteResult)(nil), errors.New("delete error"))

	err := store.DeleteReviewByID(ctx, "bad_id")

	assert.Error(t, err)
}

func TestGetReviewsByProductID_Success(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCollection)
	mockCur := new(mockCursor)

	mockCol.On("Find", ctx, bson.M{"product_id": "p1"}).Return(&mongo.Cursor{}, nil)
	mockCur.On("Next", ctx).Return(true).Once()
	mockCur.On("Decode", mock.AnythingOfType("*models.Review")).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Review)
		arg.ID = "r1"
	}).Return(nil)
	mockCur.On("Next", ctx).Return(false)
	mockCur.On("Close", ctx).Return(nil)

	// inject adapter
	orig := intmongo.NewCursorAdapter
	defer func() { intmongo.NewCursorAdapter = orig }()
	intmongo.NewCursorAdapter = func(*mongo.Cursor) intmongo.CursorInterface { return mockCur }

	store := &intmongo.ReviewMongo{Collection: mockCol}
	reviews, err := store.GetReviewsByProductID(ctx, "p1")

	assert.NoError(t, err)
	assert.Len(t, reviews, 1)
	assert.Equal(t, "r1", reviews[0].ID)
}

func TestGetReviewsByProductID_FindError(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCollection)

	mockCol.On("Find", ctx, bson.M{"product_id": "p1"}).
		Return((*mongo.Cursor)(nil), errors.New("find error"))

	store := &intmongo.ReviewMongo{Collection: mockCol}
	_, err := store.GetReviewsByProductID(ctx, "p1")

	assert.Error(t, err)
}

func TestGetReviewsByUserID_Success(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCollection)
	mockCur := new(mockCursor)

	mockCol.On("Find", ctx, bson.M{"user_id": "u1"}).Return(&mongo.Cursor{}, nil)
	mockCur.On("All", ctx, mock.Anything).Run(func(args mock.Arguments) {
		ptr := args.Get(1).(*[]*models.Review)
		*ptr = []*models.Review{{ID: "r2"}}
	}).Return(nil)
	mockCur.On("Close", ctx).Return(nil)

	orig := intmongo.NewCursorAdapter
	defer func() { intmongo.NewCursorAdapter = orig }()
	intmongo.NewCursorAdapter = func(*mongo.Cursor) intmongo.CursorInterface { return mockCur }

	store := &intmongo.ReviewMongo{Collection: mockCol}
	reviews, err := store.GetReviewsByUserID(ctx, "u1")

	assert.NoError(t, err)
	assert.Len(t, reviews, 1)
	assert.Equal(t, "r2", reviews[0].ID)
}

func TestGetReviewsByUserID_AllError(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCollection)
	mockCur := new(mockCursor)

	mockCol.On("Find", ctx, bson.M{"user_id": "u1"}).Return(&mongo.Cursor{}, nil)
	mockCur.On("All", ctx, mock.Anything).Return(errors.New("decode error"))
	mockCur.On("Close", ctx).Return(nil)

	orig := intmongo.NewCursorAdapter
	defer func() { intmongo.NewCursorAdapter = orig }()
	intmongo.NewCursorAdapter = func(*mongo.Cursor) intmongo.CursorInterface { return mockCur }

	store := &intmongo.ReviewMongo{Collection: mockCol}
	_, err := store.GetReviewsByUserID(ctx, "u1")

	assert.Error(t, err)
}

func TestGetReviewByID_Success(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCollection)
	mockRes := new(mockSingleResult)

	mockCol.On("FindOne", ctx, bson.M{"_id": "r1"}).Return(&mongo.SingleResult{})
	mockRes.On("Decode", mock.AnythingOfType("*models.Review")).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Review)
		arg.ID = "r1"
	}).Return(nil)

	orig := intmongo.NewSingleResultAdapter
	defer func() { intmongo.NewSingleResultAdapter = orig }()
	intmongo.NewSingleResultAdapter = func(*mongo.SingleResult) intmongo.SingleResultInterface { return mockRes }

	store := &intmongo.ReviewMongo{Collection: mockCol}
	review, err := store.GetReviewByID(ctx, "r1")

	assert.NoError(t, err)
	assert.Equal(t, "r1", review.ID)
}

func TestGetReviewByID_DecodeError(t *testing.T) {
	ctx := context.Background()
	mockCol := new(mockCollection)
	mockRes := new(mockSingleResult)

	mockCol.On("FindOne", ctx, bson.M{"_id": "r1"}).Return(&mongo.SingleResult{})
	mockRes.On("Decode", mock.Anything).Return(errors.New("decode fail"))

	orig := intmongo.NewSingleResultAdapter
	defer func() { intmongo.NewSingleResultAdapter = orig }()
	intmongo.NewSingleResultAdapter = func(*mongo.SingleResult) intmongo.SingleResultInterface { return mockRes }

	store := &intmongo.ReviewMongo{Collection: mockCol}
	_, err := store.GetReviewByID(ctx, "r1")

	assert.Error(t, err)
}
