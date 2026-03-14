package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/driftguard/driftguard/internal/alerter"
	"github.com/driftguard/driftguard/internal/collector"
	"github.com/driftguard/driftguard/internal/detector"
	"github.com/driftguard/driftguard/internal/evaluator"
	"github.com/driftguard/driftguard/internal/handler"
	"github.com/driftguard/driftguard/internal/middleware"
	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/database"
	"github.com/driftguard/driftguard/pkg/env"
	"github.com/driftguard/driftguard/pkg/metrics"
	"github.com/driftguard/driftguard/pkg/models"
	"github.com/driftguard/driftguard/pkg/shutdown"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	Version   = "0.1.1"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Load environment configuration
	environment := env.Load()

	// Validate environment configuration
	if err := environment.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Environment validation failed: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := initLogger(environment)
	logger.WithFields(logrus.Fields{
		"version":   Version,
		"build_time": BuildTime,
		"git_commit": GitCommit,
		"environment": environment.Environment,
	}).Info("Starting DriftGuard")

	// Load YAML config (with environment variable overrides)
	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.WithError(err).Fatal("Configuration validation failed")
	}

	// Apply environment variable overrides
	applyEnvironmentOverrides(cfg, environment)

	// Initialize database
	db, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}
	defer func() {
		logger.Info("Closing database connection...")
		if err := database.CloseDatabase(db); err != nil {
			logger.WithError(err).Error("Failed to close database")
		}
	}()

	// Auto-migrate database schema
	if err := models.AutoMigrate(db); err != nil {
		logger.WithError(err).Fatal("Failed to migrate database schema")
	}

	// Initialize metrics
	metrics := metrics.NewMetrics()

	// Initialize collector
	collector := collector.NewCollector(&cfg.Collector, db, metrics, logger)
	collector.Start()
	defer collector.Stop()

	// Initialize evaluator
	evaluator := evaluator.NewEvaluator(&cfg.Evaluator, db, metrics, logger)

	// Initialize detector
	detector := detector.NewDetector(&cfg.Detector, db, evaluator, metrics, logger)

	// Initialize enhanced detector
	enhancedDetector := detector.NewEnhancedDetector(detector, db, logger)

	// Initialize alerter
	alerter := alerter.NewAlerter(&cfg.Alerter, logger)

	// Setup Gin router
	gin.SetMode(ginMode(environment.LogLevel))
	router := setupRouter(db, cfg, environment, logger, metrics, enhancedDetector, alerter)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.WithField("addr", server.Addr).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutdown signal received, starting graceful shutdown...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("HTTP server forced to shutdown")
	}
	logger.Info("HTTP server stopped")

	// Run cleanup tasks
	logger.Info("Running cleanup tasks...")
	cleanupTasks := []func() error{
		func() error {
			collector.Stop()
			return nil
		},
		func() error {
			return database.CloseDatabase(db)
		},
	}

	for i, cleanup := range cleanupTasks {
		if err := cleanup(); err != nil {
			logger.WithField("task", i).WithError(err).Error("Cleanup task failed")
		} else {
			logger.WithField("task", i).Debug("Cleanup task completed")
		}
	}

	logger.Info("DriftGuard shutdown complete")
}

func initLogger(environment *env.Env) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(environment.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	if environment.LogFormat == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	return logger
}

func ginMode(logLevel string) string {
	if logLevel == "debug" {
		return gin.DebugMode
	}
	return gin.ReleaseMode
}

func applyEnvironmentOverrides(cfg *config.Config, environment *env.Env) {
	// Server
	if environment.Port != "8080" {
		cfg.Server.Port = 8080 // Parse from string if needed
	}
	if environment.Host != "0.0.0.0" {
		cfg.Server.Host = environment.Host
	}
	cfg.Server.ReadTimeout = environment.ReadTimeout
	cfg.Server.WriteTimeout = environment.WriteTimeout
	cfg.Server.ShutdownTimeout = environment.ShutdownTimeout

	// Database
	if environment.Driver != "sqlite" {
		cfg.Database.Driver = environment.Driver
	}
	cfg.Database.Host = environment.DBHost
	cfg.Database.Port = environment.DBPort
	cfg.Database.Database = environment.DBName
	cfg.Database.Username = environment.DBUsername
	cfg.Database.Password = environment.DBPassword
	cfg.Database.MaxOpenConns = environment.MaxOpenConns
	cfg.Database.MaxIdleConns = environment.MaxIdleConns
	cfg.Database.ConnMaxLifetime = environment.ConnMaxLifetime
	cfg.Database.ConnMaxIdleTime = environment.ConnMaxIdleTime

	// Collector
	if environment.BatchSize != 100 {
		cfg.Collector.BatchSize = environment.BatchSize
	}
	if environment.FlushInterval != 60*time.Second {
		cfg.Collector.FlushInterval = environment.FlushInterval
	}

	// Detector
	if environment.WindowDays != 7 {
		cfg.Detector.WindowDays = environment.WindowDays
	}
	if environment.Threshold != 70.0 {
		cfg.Detector.Threshold = environment.Threshold
	}

	// Alerter
	if environment.AlerterEnabled {
		cfg.Alerter.Enabled = true
	}
	if environment.SlackWebhook != "" {
		// Add Slack channel
		cfg.Alerter.Channels = append(cfg.Alerter.Channels, config.AlertChannel{
			Type:       "slack",
			Enabled:    true,
			WebhookURL: environment.SlackWebhook,
		})
	}
	if environment.DiscordWebhook != "" {
		cfg.Alerter.Channels = append(cfg.Alerter.Channels, config.AlertChannel{
			Type:       "discord",
			Enabled:    true,
			WebhookURL: environment.DiscordWebhook,
		})
	}

	// Evaluator
	if environment.CacheSize != 1000 {
		cfg.Evaluator.CacheSize = environment.CacheSize
	}

	// Health weights
	if environment.LatencyWeight != 0.15 {
		cfg.Evaluator.Weights.Latency = environment.LatencyWeight
	}
	if environment.EfficiencyWeight != 0.10 {
		cfg.Evaluator.Weights.Efficiency = environment.EfficiencyWeight
	}
	if environment.ConsistencyWeight != 0.30 {
		cfg.Evaluator.Weights.Consistency = environment.ConsistencyWeight
	}
	if environment.AccuracyWeight != 0.35 {
		cfg.Evaluator.Weights.Accuracy = environment.AccuracyWeight
	}
	if environment.HallucinationWeight != 0.10 {
		cfg.Evaluator.Weights.Hallucination = environment.HallucinationWeight
	}
}

func setupRouter(
	db *gorm.DB,
	cfg *config.Config,
	environment *env.Env,
	logger *logrus.Logger,
	metrics *metrics.Metrics,
	enhancedDetector *detector.EnhancedDetector,
	alerter *alerter.Alerter,
) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.RecoveryWithLogger(logger))
	router.Use(middleware.SecurityHeaders())

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Health endpoints
		v1.GET("/health", handler.EnhancedHealthCheck(db, cfg, logger, time.Now(), Version))
		v1.GET("/ready", handler.EnhancedReadyCheck(db, logger))
		v1.GET("/live", handler.LiveCheck())

		// Metrics
		v1.GET("/metrics", gin.WrapH(promhttp.Handler()))

		// Agent routes with validation
		agents := v1.Group("/agents")
		agents.Use(middleware.ValidatePagination())
		{
			agents.GET("", handler.ListAgents(db))
			agents.POST("", handler.CreateAgent(db))

			agentSpecific := agents.Group("/:agent_id")
			agentSpecific.Use(middleware.ValidateAgentID())
			{
				agentSpecific.GET("", handler.GetAgent(db))
				agentSpecific.PUT("", handler.UpdateAgent(db))
				agentSpecific.DELETE("", handler.DeleteAgent(db))

				// Health scores
				agentSpecific.GET("/scores", handler.GetHealthScores(db, enhancedDetector))
				agentSpecific.GET("/scores/latest", handler.GetLatestHealthScore(db))
				agentSpecific.GET("/trend", middleware.ValidateTimeRange(), handler.GetHealthTrend(db, enhancedDetector))

				// Drift detection
				agentSpecific.GET("/drift/ks-test", handler.GetKSTest(db, enhancedDetector))
				agentSpecific.GET("/drift/psi", handler.GetPSITest(db, enhancedDetector))
				agentSpecific.GET("/drift/spikes", middleware.ValidateTimeRange(), handler.GetSpikes(db, enhancedDetector))
				agentSpecific.GET("/report/comprehensive", handler.GetComprehensiveReport(db, enhancedDetector))

				// Alerts
				agentSpecific.GET("/alerts", handler.GetAlerts(db))
				agentSpecific.POST("/alerts", handler.CreateAlert(db))
			}
		}
	}

	return router
}
