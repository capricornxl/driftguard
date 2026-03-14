package handler

import (
	"fmt"
	"net/http"

	"github.com/driftguard/driftguard/internal/detector"
	"github.com/driftguard/driftguard/internal/middleware"
	"github.com/driftguard/driftguard/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RegisterAPIRoutes registers all API routes with middleware
func RegisterAPIRoutes(r *gin.Engine, db *gorm.DB, logger *logrus.Logger, enhancedDetector *detector.EnhancedDetector) {
	// API v1 group with common middleware
	v1 := r.Group("/api/v1")
	{
		// Health and readiness endpoints (no auth required)
		v1.GET("/health", HealthCheck(db))
		v1.GET("/ready", ReadyCheck(db))

		// Agent endpoints with validation middleware
		agents := v1.Group("/agents")
		agents.Use(middleware.ValidatePagination())
		{
			agents.GET("", ListAgents(db))
			agents.POST("", CreateAgent(db))

			// Agent-specific endpoints with agent_id validation
			agentSpecific := agents.Group("/:agent_id")
			agentSpecific.Use(middleware.ValidateAgentID())
			{
				agentSpecific.GET("", GetAgent(db))
				agentSpecific.PUT("", UpdateAgent(db))
				agentSpecific.DELETE("", DeleteAgent(db))

				// Health score endpoints
				agentSpecific.GET("/scores", GetHealthScores(db, enhancedDetector))
				agentSpecific.GET("/scores/latest", GetLatestHealthScore(db))
				agentSpecific.GET("/trend", middleware.ValidateTimeRange(), GetHealthTrend(db, enhancedDetector))

				// Drift detection endpoints
				agentSpecific.GET("/drift/ks-test", GetKSTest(db, enhancedDetector))
				agentSpecific.GET("/drift/psi", GetPSITest(db, enhancedDetector))
				agentSpecific.GET("/drift/spikes", middleware.ValidateTimeRange(), GetSpikes(db, enhancedDetector))
				agentSpecific.GET("/report/comprehensive", GetComprehensiveReport(db, enhancedDetector))

				// Alert endpoints
				agentSpecific.GET("/alerts", GetAlerts(db))
				agentSpecific.POST("/alerts", CreateAlert(db))
			}
		}

		// Metrics endpoint (Prometheus)
		v1.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}
}

// HealthCheck returns the health status of the service
func HealthCheck(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := "healthy"
		checks := gin.H{
			"status": status,
		}

		// Check database connection
		sqlDB, err := db.DB()
		if err != nil {
			status = "unhealthy"
			checks["database"] = "failed"
		} else if err := sqlDB.Ping(); err != nil {
			status = "unhealthy"
			checks["database"] = "failed"
		} else {
			checks["database"] = "ok"
		}

		if status == "healthy" {
			c.JSON(http.StatusOK, checks)
		} else {
			c.JSON(http.StatusServiceUnavailable, checks)
		}
	}
}

// ReadyCheck checks if the service is ready to accept traffic
func ReadyCheck(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"ready": false,
				"error": "Database connection not established",
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"ready": false,
				"error": "Database not reachable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ready": true,
		})
	}
}

// ListAgents returns all agents with pagination
func ListAgents(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := c.Get("page")
		limit, _ := c.Get("limit")

		var agents []models.Agent
		offset := (page.(int) - 1) * limit.(int)

		if err := db.Offset(offset).Limit(limit.(int)).Find(&agents).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch agents",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"agents": agents,
			"page":   page,
			"limit":  limit,
		})
	}
}

// CreateAgent creates a new agent
func CreateAgent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var agent models.Agent
		if err := c.ShouldBindJSON(&agent); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		if err := db.Create(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create agent",
			})
			return
		}

		c.JSON(http.StatusCreated, agent)
	}
}

// GetAgent returns a specific agent
func GetAgent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")

		var agent models.Agent
		if err := db.First(&agent, "id = ?", agentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Agent not found",
			})
			return
		}

		c.JSON(http.StatusOK, agent)
	}
}

// UpdateAgent updates an existing agent
func UpdateAgent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")

		var agent models.Agent
		if err := db.First(&agent, "id = ?", agentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Agent not found",
			})
			return
		}

		if err := c.ShouldBindJSON(&agent); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		if err := db.Save(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update agent",
			})
			return
		}

		c.JSON(http.StatusOK, agent)
	}
}

// DeleteAgent deletes an agent
func DeleteAgent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")

		if err := db.Delete(&models.Agent{}, "id = ?", agentID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete agent",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Agent deleted",
		})
	}
}

// GetHealthScores returns health scores for an agent
func GetHealthScores(db *gorm.DB, enhancedDetector *detector.EnhancedDetector) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")
		days, _ := c.Get("days")

		var scores []models.HealthScore
		err := db.Where("agent_id = ?", agentID).
			Order("created_at DESC").
			Limit(days.(int) * 24). // Assuming hourly scores
			Find(&scores).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch health scores",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"scores": scores,
			"count":  len(scores),
		})
	}
}

// GetLatestHealthScore returns the latest health score
func GetLatestHealthScore(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")

		var score models.HealthScore
		if err := db.Where("agent_id = ?", agentID).
			Order("created_at DESC").
			First(&score).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No health scores found",
			})
			return
		}

		c.JSON(http.StatusOK, score)
	}
}

// GetHealthTrend returns health trend analysis
func GetHealthTrend(db *gorm.DB, enhancedDetector *detector.EnhancedDetector) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")
		days, _ := c.Get("days")

		trend, err := enhancedDetector.AnalyzeHealthTrend(agentID, days.(int))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to analyze trend",
			})
			return
		}

		c.JSON(http.StatusOK, trend)
	}
}

// GetKSTest returns KS test results
func GetKSTest(db *gorm.DB, enhancedDetector *detector.EnhancedDetector) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")
		baseline := c.DefaultQuery("baseline", "7")
		current := c.DefaultQuery("current", "3")

		// Parse parameters with validation
		var baselineDays, currentDays int
		if _, err := fmt.Sscanf(baseline, "%d", &baselineDays); err != nil || baselineDays <= 0 {
			baselineDays = 7
		}
		if _, err := fmt.Sscanf(current, "%d", &currentDays); err != nil || currentDays <= 0 {
			currentDays = 3
		}

		result, err := enhancedDetector.DetectDriftWithKS(agentID, baselineDays, currentDays)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to perform KS test",
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// GetPSITest returns PSI test results
func GetPSITest(db *gorm.DB, enhancedDetector *detector.EnhancedDetector) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")
		baseline := c.DefaultQuery("baseline", "7")
		current := c.DefaultQuery("current", "7")
		buckets := c.DefaultQuery("buckets", "5")

		var baselineDays, currentDays, numBuckets int
		if _, err := fmt.Sscanf(baseline, "%d", &baselineDays); err != nil || baselineDays <= 0 {
			baselineDays = 7
		}
		if _, err := fmt.Sscanf(current, "%d", &currentDays); err != nil || currentDays <= 0 {
			currentDays = 7
		}
		if _, err := fmt.Sscanf(buckets, "%d", &numBuckets); err != nil || numBuckets <= 0 {
			numBuckets = 5
		}

		result, err := enhancedDetector.DetectDriftWithPSI(agentID, baselineDays, currentDays, numBuckets)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to calculate PSI",
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// GetSpikes returns spike detection results
func GetSpikes(db *gorm.DB, enhancedDetector *detector.EnhancedDetector) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")
		days, _ := c.Get("days")
		threshold := c.DefaultQuery("threshold", "2.5")

		var thresholdFloat float64
		if _, err := fmt.Sscanf(threshold, "%f", &thresholdFloat); err != nil || thresholdFloat <= 0 {
			thresholdFloat = 2.5
		}

		result, err := enhancedDetector.DetectSpikes(agentID, days.(int), thresholdFloat)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to detect spikes",
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// GetComprehensiveReport returns comprehensive drift analysis report
func GetComprehensiveReport(db *gorm.DB, enhancedDetector *detector.EnhancedDetector) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")

		report, err := enhancedDetector.GenerateComprehensiveReport(agentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate report",
			})
			return
		}

		c.JSON(http.StatusOK, report)
	}
}

// GetAlerts returns alerts for an agent
func GetAlerts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")
		page, _ := c.Get("page")
		limit, _ := c.Get("limit")

		var alerts []models.Alert
		offset := (page.(int) - 1) * limit.(int)

		if err := db.Where("agent_id = ?", agentID).
			Offset(offset).
			Limit(limit.(int)).
			Order("created_at DESC").
			Find(&alerts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch alerts",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"alerts": alerts,
			"page":   page,
			"limit":  limit,
		})
	}
}

// CreateAlert creates a new alert
func CreateAlert(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")

		var alert models.Alert
		if err := c.ShouldBindJSON(&alert); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		alert.AgentID = agentID

		if err := db.Create(&alert).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create alert",
			})
			return
		}

		c.JSON(http.StatusCreated, alert)
	}
}
