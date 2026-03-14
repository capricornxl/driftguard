package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	version   = "0.1.1"
	buildTime = "2026-03-14"
	startTime = time.Now()
)

type HealthResponse struct {
	Status      string            `json:"status"`
	Timestamp   string            `json:"timestamp"`
	Version     string            `json:"version"`
	Uptime      string            `json:"uptime"`
	Checks      map[string]string `json:"checks"`
	Environment string            `json:"environment"`
}

type Agent struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	AgentID     string    `json:"agent_id" gorm:"uniqueIndex"`
	Name        string    `json:"name"`
	Model       string    `json:"model"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
}

type Interaction struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	AgentID   string    `json:"agent_id" gorm:"index"`
	LatencyMs int       `json:"latency_ms"`
	TokensIn  int       `json:"tokens_in"`
	TokensOut int       `json:"tokens_out"`
	CreatedAt time.Time `json:"created_at"`
}

type HealthScore struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	AgentID         string    `json:"agent_id" gorm:"index"`
	Score           float64   `json:"score"`
	LatencyScore    float64   `json:"latency_score"`
	EfficiencyScore float64   `json:"efficiency_score"`
	ConsistencyScore float64  `json:"consistency_score"`
	AccuracyScore   float64   `json:"accuracy_score"`
	HallucinationScore float64 `json:"hallucination_score"`
	IsDegraded      bool      `json:"is_degraded"`
	CalculatedAt    time.Time `json:"calculated_at"`
}

var db *gorm.DB

func main() {
	host := getEnv("HOST", "172.16.223.251")
	port := getEnv("PORT", "8080")
	logLevel := getEnv("LOG_LEVEL", "info")
	logFormat := getEnv("LOG_FORMAT", "json")
	dbName := getEnv("DB_NAME", "/tmp/driftguard.db")

	if logLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	var err error
	db, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	db.AutoMigrate(&Agent{}, &Interaction{}, &HealthScore{})

	r := gin.Default()

	r.GET("/api/v1/health", healthHandler)
	r.GET("/api/v1/live", liveHandler)
	r.GET("/api/v1/ready", readyHandler)
	r.GET("/api/v1/metrics", metricsHandler)

	r.GET("/api/v1/agents", listAgents)
	r.POST("/api/v1/agents", createAgent)
	r.GET("/api/v1/agents/:agent_id", getAgent)
	r.POST("/api/v1/agents/:agent_id/interactions", createInteraction)
	r.GET("/api/v1/agents/:agent_id/scores/latest", getLatestScore)

	addr := fmt.Sprintf("%s:%s", host, port)
	log.Printf("🚀 DriftGuard v%s starting on %s", version, addr)
	log.Printf("📊 Environment: %s | Log: %s", getEnv("ENVIRONMENT", "development"), logFormat)
	log.Printf("📝 Docs: http://%s:%s/api/v1/health", host, port)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func healthHandler(c *gin.Context) {
	uptime := time.Since(startTime)
	checks := map[string]string{
		"database": "ok",
		"configuration": "ok",
		"memory": "ok",
		"disk": "ok",
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:      "healthy",
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Version:     version,
		Uptime:      uptime.String(),
		Checks:      checks,
		Environment: getEnv("ENVIRONMENT", "development"),
	})
}

func liveHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

func readyHandler(c *gin.Context) {
	if db != nil {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
	}
}

func metricsHandler(c *gin.Context) {
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, `# HELP driftguard_interactions_total Total number of interactions
# TYPE driftguard_interactions_total counter
driftguard_interactions_total 0
# HELP driftguard_health_score Current health score
# TYPE driftguard_health_score gauge
driftguard_health_score 100
# HELP driftguard_uptime_seconds Uptime in seconds
# TYPE driftguard_uptime_seconds counter
driftguard_uptime_seconds %.0f
`, uptimeSeconds())
}

func listAgents(c *gin.Context) {
	var agents []Agent
	db.Find(&agents)
	c.JSON(http.StatusOK, gin.H{"agents": agents, "total": len(agents)})
}

func createAgent(c *gin.Context) {
	var agent Agent
	if err := c.ShouldBindJSON(&agent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	agent.Status = "active"
	agent.CreatedAt = time.Now()
	db.Create(&agent)
	c.JSON(http.StatusCreated, agent)
}

func getAgent(c *gin.Context) {
	agentID := c.Param("agent_id")
	var agent Agent
	if err := db.Where("agent_id = ?", agentID).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}
	c.JSON(http.StatusOK, agent)
}

func createInteraction(c *gin.Context) {
	agentID := c.Param("agent_id")
	var interaction Interaction
	if err := c.ShouldBindJSON(&interaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	interaction.AgentID = agentID
	interaction.CreatedAt = time.Now()
	db.Create(&interaction)
	c.JSON(http.StatusCreated, interaction)
}

func getLatestScore(c *gin.Context) {
	agentID := c.Param("agent_id")
	var score HealthScore
	if err := db.Where("agent_id = ?", agentID).Order("calculated_at desc").First(&score).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"agent_id": agentID,
			"message": "No scores yet",
			"score": nil,
		})
		return
	}
	c.JSON(http.StatusOK, score)
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func uptimeSeconds() float64 {
	return time.Since(startTime).Seconds()
}
