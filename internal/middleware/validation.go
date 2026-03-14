package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Validation middleware for API input validation (P0 fix: 输入验证不足)

var (
	// AgentID 格式：字母数字 + 连字符/下划线，长度 3-64
	agentIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`)
	
	// 时间范围查询最大窗口（天）
	maxTimeWindowDays = 90
)

// ValidateAgentID validates agent_id path parameter
func ValidateAgentID() gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("agent_id")
		
		if agentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "agent_id is required",
			})
			c.Abort()
			return
		}
		
		if !agentIDRegex.MatchString(agentID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid agent_id format. Must be 3-64 characters, alphanumeric with underscores or hyphens",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// ValidateTimeRange validates time range query parameters
func ValidateTimeRange() gin.HandlerFunc {
	return func(c *gin.Context) {
		daysStr := c.DefaultQuery("days", "7")
		var days int
		
		if _, err := fmt.Sscanf(daysStr, "%d", &days); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid 'days' parameter",
			})
			c.Abort()
			return
		}
		
		if days <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "'days' must be positive",
			})
			c.Abort()
			return
		}
		
		if days > maxTimeWindowDays {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("'days' cannot exceed %d", maxTimeWindowDays),
			})
			c.Abort()
			return
		}
		
		c.Set("days", days)
		c.Next()
	}
}

// ValidatePagination validates pagination parameters
func ValidatePagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "20")
		
		var page, limit int
		
		if _, err := fmt.Sscanf(pageStr, "%d", &page); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid 'page' parameter",
			})
			c.Abort()
			return
		}
		
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid 'limit' parameter",
			})
			c.Abort()
			return
		}
		
		if page <= 0 {
			page = 1
		}
		
		if limit <= 0 {
			limit = 20
		}
		
		if limit > 100 {
			limit = 100
		}
		
		c.Set("page", page)
		c.Set("limit", limit)
		c.Next()
	}
}

// ValidateScoreThreshold validates score threshold parameters
func ValidateScoreThreshold() gin.HandlerFunc {
	return func(c *gin.Context) {
		thresholdStr := c.Query("threshold")
		
		if thresholdStr == "" {
			c.Next()
			return
		}
		
		var threshold float64
		if _, err := fmt.Sscanf(thresholdStr, "%f", &threshold); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid 'threshold' parameter",
			})
			c.Abort()
			return
		}
		
		if threshold < 0 || threshold > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "'threshold' must be between 0 and 100",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequestLogger adds structured logging with request ID (P1 fix: 日志结构化)
func RequestLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or extract request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		
		// Create structured logger with request context
		start := time.Now()
		path := c.Request.URL.Path
		
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       path,
			"client_ip":  c.ClientIP(),
		}).Info("Request started")
		
		c.Next()
		
		// Log completion
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		
		logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       path,
			"status":     statusCode,
			"latency":    latency.String(),
			"client_ip":  c.ClientIP(),
		}).Info("Request completed")
	}
}

// RecoveryWithLogger is a custom recovery middleware with structured logging
func RecoveryWithLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")
				
				logger.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err,
				}).Error("Panic recovered")
				
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"request_id": requestID,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// RateLimiter simple in-memory rate limiter (basic implementation)
func RateLimiter(requestsPerMinute int) gin.HandlerFunc {
	// In production, use Redis or similar for distributed rate limiting
	return func(c *gin.Context) {
		// TODO: Implement proper rate limiting
		c.Next()
	}
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}
