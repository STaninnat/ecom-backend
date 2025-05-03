package router

import (
	"github.com/STaninnat/ecom-backend/handlers"
	authhandlers "github.com/STaninnat/ecom-backend/handlers/auth_handler"
	userhandlers "github.com/STaninnat/ecom-backend/handlers/user_handler"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
)

type RouterConfig struct {
	*handlers.HandlersConfig
}

func (apicfg *RouterConfig) SetupRouter(logger *logrus.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(middlewares.RequestIDMiddleware)
	router.Use(middlewares.LoggingMiddleware(
		logger,
		map[string]struct{}{"/v1": {}},
		map[string]struct{}{"/v1/healthz": {}, "/v1/error": {}},
	))

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	authHandlers := &authhandlers.HandlersAuthConfig{HandlersConfig: apicfg.HandlersConfig}
	userHandlers := &userhandlers.HandlersUserConfig{HandlersConfig: apicfg.HandlersConfig}

	v1Router := chi.NewRouter()

	v1Router.Get("/healthz", handlers.HandlerReadiness)
	v1Router.Get("/error", handlers.HandlerError)

	v1Router.Post("/auth/signup", authHandlers.HandlerSignUp)
	v1Router.Post("/auth/signin", authHandlers.HandlerSignIn)
	v1Router.Post("/auth/signout", authHandlers.HandlerSignOut)
	v1Router.Post("/auth/refresh", authHandlers.HandlerRefreshToken)

	v1Router.Get("/auth/google/signin", authHandlers.HandlerGoogleSignIn)
	v1Router.Get("/auth/google/callback", authHandlers.HandlerGoogleCallback)

	v1Router.Get("/users", apicfg.HandlerMiddleware(userHandlers.HandlerGetUser))
	v1Router.Put("/users", apicfg.HandlerMiddleware(userHandlers.HandlerUpdateUser))

	router.Mount("/v1", v1Router)
	return router
}
