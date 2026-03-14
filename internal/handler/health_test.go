package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	return db
}

func TestEnhancedHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	db := setupTestDB()

	router.GET("/health", EnhancedHealthCheck(db, nil, nil, time.Now(), "test-v1.0.0"))

	t.Run("healthy status", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "healthy")
		assert.Contains(t, w.Body.String(), "test-v1.0.0")
	})

	t.Run("includes timestamp", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Contains(t, w.Body.String(), "timestamp")
	})

	t.Run("includes uptime", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Contains(t, w.Body.String(), "uptime")
	})

	t.Run("includes checks", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		router.ServeHTTP(w, req)

		assert.Contains(t, w.Body.String(), "checks")
		assert.Contains(t, w.Body.String(), "database")
	})
}

func TestEnhancedReadyCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	db := setupTestDB()

	router.GET("/ready", EnhancedReadyCheck(db, nil))

	t.Run("ready status", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ready", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ready")
		assert.Contains(t, w.Body.String(), "true")
	})
}

func TestLiveCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/live", LiveCheck())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/live", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alive")
	assert.Contains(t, w.Body.String(), "true")
}

func TestStartupCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	startTime := time.Now().Add(-60 * time.Second) // Started 60s ago

	router.GET("/startup", StartupCheck(startTime))

	t.Run("not starting after 30s", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/startup", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "false")
	})

	t.Run("starting within 30s", func(t *testing.T) {
		router2 := gin.New()
		recentStart := time.Now().Add(-10 * time.Second)
		router2.GET("/startup", StartupCheck(recentStart))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/startup", nil)
		router2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "true")
		assert.Contains(t, w.Body.String(), "starting up")
	})
}

func TestCheckDatabase(t *testing.T) {
	t.Run("nil database", func(t *testing.T) {
		result := checkDatabase(nil)
		assert.Equal(t, "not_configured", result)
	})

	t.Run("valid database", func(t *testing.T) {
		db := setupTestDB()
		result := checkDatabase(db)
		assert.Equal(t, "ok", result)
	})
}

func TestCheckMemory(t *testing.T) {
	result := checkMemory()
	assert.Equal(t, "ok", result)
}

func TestCheckDisk(t *testing.T) {
	result := checkDisk()
	assert.Equal(t, "ok", result)
}
