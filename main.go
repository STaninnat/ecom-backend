// Package main is the entry point for the ecom-backend application. It sets up the server and routes.
// @title           E-Commerce Backend API
// @version         1.0
// @description     This is the backend API for the e-commerce platform.
// @termsOfService  https://yourdomain.com/terms/
//
// @contact.name   API Support
// @contact.url    https://yourdomain.com/support
// @contact.email  support@yourdomain.com
//
// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT
//
// @host      localhost:8080
// @BasePath  /v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/router"
	"github.com/STaninnat/ecom-backend/utils"

	_ "github.com/lib/pq"

	_ "github.com/STaninnat/ecom-backend/docs"
)

func main() {
	err := godotenv.Load(".env.development")
	if err != nil {
		log.Printf("Warning: assuming default configuration, env unreadable: %v", err)
	}

	logger := utils.InitLogger()
	Config := handlers.SetupHandlersConfig(logger)

	port := Config.Port

	r := &router.Config{Config: Config}
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r.SetupRouter(logger),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Serving on port: %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v\n", err)
		}
	}()

	utils.GracefulShutdown(srv, Config.APIConfig, 10*time.Second)
}
