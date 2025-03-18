package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load(".env.development")
	if err != nil {
		log.Printf("Warning: assuming default configuration, env unreadable: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Warning: environment variable is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	issuerName := os.Getenv("ISSUER")
	audienceName := os.Getenv("AUDIENCE")

	redisClient := config.InitRedis()

	apicfg := &config.APIConfig{
		RedisClient: redisClient,
		JWTSecret:   jwtSecret,
		Issuer:      issuerName,
		Audience:    audienceName,
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("Warning: database url environment variable is not set")
		log.Println("Running without CRUD endpoints")
	} else {
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatalf("Warning: can't connect to database: %v\n", err)
		}

		if err := db.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v\n", err)
		}

		dbQueries := database.New(db)
		apicfg.DB = dbQueries

		log.Println("Connected to database successfully...")
	}

	authCfg := &auth.AuthConfig{
		APIConfig: apicfg,
	}
	handlersCfg := &handlers.HandlersConfig{
		APIConfig: apicfg,
		Auth:      authCfg,
	}

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

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

	v1Router.Post("/signup", handlersCfg.HandlerSignUp)

	router.Mount("/v1", v1Router)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Serving on port: %s\n", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v\n", err)
	}
}
