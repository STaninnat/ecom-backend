package handlers

import (
	"net/http"

	"github.com/STaninnat/ecom-backend/middlewares"
)

// HandlerOptionalMiddleware creates middleware that optionally authenticates users
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
