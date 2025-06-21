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

func GracefulShutdown(srv ServerWithShutdown, cfg APIConfigWithDisconnect) {
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
