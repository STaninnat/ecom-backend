package handlers

import (
	"net/http"

	"github.com/STaninnat/ecom-backend/middlewares"
)

// HandlerAdminOnlyMiddleware creates middleware that only allows admin users
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

// HandlerMiddleware creates authentication middleware that validates JWT tokens
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
