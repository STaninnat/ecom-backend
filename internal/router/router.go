package router

import (
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
)

func SetupRouter(handlersCfg *handlers.HandlersConfig, logger *logrus.Logger) *chi.Mux {
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

	v1Router := chi.NewRouter()

	v1Router.Get("/healthz", handlers.HandlerReadiness)
	v1Router.Get("/error", handlers.HandlerError)

	v1Router.Post("/auth/signup", handlersCfg.HandlerSignUp)
	v1Router.Post("/auth/signin", handlersCfg.HandlerSignIn)
	v1Router.Post("/auth/signout", handlersCfg.HandlerSignOut)
	v1Router.Post("/auth/refresh", handlersCfg.HandlerRefreshToken)

	v1Router.Get("/auth/google/signin", handlersCfg.HandlerGoogleSignIn)
	v1Router.Get("/auth/google/callback", handlersCfg.HandlerGoogleCallback)

	router.Mount("/v1", v1Router)
	return router
}
