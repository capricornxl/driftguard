package detector

import (
	"testing"
	"time"

	"github.com/driftguard/driftguard/internal/stats"
	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/models"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/stretchr/testify/assert"
)

func setupEnhancedDetectorDB(t *testing.T) (*gorm.DB, *EnhancedDetector) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = models.AutoMigrate(db)
	assert.NoError(t, err)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Suppress logs during tests

	cfg := &config.DetectorConfig{
		Enabled:         true,
		CheckInterval:   "5m",
		TrendWindowSize: 10,
	}

	detector := NewEnhancedDetector(cfg, db, logger)

	// Seed test data
	seedTestData(db, "test-agent")

	return db, detector
}

func seedTestData(db *gorm.DB, agentID string) {
	// Create agent
	agent := models.Agent{
		ID:        agentID,
		Name:      "Test Agent",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(&agent)

	// Create historical health scores (14 days)
	now := time.Now()
	for i := 0; i < 14; i++ {
		score := 85.0 - float64(i)*0.5 // Slight declining trend
		if i >= 10 {
			score = 70.0 // Sharp decline in last 4 days
		}

		healthScore := models.HealthScore{
			AgentID:     agentID,
			Score:       score,
			Latency:     90.0,
			Efficiency:  85.0,
			Consistency: 88.0,
			Accuracy:    82.0,
			Hallucination: 5.0,
			Trend:       "stable",
			CreatedAt:   now.AddDate(0, 0, -i),
		}
		db.Create(&healthScore)
	}
}

func TestNewEnhancedDetector(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	logger := logrus.New()
	cfg := &config.DetectorConfig{
		Enabled:       true,
		CheckInterval: "5m",
	}

	detector := NewEnhancedDetector(cfg, db, logger)

	assert.NotNil(t, detector)
	assert.Equal(t, 10, detector.windowSize) // Default window size
}

func TestDetectDriftWithKS(t *testing.T) {
	db, detector := setupEnhancedDetectorDB(t)

	// Add more recent data with significant drift
	now := time.Now()
	for i := 0; i < 5; i++ {
		healthScore := models.HealthScore{
			AgentID:     "test-agent",
			Score:       50.0, // Significant drop
			CreatedAt:   now.AddDate(0, 0, -i),
		}
		db.Create(&healthScore)
	}

	result, err := detector.DetectDriftWithKS("test-agent", 10, 5)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDetectDriftWithPSI(t *testing.T) {
	_, detector := setupEnhancedDetectorDB(t)

	result, err := detector.DetectDriftWithPSI("test-agent", 7, 7, 5)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, []string{"stable", "moderate", "significant", "unknown"}, result.Stability)
}

func TestAnalyzeHealthTrend(t *testing.T) {
	_, detector := setupEnhancedDetectorDB(t)

	result, err := detector.AnalyzeHealthTrend("test-agent", 14)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, []string{"increasing", "decreasing", "stable", "unknown"}, result.Direction)
}

func TestDetectSpikes(t *testing.T) {
	db, detector := setupEnhancedDetectorDB(t)

	// Add a spike
	healthScore := models.HealthScore{
		AgentID:   "test-agent",
		Score:     20.0, // Extreme low
		CreatedAt: time.Now(),
	}
	db.Create(&healthScore)

	result, err := detector.DetectSpikes("test-agent", 7, 2.0)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestGenerateComprehensiveReport(t *testing.T) {
	_, detector := setupEnhancedDetectorDB(t)

	report, err := detector.GenerateComprehensiveReport("test-agent")
	assert.NoError(t, err)
	assert.NotNil(t, report)

	assert.Equal(t, "test-agent", report.AgentID)
	assert.NotNil(t, report.KSTest)
	assert.NotNil(t, report.PSI)
	assert.NotNil(t, report.Trend)
	assert.NotNil(t, report.Spikes)
	assert.Contains(t, []string{"low", "medium", "high"}, report.OverallRisk)
	assert.NotEmpty(t, report.Recommendations)
}

func TestCalculateOverallRisk(t *testing.T) {
	_, detector := setupEnhancedDetectorDB(t)

	t.Run("low risk", func(t *testing.T) {
		report := &ComprehensiveDriftReport{
			KSTest: &stats.KSTestResult{Significant: false},
			PSI:    &stats.PSICalculation{Stability: "stable"},
			Trend:  &stats.TrendAnalysis{Direction: "stable"},
			Spikes: &stats.SpikeResult{HasSpikes: false},
		}

		risk := detector.calculateOverallRisk(report)
		assert.Equal(t, "low", risk)
	})

	t.Run("high risk", func(t *testing.T) {
		report := &ComprehensiveDriftReport{
			KSTest: &stats.KSTestResult{Significant: true, Statistic: 0.6},
			PSI:    &stats.PSICalculation{Stability: "significant"},
			Trend:  &stats.TrendAnalysis{Direction: "decreasing", Significant: true},
			Spikes: &stats.SpikeResult{HasSpikes: true},
		}

		risk := detector.calculateOverallRisk(report)
		assert.Equal(t, "high", risk)
	})
}

func TestGenerateRecommendations(t *testing.T) {
	_, detector := setupEnhancedDetectorDB(t)

	t.Run("no issues", func(t *testing.T) {
		report := &ComprehensiveDriftReport{
			KSTest: &stats.KSTestResult{Significant: false},
			PSI:    &stats.PSICalculation{Stability: "stable"},
			Trend:  &stats.TrendAnalysis{Direction: "stable"},
			Spikes: &stats.SpikeResult{HasSpikes: false},
		}

		recs := detector.generateRecommendations(report)
		assert.NotEmpty(t, recs)
		assert.Contains(t, recs[0], "No significant issues")
	})

	t.Run("multiple issues", func(t *testing.T) {
		report := &ComprehensiveDriftReport{
			KSTest: &stats.KSTestResult{Significant: true, Statistic: 0.6},
			PSI:    &stats.PSICalculation{Stability: "significant", PSI: 0.25},
			Trend:  &stats.TrendAnalysis{Direction: "decreasing", Significant: true, Slope: -0.5},
			Spikes: &stats.SpikeResult{HasSpikes: true, Spikes: []int{1, 2}},
		}

		recs := detector.generateRecommendations(report)
		assert.Greater(t, len(recs), 1)
	})
}

func TestGetRecentHealthScores(t *testing.T) {
	db, detector := setupEnhancedDetectorDB(t)

	scores, err := detector.getRecentHealthScores("test-agent", 5)
	assert.NoError(t, err)
	assert.Len(t, scores, 5)
}

func TestGetHealthScoresInRange(t *testing.T) {
	_, detector := setupEnhancedDetectorDB(t)

	start := time.Now().AddDate(0, 0, -7)
	end := time.Now()

	scores, err := detector.getHealthScoresInRange("test-agent", start, end)
	assert.NoError(t, err)
	assert.Greater(t, len(scores), 0)
}
