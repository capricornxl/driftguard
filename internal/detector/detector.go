package detector

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/metrics"
	"github.com/driftguard/driftguard/pkg/models"
	"gorm.io/gorm"
)

// Detector 退化检测器
type Detector struct {
	cfg *config.DetectorConfig
	db  *gorm.DB
}

// DetectionResult 检测结果
type DetectionResult struct {
	AgentID      string    `json:"agent_id"`
	IsDegraded   bool      `json:"is_degraded"`
	DegradedType string    `json:"degraded_type,omitempty"`
	Confidence   float64   `json:"confidence"`
	Trend        string    `json:"trend"` // improving, stable, degrading
	TrendSlope   float64   `json:"trend_slope"`
	Alerts       []AlertInfo `json:"alerts"`
	DetectedAt   time.Time `json:"detected_at"`
}

// AlertInfo 告警信息
type AlertInfo struct {
	Type      string  `json:"type"`      // threshold, trend, anomaly, spike
	Level     string  `json:"level"`     // info, warning, critical
	Message   string  `json:"message"`
	Score     float64 `json:"score"`
	Threshold float64 `json:"threshold"`
}

// NewDetector 创建检测器
func NewDetector(cfg *config.DetectorConfig, db *gorm.DB) *Detector {
	return &Detector{
		cfg: cfg,
		db:  db,
	}
}

// Detect 执行退化检测
func (d *Detector) Detect(agentID string) (*DetectionResult, error) {
	result := &DetectionResult{
		AgentID:    agentID,
		DetectedAt: time.Now(),
		Alerts:     []AlertInfo{},
	}

	// 获取历史健康度数据
	windowStart := time.Now().Add(-time.Duration(d.cfg.WindowDays) * 24 * time.Hour)
	var scores []models.HealthScore
	if err := d.db.Where("agent_id = ? AND created_at >= ?", agentID, windowStart).
		Order("created_at ASC").Find(&scores).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch scores: %w", err)
	}

	if len(scores) < 3 {
		result.Trend = "insufficient_data"
		return result, nil
	}

	// 1. 阈值检测
	thresholdAlert := d.detectThreshold(scores)
	if thresholdAlert != nil {
		result.Alerts = append(result.Alerts, *thresholdAlert)
		result.IsDegraded = true
		result.DegradedType = "threshold"
	}

	// 2. 趋势检测
	trendAlert, trend, slope := d.detectTrend(scores)
	result.Trend = trend
	result.TrendSlope = slope
	if trendAlert != nil {
		result.Alerts = append(result.Alerts, *trendAlert)
		if !result.IsDegraded {
			result.IsDegraded = true
			result.DegradedType = "trend"
		}
	}

	// 3. 异常检测
	anomalyAlert := d.detectAnomaly(scores)
	if anomalyAlert != nil {
		result.Alerts = append(result.Alerts, *anomalyAlert)
		if !result.IsDegraded {
			result.IsDegraded = true
			result.DegradedType = "anomaly"
		}
	}

	// 4. 突变检测
	spikeAlert := d.detectSpike(scores)
	if spikeAlert != nil {
		result.Alerts = append(result.Alerts, *spikeAlert)
		if !result.IsDegraded {
			result.IsDegraded = true
			result.DegradedType = "spike"
		}
	}

	// 计算置信度
	result.Confidence = d.calculateConfidence(result.Alerts)

	return result, nil
}

// detectThreshold 阈值检测
func (d *Detector) detectThreshold(scores []models.HealthScore) *AlertInfo {
	if len(scores) == 0 {
		return nil
	}

	latest := scores[len(scores)-1]
	currentScore := latest.Score

	// 严重阈值 < 50
	if currentScore < 50 {
		return &AlertInfo{
			Type:      "threshold",
			Level:     "critical",
			Message:   fmt.Sprintf("健康度严重低于阈值 (当前：%.1f, 阈值：50)", currentScore),
			Score:     currentScore,
			Threshold: 50,
		}
	}

	// 警告阈值 < 70
	if currentScore < 70 {
		return &AlertInfo{
			Type:      "threshold",
			Level:     "warning",
			Message:   fmt.Sprintf("健康度低于警告阈值 (当前：%.1f, 阈值：70)", currentScore),
			Score:     currentScore,
			Threshold: 70,
		}
	}

	return nil
}

// detectTrend 趋势检测 (线性回归)
func (d *Detector) detectTrend(scores []models.HealthScore) (*AlertInfo, string, float64) {
	if len(scores) < 3 {
		return nil, "insufficient_data", 0
	}

	// 简单线性回归
	n := float64(len(scores))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i, score := range scores {
		x := float64(i)
		y := score.Score
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// 判断趋势
	var trend string
	if slope > 0.5 {
		trend = "improving"
	} else if slope < -0.5 {
		trend = "degrading"
	} else {
		trend = "stable"
	}

	// 如果持续下降，发出告警
	if slope < -1.0 {
		latest := scores[len(scores)-1]
		return &AlertInfo{
			Type:      "trend",
			Level:     "warning",
			Message:   fmt.Sprintf("检测到持续退化趋势 (斜率：%.2f/天)", slope),
			Score:     latest.Score,
			Threshold: -1.0,
		}, trend, slope
	}

	return nil, trend, slope
}

// detectAnomaly 异常检测 (基于标准差)
func (d *Detector) detectAnomaly(scores []models.HealthScore) *AlertInfo {
	if len(scores) < 5 {
		return nil
	}

	// 计算均值和标准差
	sum := 0.0
	for _, s := range scores {
		sum += s.Score
	}
	mean := sum / float64(len(scores))

	variance := 0.0
	for _, s := range scores {
		diff := s.Score - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(scores)))

	// 检查最新点是否异常
	latest := scores[len(scores)-1]
	zScore := (latest.Score - mean) / stdDev

	if zScore < -2.0 { // 低于 2 个标准差
		return &AlertInfo{
			Type:      "anomaly",
			Level:     "warning",
			Message:   fmt.Sprintf("检测到异常下降 (Z-Score: %.2f)", zScore),
			Score:     latest.Score,
			Threshold: -2.0,
		}
	}

	return nil
}

// detectSpike 突变检测 (相邻点差异)
func (d *Detector) detectSpike(scores []models.HealthScore) *AlertInfo {
	if len(scores) < 2 {
		return nil
	}

	latest := scores[len(scores)-1]
	previous := scores[len(scores)-2]

	diff := latest.Score - previous.Score

	// 突然下降超过 15 分
	if diff < -15 {
		return &AlertInfo{
			Type:      "spike",
			Level:     "critical",
			Message:   fmt.Sprintf("检测到健康度骤降 (变化：%.1f 分)", diff),
			Score:     latest.Score,
			Threshold: -15,
		}
	}

	return nil
}

// calculateConfidence 计算检测置信度
func (d *Detector) calculateConfidence(alerts []AlertInfo) float64 {
	if len(alerts) == 0 {
		return 0
	}

	// 每个告警类型增加置信度
	confidence := 0.0
	criticalCount := 0
	warningCount := 0

	for _, alert := range alerts {
		if alert.Level == "critical" {
			criticalCount++
			confidence += 40
		} else if alert.Level == "warning" {
			warningCount++
			confidence += 20
		} else {
			confidence += 10
		}
	}

	// 多种告警类型同时出现，置信度更高
	if len(alerts) >= 3 {
		confidence += 20
	}

	if confidence > 100 {
		confidence = 100
	}

	return confidence
}

// CreateAlert 创建告警记录
func (d *Detector) CreateAlert(agentID string, alert AlertInfo) error {
	modelAlert := models.Alert{
		AgentID:   agentID,
		Level:     alert.Level,
		Type:      alert.Type,
		Score:     alert.Score,
		Threshold: alert.Threshold,
		Message:   alert.Message,
		Acked:     false,
		Resolved:  false,
		CreatedAt: time.Now(),
	}

	return d.db.Create(&modelAlert).Error
}

// GetActiveAlerts 获取活跃告警
func (d *Detector) GetActiveAlerts(agentID string) ([]models.Alert, error) {
	var alerts []models.Alert
	err := d.db.Where("agent_id = ? AND resolved = ?", agentID, false).
		Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

// LogResult 打印检测结果并导出 Metrics
func (d *Detector) LogResult(result *DetectionResult) {
	status := "healthy"
	if result.IsDegraded {
		status = fmt.Sprintf("DEGRADED (%s)", result.DegradedType)
	}
	log.Printf("[Detector] Agent=%s Status=%s Trend=%s (slope=%.2f) Confidence=%.1f%% Alerts=%d",
		result.AgentID, status, result.Trend, result.TrendSlope,
		result.Confidence, len(result.Alerts))

	// 导出 Prometheus Metrics
	if d.metrics != nil {
		d.metrics.UpdateDetection(
			result.AgentID,
			result.IsDegraded,
			result.DegradedType,
			result.TrendSlope,
			result.Confidence,
		)
	}
}
