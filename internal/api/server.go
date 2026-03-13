package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/driftguard/driftguard/internal/alerter"
	"github.com/driftguard/driftguard/internal/collector"
	"github.com/driftguard/driftguard/internal/detector"
	"github.com/driftguard/driftguard/internal/evaluator"
	"github.com/driftguard/driftguard/pkg/metrics"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server API 服务器
type Server struct {
	collector  *collector.Collector
	evaluator  *evaluator.Evaluator
	detector   *detector.Detector
	alerter    *alerter.Alerter
	metrics    *metrics.Metrics
}

// NewServer 创建 API 服务器
func NewServer(c *collector.Collector, e *evaluator.Evaluator, d *detector.Detector, a *alerter.Alerter, m *metrics.Metrics) *Server {
	return &Server{
		collector:  c,
		evaluator:  e,
		detector:   d,
		alerter:    a,
		metrics:    m,
	}
}

// Run 启动服务器
func (s *Server) Run(host, port string) error {
	r := gin.Default()

	// 健康检查
	r.GET("/health", s.healthHandler)

	// Prometheus Metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 数据采集
	r.POST("/api/v1/interactions", s.collectHandler)

	// 健康度查询
	r.GET("/api/v1/agents/:agent_id/health", s.healthScoreHandler)
	r.POST("/api/v1/agents/:agent_id/evaluate", s.evaluateHandler)

	// 退化检测
	r.GET("/api/v1/agents/:agent_id/detect", s.detectHandler)

	// 告警管理
	r.GET("/api/v1/alerts", s.listAlertsHandler)
	r.POST("/api/v1/alerts/:id/ack", s.ackAlertHandler)
	r.POST("/api/v1/alerts/:id/resolve", s.resolveAlertHandler)

	// 统计信息
	r.GET("/api/v1/stats", s.statsHandler)

	log.Printf("[API] Starting server on %s:%s", host, port)
	log.Printf("[API] Metrics endpoint: http://%s:%s/metrics", host, port)
	return r.Run(host + ":" + port)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "driftguard",
	})
}

func (s *Server) collectHandler(c *gin.Context) {
	var interaction InteractionRequest
	if err := c.ShouldBindJSON(&interaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 转换为 models.Interaction 并存储
	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

func (s *Server) healthScoreHandler(c *gin.Context) {
	agentID := c.Param("agent_id")
	windowHours := c.DefaultQuery("window", "24")

	// TODO: 调用 evaluator.GetLatestScore
	c.JSON(http.StatusOK, gin.H{
		"agent_id": agentID,
		"window_hours": windowHours,
		"status": "not_implemented",
	})
}

func (s *Server) evaluateHandler(c *gin.Context) {
	agentID := c.Param("agent_id")
	windowHours := 24

	result, err := s.evaluator.Evaluate(agentID, windowHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.evaluator.LogResult(result)
	s.evaluator.SaveResult(result)

	c.JSON(http.StatusOK, result)
}

func (s *Server) detectHandler(c *gin.Context) {
	agentID := c.Param("agent_id")

	result, err := s.detector.Detect(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.detector.LogResult(result)

	// 如果有告警，创建并推送
	for _, alert := range result.Alerts {
		if err := s.detector.CreateAlert(agentID, alert); err != nil {
			log.Printf("[API] Failed to create alert: %v", err)
		}
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) listAlertsHandler(c *gin.Context) {
	agentID := c.Query("agent_id")
	unacked := c.Query("unacked") == "true"

	var alerts []AlertResponse
	// TODO: 从数据库查询
	_ = agentID
	_ = unacked

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total": len(alerts),
	})
}

func (s *Server) ackAlertHandler(c *gin.Context) {
	// TODO: 实现告警确认
	c.JSON(http.StatusOK, gin.H{"status": "acked"})
}

func (s *Server) resolveAlertHandler(c *gin.Context) {
	// TODO: 实现告警解决
	c.JSON(http.StatusOK, gin.H{"status": "resolved"})
}

func (s *Server) statsHandler(c *gin.Context) {
	collectorStats := s.collector.Stats()

	alerterStats, err := s.alerter.GetStats()
	if err != nil {
		alerterStats = map[string]interface{}{"error": err.Error()}
	}

	c.JSON(http.StatusOK, gin.H{
		"collector": collectorStats,
		"alerter":   alerterStats,
	})
}

// InteractionRequest 交互数据请求
type InteractionRequest struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
	Input     string `json:"input"`
	Output    string `json:"output"`
	LatencyMs int    `json:"latency_ms"`
	TokensIn  int    `json:"tokens_in"`
	TokensOut int    `json:"tokens_out"`
}

// AlertResponse 告警响应
type AlertResponse struct {
	ID        uint   `json:"id"`
	AgentID   string `json:"agent_id"`
	Level     string `json:"level"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// Helper to write JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
