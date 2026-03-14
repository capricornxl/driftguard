package shutdown

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewGracefulShutdown(t *testing.T) {
	logger := logrus.New()
	server := &http.Server{
		Addr: ":8080",
	}
	timeout := 30 * time.Second

	gs := NewGracefulShutdown(server, logger, timeout)

	assert.NotNil(t, gs)
	assert.Equal(t, server, gs.server)
	assert.Equal(t, timeout, gs.shutdownTimeout)
	assert.NotNil(t, gs.done)
}

func TestGracefulShutdown_Done(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Suppress logs during tests

	server := &http.Server{Addr: ":8080"}
	gs := NewGracefulShutdown(server, logger, 30*time.Second)

	doneChan := gs.Done()
	assert.NotNil(t, doneChan)

	// Channel should not be closed initially
	select {
	case <-doneChan:
		t.Error("Done channel should not be closed initially")
	default:
		// Expected
	}
}

func TestGracefulShutdown_Cleanup(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	server := &http.Server{Addr: ":8080"}
	gs := NewGracefulShutdown(server, logger, 30*time.Second)

	cleanupCalled := 0
	cleanupFuncs := []func() error{
		func() error {
			cleanupCalled++
			return nil
		},
		func() error {
			cleanupCalled++
			return nil
		},
	}

	gs.Cleanup(cleanupFuncs...)
	assert.Equal(t, 2, cleanupCalled)
}

func TestGracefulShutdown_CleanupWithError(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	server := &http.Server{Addr: ":8080"}
	gs := NewGracefulShutdown(server, logger, 30*time.Second)

	cleanupCalled := 0
	cleanupFuncs := []func() error{
		func() error {
			cleanupCalled++
			return nil
		},
		func() error {
			cleanupCalled++
			return assert.AnError
		},
		func() error {
			cleanupCalled++
			return nil
		},
	}

	// Should continue executing cleanup funcs even if one fails
	gs.Cleanup(cleanupFuncs...)
	assert.Equal(t, 3, cleanupCalled)
}

func TestGracefulShutdown_WaitForShutdown(t *testing.T) {
	t.Run("timeout before shutdown", func(t *testing.T) {
		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		server := &http.Server{Addr: ":8080"}
		gs := NewGracefulShutdown(server, logger, 30*time.Second)

		// Wait for 100ms, should timeout since shutdown hasn't been triggered
		result := gs.WaitForShutdown(100 * time.Millisecond)
		assert.False(t, result)
	})
}

func TestGracefulShutdown_Listen(t *testing.T) {
	// This test would require sending actual signals which is complex
	// In production, the Listen method is tested via integration tests
	t.Skip("Integration test - requires signal handling")
}
