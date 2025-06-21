package intmongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ===== Review =====

// --- INTERFACES ---
type CollectionInterface interface {
	InsertOne(ctx context.Context, document any) (*mongo.InsertOneResult, error)
	Find(ctx context.Context, filter any) (*mongo.Cursor, error)
	FindOne(ctx context.Context, filter any) *mongo.SingleResult
	UpdateOne(ctx context.Context, filter any, update any) (*mongo.UpdateResult, error)
	DeleteOne(ctx context.Context, filter any) (*mongo.DeleteResult, error)
}

type CursorInterface interface {
	Next(ctx context.Context) bool
	Decode(val any) error
	All(ctx context.Context, results any) error
	Close(ctx context.Context) error
}

type SingleResultInterface interface {
	Decode(val any) error
}

// --- ADAPTERS ---
type MongoCollectionAdapter struct {
	Inner *mongo.Collection
}

func (m *MongoCollectionAdapter) InsertOne(ctx context.Context, doc any) (*mongo.InsertOneResult, error) {
	return m.Inner.InsertOne(ctx, doc)
}

func (m *MongoCollectionAdapter) Find(ctx context.Context, filter any) (*mongo.Cursor, error) {
	return m.Inner.Find(ctx, filter)
}

func (m *MongoCollectionAdapter) FindOne(ctx context.Context, filter any) *mongo.SingleResult {
	return m.Inner.FindOne(ctx, filter)
}

func (m *MongoCollectionAdapter) UpdateOne(ctx context.Context, filter any, update any) (*mongo.UpdateResult, error) {
	return m.Inner.UpdateOne(ctx, filter, update)
}

func (m *MongoCollectionAdapter) DeleteOne(ctx context.Context, filter any) (*mongo.DeleteResult, error) {
	return m.Inner.DeleteOne(ctx, filter)
}

type MongoCursorAdapter struct {
	Inner *mongo.Cursor
}

func (c *MongoCursorAdapter) Next(ctx context.Context) bool {
	return c.Inner.Next(ctx)
}

func (c *MongoCursorAdapter) Decode(val any) error {
	return c.Inner.Decode(val)
}

func (c *MongoCursorAdapter) All(ctx context.Context, results any) error {
	return c.Inner.All(ctx, results)
}

func (c *MongoCursorAdapter) Close(ctx context.Context) error {
	return c.Inner.Close(ctx)
}

type MongoSingleResultAdapter struct {
	Inner *mongo.SingleResult
}

func (r *MongoSingleResultAdapter) Decode(val any) error {
	return r.Inner.Decode(val)
}

// -- Allow mocking adapters --
var (
	NewCursorAdapter       = func(cur *mongo.Cursor) CursorInterface { return &MongoCursorAdapter{Inner: cur} }
	NewSingleResultAdapter = func(res *mongo.SingleResult) SingleResultInterface { return &MongoSingleResultAdapter{Inner: res} }
)

// ===== Cart =====

// --- INTERFACES & ADAPTERS ---
type CartCollectionInterface interface {
	FindOne(ctx context.Context, filter any) CartSingleResultInterface
	UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error)
}

type CartSingleResultInterface interface {
	Decode(val any) error
}

// --- ADAPTER ---
type MongoCartCollectionAdapter struct {
	Inner *mongo.Collection
}

func (m *MongoCartCollectionAdapter) FindOne(ctx context.Context, filter any) CartSingleResultInterface {
	return &MongoCartSingleResultAdapter{Inner: m.Inner.FindOne(ctx, filter)}
}

func (m *MongoCartCollectionAdapter) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	return m.Inner.UpdateOne(ctx, filter, update, opts...)
}

type MongoCartSingleResultAdapter struct {
	Inner *mongo.SingleResult
}

func (r *MongoCartSingleResultAdapter) Decode(val any) error {
	return r.Inner.Decode(val)
}

// -- Allow mocking adapters --
var NewCartSingleResultAdapter = func(res *mongo.SingleResult) CartSingleResultInterface {
	return &MongoCartSingleResultAdapter{Inner: res}
}
