package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/driftguard/driftguard/internal/api"
	"github.com/driftguard/driftguard/internal/collector"
	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto migrate all models
	err = models.AutoMigrate(db)
	assert.NoError(t, err)

	return db
}

// setupTestServer creates a test HTTP server
func setupTestServer(t *testing.T, db *gorm.DB) *httptest.Server {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 0, // Auto-assign port
		},
	}

	router := gin.New()
	api.SetupRoutes(router, db, cfg)

	return httptest.NewServer(router)
}

// TestHealthEndpoint tests the /health endpoint
func TestHealthEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := setupTestServer(t, db)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", health["status"])
	assert.Contains(t, health, "timestamp")
	assert.Contains(t, health, "version")
}

// TestReadyEndpoint tests the /ready endpoint
func TestReadyEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := setupTestServer(t, db)
	defer server.Close()

	resp, err := http.Get(server.URL + "/ready")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var ready map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&ready)
	assert.NoError(t, err)

	assert.Equal(t, true, ready["ready"])
}

// TestListAgentsEndpoint tests the /api/v1/agents endpoint
func TestListAgentsEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := setupTestServer(t, db)
	defer server.Close()

	// Create test agent
	agent := models.Agent{
		ID:        "test-agent-1",
		Name:      "Test Agent",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(&agent)

	resp, err := http.Get(server.URL + "/api/v1/agents")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var agents []models.Agent
	err = json.NewDecoder(resp.Body).Decode(&agents)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(agents), 1)
}

// TestGetAgentHealthEndpoint tests the /api/v1/agents/{id}/health endpoint
func TestGetAgentHealthEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := setupTestServer(t, db)
	defer server.Close()

	// Create test agent and health score
	agent := models.Agent{
		ID:        "test-agent-health",
		Name:      "Test Agent Health",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(&agent)

	healthScore := models.HealthScore{
		AgentID:   agent.ID,
		Score:     85.5,
		Latency:   90.0,
		Efficiency: 85.0,
		Consistency: 88.0,
		Accuracy:  82.0,
		Hallucination: 5.0,
		Trend:     "stable",
		CreatedAt: time.Now(),
	}
	db.Create(&healthScore)

	resp, err := http.Get(server.URL + "/api/v1/agents/test-agent-health/health")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, "test-agent-health", result["agentId"])
	assert.Equal(t, 85.5, result["score"])
}

// TestMetricsEndpoint tests the /metrics endpoint
func TestMetricsEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := setupTestServer(t, db)
	defer server.Close()

	resp, err := http.Get(server.URL + "/metrics")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", resp.Header.Get("Content-Type"))
}

// TestAgentHealthHistoryEndpoint tests the /api/v1/agents/{id}/history endpoint
func TestAgentHealthHistoryEndpoint(t *testing.T) {
	db := setupTestDB(t)
	server := setupTestServer(t, db)
	defer server.Close()

	// Create test agent
	agent := models.Agent{
		ID:        "test-agent-history",
		Name:      "Test Agent History",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(&agent)

	// Create historical health scores
	now := time.Now()
	for i := 0; i < 5; i++ {
		healthScore := models.HealthScore{
			AgentID:     agent.ID,
			Score:       80.0 + float64(i)*2,
			Latency:     85.0,
			Efficiency:  80.0,
			Consistency: 82.0,
			Accuracy:    78.0,
			Hallucination: 5.0,
			Trend:       "stable",
			CreatedAt:   now.Add(-time.Duration(i) * time.Hour),
		}
		db.Create(&healthScore)
	}

	resp, err := http.Get(server.URL + "/api/v1/agents/test-agent-history/history?limit=10")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var history []models.HealthScore
	err = json.NewDecoder(resp.Body).Decode(&history)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(history), 5)
}

// TestCollectorIntegration tests the collector component integration
func TestCollectorIntegration(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.CollectorConfig{
		Enabled:  true,
		Interval: "1m",
		BatchSize: 100,
	}

	collector := collector.NewCollector(cfg, db, nil)
	assert.NotNil(t, collector)
}
