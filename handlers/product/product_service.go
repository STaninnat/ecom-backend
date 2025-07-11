package producthandlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/google/uuid"
)

// --- Interfaces for DB and Transaction ---
type ProductDBQueries interface {
	WithTx(tx ProductDBTx) ProductDBQueries
	CreateProduct(ctx context.Context, params database.CreateProductParams) error
	UpdateProduct(ctx context.Context, params database.UpdateProductParams) error
	DeleteProductByID(ctx context.Context, id string) error
	GetAllProducts(ctx context.Context) ([]database.Product, error)
	GetAllActiveProducts(ctx context.Context) ([]database.Product, error)
	GetProductByID(ctx context.Context, id string) (database.Product, error)
	GetActiveProductByID(ctx context.Context, id string) (database.Product, error)
	FilterProducts(ctx context.Context, params database.FilterProductsParams) ([]database.Product, error)
}

type ProductDBConn interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (ProductDBTx, error)
}

type ProductDBTx interface {
	Commit() error
	Rollback() error
}

// --- Adapters for sqlc-generated types ---
type ProductDBQueriesAdapter struct {
	*database.Queries
}

func (a *ProductDBQueriesAdapter) WithTx(tx ProductDBTx) ProductDBQueries {
	return &ProductDBQueriesAdapter{a.Queries.WithTx(tx.(*sql.Tx))}
}
func (a *ProductDBQueriesAdapter) CreateProduct(ctx context.Context, params database.CreateProductParams) error {
	return a.Queries.CreateProduct(ctx, params)
}
func (a *ProductDBQueriesAdapter) UpdateProduct(ctx context.Context, params database.UpdateProductParams) error {
	return a.Queries.UpdateProduct(ctx, params)
}
func (a *ProductDBQueriesAdapter) DeleteProductByID(ctx context.Context, id string) error {
	return a.Queries.DeleteProductByID(ctx, id)
}
func (a *ProductDBQueriesAdapter) GetAllProducts(ctx context.Context) ([]database.Product, error) {
	return a.Queries.GetAllProducts(ctx)
}
func (a *ProductDBQueriesAdapter) GetAllActiveProducts(ctx context.Context) ([]database.Product, error) {
	return a.Queries.GetAllActiveProducts(ctx)
}
func (a *ProductDBQueriesAdapter) GetProductByID(ctx context.Context, id string) (database.Product, error) {
	return a.Queries.GetProductByID(ctx, id)
}
func (a *ProductDBQueriesAdapter) GetActiveProductByID(ctx context.Context, id string) (database.Product, error) {
	return a.Queries.GetActiveProductByID(ctx, id)
}
func (a *ProductDBQueriesAdapter) FilterProducts(ctx context.Context, params database.FilterProductsParams) ([]database.Product, error) {
	return a.Queries.FilterProducts(ctx, params)
}

type ProductDBConnAdapter struct {
	*sql.DB
}

func (a *ProductDBConnAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (ProductDBTx, error) {
	tx, err := a.DB.BeginTx(ctx, opts)
	return tx, err
}

// --- Service Implementation ---
type productServiceImpl struct {
	db     ProductDBQueries
	dbConn ProductDBConn
}

// ProductService defines the business logic interface for product operations
// Add more methods as needed for product-related features
// (e.g., Create, Update, Delete, Get, Filter, etc.)
type ProductService interface {
	CreateProduct(ctx context.Context, params ProductRequest) (string, error)
	UpdateProduct(ctx context.Context, params ProductRequest) error
	DeleteProduct(ctx context.Context, productID string) error
	GetAllProducts(ctx context.Context, isAdmin bool) ([]database.Product, error)
	GetProductByID(ctx context.Context, productID string, isAdmin bool) (database.Product, error)
	FilterProducts(ctx context.Context, params FilterProductsRequest) ([]database.Product, error)
}

func NewProductService(db *database.Queries, dbConn *sql.DB) ProductService {
	return &productServiceImpl{
		db:     &ProductDBQueriesAdapter{db},
		dbConn: &ProductDBConnAdapter{dbConn},
	}
}

// CreateProduct creates a new product
func (s *productServiceImpl) CreateProduct(ctx context.Context, params ProductRequest) (string, error) {
	if s.dbConn == nil {
		return "", &handlers.AppError{Code: "transaction_error", Message: "DB connection is nil", Err: fmt.Errorf("dbConn is nil")}
	}
	if params.CategoryID == "" || params.Name == "" || params.Price <= 0 || params.Stock < 0 {
		return "", &handlers.AppError{Code: "invalid_request", Message: "Missing or invalid required fields"}
	}
	id := uuid.New().String()
	timeNow := time.Now().UTC()
	isActive := true
	if params.IsActive != nil {
		isActive = *params.IsActive
	}
	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return "", &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer tx.Rollback()
	queries := s.db.WithTx(tx)
	err = queries.CreateProduct(ctx, database.CreateProductParams{
		ID:          id,
		CategoryID:  utils.ToNullString(params.CategoryID),
		Name:        params.Name,
		Description: utils.ToNullString(params.Description),
		Price:       fmt.Sprintf("%.2f", params.Price),
		Stock:       params.Stock,
		ImageUrl:    utils.ToNullString(params.ImageURL),
		IsActive:    isActive,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	})
	if err != nil {
		return "", &handlers.AppError{Code: "create_product_error", Message: "Error creating product", Err: err}
	}
	if err = tx.Commit(); err != nil {
		return "", &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}
	return id, nil
}

// UpdateProduct updates an existing product
func (s *productServiceImpl) UpdateProduct(ctx context.Context, params ProductRequest) error {
	if s.dbConn == nil {
		return &handlers.AppError{Code: "transaction_error", Message: "DB connection is nil", Err: fmt.Errorf("dbConn is nil")}
	}
	if params.ID == "" || params.CategoryID == "" || params.Name == "" || params.Price <= 0 || params.Stock < 0 {
		return &handlers.AppError{Code: "invalid_request", Message: "Missing or invalid required fields"}
	}
	isActive := true
	if params.IsActive != nil {
		isActive = *params.IsActive
	}
	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer tx.Rollback()
	queries := s.db.WithTx(tx)
	err = queries.UpdateProduct(ctx, database.UpdateProductParams{
		ID:          params.ID,
		CategoryID:  utils.ToNullString(params.CategoryID),
		Name:        params.Name,
		Description: utils.ToNullString(params.Description),
		Price:       fmt.Sprintf("%.2f", params.Price),
		Stock:       params.Stock,
		ImageUrl:    utils.ToNullString(params.ImageURL),
		IsActive:    isActive,
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		return &handlers.AppError{Code: "update_failed", Message: "Error updating product", Err: err}
	}
	if err = tx.Commit(); err != nil {
		return &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}
	return nil
}

// DeleteProduct deletes a product by ID
func (s *productServiceImpl) DeleteProduct(ctx context.Context, productID string) error {
	if s.dbConn == nil {
		return &handlers.AppError{Code: "transaction_error", Message: "DB connection is nil", Err: fmt.Errorf("dbConn is nil")}
	}
	if productID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Product ID is required"}
	}
	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer tx.Rollback()
	queries := s.db.WithTx(tx)
	_, err = queries.GetProductByID(ctx, productID)
	if err != nil {
		return &handlers.AppError{Code: "product_not_found", Message: "Product not found", Err: err}
	}
	err = queries.DeleteProductByID(ctx, productID)
	if err != nil {
		return &handlers.AppError{Code: "delete_product_error", Message: "Error deleting product", Err: err}
	}
	if err = tx.Commit(); err != nil {
		return &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}
	return nil
}

// GetAllProducts returns all products (admin: all, non-admin: only active)
func (s *productServiceImpl) GetAllProducts(ctx context.Context, isAdmin bool) ([]database.Product, error) {
	if s.db == nil {
		return nil, &handlers.AppError{Code: "transaction_error", Message: "DB is nil", Err: fmt.Errorf("db is nil")}
	}
	if isAdmin {
		return s.db.GetAllProducts(ctx)
	}
	return s.db.GetAllActiveProducts(ctx)
}

// GetProductByID returns a product by ID (admin: all, non-admin: only active)
func (s *productServiceImpl) GetProductByID(ctx context.Context, productID string, isAdmin bool) (database.Product, error) {
	if s.db == nil {
		return database.Product{}, &handlers.AppError{Code: "transaction_error", Message: "DB is nil", Err: fmt.Errorf("db is nil")}
	}
	if productID == "" {
		return database.Product{}, &handlers.AppError{Code: "invalid_request", Message: "Missing product ID"}
	}
	if isAdmin {
		return s.db.GetProductByID(ctx, productID)
	}
	return s.db.GetActiveProductByID(ctx, productID)
}

// FilterProducts filters products by various criteria
func (s *productServiceImpl) FilterProducts(ctx context.Context, params FilterProductsRequest) ([]database.Product, error) {
	if s.db == nil {
		return nil, &handlers.AppError{Code: "transaction_error", Message: "DB is nil", Err: fmt.Errorf("db is nil")}
	}
	return s.db.FilterProducts(ctx, database.FilterProductsParams{
		CategoryID: params.CategoryID.NullString,
		IsActive:   params.IsActive.NullBool,
		MinPrice: sql.NullString{
			String: fmt.Sprintf("%f", params.MinPrice.Float64),
			Valid:  params.MinPrice.Valid,
		},
		MaxPrice: sql.NullString{
			String: fmt.Sprintf("%f", params.MaxPrice.Float64),
			Valid:  params.MaxPrice.Valid,
		},
	})
}
