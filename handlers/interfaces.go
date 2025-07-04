package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	ValidateAccessToken(tokenString, secret string) (*Claims, error)
	GenerateAccessToken(userID string, expiresAt time.Time) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	GenerateTokens(userID string, accessTokenExpiresAt time.Time) (string, string, error)
	StoreRefreshTokenInRedis(r *http.Request, userID, refreshToken, provider string, ttl time.Duration) error
	ValidateRefreshToken(refreshToken string) (string, error)
	ValidateCookieRefreshTokenData(w http.ResponseWriter, r *http.Request) (string, *RefreshTokenData, error)
}

// UserService defines the interface for user database operations
type UserService interface {
	GetUserByID(ctx context.Context, id string) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
	CheckUserExistsByEmail(ctx context.Context, email string) (bool, error)
	CheckUserExistsByName(ctx context.Context, name string) (bool, error)
	CreateUser(ctx context.Context, arg database.CreateUserParams) error
	UpdateUserInfo(ctx context.Context, arg database.UpdateUserInfoParams) error
	UpdateUserRole(ctx context.Context, arg database.UpdateUserRoleParams) error
	CheckExistsAndGetIDByEmail(ctx context.Context, email string) (database.CheckExistsAndGetIDByEmailRow, error)
}

// LoggerService defines the interface for logging operations
type LoggerService interface {
	WithError(err error) *logrus.Entry
	Error(args ...any)
	Info(args ...any)
	Debug(args ...any)
	Warn(args ...any)
}

// RequestMetadataService defines the interface for extracting request metadata
type RequestMetadataService interface {
	GetIPAddress(r *http.Request) string
	GetUserAgent(r *http.Request) string
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	// Add other JWT claims as needed
}

// RefreshTokenData represents refresh token data structure
type RefreshTokenData struct {
	Token    string `json:"token"`
	Provider string `json:"provider"`
}

// HandlerConfig defines the configuration for handlers with interfaces
type HandlerConfig struct {
	AuthService            AuthService
	UserService            UserService
	LoggerService          LoggerService
	RequestMetadataService RequestMetadataService
	JWTSecret              string
	RefreshSecret          string
	Issuer                 string
	Audience               string
	OAuth                  *OAuthConfig
	CustomTokenSource      func(ctx context.Context, refreshToken string) oauth2.TokenSource
}

// OAuthConfig represents OAuth configuration
type OAuthConfig struct {
	Google *oauth2.Config
}

// Handler types for different middleware patterns
type AuthHandler func(http.ResponseWriter, *http.Request, database.User)
type OptionalHandler func(http.ResponseWriter, *http.Request, *database.User)
