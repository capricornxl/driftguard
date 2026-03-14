package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{
				Port:         8080,
				Host:         "0.0.0.0",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
			},
			Database: DatabaseConfig{
				Driver:         "sqlite",
				Database:       "driftguard.db",
				MaxOpenConns:   25,
				MaxIdleConns:   5,
				ConnMaxLifetime: 5 * time.Minute,
				ConnMaxIdleTime: 2 * time.Minute,
			},
			Collector: CollectorConfig{
				BatchSize:     100,
				FlushInterval: 60 * time.Second,
			},
			Detector: DetectorConfig{
				CheckInterval:  300 * time.Second,
				WindowDays:     7,
				Threshold:      70.0,
				KSThreshold:    0.05,
				PSIThreshold:   0.2,
				SpikeThreshold: 2.5,
			},
			Alerter: AlerterConfig{
				Enabled: false,
			},
			Evaluator: EvaluatorConfig{
				Weights: HealthWeights{
					Latency:     0.15,
					Efficiency:  0.10,
					Consistency: 0.30,
					Accuracy:    0.35,
					Hallucination: 0.10,
				},
				CacheSize: 1000,
			},
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid server config", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{
				Port: 70000, // Invalid port
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server config")
	})

	t.Run("invalid database config", func(t *testing.T) {
		cfg := &Config{
			Server: ServerConfig{
				Port: 8080,
				Host: "0.0.0.0",
			},
			Database: DatabaseConfig{
				Driver: "invalid",
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database config")
	})
}

func TestServerConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := ServerConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("invalid port - too low", func(t *testing.T) {
		cfg := ServerConfig{Port: 0}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "port must be between 1 and 65535")
	})

	t.Run("invalid port - too high", func(t *testing.T) {
		cfg := ServerConfig{Port: 70000}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "port must be between 1 and 65535")
	})

	t.Run("empty host", func(t *testing.T) {
		cfg := ServerConfig{Port: 8080, Host: ""}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "host cannot be empty")
	})

	t.Run("negative read timeout", func(t *testing.T) {
		cfg := ServerConfig{Port: 8080, Host: "0.0.0.0", ReadTimeout: -1 * time.Second}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read_timeout must be positive")
	})
}

func TestDatabaseConfig_Validate(t *testing.T) {
	t.Run("valid sqlite config", func(t *testing.T) {
		cfg := DatabaseConfig{
			Driver:         "sqlite",
			Database:       "driftguard.db",
			MaxOpenConns:   25,
			MaxIdleConns:   5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 2 * time.Minute,
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("valid postgres config", func(t *testing.T) {
		cfg := DatabaseConfig{
			Driver:         "postgres",
			Host:           "localhost",
			Port:           "5432",
			Username:       "driftguard",
			Password:       "secret",
			Database:       "driftguard",
			MaxOpenConns:   25,
			MaxIdleConns:   5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 2 * time.Minute,
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("empty driver", func(t *testing.T) {
		cfg := DatabaseConfig{Driver: ""}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "driver cannot be empty")
	})

	t.Run("unsupported driver", func(t *testing.T) {
		cfg := DatabaseConfig{Driver: "mysql"}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported driver")
	})

	t.Run("postgres missing host", func(t *testing.T) {
		cfg := DatabaseConfig{
			Driver:   "postgres",
			Database: "driftguard",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "host cannot be empty for postgres")
	})

	t.Run("idle conns exceeds open conns", func(t *testing.T) {
		cfg := DatabaseConfig{
			Driver:         "sqlite",
			Database:       "test.db",
			MaxOpenConns:   5,
			MaxIdleConns:   10,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 2 * time.Minute,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max_idle_conns")
	})
}

func TestCollectorConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := CollectorConfig{
			BatchSize:     100,
			FlushInterval: 60 * time.Second,
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("batch size too large", func(t *testing.T) {
		cfg := CollectorConfig{BatchSize: 2000}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "batch_size should not exceed 1000")
	})

	t.Run("negative batch size", func(t *testing.T) {
		cfg := CollectorConfig{BatchSize: -1}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "batch_size must be positive")
	})

	t.Run("flush interval too long", func(t *testing.T) {
		cfg := CollectorConfig{
			BatchSize:     100,
			FlushInterval: 15 * time.Minute,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "flush_interval should not exceed 10 minutes")
	})
}

func TestDetectorConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := DetectorConfig{
			CheckInterval:  300 * time.Second,
			WindowDays:     7,
			Threshold:      70.0,
			KSThreshold:    0.05,
			PSIThreshold:   0.2,
			SpikeThreshold: 2.5,
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("window days too large", func(t *testing.T) {
		cfg := DetectorConfig{WindowDays: 400}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "window_days should not exceed 365")
	})

	t.Run("threshold out of range", func(t *testing.T) {
		cfg := DetectorConfig{Threshold: 150.0}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "threshold must be between 0 and 100")
	})

	t.Run("ks threshold out of range", func(t *testing.T) {
		cfg := DetectorConfig{KSThreshold: 1.5}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ks_threshold must be between 0 and 1")
	})

	t.Run("psi threshold too low", func(t *testing.T) {
		cfg := DetectorConfig{PSIThreshold: 0.05}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "psi_threshold must be between 0.1 and 1")
	})
}

func TestAlerterConfig_Validate(t *testing.T) {
	t.Run("disabled alerter", func(t *testing.T) {
		cfg := AlerterConfig{Enabled: false}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("enabled without channels", func(t *testing.T) {
		cfg := AlerterConfig{Enabled: true, Channels: []AlertChannel{}}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one alert channel")
	})

	t.Run("valid slack channel", func(t *testing.T) {
		cfg := AlerterConfig{
			Enabled: true,
			Channels: []AlertChannel{
				{
					Type:      "slack",
					Enabled:   true,
					WebhookURL: "https://hooks.slack.com/services/xxx",
				},
			},
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("webhook without https", func(t *testing.T) {
		cfg := AlerterConfig{
			Enabled: true,
			Channels: []AlertChannel{
				{
					Type:       "webhook",
					Enabled:    true,
					WebhookURL: "http://example.com/webhook",
				},
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must use HTTPS")
	})

	t.Run("invalid channel type", func(t *testing.T) {
		cfg := AlerterConfig{
			Enabled: true,
			Channels: []AlertChannel{
				{Type: "invalid", Enabled: true},
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid channel type")
	})
}

func TestHealthWeights_Validate(t *testing.T) {
	t.Run("valid weights", func(t *testing.T) {
		cfg := HealthWeights{
			Latency:     0.15,
			Efficiency:  0.10,
			Consistency: 0.30,
			Accuracy:    0.35,
			Hallucination: 0.10,
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("weights do not sum to 1", func(t *testing.T) {
		cfg := HealthWeights{
			Latency:     0.20,
			Efficiency:  0.20,
			Consistency: 0.20,
			Accuracy:    0.20,
			Hallucination: 0.20,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must sum to 1.0")
	})

	t.Run("negative weight", func(t *testing.T) {
		cfg := HealthWeights{
			Latency:     -0.10,
			Efficiency:  0.10,
			Consistency: 0.30,
			Accuracy:    0.35,
			Hallucination: 0.35,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be between 0 and 1")
	})
}

func TestEvaluatorConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := EvaluatorConfig{
			Weights: HealthWeights{
				Latency:     0.15,
				Efficiency:  0.10,
				Consistency: 0.30,
				Accuracy:    0.35,
				Hallucination: 0.10,
			},
			CacheSize: 1000,
		}
		assert.NoError(t, cfg.Validate())
	})

	t.Run("invalid weights", func(t *testing.T) {
		cfg := EvaluatorConfig{
			Weights: HealthWeights{
				Latency: 0.50,
			},
			CacheSize: 1000,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "weights")
	})

	t.Run("invalid cache size", func(t *testing.T) {
		cfg := EvaluatorConfig{
			Weights: HealthWeights{
				Latency:     0.15,
				Efficiency:  0.10,
				Consistency: 0.30,
				Accuracy:    0.35,
				Hallucination: 0.10,
			},
			CacheSize: 0,
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache_size must be positive")
	})
}
