package env

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Clear environment variables first
	os.Clearenv()

	t.Run("default values", func(t *testing.T) {
		env := Load()

		assert.Equal(t, "8080", env.Port)
		assert.Equal(t, "0.0.0.0", env.Host)
		assert.Equal(t, 30*time.Second, env.ReadTimeout)
		assert.Equal(t, "sqlite", env.Driver)
		assert.Equal(t, 25, env.MaxOpenConns)
		assert.Equal(t, 5, env.MaxIdleConns)
		assert.Equal(t, 100, env.BatchSize)
		assert.Equal(t, 7, env.WindowDays)
		assert.Equal(t, 70.0, env.Threshold)
		assert.Equal(t, false, env.AlerterEnabled)
		assert.Equal(t, 1000, env.CacheSize)
		assert.Equal(t, "info", env.LogLevel)
		assert.Equal(t, "development", env.Environment)
		assert.Equal(t, false, env.Debug)
	})

	t.Run("custom values", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("PORT", "9090")
		os.Setenv("DB_DRIVER", "postgres")
		os.Setenv("DB_PASSWORD", "secret123")
		os.Setenv("COLLECTOR_BATCH_SIZE", "500")
		os.Setenv("ALERTER_ENABLED", "true")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("DEBUG", "true")

		env := Load()

		assert.Equal(t, "9090", env.Port)
		assert.Equal(t, "postgres", env.Driver)
		assert.Equal(t, "secret123", env.DBPassword)
		assert.Equal(t, 500, env.BatchSize)
		assert.Equal(t, true, env.AlerterEnabled)
		assert.Equal(t, "debug", env.LogLevel)
		assert.Equal(t, true, env.Debug)
	})
}

func TestEnv_Validate(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		os.Clearenv()
		env := Load()
		err := env.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid driver", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("DB_DRIVER", "mysql")
		env := Load()
		err := env.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DB_DRIVER must be sqlite or postgres")
	})

	t.Run("postgres without password", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("DB_DRIVER", "postgres")
		os.Setenv("DB_PASSWORD", "")
		env := Load()
		err := env.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DB_PASSWORD is required for postgres")
	})

	t.Run("weights do not sum to 1", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("WEIGHT_LATENCY", "0.50")
		os.Setenv("WEIGHT_EFFICIENCY", "0.50")
		env := Load()
		err := env.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must sum to 1.0")
	})

	t.Run("invalid threshold", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("DETECTOR_THRESHOLD", "150")
		env := Load()
		err := env.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be between 0 and 100")
	})

	t.Run("invalid KS threshold", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("DETECTOR_KS_THRESHOLD", "1.5")
		env := Load()
		err := env.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be between 0 and 1")
	})

	t.Run("invalid PSI threshold", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("DETECTOR_PSI_THRESHOLD", "0.05")
		env := Load()
		err := env.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be between 0.1 and 1")
	})

	t.Run("http webhook url", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SLACK_WEBHOOK_URL", "http://hooks.slack.com/test")
		env := Load()
		err := env.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must use HTTPS")
	})

	t.Run("https webhook url", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/test")
		env := Load()
		err := env.Validate()
		assert.NoError(t, err)
	})
}

func TestGetDurationEnv(t *testing.T) {
	t.Run("duration format", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_DURATION", "30s")
		result := getDurationEnv("TEST_DURATION", time.Minute)
		assert.Equal(t, 30*time.Second, result)
	})

	t.Run("minutes format", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_DURATION", "5m")
		result := getDurationEnv("TEST_DURATION", time.Minute)
		assert.Equal(t, 5*time.Minute, result)
	})

	t.Run("seconds integer", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_DURATION", "300")
		result := getDurationEnv("TEST_DURATION", time.Minute)
		assert.Equal(t, 300*time.Second, result)
	})

	t.Run("invalid format uses default", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_DURATION", "invalid")
		result := getDurationEnv("TEST_DURATION", 2*time.Minute)
		assert.Equal(t, 2*time.Minute, result)
	})

	t.Run("empty uses default", func(t *testing.T) {
		os.Clearenv()
		result := getDurationEnv("TEST_DURATION", 3*time.Minute)
		assert.Equal(t, 3*time.Minute, result)
	})
}

func TestGetBoolEnv(t *testing.T) {
	t.Run("true value", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_BOOL", "true")
		result := getBoolEnv("TEST_BOOL", false)
		assert.Equal(t, true, result)
	})

	t.Run("false value", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_BOOL", "false")
		result := getBoolEnv("TEST_BOOL", true)
		assert.Equal(t, false, result)
	})

	t.Run("invalid uses default", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_BOOL", "maybe")
		result := getBoolEnv("TEST_BOOL", true)
		assert.Equal(t, true, result)
	})
}

func TestGetIntEnv(t *testing.T) {
	t.Run("valid integer", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_INT", "42")
		result := getIntEnv("TEST_INT", 0)
		assert.Equal(t, 42, result)
	})

	t.Run("invalid uses default", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_INT", "not-a-number")
		result := getIntEnv("TEST_INT", 100)
		assert.Equal(t, 100, result)
	})
}

func TestGetFloatEnv(t *testing.T) {
	t.Run("valid float", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_FLOAT", "3.14")
		result := getFloatEnv("TEST_FLOAT", 0.0)
		assert.Equal(t, 3.14, result)
	})

	t.Run("invalid uses default", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_FLOAT", "not-a-number")
		result := getFloatEnv("TEST_FLOAT", 2.71)
		assert.Equal(t, 2.71, result)
	})
}
