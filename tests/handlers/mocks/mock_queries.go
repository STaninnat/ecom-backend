package mocks

import (
	"context"

	"github.com/STaninnat/ecom-backend/internal/database"
)

type DBQuerier interface {
	CheckUserExistsByName(ctx context.Context, name string) (bool, error)
	CheckUserExistsByEmail(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, arg database.CreateUserParams) error
}

type MockQueries struct {
	CheckUserExistsByNameFn  func(ctx context.Context, name string) (bool, error)
	CheckUserExistsByEmailFn func(ctx context.Context, email string) (bool, error)
	CreateUserFn             func(ctx context.Context, arg database.CreateUserParams) error
}

func (m *MockQueries) CheckUserExistsByName(ctx context.Context, name string) (bool, error) {
	return m.CheckUserExistsByNameFn(ctx, name)
}

func (m *MockQueries) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return m.CheckUserExistsByEmailFn(ctx, email)
}

func (m *MockQueries) CreateUser(ctx context.Context, arg database.CreateUserParams) error {
	return m.CreateUserFn(ctx, arg)
}
