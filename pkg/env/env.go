package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Env holds all environment configuration for DriftGuard
type Env struct {
	// Server
	Port              string
	Host              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	ShutdownTimeout   time.Duration

	// Database
	Driver          string
	DatabaseURL     string
	DBHost          string
	DBPort          string
	DBName          string
	DBUsername      string
	DBPassword      string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Collector
	BatchSize     int
	FlushInterval time.Duration
	AgentEndpoint string

	// Detector
	CheckInterval  time.Duration
	WindowDays     int
	Threshold      float64
	KSThreshold    float64
	PSIThreshold   float64
	SpikeThreshold float64

	// Alerter
	AlerterEnabled bool
	SlackWebhook   string
	DiscordWebhook string
	WebhookURL     string

	// Evaluator
	CacheSize int

	// Health Weights
	LatencyWeight     float64
	EfficiencyWeight  float64
	ConsistencyWeight float64
	AccuracyWeight    float64
	HallucinationWeight float64

	// Logging
	LogLevel string
	LogFormat string

	// External APIs
	TavilyAPIKey     string
	ConfluenceCookie string

	// Environment
	Environment string
	Debug       bool
}

// Load loads environment variables with defaults
func Load() *Env {
	return &Env{
		// Server
		Port:            getEnv("PORT", "8080"),
		Host:            getEnv("HOST", "0.0.0.0"),
		ReadTimeout:     getDurationEnv("READ_TIMEOUT", 30*time.Second),
		WriteTimeout:    getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
		ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 30*time.Second),

		// Database
		Driver:          getEnv("DB_DRIVER", "sqlite"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBName:          getEnv("DB_NAME", "driftguard"),
		DBUsername:      getEnv("DB_USERNAME", "driftguard"),
		DBPassword:      getEnv("DB_PASSWORD", "driftguard123"),
		MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		ConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", 2*time.Minute),

		// Collector
		BatchSize:     getIntEnv("COLLECTOR_BATCH_SIZE", 100),
		FlushInterval: getDurationEnv("COLLECTOR_FLUSH_INTERVAL", 60*time.Second),
		AgentEndpoint: getEnv("AGENT_ENDPOINT", ""),

		// Detector
		CheckInterval:  getDurationEnv("DETECTOR_CHECK_INTERVAL", 300*time.Second),
		WindowDays:     getIntEnv("DETECTOR_WINDOW_DAYS", 7),
		Threshold:      getFloatEnv("DETECTOR_THRESHOLD", 70.0),
		KSThreshold:    getFloatEnv("DETECTOR_KS_THRESHOLD", 0.05),
		PSIThreshold:   getFloatEnv("DETECTOR_PSI_THRESHOLD", 0.2),
		SpikeThreshold: getFloatEnv("DETECTOR_SPIKE_THRESHOLD", 2.5),

		// Alerter
		AlerterEnabled: getBoolEnv("ALERTER_ENABLED", false),
		SlackWebhook:   getEnv("SLACK_WEBHOOK_URL", ""),
		DiscordWebhook: getEnv("DISCORD_WEBHOOK_URL", ""),
		WebhookURL:     getEnv("WEBHOOK_URL", ""),

		// Evaluator
		CacheSize: getIntEnv("EVALUATOR_CACHE_SIZE", 1000),

		// Health Weights
		LatencyWeight:     getFloatEnv("WEIGHT_LATENCY", 0.15),
		EfficiencyWeight:  getFloatEnv("WEIGHT_EFFICIENCY", 0.10),
		ConsistencyWeight: getFloatEnv("WEIGHT_CONSISTENCY", 0.30),
		AccuracyWeight:    getFloatEnv("WEIGHT_ACCURACY", 0.35),
		HallucinationWeight: getFloatEnv("WEIGHT_HALLUCINATION", 0.10),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		// External APIs
		TavilyAPIKey:     getEnv("TAVILY_API_KEY", ""),
		ConfluenceCookie: getEnv("CONFLUENCE_COOKIE", ""),

		// Environment
		Environment: getEnv("ENVIRONMENT", "development"),
		Debug:       getBoolEnv("DEBUG", false),
	}
}

// Validate validates environment configuration
func (e *Env) Validate() error {
	// Validate server
	if e.Port == "" {
		return fmt.Errorf("PORT cannot be empty")
	}

	// Validate database
	if e.Driver != "sqlite" && e.Driver != "postgres" {
		return fmt.Errorf("DB_DRIVER must be sqlite or postgres, got: %s", e.Driver)
	}

	if e.Driver == "postgres" && e.DBPassword == "" {
		return fmt.Errorf("DB_PASSWORD is required for postgres")
	}

	// Validate health weights sum to 1.0
	total := e.LatencyWeight + e.EfficiencyWeight + e.ConsistencyWeight + e.AccuracyWeight + e.HallucinationWeight
	if total < 0.99 || total > 1.01 {
		return fmt.Errorf("health weights must sum to 1.0, got %.2f", total)
	}

	// Validate thresholds
	if e.Threshold < 0 || e.Threshold > 100 {
		return fmt.Errorf("DETECTOR_THRESHOLD must be between 0 and 100, got %.2f", e.Threshold)
	}

	if e.KSThreshold <= 0 || e.KSThreshold > 1 {
		return fmt.Errorf("DETECTOR_KS_THRESHOLD must be between 0 and 1, got %.4f", e.KSThreshold)
	}

	if e.PSIThreshold < 0.1 || e.PSIThreshold > 1 {
		return fmt.Errorf("DETECTOR_PSI_THRESHOLD must be between 0.1 and 1, got %.4f", e.PSIThreshold)
	}

	// Validate webhook URLs use HTTPS
	if e.SlackWebhook != "" && !startsWithHTTPS(e.SlackWebhook) {
		return fmt.Errorf("SLACK_WEBHOOK_URL must use HTTPS")
	}

	if e.DiscordWebhook != "" && !startsWithHTTPS(e.DiscordWebhook) {
		return fmt.Errorf("DISCORD_WEBHOOK_URL must use HTTPS")
	}

	if e.WebhookURL != "" && !startsWithHTTPS(e.WebhookURL) {
		return fmt.Errorf("WEBHOOK_URL must use HTTPS")
	}

	return nil
}

// getEnv gets environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getBoolEnv gets boolean environment variable with default
func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return result
}

// getIntEnv gets integer environment variable with default
func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return result
}

// getFloatEnv gets float environment variable with default
func getFloatEnv(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return result
}

// getDurationEnv gets duration environment variable with default
// Supports formats: "30s", "5m", "1h", "300" (seconds)
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	// Try parsing as duration first
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}

	// Try parsing as seconds (integer)
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}

	return defaultValue
}

// startsWithHTTPS checks if URL starts with https://
func startsWithHTTPS(url string) bool {
	return len(url) >= 8 && url[:8] == "https://"
}
