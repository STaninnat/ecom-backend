package utils

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ServerWithShutdown interface {
	Shutdown(ctx context.Context) error
}

type APIConfigWithDisconnect interface {
	DisconnectMongoDB(ctx context.Context) error
}

func GracefulShutdown(srv ServerWithShutdown, cfg APIConfigWithDisconnect, timeout time.Duration) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("Shutdown signal received")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server shutdown gracefully.")
	}

	if err := cfg.DisconnectMongoDB(context.Background()); err != nil {
		log.Printf("Error disconnecting MongoDB: %v", err)
	} else {
		log.Println("MongoDB disconnected.")
	}
}
