package shutdown

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// GracefulShutdown handles graceful shutdown of the application
type GracefulShutdown struct {
	server         *http.Server
	logger         *logrus.Logger
	shutdownTimeout time.Duration
	done           chan struct{}
}

// NewGracefulShutdown creates a new graceful shutdown handler
func NewGracefulShutdown(server *http.Server, logger *logrus.Logger, shutdownTimeout time.Duration) *GracefulShutdown {
	return &GracefulShutdown{
		server:          server,
		logger:          logger,
		shutdownTimeout: shutdownTimeout,
		done:            make(chan struct{}),
	}
}

// Listen listens for shutdown signals and handles graceful shutdown
func (gs *GracefulShutdown) Listen() {
	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-quit
	gs.logger.WithField("signal", sig).Info("Shutdown signal received")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), gs.shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	gs.logger.Info("Starting graceful shutdown...")

	if err := gs.server.Shutdown(ctx); err != nil {
		gs.logger.WithError(err).Error("Server forced to shutdown")
	}

	gs.logger.Info("HTTP server stopped")
	close(gs.done)
}

// Done returns a channel that's closed when shutdown is complete
func (gs *GracefulShutdown) Done() <-chan struct{} {
	return gs.done
}

// Shutdown manually triggers shutdown (for testing or programmatic shutdown)
func (gs *GracefulShutdown) Shutdown() {
	gs.logger.Info("Manual shutdown triggered")
	quit := make(chan os.Signal, 1)
	quit <- syscall.SIGTERM
}

// WaitForShutdown waits for shutdown to complete with timeout
func (gs *GracefulShutdown) WaitForShutdown(timeout time.Duration) bool {
	select {
	case <-gs.done:
		return true
	case <-time.After(timeout):
		gs.logger.Error("Shutdown timeout exceeded")
		return false
	}
}

// Cleanup performs cleanup tasks after shutdown
func (gs *GracefulShutdown) Cleanup(cleanupFuncs ...func() error) {
	gs.logger.Info("Running cleanup tasks...")
	for i, cleanup := range cleanupFuncs {
		if err := cleanup(); err != nil {
			gs.logger.WithField("task", i).WithError(err).Error("Cleanup task failed")
		} else {
			gs.logger.WithField("task", i).Debug("Cleanup task completed")
		}
	}
	gs.logger.Info("All cleanup tasks completed")
}
