package evaluator

import (
	"testing"
	"time"

	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	
	if err := db.AutoMigrate(&models.Interaction{}, &models.HealthScore{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
	
	return db
}

func TestEvaluateLatency(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.EvaluatorConfig{
		Weights: config.HealthWeights{
			Latency:       0.15,
			Efficiency:    0.10,
			Consistency:   0.30,
			Accuracy:      0.35,
			Hallucination: 0.10,
		},
	}
	
	eval := NewEvaluator(cfg, db)
	
	// 测试 P95 < 500ms (应得 100 分)
	interactions := []models.Interaction{
		{LatencyMs: 100},
		{LatencyMs: 200},
		{LatencyMs: 300},
		{LatencyMs: 400},
		{LatencyMs: 450},
	}
	
	score := eval.evaluateLatency(interactions)
	if score != 100 {
		t.Errorf("Expected latency score 100 for P95<500ms, got %.2f", score)
	}
	
	// 测试 P95 > 3000ms (应得 0 分)
	interactions2 := []models.Interaction{
		{LatencyMs: 1000},
		{LatencyMs: 2000},
		{LatencyMs: 2500},
		{LatencyMs: 3000},
		{LatencyMs: 3500},
	}
	
	score2 := eval.evaluateLatency(interactions2)
	if score2 != 0 {
		t.Errorf("Expected latency score 0 for P95>3000ms, got %.2f", score2)
	}
}

func TestEvaluateEfficiency(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.EvaluatorConfig{}
	eval := NewEvaluator(cfg, db)
	
	// 测试高效率 (>100 tokens/s)
	interactions := []models.Interaction{
		{TokensIn: 100, TokensOut: 100, LatencyMs: 1000},
		{TokensIn: 100, TokensOut: 100, LatencyMs: 1000},
	}
	
	score := eval.evaluateEfficiency(interactions)
	if score < 100 {
		t.Errorf("Expected efficiency score >=100 for 200 tokens/s, got %.2f", score)
	}
	
	// 测试低效率 (<10 tokens/s)
	interactions2 := []models.Interaction{
		{TokensIn: 5, TokensOut: 5, LatencyMs: 2000},
		{TokensIn: 5, TokensOut: 5, LatencyMs: 2000},
	}
	
	score2 := eval.evaluateEfficiency(interactions2)
	if score2 > 0 {
		t.Errorf("Expected efficiency score 0 for <10 tokens/s, got %.2f", score2)
	}
}

func TestEvaluateConsistency(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.EvaluatorConfig{}
	eval := NewEvaluator(cfg, db)
	
	// 测试高一致性 (相同长度输出)
	interactions := []models.Interaction{
		{Input: "test", Output: "hello world"},
		{Input: "test", Output: "hello again"},
		{Input: "test", Output: "hi there"},
	}
	
	score := eval.evaluateConsistency(interactions)
	if score < 80 {
		t.Errorf("Expected high consistency score, got %.2f", score)
	}
}

func TestContainsErrorPatterns(t *testing.T) {
	tests := []struct {
		text     string
		expected bool
	}{
		{"抱歉我不知道", true},
		{"无法回答", true},
		{"error occurred", true},
		{"failed to process", true},
		{"你好，这是正常回答", false},
		{"I'm happy to help", false},
	}
	
	for _, tt := range tests {
		result := containsErrorPatterns(tt.text)
		if result != tt.expected {
			t.Errorf("containsErrorPatterns(%q) = %v, expected %v", tt.text, result, tt.expected)
		}
	}
}

func TestContainsHallucinationPatterns(t *testing.T) {
	tests := []struct {
		text     string
		expected bool
	}{
		{"据我所知", true},
		{"可能吧", true},
		{"据说", true},
		{"这是确定的事实", false},
		{"根据文档记载", false},
	}
	
	for _, tt := range tests {
		result := containsHallucinationPatterns(tt.text)
		if result != tt.expected {
			t.Errorf("containsHallucinationPatterns(%q) = %v, expected %v", tt.text, result, tt.expected)
		}
	}
}

func TestEvaluateAccuracy(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.EvaluatorConfig{}
	eval := NewEvaluator(cfg, db)
	
	// 测试正常回答 (应得高分)
	interactions := []models.Interaction{
		{Output: "这是一个详细的回答，包含了完整的信息和解释。"},
		{Output: "根据您的问题，我找到了以下答案..."},
	}
	
	score := eval.evaluateAccuracy(interactions)
	if score < 80 {
		t.Errorf("Expected high accuracy score for normal answers, got %.2f", score)
	}
	
	// 测试错误回答 (应得低分)
	interactions2 := []models.Interaction{
		{Output: "抱歉我不知道"},
		{Output: "error: failed"},
	}
	
	score2 := eval.evaluateAccuracy(interactions2)
	if score2 > 50 {
		t.Errorf("Expected low accuracy score for error answers, got %.2f", score2)
	}
}

func TestEvaluateHallucination(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.EvaluatorConfig{}
	eval := NewEvaluator(cfg, db)
	
	// 测试无幻觉 (应得高分)
	interactions := []models.Interaction{
		{Output: "根据官方文档，这个 API 的用法是..."},
		{Output: "代码示例如下：..."},
	}
	
	score := eval.evaluateHallucination(interactions)
	if score < 80 {
		t.Errorf("Expected high hallucination score for factual answers, got %.2f", score)
	}
	
	// 测试有幻觉 (应得低分)
	interactions2 := []models.Interaction{
		{Output: "据我所知，可能..."},
		{Output: "也许是这样，据说..."},
	}
	
	score2 := eval.evaluateHallucination(interactions2)
	if score2 < 50 {
		t.Errorf("Expected low hallucination score for uncertain answers, got %.2f", score2)
	}
}

func TestHealthResult(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.EvaluatorConfig{
		Weights: config.HealthWeights{
			Latency:       0.15,
			Efficiency:    0.10,
			Consistency:   0.30,
			Accuracy:      0.35,
			Hallucination: 0.10,
		},
	}
	
	eval := NewEvaluator(cfg, db)
	
	// 插入测试数据
	now := time.Now()
	for i := 0; i < 20; i++ {
		interaction := models.Interaction{
			AgentID:   "test-agent",
			SessionID: "session-1",
			Input:     "test input",
			Output:    "test output",
			LatencyMs: 300,
			TokensIn:  10,
			TokensOut: 20,
			Timestamp: now.Add(-time.Duration(i) * time.Minute),
		}
		db.Create(&interaction)
	}
	
	result, err := eval.Evaluate("test-agent", 24)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	
	if result.Score < 0 || result.Score > 100 {
		t.Errorf("Health score out of range: %.2f", result.Score)
	}
	
	if result.WindowSize != 20 {
		t.Errorf("Expected window size 20, got %d", result.WindowSize)
	}
	
	t.Logf("Health Result: Score=%.2f Level=%s", result.Score, result.Level)
}
