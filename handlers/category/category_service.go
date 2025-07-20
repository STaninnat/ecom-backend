package categoryhandlers

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
// CategoryDBQueries defines the interface for category-related database operations.
type CategoryDBQueries interface {
	WithTx(tx CategoryDBTx) CategoryDBQueries
	CreateCategory(ctx context.Context, params database.CreateCategoryParams) error
	UpdateCategories(ctx context.Context, params database.UpdateCategoriesParams) error
	DeleteCategory(ctx context.Context, id string) error
	GetAllCategories(ctx context.Context) ([]database.Category, error)
}

// CategoryDBConn defines the interface for beginning database transactions for category operations.
type CategoryDBConn interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (CategoryDBTx, error)
}

// CategoryDBTx defines the interface for a database transaction used in category operations.
type CategoryDBTx interface {
	Commit() error
	Rollback() error
}

// --- Adapters for sqlc-generated types ---
// CategoryDBQueriesAdapter adapts sqlc-generated Queries to the CategoryDBQueries interface.
type CategoryDBQueriesAdapter struct {
	*database.Queries
}

func (a *CategoryDBQueriesAdapter) WithTx(tx CategoryDBTx) CategoryDBQueries {
	return &CategoryDBQueriesAdapter{a.Queries.WithTx(tx.(*sql.Tx))}
}

func (a *CategoryDBQueriesAdapter) CreateCategory(ctx context.Context, params database.CreateCategoryParams) error {
	return a.Queries.CreateCategory(ctx, params)
}

func (a *CategoryDBQueriesAdapter) UpdateCategories(ctx context.Context, params database.UpdateCategoriesParams) error {
	return a.Queries.UpdateCategories(ctx, params)
}

func (a *CategoryDBQueriesAdapter) DeleteCategory(ctx context.Context, id string) error {
	return a.Queries.DeleteCategory(ctx, id)
}

func (a *CategoryDBQueriesAdapter) GetAllCategories(ctx context.Context) ([]database.Category, error) {
	return a.Queries.GetAllCategories(ctx)
}

// CategoryDBConnAdapter adapts a sql.DB to the CategoryDBConn interface.
type CategoryDBConnAdapter struct {
	*sql.DB
}

func (a *CategoryDBConnAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (CategoryDBTx, error) {
	tx, err := a.DB.BeginTx(ctx, opts)
	return tx, err
}

// --- Service Implementation ---
type categoryServiceImpl struct {
	db     CategoryDBQueries
	dbConn CategoryDBConn
}

// CategoryService defines the business logic interface for category operations.
// Provides methods for creating, updating, deleting, and retrieving categories.
type CategoryService interface {
	CreateCategory(ctx context.Context, params CategoryRequest) (string, error)
	UpdateCategory(ctx context.Context, params CategoryRequest) error
	DeleteCategory(ctx context.Context, categoryID string) error
	GetAllCategories(ctx context.Context) ([]database.Category, error)
}

// CategoryRequest represents the request parameters for category operations.
type CategoryRequest struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CategoryResponse represents the category data returned to the client.
type CategoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewCategoryService creates a new CategoryService with the provided database query and connection adapters.
// Returns a CategoryService implementation.
func NewCategoryService(db *database.Queries, dbConn *sql.DB) CategoryService {
	var dbQueries CategoryDBQueries
	var dbConnection CategoryDBConn

	if db != nil {
		dbQueries = &CategoryDBQueriesAdapter{db}
	}
	if dbConn != nil {
		dbConnection = &CategoryDBConnAdapter{dbConn}
	}

	return &categoryServiceImpl{
		db:     dbQueries,
		dbConn: dbConnection,
	}
}

// CreateCategory creates a new category.
// Validates the request, creates the category in a transaction, and returns the new category ID or an error.
func (s *categoryServiceImpl) CreateCategory(ctx context.Context, params CategoryRequest) (string, error) {
	if s.dbConn == nil {
		return "", &handlers.AppError{Code: "transaction_error", Message: "DB connection is nil", Err: fmt.Errorf("dbConn is nil")}
	}
	if params.Name == "" {
		return "", &handlers.AppError{Code: "invalid_request", Message: "Category name is required"}
	}
	if len(params.Name) > 100 {
		return "", &handlers.AppError{Code: "invalid_request", Message: "Category name too long (max 100 characters)"}
	}
	if len(params.Description) > 500 {
		return "", &handlers.AppError{Code: "invalid_request", Message: "Category description too long (max 500 characters)"}
	}

	id := uuid.New().String()
	timeNow := time.Now().UTC()

	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return "", &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer tx.Rollback()

	queries := s.db.WithTx(tx)

	err = queries.CreateCategory(ctx, database.CreateCategoryParams{
		ID:          id,
		Name:        params.Name,
		Description: utils.ToNullString(params.Description),
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	})
	if err != nil {
		return "", &handlers.AppError{Code: "create_category_error", Message: "Error creating category", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return "", &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return id, nil
}

// UpdateCategory updates an existing category.
// Validates the request, updates the category in a transaction, and returns an error if unsuccessful.
func (s *categoryServiceImpl) UpdateCategory(ctx context.Context, params CategoryRequest) error {
	if s.dbConn == nil {
		return &handlers.AppError{Code: "transaction_error", Message: "DB connection is nil", Err: fmt.Errorf("dbConn is nil")}
	}
	if params.ID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Category ID is required"}
	}
	if params.Name == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Category name is required"}
	}
	if len(params.Name) > 100 {
		return &handlers.AppError{Code: "invalid_request", Message: "Category name too long (max 100 characters)"}
	}
	if len(params.Description) > 500 {
		return &handlers.AppError{Code: "invalid_request", Message: "Category description too long (max 500 characters)"}
	}

	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer tx.Rollback()

	queries := s.db.WithTx(tx)

	err = queries.UpdateCategories(ctx, database.UpdateCategoriesParams{
		ID:          params.ID,
		Name:        params.Name,
		Description: utils.ToNullString(params.Description),
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		return &handlers.AppError{Code: "update_category_error", Message: "Error updating category", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return nil
}

// DeleteCategory deletes a category by ID.
// Validates the ID, deletes the category in a transaction, and returns an error if unsuccessful.
func (s *categoryServiceImpl) DeleteCategory(ctx context.Context, categoryID string) error {
	if s.dbConn == nil {
		return &handlers.AppError{Code: "transaction_error", Message: "DB connection is nil", Err: fmt.Errorf("dbConn is nil")}
	}
	if categoryID == "" {
		return &handlers.AppError{Code: "invalid_request", Message: "Category ID is required"}
	}

	tx, err := s.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return &handlers.AppError{Code: "transaction_error", Message: "Error starting transaction", Err: err}
	}
	defer tx.Rollback()

	queries := s.db.WithTx(tx)

	err = queries.DeleteCategory(ctx, categoryID)
	if err != nil {
		return &handlers.AppError{Code: "delete_category_error", Message: "Error deleting category", Err: err}
	}

	if err = tx.Commit(); err != nil {
		return &handlers.AppError{Code: "commit_error", Message: "Error committing transaction", Err: err}
	}

	return nil
}

// GetAllCategories returns all categories.
// Returns a list of all categories or an error.
func (s *categoryServiceImpl) GetAllCategories(ctx context.Context) ([]database.Category, error) {
	if s.db == nil {
		return nil, &handlers.AppError{Code: "database_error", Message: "DB is nil", Err: fmt.Errorf("db is nil")}
	}

	return s.db.GetAllCategories(ctx)
}

// CategoryError is an alias for handlers.AppError, used for category-related errors.
type CategoryError = handlers.AppError
