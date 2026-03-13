package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics Prometheus 指标集合
type Metrics struct {
	// 健康度评分
	HealthScore *prometheus.GaugeVec
	
	// 健康度各维度得分
	HealthScoreLatency *prometheus.GaugeVec
	HealthScoreEfficiency *prometheus.GaugeVec
	HealthScoreConsistency *prometheus.GaugeVec
	HealthScoreAccuracy *prometheus.GaugeVec
	HealthScoreHallucination *prometheus.GaugeVec
	
	// 交互统计
	InteractionsTotal *prometheus.CounterVec
	InteractionLatency *prometheus.HistogramVec
	InteractionTokens *prometheus.HistogramVec
	
	// 退化检测
	DegradedAgents *prometheus.GaugeVec
	DetectionTrend *prometheus.GaugeVec
	DetectionConfidence *prometheus.GaugeVec
	
	// 告警统计
	AlertsTotal *prometheus.CounterVec
	AlertsActive *prometheus.GaugeVec
	AlertsByLevel *prometheus.GaugeVec
}

// NewMetrics 创建 Metrics 实例
func NewMetrics() *Metrics {
	return &Metrics{
		// 健康度评分
		HealthScore: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "driftguard_health_score",
			Help: "Agent health score (0-100)",
		}, []string{"agent_id", "level"}),
		
		HealthScoreLatency: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "driftguard_health_score_latency",
			Help: "Agent health score - latency dimension",
		}, []string{"agent_id"}),
		
		HealthScoreEfficiency: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "driftguard_health_score_efficiency",
			Help: "Agent health score - efficiency dimension",
		}, []string{"agent_id"}),
		
		HealthScoreConsistency: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "driftguard_health_score_consistency",
			Help: "Agent health score - consistency dimension",
		}, []string{"agent_id"}),
		
		HealthScoreAccuracy: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "driftguard_health_score_accuracy",
			Help: "Agent health score - accuracy dimension",
		}, []string{"agent_id"}),
		
		HealthScoreHallucination: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "driftguard_health_score_hallucination",
			Help: "Agent health score - hallucination dimension",
		}, []string{"agent_id"}),
		
		// 交互统计
		InteractionsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "driftguard_interactions_total",
			Help: "Total number of interactions",
		}, []string{"agent_id"}),
		
		InteractionLatency: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "driftguard_interaction_latency_ms",
			Help:    "Interaction latency distribution (ms)",
			Buckets: []float64{50, 100, 200, 500, 1000, 2000, 3000, 5000},
		}, []string{"agent_id"}),
		
		InteractionTokens: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "driftguard_interaction_tokens",
			Help:    "Interaction tokens distribution",
			Buckets: []float64{10, 50, 100, 200, 500, 1000, 2000},
		}, []string{"agent_id", "type"}), // type: input or output
		
		// 退化检测
		DegradedAgents: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "driftguard_degraded_agents",
			Help: "Number of degraded agents",
		}, []string{"agent_id", "degraded_type"}),
		
		DetectionTrend: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "driftguard_detection_trend",
			Help: "Agent detection trend slope",
		}, []string{"agent_id"}),
		
		DetectionConfidence: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "driftguard_detection_confidence",
			Help: "Detection confidence percentage",
		}, []string{"agent_id"}),
		
		// 告警统计
		AlertsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "driftguard_alerts_total",
			Help: "Total number of alerts",
		}, []string{"agent_id", "level", "type"}),
		
		AlertsActive: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "driftguard_alerts_active",
			Help: "Number of active (unresolved) alerts",
		}, []string{"agent_id"}),
		
		AlertsByLevel: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "driftguard_alerts_by_level",
			Help: "Alerts grouped by level",
		}, []string{"level"}),
	}
}

// UpdateHealthScore 更新健康度指标
func (m *Metrics) UpdateHealthScore(agentID string, score float64, level string) {
	m.HealthScore.WithLabelValues(agentID, level).Set(score)
}

// UpdateHealthScoreDimensions 更新健康度各维度
func (m *Metrics) UpdateHealthScoreDimensions(agentID string, latency, efficiency, consistency, accuracy, hallucination float64) {
	m.HealthScoreLatency.WithLabelValues(agentID).Set(latency)
	m.HealthScoreEfficiency.WithLabelValues(agentID).Set(efficiency)
	m.HealthScoreConsistency.WithLabelValues(agentID).Set(consistency)
	m.HealthScoreAccuracy.WithLabelValues(agentID).Set(accuracy)
	m.HealthScoreHallucination.WithLabelValues(agentID).Set(hallucination)
}

// RecordInteraction 记录交互
func (m *Metrics) RecordInteraction(agentID string, latencyMs, tokensIn, tokensOut int) {
	m.InteractionsTotal.WithLabelValues(agentID).Inc()
	m.InteractionLatency.WithLabelValues(agentID).Observe(float64(latencyMs))
	m.InteractionTokens.WithLabelValues(agentID, "input").Observe(float64(tokensIn))
	m.InteractionTokens.WithLabelValues(agentID, "output").Observe(float64(tokensOut))
}

// UpdateDetection 更新检测结果
func (m *Metrics) UpdateDetection(agentID string, isDegraded bool, degradedType string, trend float64, confidence float64) {
	degraded := 0.0
	if isDegraded {
		degraded = 1.0
	}
	m.DegradedAgents.WithLabelValues(agentID, degradedType).Set(degraded)
	m.DetectionTrend.WithLabelValues(agentID).Set(trend)
	m.DetectionConfidence.WithLabelValues(agentID).Set(confidence)
}

// RecordAlert 记录告警
func (m *Metrics) RecordAlert(agentID, level, alertType string) {
	m.AlertsTotal.WithLabelValues(agentID, level, alertType).Inc()
	m.AlertsByLevel.WithLabelValues(level).Inc()
}

// UpdateActiveAlerts 更新活跃告警数
func (m *Metrics) UpdateActiveAlerts(agentID string, count int) {
	m.AlertsActive.WithLabelValues(agentID).Set(float64(count))
}
