package handlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/sirupsen/logrus"
)

// Adapters for HandlerConfig
// --------------------------
type handlerConfigAuthAdapter struct {
	authService AuthService
}

func (h *handlerConfigAuthAdapter) ValidateAccessToken(tokenString, secret string) (*middlewares.Claims, error) {
	claims, err := h.authService.ValidateAccessToken(tokenString, secret)
	if err != nil {
		return nil, err
	}
	return &middlewares.Claims{UserID: claims.UserID}, nil
}

type handlerConfigUserAdapter struct {
	userService UserService
}

func (h *handlerConfigUserAdapter) GetUserByID(ctx context.Context, id string) (database.User, error) {
	return h.userService.GetUserByID(ctx, id)
}

type handlerConfigLoggerAdapter struct {
	loggerService LoggerService
}

func (h *handlerConfigLoggerAdapter) WithError(err error) interface{ Error(args ...any) } {
	return h.loggerService.WithError(err)
}

func (h *handlerConfigLoggerAdapter) Error(args ...any) {
	h.loggerService.Error(args...)
}

type handlerConfigMetadataAdapter struct {
	metadataService RequestMetadataService
}

func (h *handlerConfigMetadataAdapter) GetIPAddress(r *http.Request) string {
	return h.metadataService.GetIPAddress(r)
}

func (h *handlerConfigMetadataAdapter) GetUserAgent(r *http.Request) string {
	return h.metadataService.GetUserAgent(r)
}

// Legacy adapters for HandlersConfig
// ----------------------------------
type legacyAuthService struct {
	auth interface {
		ValidateAccessToken(tokenString, secret string) (*auth.Claims, error)
	}
}

func (l *legacyAuthService) ValidateAccessToken(tokenString, secret string) (*middlewares.Claims, error) {
	claims, err := l.auth.ValidateAccessToken(tokenString, secret)
	if err != nil {
		return nil, err
	}
	return &middlewares.Claims{UserID: claims.UserID}, nil
}

type legacyUserService struct {
	db *database.Queries
}

func (l *legacyUserService) GetUserByID(ctx context.Context, id string) (database.User, error) {
	return l.db.GetUserByID(ctx, id)
}

type legacyLoggerService struct {
	logger *logrus.Logger
}

func (l *legacyLoggerService) WithError(err error) interface{ Error(args ...any) } {
	return l.logger.WithError(err)
}

func (l *legacyLoggerService) Error(args ...any) {
	l.logger.Error(args...)
}

type legacyMetadataService struct{}

func (l *legacyMetadataService) GetIPAddress(r *http.Request) string {
	ip, _ := GetRequestMetadata(r)
	return ip
}

func (l *legacyMetadataService) GetUserAgent(r *http.Request) string {
	return r.UserAgent()
}
