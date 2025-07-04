package middlewares

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// AuthHandler type for authenticated handlers
type AuthHandler func(http.ResponseWriter, *http.Request, database.User)

// OptionalHandler type for optionally authenticated handlers
type OptionalHandler func(http.ResponseWriter, *http.Request, *database.User)

// LoggerService interface for logging operations
type LoggerService interface {
	WithError(err error) interface{ Error(args ...any) }
	Error(args ...any)
}

// AuthService interface for authentication operations
type AuthService interface {
	ValidateAccessToken(tokenString, secret string) (*Claims, error)
}

// UserService interface for user operations
type UserService interface {
	GetUserByID(ctx context.Context, id string) (database.User, error)
}

// RequestMetadataService interface for request metadata
type RequestMetadataService interface {
	GetIPAddress(r *http.Request) string
	GetUserAgent(r *http.Request) string
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
}

// LogHandlerError logs an error with structured logging
func LogHandlerError(logger LoggerService, ctx context.Context, action, details, logMsg, ip, ua string, err error) {
	if err != nil {
		logger.WithError(err).Error(logMsg)
	} else {
		logger.Error(logMsg)
	}
}

// GetRequestMetadata extracts IP address and user agent from the request
func GetRequestMetadata(metadataService RequestMetadataService, r *http.Request) (ip string, userAgent string) {
	ip = metadataService.GetIPAddress(r)
	userAgent = metadataService.GetUserAgent(r)
	return
}

// CreateAuthMiddleware creates authentication middleware that validates JWT tokens
func CreateAuthMiddleware(
	authService AuthService,
	userService UserService,
	loggerService LoggerService,
	metadataService RequestMetadataService,
	jwtSecret string,
) func(AuthHandler) http.HandlerFunc {
	return func(handler AuthHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip, userAgent := GetRequestMetadata(metadataService, r)
			ctx := r.Context()

			cookie, err := r.Cookie("access_token")
			if err != nil {
				LogHandlerError(
					loggerService,
					ctx,
					"auth_middleware",
					"missing access token cookie",
					"Access token cookie not found",
					ip, userAgent, err,
				)
				RespondWithError(w, http.StatusUnauthorized, "Couldn't find token")
				return
			}

			token := cookie.Value

			claims, err := authService.ValidateAccessToken(token, jwtSecret)
			if err != nil {
				LogHandlerError(
					loggerService,
					ctx,
					"auth_middleware",
					"invalid access token",
					"Access token validation failed",
					ip, userAgent, err,
				)
				RespondWithError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			user, err := userService.GetUserByID(ctx, claims.UserID)
			if err != nil {
				LogHandlerError(
					loggerService,
					ctx,
					"auth_middleware",
					"user lookup failed",
					"Failed to fetch user from database",
					ip, userAgent, err,
				)
				RespondWithError(w, http.StatusInternalServerError, "Couldn't get user info")
				return
			}

			handler(w, r, user)
		}
	}
}

// CreateAdminOnlyMiddleware creates middleware that only allows admin users
func CreateAdminOnlyMiddleware(
	authService AuthService,
	userService UserService,
	loggerService LoggerService,
	metadataService RequestMetadataService,
	jwtSecret string,
) func(AuthHandler) http.HandlerFunc {
	authMiddleware := CreateAuthMiddleware(authService, userService, loggerService, metadataService, jwtSecret)

	return func(handler AuthHandler) http.HandlerFunc {
		return authMiddleware(func(w http.ResponseWriter, r *http.Request, user database.User) {
			ip, userAgent := GetRequestMetadata(metadataService, r)
			ctx := r.Context()

			if user.Role != "admin" {
				LogHandlerError(
					loggerService,
					ctx,
					"admin_middleware",
					"user is not admin",
					"unauthorized access attempt",
					ip, userAgent, nil,
				)
				RespondWithError(w, http.StatusForbidden, "Access Denied")
				return
			}

			handler(w, r, user)
		})
	}
}

// CreateOptionalAuthMiddleware creates middleware that optionally authenticates users
func CreateOptionalAuthMiddleware(
	authService AuthService,
	userService UserService,
	loggerService LoggerService,
	metadataService RequestMetadataService,
	jwtSecret string,
) func(OptionalHandler) http.HandlerFunc {
	return func(handler OptionalHandler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var user *database.User
			ip, userAgent := GetRequestMetadata(metadataService, r)
			ctx := r.Context()

			cookie, err := r.Cookie("access_token")
			if err == nil {
				token := cookie.Value

				claims, err := authService.ValidateAccessToken(token, jwtSecret)
				if err != nil {
					LogHandlerError(
						loggerService,
						ctx,
						"optional_auth",
						"invalid token",
						"token validation failed",
						ip, userAgent, err,
					)
				} else {
					u, err := userService.GetUserByID(ctx, claims.UserID)
					if err != nil {
						LogHandlerError(
							loggerService,
							ctx,
							"optional_auth",
							"user not found",
							"user lookup failed",
							ip, userAgent, err,
						)
					} else {
						user = &u
					}
				}
			}

			handler(w, r, user)
		}
	}
}
