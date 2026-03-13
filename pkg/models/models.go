package models

import (
	"time"
)

// Interaction 代表一次 Agent 交互记录
type Interaction struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AgentID       string    `gorm:"index" json:"agent_id"`
	SessionID     string    `json:"session_id"`
	Input         string    `gorm:"type:text" json:"input"`
	Output        string    `gorm:"type:text" json:"output"`
	LatencyMs     int       `json:"latency_ms"`
	TokensIn      int       `json:"tokens_in"`
	TokensOut     int       `json:"tokens_out"`
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
	Metadata      string    `gorm:"type:json" json:"metadata,omitempty"`
	HealthScore   float64   `gorm:"index" json:"health_score"`
	IsDegraded    bool      `json:"is_degraded"`
	DegradedType  string    `json:"degraded_type,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// HealthScore 健康度评分记录
type HealthScore struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AgentID     string    `gorm:"index" json:"agent_id"`
	Score       float64   `json:"score"`
	LatencyScore    float64 `json:"latency_score"`
	EfficiencyScore float64 `json:"efficiency_score"`
	ConsistencyScore float64 `json:"consistency_score"`
	AccuracyScore    float64 `json:"accuracy_score"`
	HallucinationScore float64 `json:"hallucination_score"`
	WindowSize  int       `json:"window_size"` // 统计窗口大小 (交互次数)
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedAt   time.Time `json:"created_at"`
}

// Alert 告警记录
type Alert struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AgentID     string    `gorm:"index" json:"agent_id"`
	Level       string    `json:"level"` // info, warning, critical
	Type        string    `json:"type"`  // threshold, trend, anomaly, spike
	Score       float64   `json:"score"`
	Threshold   float64   `json:"threshold"`
	Message     string    `gorm:"type:text" json:"message"`
	Context     string    `gorm:"type:json" json:"context,omitempty"`
	Acked       bool      `json:"acked"`
	AckedBy     string    `json:"acked_by,omitempty"`
	AckedAt     *time.Time `json:"acked_at,omitempty"`
	Resolved    bool      `json:"resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
}

// AgentConfig Agent 配置
type AgentConfig struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	AgentID         string    `gorm:"uniqueIndex" json:"agent_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Enabled         bool      `json:"enabled"`
	ThresholdWarning  float64   `json:"threshold_warning"`  // 警告阈值
	ThresholdCritical float64   `json:"threshold_critical"` // 严重阈值
	CheckInterval   int       `json:"check_interval"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Interaction) TableName() string {
	return "interactions"
}

func (HealthScore) TableName() string {
	return "health_scores"
}

func (Alert) TableName() string {
	return "alerts"
}

func (AgentConfig) TableName() string {
	return "agent_configs"
}
