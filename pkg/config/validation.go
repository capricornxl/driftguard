package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Validate validates the entire configuration
func (c *Config) Validate() error {
	var errors []string

	if err := c.Server.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("server config: %v", err))
	}

	if err := c.Database.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("database config: %v", err))
	}

	if err := c.Collector.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("collector config: %v", err))
	}

	if err := c.Detector.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("detector config: %v", err))
	}

	if err := c.Alerter.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("alerter config: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// Validate validates server configuration
func (c *ServerConfig) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}

	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be positive, got %v", c.ReadTimeout)
	}

	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be positive, got %v", c.WriteTimeout)
	}

	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown_timeout must be positive, got %v", c.ShutdownTimeout)
	}

	return nil
}

// Validate validates database configuration
func (c *DatabaseConfig) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("driver cannot be empty")
	}

	if c.Driver != "sqlite" && c.Driver != "postgres" {
		return fmt.Errorf("unsupported driver: %s (supported: sqlite, postgres)", c.Driver)
	}

	if c.Database == "" {
		return fmt.Errorf("database name/path cannot be empty")
	}

	if c.Driver == "postgres" {
		if c.Host == "" {
			return fmt.Errorf("host cannot be empty for postgres")
		}
		if c.Username == "" {
			return fmt.Errorf("username cannot be empty for postgres")
		}
		if c.Password == "" {
			return fmt.Errorf("password cannot be empty for postgres")
		}
		if c.Port == "" {
			return fmt.Errorf("port cannot be empty for postgres")
		}
	}

	if c.MaxOpenConns <= 0 {
		return fmt.Errorf("max_open_conns must be positive, got %d", c.MaxOpenConns)
	}

	if c.MaxIdleConns <= 0 {
		return fmt.Errorf("max_idle_conns must be positive, got %d", c.MaxIdleConns)
	}

	if c.MaxIdleConns > c.MaxOpenConns {
		return fmt.Errorf("max_idle_conns (%d) cannot exceed max_open_conns (%d)", c.MaxIdleConns, c.MaxOpenConns)
	}

	if c.ConnMaxLifetime <= 0 {
		return fmt.Errorf("conn_max_lifetime must be positive, got %v", c.ConnMaxLifetime)
	}

	if c.ConnMaxIdleTime <= 0 {
		return fmt.Errorf("conn_max_idle_time must be positive, got %v", c.ConnMaxIdleTime)
	}

	return nil
}

// Validate validates collector configuration
func (c *CollectorConfig) Validate() error {
	if c.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be positive, got %d", c.BatchSize)
	}

	if c.BatchSize > 1000 {
		return fmt.Errorf("batch_size should not exceed 1000, got %d", c.BatchSize)
	}

	if c.FlushInterval <= 0 {
		return fmt.Errorf("flush_interval must be positive, got %v", c.FlushInterval)
	}

	if c.FlushInterval > 10*time.Minute {
		return fmt.Errorf("flush_interval should not exceed 10 minutes, got %v", c.FlushInterval)
	}

	if c.AgentEndpoint != "" {
		if _, err := url.ParseRequestURI(c.AgentEndpoint); err != nil {
			return fmt.Errorf("invalid agent_endpoint URL: %v", err)
		}
	}

	return nil
}

// Validate validates detector configuration
func (c *DetectorConfig) Validate() error {
	if c.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be positive, got %v", c.CheckInterval)
	}

	if c.WindowDays <= 0 {
		return fmt.Errorf("window_days must be positive, got %d", c.WindowDays)
	}

	if c.WindowDays > 365 {
		return fmt.Errorf("window_days should not exceed 365, got %d", c.WindowDays)
	}

	if c.Threshold < 0 || c.Threshold > 100 {
		return fmt.Errorf("threshold must be between 0 and 100, got %.2f", c.Threshold)
	}

	if c.KSThreshold <= 0 || c.KSThreshold > 1 {
		return fmt.Errorf("ks_threshold must be between 0 and 1, got %.4f", c.KSThreshold)
	}

	if c.PSIThreshold < 0.1 || c.PSIThreshold > 1 {
		return fmt.Errorf("psi_threshold must be between 0.1 and 1, got %.4f", c.PSIThreshold)
	}

	if c.SpikeThreshold <= 0 {
		return fmt.Errorf("spike_threshold must be positive, got %.2f", c.SpikeThreshold)
	}

	return nil
}

// Validate validates alerter configuration
func (c *AlerterConfig) Validate() error {
	if !c.Enabled {
		return nil // Skip validation if disabled
	}

	if len(c.Channels) == 0 {
		return fmt.Errorf("at least one alert channel must be configured when alerter is enabled")
	}

	for i, ch := range c.Channels {
		if err := ch.Validate(); err != nil {
			return fmt.Errorf("channel[%d]: %v", i, err)
		}
	}

	return nil
}

// Validate validates alert channel configuration
func (c *AlertChannel) Validate() error {
	if !c.Enabled {
		return nil // Skip validation if disabled
	}

	if c.Type == "" {
		return fmt.Errorf("channel type cannot be empty")
	}

	validTypes := []string{"slack", "discord", "webhook", "log"}
	isValid := false
	for _, t := range validTypes {
		if c.Type == t {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid channel type: %s (supported: %s)", c.Type, strings.Join(validTypes, ", "))
	}

	// Validate webhook URL for slack, discord, and webhook types
	if c.Type == "slack" || c.Type == "discord" || c.Type == "webhook" {
		if c.WebhookURL == "" {
			return fmt.Errorf("webhook_url is required for %s channel", c.Type)
		}

		// Validate URL format
		parsedURL, err := url.ParseRequestURI(c.WebhookURL)
		if err != nil {
			return fmt.Errorf("invalid webhook_url format: %v", err)
		}

		// Enforce HTTPS for webhook URLs
		if parsedURL.Scheme != "https" {
			return fmt.Errorf("webhook_url must use HTTPS for security, got: %s", parsedURL.Scheme)
		}
	}

	// Validate severity filter
	if c.SeverityFilter != "" {
		validSeverities := []string{"info", "warning", "critical", "all"}
		isValid := false
		for _, s := range validSeverities {
			if c.SeverityFilter == s {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid severity_filter: %s (supported: %s)", c.SeverityFilter, strings.Join(validSeverities, ", "))
		}
	}

	return nil
}

// Validate validates health weights configuration
func (c *HealthWeights) Validate() error {
	total := c.Latency + c.Efficiency + c.Consistency + c.Accuracy + c.Hallucination

	if total < 0.99 || total > 1.01 {
		return fmt.Errorf("health weights must sum to 1.0, got %.2f (latency=%.2f, efficiency=%.2f, consistency=%.2f, accuracy=%.2f, hallucination=%.2f)",
			total, c.Latency, c.Efficiency, c.Consistency, c.Accuracy, c.Hallucination)
	}

	// Validate individual weights are between 0 and 1
	weights := map[string]float64{
		"latency":     c.Latency,
		"efficiency":  c.Efficiency,
		"consistency": c.Consistency,
		"accuracy":    c.Accuracy,
		"hallucination": c.Hallucination,
	}

	for name, weight := range weights {
		if weight < 0 || weight > 1 {
			return fmt.Errorf("%s weight must be between 0 and 1, got %.2f", name, weight)
		}
	}

	return nil
}

// Validate validates evaluator configuration
func (c *EvaluatorConfig) Validate() error {
	if err := c.Weights.Validate(); err != nil {
		return fmt.Errorf("weights: %v", err)
	}

	if c.CacheSize <= 0 {
		return fmt.Errorf("cache_size must be positive, got %d", c.CacheSize)
	}

	return nil
}
