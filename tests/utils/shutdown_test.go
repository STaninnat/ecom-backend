package utils_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/utils"
)

// mockServer embeds http.Server and overrides Shutdown for testing
type mockServer struct {
	http.Server
	ShutdownCalled bool
}

func (m *mockServer) Shutdown(ctx context.Context) error {
	m.ShutdownCalled = true
	return nil
}

// mockConfig mocks config.APIConfig
type mockConfig struct {
	config.APIConfig
	MongoDisconnected bool
	ShouldFail        bool
}

func (m *mockConfig) DisconnectMongoDB(ctx context.Context) error {
	m.MongoDisconnected = true
	if m.ShouldFail {
		return errors.New("mock disconnect error")
	}
	return nil
}

func TestGracefulShutdown(t *testing.T) {
	// setup mock
	mockSrv := &mockServer{}
	mockCfg := &mockConfig{}

	// simulate sending signal after short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGTERM)
	}()

	utils.GracefulShutdown(mockSrv, mockCfg)

	// assert that shutdown and disconnect happened
	if !mockSrv.ShutdownCalled {
		t.Error("expected server.Shutdown to be called")
	}
	if !mockCfg.MongoDisconnected {
		t.Error("expected DisconnectMongoDB to be called")
	}
}
