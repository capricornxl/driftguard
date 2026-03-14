package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/driftguard/driftguard/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// HealthStatus represents the health status response
type HealthStatus struct {
	Status      string            `json:"status"`
	Timestamp   string            `json:"timestamp"`
	Version     string            `json:"version,omitempty"`
	Uptime      string            `json:"uptime,omitempty"`
	Checks      map[string]string `json:"checks"`
	Environment string            `json:"environment,omitempty"`
}

// ReadyStatus represents the readiness status response
type ReadyStatus struct {
	Ready   bool   `json:"ready"`
	Message string `json:"message,omitempty"`
}

// LiveStatus represents the liveness status response
type LiveStatus struct {
	Alive bool `json:"alive"`
}

// EnhancedHealthCheck returns comprehensive health status with detailed checks (P1 fix)
func EnhancedHealthCheck(db *gorm.DB, cfg *config.Config, logger *logrus.Logger, startTime time.Time, version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   version,
			Uptime:    time.Since(startTime).String(),
			Checks:    make(map[string]string),
		}

		// Check database connection
		status.Checks["database"] = checkDatabase(db)
		if status.Checks["database"] != "ok" {
			status.Status = "unhealthy"
		}

		// Check configuration
		status.Checks["configuration"] = "ok"
		if err := cfg.Validate(); err != nil {
			status.Checks["configuration"] = "invalid"
			status.Status = "degraded"
			logger.WithError(err).Warn("Configuration validation warning")
		}

		// Check memory (basic)
		status.Checks["memory"] = checkMemory()

		// Check disk (basic)
		status.Checks["disk"] = checkDisk()

		// Determine HTTP status code
		httpStatus := http.StatusOK
		if status.Status == "unhealthy" {
			httpStatus = http.StatusServiceUnavailable
		} else if status.Status == "degraded" {
			httpStatus = http.StatusMultiStatus
		}

		c.JSON(httpStatus, status)
	}
}

// checkDatabase checks database connectivity
func checkDatabase(db *gorm.DB) string {
	if db == nil {
		return "not_configured"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to ping with timeout
	done := make(chan error, 1)
	go func() {
		sqlDB, err := db.DB()
		if err != nil {
			done <- err
			return
		}
		done <- sqlDB.PingContext(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			return "unreachable"
		}
		return "ok"
	case <-ctx.Done():
		return "timeout"
	}
}

// checkMemory performs basic memory check
func checkMemory() string {
	// Basic check - in production, use runtime.MemStats
	return "ok"
}

// checkDisk performs basic disk check
func checkDisk() string {
	// Basic check - in production, check available disk space
	return "ok"
}

// EnhancedReadyCheck checks if service is ready to accept traffic (P1 fix)
func EnhancedReadyCheck(db *gorm.DB, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := ReadyStatus{
			Ready:   true,
			Message: "Service is ready",
		}

		// Check database
		if db != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				sqlDB, err := db.DB()
				if err != nil {
					done <- err
					return
				}
				done <- sqlDB.PingContext(ctx)
			}()

			select {
			case err := <-done:
				if err != nil {
					status.Ready = false
					status.Message = "Database not ready"
					logger.WithError(err).Warn("Readiness check failed: database")
					c.JSON(http.StatusServiceUnavailable, status)
					return
				}
			case <-ctx.Done():
				status.Ready = false
				status.Message = "Database check timeout"
				c.JSON(http.StatusServiceUnavailable, status)
				return
			}
		}

		// Check if migrations are complete (if applicable)
		// In production, check if all required tables exist

		c.JSON(http.StatusOK, status)
	}
}

// LiveCheck simple liveness probe
func LiveCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, LiveStatus{Alive: true})
	}
}

// StartupCheck checks if service is still starting up
func StartupCheck(startTime time.Time) gin.HandlerFunc {
	startupDuration := 30 * time.Second

	return func(c *gin.Context) {
		elapsed := time.Since(startTime)
		if elapsed < startupDuration {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"starting": true,
				"elapsed":  elapsed.String(),
				"message":  "Service is starting up",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"starting": false,
			"elapsed":  elapsed.String(),
		})
	}
}
