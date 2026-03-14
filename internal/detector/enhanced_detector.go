package detector

import (
	"fmt"
	"time"

	"github.com/driftguard/driftguard/internal/stats"
	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// EnhancedDetector extends the base detector with statistical tests
type EnhancedDetector struct {
	config     *config.DetectorConfig
	db         *gorm.DB
	logger     *logrus.Logger
	windowSize int
}

// NewEnhancedDetector creates a new enhanced detector with statistical tests
func NewEnhancedDetector(cfg *config.DetectorConfig, db *gorm.DB, logger *logrus.Logger) *EnhancedDetector {
	windowSize := cfg.TrendWindowSize
	if windowSize == 0 {
		windowSize = 10
	}

	return &EnhancedDetector{
		config:     cfg,
		db:         db,
		logger:     logger,
		windowSize: windowSize,
	}
}

// DetectDriftWithKS performs drift detection using KS test
func (d *EnhancedDetector) DetectDriftWithKS(agentID string, baselineWindow, currentWindow int) (*stats.KSTestResult, error) {
	// Get baseline health scores
	baselineScores, err := d.getRecentHealthScores(agentID, baselineWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to get baseline scores: %w", err)
	}

	// Get current health scores
	currentScores, err := d.getRecentHealthScores(agentID, currentWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to get current scores: %w", err)
	}

	if len(baselineScores) < 5 || len(currentScores) < 5 {
		d.logger.Warn("Insufficient data for KS test")
		return &stats.KSTestResult{
			Statistic:   0,
			PValue:      1,
			Significant: false,
		}, nil
	}

	// Perform KS test
	result := stats.KSTest(baselineScores, currentScores)

	d.logger.WithFields(logrus.Fields{
		"agent_id":  agentID,
		"ks_stat":   result.Statistic,
		"p_value":   result.PValue,
		"significant": result.Significant,
	}).Info("KS test completed")

	return &result, nil
}

// DetectDriftWithPSI performs drift detection using PSI
func (d *EnhancedDetector) DetectDriftWithPSI(agentID string, baselineDays, currentDays int, numBuckets int) (*stats.PSICalculation, error) {
	// Get baseline health scores
	baselineEnd := time.Now().AddDate(0, 0, -currentDays)
	baselineStart := baselineEnd.AddDate(0, 0, -baselineDays)

	baselineScores, err := d.getHealthScoresInRange(agentID, baselineStart, baselineEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get baseline scores: %w", err)
	}

	// Get current health scores
	currentScores, err := d.getHealthScoresInRange(agentID, baselineEnd, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get current scores: %w", err)
	}

	if len(baselineScores) < 10 || len(currentScores) < 10 {
		d.logger.Warn("Insufficient data for PSI calculation")
		return &stats.PSICalculation{
			PSI:       0,
			Stability: "unknown",
			Details:   []stats.BucketDetail{},
		}, nil
	}

	// Calculate PSI
	result := stats.CalculatePSI(baselineScores, currentScores, numBuckets)

	d.logger.WithFields(logrus.Fields{
		"agent_id":  agentID,
		"psi":       result.PSI,
		"stability": result.Stability,
	}).Info("PSI calculation completed")

	return &result, nil
}

// AnalyzeHealthTrend analyzes the trend of health scores
func (d *EnhancedDetector) AnalyzeHealthTrend(agentID string, windowDays int) (*stats.TrendAnalysis, error) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -windowDays)

	scores, err := d.getHealthScoresInRange(agentID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get health scores: %w", err)
	}

	if len(scores) < 3 {
		return &stats.TrendAnalysis{
			Slope:       0,
			Direction:   "unknown",
			Strength:    0,
			Significant: false,
		}, nil
	}

	result := stats.AnalyzeTrend(scores)

	d.logger.WithFields(logrus.Fields{
		"agent_id":  agentID,
		"slope":     result.Slope,
		"direction": result.Direction,
		"strength":  result.Strength,
	}).Info("Trend analysis completed")

	return &result, nil
}

// DetectSpikes detects sudden spikes in health scores
func (d *EnhancedDetector) DetectSpikes(agentID string, windowDays int, threshold float64) (*stats.SpikeResult, error) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -windowDays)

	scores, err := d.getHealthScoresInRange(agentID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get health scores: %w", err)
	}

	if len(scores) < 3 {
		return &stats.SpikeResult{
			HasSpikes: false,
			Spikes:    []int{},
			Threshold: threshold,
		}, nil
	}

	result := stats.DetectSpikes(scores, threshold)

	d.logger.WithFields(logrus.Fields{
		"agent_id":  agentID,
		"has_spikes": result.HasSpikes,
		"spike_count": len(result.Spikes),
		"mean":      result.Mean,
		"std_dev":   result.StdDev,
	}).Info("Spike detection completed")

	return &result, nil
}

// ComprehensiveDriftReport provides a complete drift analysis
type ComprehensiveDriftReport struct {
	AgentID      string
	AnalysisTime time.Time
	KSTest       *stats.KSTestResult
	PSI          *stats.PSICalculation
	Trend        *stats.TrendAnalysis
	Spikes       *stats.SpikeResult
	OverallRisk  string // "low", "medium", "high"
	Recommendations []string
}

// GenerateComprehensiveReport generates a comprehensive drift analysis report
func (d *EnhancedDetector) GenerateComprehensiveReport(agentID string) (*ComprehensiveDriftReport, error) {
	report := &ComprehensiveDriftReport{
		AgentID:      agentID,
		AnalysisTime: time.Now(),
		Recommendations: []string{},
	}

	// KS Test (baseline: 7 days, current: 3 days)
	ksResult, err := d.DetectDriftWithKS(agentID, 7, 3)
	if err != nil {
		d.logger.WithError(err).Warn("KS test failed")
	} else {
		report.KSTest = ksResult
	}

	// PSI (baseline: 7 days, current: 7 days)
	psiResult, err := d.DetectDriftWithPSI(agentID, 7, 7, 5)
	if err != nil {
		d.logger.WithError(err).Warn("PSI calculation failed")
	} else {
		report.PSI = psiResult
	}

	// Trend Analysis (14 days)
	trendResult, err := d.AnalyzeHealthTrend(agentID, 14)
	if err != nil {
		d.logger.WithError(err).Warn("Trend analysis failed")
	} else {
		report.Trend = trendResult
	}

	// Spike Detection (7 days)
	spikeResult, err := d.DetectSpikes(agentID, 7, 2.5)
	if err != nil {
		d.logger.WithError(err).Warn("Spike detection failed")
	} else {
		report.Spikes = spikeResult
	}

	// Determine overall risk
	report.OverallRisk = d.calculateOverallRisk(report)

	// Generate recommendations
	report.Recommendations = d.generateRecommendations(report)

	return report, nil
}

// Helper methods

func (d *EnhancedDetector) getRecentHealthScores(agentID string, limit int) ([]float64, error) {
	var scores []models.HealthScore
	err := d.db.Where("agent_id = ?", agentID).
		Order("created_at DESC").
		Limit(limit).
		Find(&scores).Error

	if err != nil {
		return nil, err
	}

	result := make([]float64, len(scores))
	for i, score := range scores {
		result[i] = score.Score
	}

	return result, nil
}

func (d *EnhancedDetector) getHealthScoresInRange(agentID string, start, end time.Time) ([]float64, error) {
	var scores []models.HealthScore
	err := d.db.Where("agent_id = ? AND created_at BETWEEN ? AND ?", agentID, start, end).
		Order("created_at ASC").
		Find(&scores).Error

	if err != nil {
		return nil, err
	}

	result := make([]float64, len(scores))
	for i, score := range scores {
		result[i] = score.Score
	}

	return result, nil
}

func (d *EnhancedDetector) calculateOverallRisk(report *ComprehensiveDriftReport) string {
	riskScore := 0

	// KS Test contribution
	if report.KSTest != nil && report.KSTest.Significant {
		riskScore += 2
		if report.KSTest.Statistic > 0.5 {
			riskScore += 1
		}
	}

	// PSI contribution
	if report.PSI != nil {
		if report.PSI.Stability == "significant" {
			riskScore += 3
		} else if report.PSI.Stability == "moderate" {
			riskScore += 1
		}
	}

	// Trend contribution
	if report.Trend != nil && report.Trend.Significant {
		if report.Trend.Direction == "decreasing" {
			riskScore += 2
		}
	}

	// Spikes contribution
	if report.Spikes != nil && report.Spikes.HasSpikes {
		riskScore += 1
	}

	// Determine risk level
	if riskScore >= 5 {
		return "high"
	} else if riskScore >= 3 {
		return "medium"
	}
	return "low"
}

func (d *EnhancedDetector) generateRecommendations(report *ComprehensiveDriftReport) []string {
	recommendations := []string{}

	if report.KSTest != nil && report.KSTest.Significant {
		recommendations = append(recommendations,
			"Significant distribution drift detected. Investigate recent changes to agent configuration or model.")
	}

	if report.PSI != nil && report.PSI.Stability == "significant" {
		recommendations = append(recommendations,
			fmt.Sprintf("PSI of %.3f indicates significant population shift. Review input data distribution.", report.PSI.PSI))
	}

	if report.Trend != nil && report.Trend.Direction == "decreasing" && report.Trend.Significant {
		recommendations = append(recommendations,
			fmt.Sprintf("Health score declining trend detected (slope: %.3f). Consider model retraining or configuration adjustment.", report.Trend.Slope))
	}

	if report.Spikes != nil && report.Spikes.HasSpikes {
		recommendations = append(recommendations,
			fmt.Sprintf("Detected %d anomalous spikes. Review logs around spike timestamps for root cause.", len(report.Spikes.Spikes)))
	}

	if report.OverallRisk == "high" {
		recommendations = append(recommendations,
			"HIGH RISK: Immediate attention required. Consider rolling back recent changes or increasing monitoring frequency.")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "No significant issues detected. Continue regular monitoring.")
	}

	return recommendations
}
