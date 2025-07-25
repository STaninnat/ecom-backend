// Package utils provides utility functions and helpers used throughout the ecom-backend project.
package utils

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// shutdown.go: Implements graceful server shutdown and MongoDB disconnect logic on OS signals.

// ServerWithShutdown is an interface for servers that support graceful shutdown via a Shutdown method.
type ServerWithShutdown interface {
	Shutdown(ctx context.Context) error
}

// APIConfigWithDisconnect is an interface for configs that can disconnect from MongoDB via a DisconnectMongoDB method.
type APIConfigWithDisconnect interface {
	DisconnectMongoDB(ctx context.Context) error
}

// GracefulShutdown handles OS signals to gracefully shut down the server and disconnect from MongoDB with a timeout.
// It listens for interrupt or termination signals, shuts down the server, and disconnects MongoDB, logging the results.
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
