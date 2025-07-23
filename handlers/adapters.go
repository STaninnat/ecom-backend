// Package handlers provides core interfaces, configurations, middleware, and utilities to support HTTP request handling, authentication, logging, and user management in the ecom-backend project.
package handlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/sirupsen/logrus"
)

// adapters.go: Provides adapter implementations to integrate various services with handler middleware, including both current and legacy systems.

// Adapters for HandlerConfig
// --------------------------
// handlerConfigAuthAdapter adapts AuthService for use with handler middleware.
type handlerConfigAuthAdapter struct {
	authService AuthService
}

// ValidateAccessToken validates the access token and returns middleware claims.
func (h *handlerConfigAuthAdapter) ValidateAccessToken(tokenString, secret string) (*middlewares.Claims, error) {
	claims, err := h.authService.ValidateAccessToken(tokenString, secret)
	if err != nil {
		return nil, err
	}
	return &middlewares.Claims{UserID: claims.UserID}, nil
}

// handlerConfigUserAdapter adapts UserService for use with handler middleware.
type handlerConfigUserAdapter struct {
	userService UserService
}

// GetUserByID fetches a user by ID from the user service.
func (h *handlerConfigUserAdapter) GetUserByID(ctx context.Context, id string) (database.User, error) {
	return h.userService.GetUserByID(ctx, id)
}

// handlerConfigLoggerAdapter adapts LoggerService for use with handler middleware.
type handlerConfigLoggerAdapter struct {
	loggerService LoggerService
}

// WithError attaches error context to the logger.
func (h *handlerConfigLoggerAdapter) WithError(err error) interface{ Error(args ...any) } {
	return h.loggerService.WithError(err)
}

// Error logs an error message.
func (h *handlerConfigLoggerAdapter) Error(args ...any) {
	h.loggerService.Error(args...)
}

// handlerConfigMetadataAdapter adapts RequestMetadataService for use with handler middleware.
type handlerConfigMetadataAdapter struct {
	metadataService RequestMetadataService
}

// GetIPAddress extracts the IP address from the request.
func (h *handlerConfigMetadataAdapter) GetIPAddress(r *http.Request) string {
	return h.metadataService.GetIPAddress(r)
}

// GetUserAgent extracts the User-Agent header from the request.
func (h *handlerConfigMetadataAdapter) GetUserAgent(r *http.Request) string {
	return h.metadataService.GetUserAgent(r)
}

// Legacy adapters for HandlersConfig
// ----------------------------------
// legacyAuthService adapts legacy Auth for use with handler middleware.
type legacyAuthService struct {
	auth interface {
		ValidateAccessToken(tokenString, secret string) (*auth.Claims, error)
	}
}

// ValidateAccessToken validates the access token and returns middleware claims.
func (l *legacyAuthService) ValidateAccessToken(tokenString, secret string) (*middlewares.Claims, error) {
	claims, err := l.auth.ValidateAccessToken(tokenString, secret)
	if err != nil {
		return nil, err
	}
	return &middlewares.Claims{UserID: claims.UserID}, nil
}

// legacyUserService adapts legacy database.Queries for use with handler middleware.
type legacyUserService struct {
	db *database.Queries
}

// GetUserByID fetches a user by ID from the database.
func (l *legacyUserService) GetUserByID(ctx context.Context, id string) (database.User, error) {
	return l.db.GetUserByID(ctx, id)
}

// legacyLoggerService adapts legacy logrus.Logger for use with handler middleware.
type legacyLoggerService struct {
	logger *logrus.Logger
}

// WithError attaches error context to the logger.
func (l *legacyLoggerService) WithError(err error) interface{ Error(args ...any) } {
	return l.logger.WithError(err)
}

// Error logs an error message.
func (l *legacyLoggerService) Error(args ...any) {
	l.logger.Error(args...)
}

// legacyMetadataService adapts legacy request metadata extraction for use with handler middleware.
type legacyMetadataService struct{}

// GetIPAddress extracts the IP address from the request.
func (l *legacyMetadataService) GetIPAddress(r *http.Request) string {
	ip, _ := GetRequestMetadata(r)
	return ip
}

// GetUserAgent extracts the User-Agent header from the request.
func (l *legacyMetadataService) GetUserAgent(r *http.Request) string {
	return r.UserAgent()
}
