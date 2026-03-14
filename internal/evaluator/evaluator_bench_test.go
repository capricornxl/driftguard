package evaluator

import (
	"testing"
	"time"

	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupBenchmarkDB creates an in-memory SQLite database for benchmarks
func setupBenchmarkDB(b *testing.B) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}

	err = models.AutoMigrate(db)
	if err != nil {
		b.Fatal(err)
	}

	return db
}

// BenchmarkEvaluateLatency benchmarks the latency evaluation function
func BenchmarkEvaluateLatency(b *testing.B) {
	cfg := &config.EvaluatorConfig{
		HealthScoreWeights: config.HealthScoreWeights{
			Latency:     0.15,
			Efficiency:  0.10,
			Consistency: 0.30,
			Accuracy:    0.35,
			Hallucination: 0.10,
		},
	}

	db := setupBenchmarkDB(b)
	eval := NewEvaluator(cfg, db, nil)

	// Create test data
	responses := make([]models.AgentResponse, 100)
	for i := 0; i < 100; i++ {
		responses[i] = models.AgentResponse{
			AgentID:     "bench-agent",
			Input:       "test input",
			Output:      "test output",
			LatencyMs:   100 + i,
			TokenUsage:  50,
			Timestamp:   time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eval.evaluateLatency(responses)
	}
}

// BenchmarkEvaluateEfficiency benchmarks the efficiency evaluation function
func BenchmarkEvaluateEfficiency(b *testing.B) {
	cfg := &config.EvaluatorConfig{
		HealthScoreWeights: config.HealthScoreWeights{
			Latency:     0.15,
			Efficiency:  0.10,
			Consistency: 0.30,
			Accuracy:    0.35,
			Hallucination: 0.10,
		},
	}

	db := setupBenchmarkDB(b)
	eval := NewEvaluator(cfg, db, nil)

	responses := make([]models.AgentResponse, 100)
	for i := 0; i < 100; i++ {
		responses[i] = models.AgentResponse{
			AgentID:     "bench-agent",
			Input:       "test input",
			Output:      "test output",
			LatencyMs:   100,
			TokenUsage:  50 + i,
			Timestamp:   time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eval.evaluateEfficiency(responses)
	}
}

// BenchmarkEvaluateConsistency benchmarks the consistency evaluation function
func BenchmarkEvaluateConsistency(b *testing.B) {
	cfg := &config.EvaluatorConfig{
		HealthScoreWeights: config.HealthScoreWeights{
			Latency:     0.15,
			Efficiency:  0.10,
			Consistency: 0.30,
			Accuracy:    0.35,
			Hallucination: 0.10,
		},
	}

	db := setupBenchmarkDB(b)
	eval := NewEvaluator(cfg, db, nil)

	responses := make([]models.AgentResponse, 100)
	for i := 0; i < 100; i++ {
		responses[i] = models.AgentResponse{
			AgentID:     "bench-agent",
			Input:       "same input",
			Output:      "similar output " + string(rune(i)),
			LatencyMs:   100,
			TokenUsage:  50,
			Timestamp:   time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eval.evaluateConsistency(responses)
	}
}

// BenchmarkEvaluateAccuracy benchmarks the accuracy evaluation function
func BenchmarkEvaluateAccuracy(b *testing.B) {
	cfg := &config.EvaluatorConfig{
		HealthScoreWeights: config.HealthScoreWeights{
			Latency:     0.15,
			Efficiency:  0.10,
			Consistency: 0.30,
			Accuracy:    0.35,
			Hallucination: 0.10,
		},
	}

	db := setupBenchmarkDB(b)
	eval := NewEvaluator(cfg, db, nil)

	responses := make([]models.AgentResponse, 100)
	for i := 0; i < 100; i++ {
		responses[i] = models.AgentResponse{
			AgentID:      "bench-agent",
			Input:        "test input",
			Output:       "test output",
			ExpectedOutput: "expected output",
			LatencyMs:    100,
			TokenUsage:   50,
			Timestamp:    time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eval.evaluateAccuracy(responses)
	}
}

// BenchmarkEvaluateHallucination benchmarks the hallucination evaluation function
func BenchmarkEvaluateHallucination(b *testing.B) {
	cfg := &config.EvaluatorConfig{
		HealthScoreWeights: config.HealthScoreWeights{
			Latency:     0.15,
			Efficiency:  0.10,
			Consistency: 0.30,
			Accuracy:    0.35,
			Hallucination: 0.10,
		},
	}

	db := setupBenchmarkDB(b)
	eval := NewEvaluator(cfg, db, nil)

	responses := make([]models.AgentResponse, 100)
	for i := 0; i < 100; i++ {
		responses[i] = models.AgentResponse{
			AgentID:     "bench-agent",
			Input:       "test input",
			Output:      "test output with possible hallucination",
			Context:     "relevant context",
			LatencyMs:   100,
			TokenUsage:  50,
			Timestamp:   time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eval.evaluateHallucination(responses)
	}
}

// BenchmarkFullEvaluation benchmarks the complete evaluation pipeline
func BenchmarkFullEvaluation(b *testing.B) {
	cfg := &config.EvaluatorConfig{
		HealthScoreWeights: config.HealthScoreWeights{
			Latency:     0.15,
			Efficiency:  0.10,
			Consistency: 0.30,
			Accuracy:    0.35,
			Hallucination: 0.10,
		},
	}

	db := setupBenchmarkDB(b)
	eval := NewEvaluator(cfg, db, nil)

	responses := make([]models.AgentResponse, 100)
	for i := 0; i < 100; i++ {
		responses[i] = models.AgentResponse{
			AgentID:      "bench-agent",
			Input:        "test input",
			Output:       "test output",
			ExpectedOutput: "expected output",
			Context:      "relevant context",
			LatencyMs:    100 + i,
			TokenUsage:   50 + i,
			Timestamp:    time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eval.Evaluate(responses)
	}
}

// BenchmarkCalculateHealthScore benchmarks the health score calculation
func BenchmarkCalculateHealthScore(b *testing.B) {
	cfg := &config.EvaluatorConfig{
		HealthScoreWeights: config.HealthScoreWeights{
			Latency:     0.15,
			Efficiency:  0.10,
			Consistency: 0.30,
			Accuracy:    0.35,
			Hallucination: 0.10,
		},
	}

	db := setupBenchmarkDB(b)
	eval := NewEvaluator(cfg, db, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eval.calculateHealthScore(90.0, 85.0, 88.0, 82.0, 5.0)
	}
}
