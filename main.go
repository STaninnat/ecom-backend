package main

import (
	"log"
	"net/http"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/internal/router"

	_ "github.com/lib/pq"
)

func main() {
	apicfg := config.LoadConfig()
	apicfg.DB = database.ConnectDB()

	authCfg := &auth.AuthConfig{
		APIConfig: apicfg,
	}
	handlersCfg := &handlers.HandlersConfig{
		APIConfig: apicfg,
		Auth:      authCfg,
	}

	r := router.SetupRouter(handlersCfg)
	srv := &http.Server{
		Addr:         ":" + apicfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Serving on port: %s\n", apicfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v\n", err)
	}
}
