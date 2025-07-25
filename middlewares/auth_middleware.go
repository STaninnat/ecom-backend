// Package middlewares provides HTTP middleware components for request processing in the ecom-backend project.
package middlewares

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// auth_middleware.go: Implements authentication and authorization middleware for protected routes.

// AuthHandler is a handler type for authenticated endpoints, receiving a database.User.
type AuthHandler func(http.ResponseWriter, *http.Request, database.User)

// OptionalHandler is a handler type for optionally authenticated endpoints, receiving a pointer to database.User (or nil).
type OptionalHandler func(http.ResponseWriter, *http.Request, *database.User)

// LoggerService is an interface for logging operations, supporting structured error logging.
type LoggerService interface {
	WithError(err error) interface{ Error(args ...any) }
	Error(args ...any)
}

// AuthService is an interface for authentication operations, such as validating access tokens.
type AuthService interface {
	ValidateAccessToken(tokenString, secret string) (*Claims, error)
}

// UserService is an interface for user operations, such as fetching a user by ID.
type UserService interface {
	GetUserByID(ctx context.Context, id string) (database.User, error)
}

// RequestMetadataService is an interface for extracting request metadata such as IP and user agent.
type RequestMetadataService interface {
	GetIPAddress(r *http.Request) string
	GetUserAgent(r *http.Request) string
}

// Claims represents JWT claims for authentication.
type Claims struct {
	UserID string `json:"user_id"`
}

// LogHandlerError logs an error with structured logging, using the logger service and additional context information.
func LogHandlerError(_ context.Context, logger LoggerService, _ string, _ string, logMsg, _ string, _ string, err error) {
	if err != nil {
		logger.WithError(err).Error(logMsg)
	} else {
		logger.Error(logMsg)
	}
}

// GetRequestMetadata extracts IP address and user agent from the request using the metadata service.
func GetRequestMetadata(metadataService RequestMetadataService, r *http.Request) (ip string, userAgent string) {
	ip = metadataService.GetIPAddress(r)
	userAgent = metadataService.GetUserAgent(r)
	return
}

// CreateAuthMiddleware creates authentication middleware that validates JWT tokens and fetches the user from the database.
// Logs authentication failures and returns appropriate HTTP error responses.
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
					ctx,
					loggerService,
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
					ctx,
					loggerService,
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
					ctx,
					loggerService,
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

// CreateAdminOnlyMiddleware creates middleware that only allows admin users, wrapping the standard auth middleware.
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
					ctx,
					loggerService,
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

// CreateOptionalAuthMiddleware creates middleware that optionally authenticates users, passing nil for unauthenticated requests.
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
						ctx,
						loggerService,
						"optional_auth",
						"invalid token",
						"token validation failed",
						ip, userAgent, err,
					)
				} else {
					u, err := userService.GetUserByID(ctx, claims.UserID)
					if err != nil {
						LogHandlerError(
							ctx,
							loggerService,
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
