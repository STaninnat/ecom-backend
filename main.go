package main

import (
	"log"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/router"
	"github.com/STaninnat/ecom-backend/utils"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env.development")
	if err != nil {
		log.Printf("Warning: assuming default configuration, env unreadable: %v", err)
	}

	logger := utils.InitLogger()
	handlersConfig := handlers.SetupHandlersConfig(logger)

	port := handlersConfig.APIConfig.Port

	r := &router.RouterConfig{HandlersConfig: handlersConfig}
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r.SetupRouter(logger),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Serving on port: %s\n", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v\n", err)
	}
}
