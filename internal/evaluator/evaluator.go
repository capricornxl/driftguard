package evaluator

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/metrics"
	"github.com/driftguard/driftguard/pkg/models"
	"gorm.io/gorm"
)

// Evaluator 健康度评估器
type Evaluator struct {
	cfg     *config.EvaluatorConfig
	db      *gorm.DB
	metrics *metrics.Metrics
}

// NewEvaluator 创建评估器
func NewEvaluator(cfg *config.EvaluatorConfig, db *gorm.DB, m *metrics.Metrics) *Evaluator {
	return &Evaluator{
		cfg:     cfg,
		db:      db,
		metrics: m,
	}
}

// HealthResult 健康度评估结果
type HealthResult struct {
	AgentID       string  `json:"agent_id"`
	Score         float64 `json:"score"`         // 总分 0-100
	LatencyScore  float64 `json:"latency_score"` // 延迟得分
	EfficiencyScore float64 `json:"efficiency_score"` // 效率得分
	ConsistencyScore float64 `json:"consistency_score"` // 一致性得分
	AccuracyScore    float64 `json:"accuracy_score"` // 准确性得分
	HallucinationScore float64 `json:"hallucination_score"` // 幻觉得分
	WindowSize    int     `json:"window_size"`   // 统计窗口大小
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Level         string  `json:"level"` // healthy, warning, critical
}

// Evaluate 评估 Agent 健康度
func (e *Evaluator) Evaluate(agentID string, windowHours int) (*HealthResult, error) {
	startTime := time.Now().Add(-time.Duration(windowHours) * time.Hour)

	// 获取窗口内的交互记录
	var interactions []models.Interaction
	if err := e.db.Where("agent_id = ? AND timestamp >= ?", agentID, startTime).
		Order("timestamp ASC").Find(&interactions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch interactions: %w", err)
	}

	if len(interactions) < 10 {
		return nil, fmt.Errorf("insufficient data: only %d interactions", len(interactions))
	}

	result := &HealthResult{
		AgentID:    agentID,
		WindowSize: len(interactions),
		StartTime:  startTime,
		EndTime:    time.Now(),
	}

	// 计算各维度得分
	result.LatencyScore = e.evaluateLatency(interactions)
	result.EfficiencyScore = e.evaluateEfficiency(interactions)
	result.ConsistencyScore = e.evaluateConsistency(interactions)
	result.AccuracyScore = e.evaluateAccuracy(interactions)
	result.HallucinationScore = e.evaluateHallucination(interactions)

	// 加权计算总分
	w := e.cfg.Weights
	result.Score = result.LatencyScore*w.Latency +
		result.EfficiencyScore*w.Efficiency +
		result.ConsistencyScore*w.Consistency +
		result.AccuracyScore*w.Accuracy +
		result.HallucinationScore*w.Hallucination

	// 确定健康等级
	if result.Score >= 85 {
		result.Level = "healthy"
	} else if result.Score >= 70 {
		result.Level = "warning"
	} else {
		result.Level = "critical"
	}

	return result, nil
}

// evaluateLatency 评估延迟得分 (基于 P95)
func (e *Evaluator) evaluateLatency(interactions []models.Interaction) float64 {
	if len(interactions) == 0 {
		return 100
	}

	// 排序计算 P95
	latencies := make([]int, len(interactions))
	for i, inter := range interactions {
		latencies[i] = inter.LatencyMs
	}
	sort.Ints(latencies)
	p95Index := int(float64(len(latencies)) * 0.95)
	if p95Index >= len(latencies) {
		p95Index = len(latencies) - 1
	}
	p95 := latencies[p95Index]

	// P95 < 500ms: 100 分, P95 > 3000ms: 0 分
	if p95 <= 500 {
		return 100
	}
	if p95 >= 3000 {
		return 0
	}
	return 100 - float64(p95-500)/25
}

// evaluateEfficiency 评估效率得分 (tokens/s)
func (e *Evaluator) evaluateEfficiency(interactions []models.Interaction) float64 {
	if len(interactions) == 0 {
		return 100
	}

	totalTokens := 0
	totalTime := 0.0
	for _, inter := range interactions {
		totalTokens += inter.TokensIn + inter.TokensOut
		totalTime += float64(inter.LatencyMs) / 1000
	}

	if totalTime == 0 {
		return 100
	}

	tokensPerSec := float64(totalTokens) / totalTime

	// > 100 tokens/s: 100 分, < 10 tokens/s: 0 分
	if tokensPerSec >= 100 {
		return 100
	}
	if tokensPerSec <= 10 {
		return 0
	}
	return (tokensPerSec - 10) / 0.9
}

// evaluateConsistency 评估一致性得分 (基于输出相似度)
func (e *Evaluator) evaluateConsistency(interactions []models.Interaction) float64 {
	if len(interactions) < 2 {
		return 100
	}

	// 简化：基于输入哈希分组，检查相同输入的输出差异
	// 实际应使用 embedding 相似度
	inputGroups := make(map[string][]string)
	for _, inter := range interactions {
		inputKey := fmt.Sprintf("%d", len(inter.Input)%100) // 简化哈希
		inputGroups[inputKey] = append(inputGroups[inputKey], inter.Output)
	}

	totalVariance := 0.0
	groupCount := 0
	for _, outputs := range inputGroups {
		if len(outputs) < 2 {
			continue
		}
		// 计算组内输出长度方差
		mean := 0
		for _, o := range outputs {
			mean += len(o)
		}
		mean /= len(outputs)

		variance := 0.0
		for _, o := range outputs {
			diff := float64(len(o)) - float64(mean)
			variance += diff * diff
		}
		variance /= float64(len(outputs))
		totalVariance += math.Sqrt(variance)
		groupCount++
	}

	if groupCount == 0 {
		return 100
	}

	avgVariance := totalVariance / float64(groupCount)

	// 方差越小越好
	if avgVariance <= 10 {
		return 100
	}
	if avgVariance >= 200 {
		return 0
	}
	return 100 - (avgVariance-10)/1.9
}

// evaluateAccuracy 评估准确性得分 (需要标注数据或 LLM-as-Judge)
func (e *Evaluator) evaluateAccuracy(interactions []models.Interaction) float64 {
	// 简化实现：基于输出长度和关键词检测
	// 实际应使用：1) 用户反馈 2) LLM-as-Judge 3) 规则匹配

	if len(interactions) == 0 {
		return 100
	}

	accuracyScore := 0.0
	for _, inter := range interactions {
		score := 100.0

		// 检测明显错误信号
		output := inter.Output
		if len(output) < 10 {
			score -= 30 // 输出过短
		}
		if containsErrorPatterns(output) {
			score -= 40 // 包含错误模式
		}

		accuracyScore += score
	}

	return accuracyScore / float64(len(interactions))
}

// evaluateHallucination 评估幻觉得分
func (e *Evaluator) evaluateHallucination(interactions []models.Interaction) float64 {
	if len(interactions) == 0 {
		return 100
	}

	hallucinationPenalty := 0.0
	for _, inter := range interactions {
		output := inter.Output

		// 检测幻觉信号
		if containsHallucinationPatterns(output) {
			hallucinationPenalty += 50 // 每个幻觉模式惩罚 50 分
		}
	}

	avgPenalty := hallucinationPenalty / float64(len(interactions))
	if avgPenalty >= 100 {
		return 0
	}
	return 100 - avgPenalty
}

// containsErrorPatterns 检测错误模式
func containsErrorPatterns(text string) bool {
	patterns := []string{
		"抱歉我不知道",
		"无法回答",
		"我不确定",
		"error",
		"failed",
	}
	for _, p := range patterns {
		if len(text) > 0 && contains(text, p) {
			return true
		}
	}
	return false
}

// containsHallucinationPatterns 检测幻觉模式
func containsHallucinationPatterns(text string) bool {
	patterns := []string{
		"据我所知",
		"可能",
		"也许",
		"据说",
		"unverified",
		"hypothetical",
	}
	for _, p := range patterns {
		if contains(text, p) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SaveResult 保存评估结果
func (e *Evaluator) SaveResult(result *HealthResult) error {
	score := models.HealthScore{
		AgentID:            result.AgentID,
		Score:              result.Score,
		LatencyScore:       result.LatencyScore,
		EfficiencyScore:    result.EfficiencyScore,
		ConsistencyScore:   result.ConsistencyScore,
		AccuracyScore:      result.AccuracyScore,
		HallucinationScore: result.HallucinationScore,
		WindowSize:         result.WindowSize,
		StartTime:          result.StartTime,
		EndTime:            result.EndTime,
		CreatedAt:          time.Now(),
	}

	return e.db.Create(&score).Error
}

// GetLatestScore 获取最新健康度评分
func (e *Evaluator) GetLatestScore(agentID string) (*models.HealthScore, error) {
	var score models.HealthScore
	err := e.db.Where("agent_id = ?", agentID).
		Order("created_at DESC").
		First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

// LogResult 打印评估结果并导出 Metrics
func (e *Evaluator) LogResult(result *HealthResult) {
	log.Printf("[Evaluator] Agent=%s Score=%.1f Level=%s [L:%.1f E:%.1f C:%.1f A:%.1f H:%.1f]",
		result.AgentID, result.Score, result.Level,
		result.LatencyScore, result.EfficiencyScore,
		result.ConsistencyScore, result.AccuracyScore,
		result.HallucinationScore)

	// 导出 Prometheus Metrics
	if e.metrics != nil {
		e.metrics.UpdateHealthScore(result.AgentID, result.Score, result.Level)
		e.metrics.UpdateHealthScoreDimensions(
			result.AgentID,
			result.LatencyScore,
			result.EfficiencyScore,
			result.ConsistencyScore,
			result.AccuracyScore,
			result.HallucinationScore,
		)
	}
}
