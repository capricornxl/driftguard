package alerter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/metrics"
	"github.com/driftguard/driftguard/pkg/models"
	"gorm.io/gorm"
)

// Alerter 告警处理器
type Alerter struct {
	cfg     *config.AlerterConfig
	db      *gorm.DB
	metrics *metrics.Metrics
}

// NewAlerter 创建告警器
func NewAlerter(cfg *config.AlerterConfig, db *gorm.DB, m *metrics.Metrics) *Alerter {
	return &Alerter{
		cfg:     cfg,
		db:      db,
		metrics: m,
	}
}

// AlertPayload 告警推送负载
type AlertPayload struct {
	AgentID   string `json:"agent_id"`
	Level     string `json:"level"`
	Type      string `json:"type"`
	Score     float64 `json:"score"`
	Threshold float64 `json:"threshold"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// Process 处理告警
func (a *Alerter) Process(alert *models.Alert) error {
	if !a.cfg.Enabled {
		log.Printf("[Alerter] Disabled, skipping alert: %s", alert.Message)
		return nil
	}

	log.Printf("[Alerter] Processing alert: level=%s type=%s agent=%s",
		alert.Level, alert.Type, alert.AgentID)

	// 导出 Prometheus Metrics
	if a.metrics != nil {
		a.metrics.RecordAlert(alert.AgentID, alert.Level, alert.Type)
	}

	// 遍历配置的渠道
	for _, channel := range a.cfg.Channels {
		switch channel {
		case "webhook":
			if err := a.sendWebhook(alert); err != nil {
				log.Printf("[Alerter] Webhook failed: %v", err)
			}
		case "log":
			a.logAlert(alert)
		case "database":
			// 已存储，无需额外操作
		default:
			log.Printf("[Alerter] Unknown channel: %s", channel)
		}
	}

	return nil
}

// sendWebhook 发送 Webhook 告警
func (a *Alerter) sendWebhook(alert *models.Alert) error {
	if a.cfg.WebhookURL == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	payload := AlertPayload{
		AgentID:   alert.AgentID,
		Level:     alert.Level,
		Type:      alert.Type,
		Score:     alert.Score,
		Threshold: alert.Threshold,
		Message:   alert.Message,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(a.cfg.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("webhook returned status: %d", resp.StatusCode)
	}

	log.Printf("[Alerter] Webhook sent successfully to %s", a.cfg.WebhookURL)
	return nil
}

// logAlert 打印告警日志
func (a *Alerter) logAlert(alert *models.Alert) {
	levelEmoji := "ℹ️"
	switch alert.Level {
	case "warning":
		levelEmoji = "⚠️"
	case "critical":
		levelEmoji = "🚨"
	}

	log.Printf("%s [ALERT] Agent=%s Level=%s Type=%s Score=%.1f Message=%s",
		levelEmoji, alert.AgentID, alert.Level, alert.Type,
		alert.Score, alert.Message)
}

// Acknowledge 确认告警
func (a *Alerter) Acknowledge(alertID uint, ackedBy string) error {
	now := time.Now()
	return a.db.Model(&models.Alert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"acked":    true,
			"acked_by": ackedBy,
			"acked_at": now,
		}).Error
}

// Resolve 解决告警
func (a *Alerter) Resolve(alertID uint) error {
	now := time.Now()
	return a.db.Model(&models.Alert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"resolved":    true,
			"resolved_at": now,
		}).Error
}

// GetUnackedAlerts 获取未确认告警
func (a *Alerter) GetUnackedAlerts(agentID string) ([]models.Alert, error) {
	var alerts []models.Alert
	query := a.db.Where("acked = ?", false)
	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}
	err := query.Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

// GetStats 获取告警统计
func (a *Alerter) GetStats() (map[string]interface{}, error) {
	var total, acked, resolved, critical, warning int64

	a.db.Model(&models.Alert{}).Count(&total)
	a.db.Model(&models.Alert{}).Where("acked = ?", true).Count(&acked)
	a.db.Model(&models.Alert{}).Where("resolved = ?", true).Count(&resolved)
	a.db.Model(&models.Alert{}).Where("level = ?", "critical").Count(&critical)
	a.db.Model(&models.Alert{}).Where("level = ?", "warning").Count(&warning)

	return map[string]interface{}{
		"total":     total,
		"acked":     acked,
		"resolved":  resolved,
		"critical":  critical,
		"warning":   warning,
		"unacked":   total - acked,
		"unresolved": total - resolved,
	}, nil
}

// CleanupOldAlerts 清理旧告警 (超过 30 天已解决的)
func (a *Alerter) CleanupOldAlerts() (int64, error) {
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	result := a.db.Where("resolved = ? AND created_at < ?", true, cutoff).
		Delete(&models.Alert{})
	return result.RowsAffected, result.Error
}
