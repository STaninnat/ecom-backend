// Package utils provides utility functions and helpers used throughout the ecom-backend project.
package utils

import (
	"bytes"
	"context"
	"log"
	"os"
	"syscall"
	"testing"
	"time"
)

// shutdown_test.go: Tests for graceful server shutdown and MongoDB disconnect logic.

// mockServer is a mock implementation of a server for testing graceful shutdown.
type mockServer struct {
	shutdownCalled bool
	shutdownErr    error
}

// Shutdown simulates shutting down the server and records if it was called.
func (m *mockServer) Shutdown(_ context.Context) error {
	m.shutdownCalled = true
	return m.shutdownErr
}

// mockConfig is a mock implementation for MongoDB disconnect logic in shutdown tests.
type mockConfig struct {
	disconnectCalled bool
	disconnectErr    error
}

// DisconnectMongoDB simulates disconnecting MongoDB and records if it was called.
func (m *mockConfig) DisconnectMongoDB(_ context.Context) error {
	m.disconnectCalled = true
	return m.disconnectErr
}

// TestGracefulShutdown_Success tests GracefulShutdown for a successful shutdown sequence.
func TestGracefulShutdown_Success(t *testing.T) {
	// Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	srv := &mockServer{}
	cfg := &mockConfig{}

	// Simulate signal in goroutine
	done := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		if err := p.Signal(syscall.SIGTERM); err != nil {
			t.Errorf("p.Signal failed: %v", err)
		}
		close(done)
	}()

	GracefulShutdown(srv, cfg, 100*time.Millisecond)
	<-done

	if !srv.shutdownCalled {
		t.Error("expected Shutdown to be called")
	}
	if !cfg.disconnectCalled {
		t.Error("expected DisconnectMongoDB to be called")
	}
	out := buf.String()
	if !containsAll(out, "Shutdown signal received", "Server shutdown gracefully.", "MongoDB disconnected.") {
		t.Errorf("unexpected log output: %q", out)
	}
}

// TestGracefulShutdown_Errors tests GracefulShutdown for error scenarios during shutdown and disconnect.
func TestGracefulShutdown_Errors(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	srv := &mockServer{shutdownErr: context.DeadlineExceeded}
	cfg := &mockConfig{disconnectErr: context.DeadlineExceeded}

	done := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		if err := p.Signal(syscall.SIGTERM); err != nil {
			t.Errorf("p.Signal failed: %v", err)
		}
		close(done)
	}()

	GracefulShutdown(srv, cfg, 100*time.Millisecond)
	<-done

	out := buf.String()
	if !containsAll(out, "Shutdown signal received", "Server forced to shutdown", "Error disconnecting MongoDB") {
		t.Errorf("unexpected log output: %q", out)
	}
}

// containsAll checks if all substrings are present in the given string.
func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !bytes.Contains([]byte(s), []byte(sub)) {
			return false
		}
	}
	return true
}
