package handlers

import (
	"net/http"

	"github.com/STaninnat/ecom-backend/middlewares"
)

// HandlerAdminOnlyMiddleware creates middleware that only allows admin users for HandlerConfig.
func (cfg *HandlerConfig) HandlerAdminOnlyMiddleware(handler AuthHandler) http.HandlerFunc {
	authService := &handlerConfigAuthAdapter{authService: cfg.AuthService}
	userService := &handlerConfigUserAdapter{userService: cfg.UserService}
	loggerService := &handlerConfigLoggerAdapter{loggerService: cfg.LoggerService}
	metadataService := &handlerConfigMetadataAdapter{metadataService: cfg.RequestMetadataService}
	authMiddleware := middlewares.CreateAdminOnlyMiddleware(
		authService,
		userService,
		loggerService,
		metadataService,
		cfg.JWTSecret,
	)
	return authMiddleware(middlewares.AuthHandler(handler))
}

// HandlerMiddleware creates authentication middleware that validates JWT tokens for HandlerConfig.
func (cfg *HandlerConfig) HandlerMiddleware(handler AuthHandler) http.HandlerFunc {
	authService := &handlerConfigAuthAdapter{authService: cfg.AuthService}
	userService := &handlerConfigUserAdapter{userService: cfg.UserService}
	loggerService := &handlerConfigLoggerAdapter{loggerService: cfg.LoggerService}
	metadataService := &handlerConfigMetadataAdapter{metadataService: cfg.RequestMetadataService}
	authMiddleware := middlewares.CreateAuthMiddleware(
		authService,
		userService,
		loggerService,
		metadataService,
		cfg.JWTSecret,
	)
	return authMiddleware(middlewares.AuthHandler(handler))
}

// HandlerAdminOnlyMiddleware creates middleware that only allows admin users for HandlersConfig (legacy compatibility).
// Legacy compatibility methods for existing HandlersConfig
func (apicfg *HandlersConfig) HandlerAdminOnlyMiddleware(handler AuthHandler) http.HandlerFunc {
	authService := &legacyAuthService{auth: apicfg.Auth}
	userService := &legacyUserService{db: apicfg.DB}
	loggerService := &legacyLoggerService{logger: apicfg.Logger}
	metadataService := &legacyMetadataService{}
	authMiddleware := middlewares.CreateAdminOnlyMiddleware(
		authService,
		userService,
		loggerService,
		metadataService,
		apicfg.JWTSecret,
	)
	return authMiddleware(middlewares.AuthHandler(handler))
}

// HandlerMiddleware returns an HTTP handler with authentication and related middlewares applied.
func (apicfg *HandlersConfig) HandlerMiddleware(handler AuthHandler) http.HandlerFunc {
	authService := &legacyAuthService{auth: apicfg.Auth}
	userService := &legacyUserService{db: apicfg.DB}
	loggerService := &legacyLoggerService{logger: apicfg.Logger}
	metadataService := &legacyMetadataService{}
	authMiddleware := middlewares.CreateAuthMiddleware(
		authService,
		userService,
		loggerService,
		metadataService,
		apicfg.JWTSecret,
	)
	return authMiddleware(middlewares.AuthHandler(handler))
}

// HandlerOptionalMiddleware creates middleware that optionally authenticates users for HandlerConfig.
func (cfg *HandlerConfig) HandlerOptionalMiddleware(handler OptionalHandler) http.HandlerFunc {
	authService := &handlerConfigAuthAdapter{authService: cfg.AuthService}
	userService := &handlerConfigUserAdapter{userService: cfg.UserService}
	loggerService := &handlerConfigLoggerAdapter{loggerService: cfg.LoggerService}
	metadataService := &handlerConfigMetadataAdapter{metadataService: cfg.RequestMetadataService}
	optionalAuthMiddleware := middlewares.CreateOptionalAuthMiddleware(
		authService,
		userService,
		loggerService,
		metadataService,
		cfg.JWTSecret,
	)
	return optionalAuthMiddleware(middlewares.OptionalHandler(handler))
}

// HandlerOptionalMiddleware creates middleware that optionally authenticates users for HandlersConfig (legacy compatibility).
// Legacy compatibility method for existing HandlersConfig
func (apicfg *HandlersConfig) HandlerOptionalMiddleware(handler OptionalHandler) http.HandlerFunc {
	authService := &legacyAuthService{auth: apicfg.Auth}
	userService := &legacyUserService{db: apicfg.DB}
	loggerService := &legacyLoggerService{logger: apicfg.Logger}
	metadataService := &legacyMetadataService{}
	optionalAuthMiddleware := middlewares.CreateOptionalAuthMiddleware(
		authService,
		userService,
		loggerService,
		metadataService,
		apicfg.JWTSecret,
	)
	return optionalAuthMiddleware(middlewares.OptionalHandler(handler))
}
