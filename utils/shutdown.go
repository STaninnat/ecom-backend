package utils

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/STaninnat/ecom-backend/internal/config"
)

func GracefulShutdown(srv *http.Server, cfg *config.APIConfig) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("Shutdown signal received")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	if err := cfg.DisconnectMongoDB(context.Background()); err != nil {
		log.Println("Error disconnecting MongoDB:", err)
	} else {
		log.Println("MongoDB disconnected.")
	}
}
