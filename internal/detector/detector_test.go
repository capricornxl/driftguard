package detector

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
	
	if err := db.AutoMigrate(&models.HealthScore{}, &models.Alert{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
	
	return db
}

func TestDetectThreshold(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.DetectorConfig{
		CheckInterval: 300,
		WindowDays:    7,
		Threshold:     70.0,
	}
	
	det := NewDetector(cfg, db, nil)
	
	// 测试严重阈值 (<50)
	scores := []models.HealthScore{
		{Score: 60},
		{Score: 55},
		{Score: 45}, // 低于 50
	}
	
	alert := det.detectThreshold(scores)
	if alert == nil {
		t.Error("Expected critical alert for score<50")
	} else if alert.Level != "critical" {
		t.Errorf("Expected critical level, got %s", alert.Level)
	}
	
	// 测试警告阈值 (<70)
	scores2 := []models.HealthScore{
		{Score: 80},
		{Score: 75},
		{Score: 65}, // 低于 70
	}
	
	alert2 := det.detectThreshold(scores2)
	if alert2 == nil {
		t.Error("Expected warning alert for score<70")
	} else if alert2.Level != "warning" {
		t.Errorf("Expected warning level, got %s", alert2.Level)
	}
	
	// 测试正常分数 (>=85)
	scores3 := []models.HealthScore{
		{Score: 90},
		{Score: 88},
		{Score: 92},
	}
	
	alert3 := det.detectThreshold(scores3)
	if alert3 != nil {
		t.Error("Expected no alert for healthy scores")
	}
}

func TestDetectTrend(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.DetectorConfig{}
	det := NewDetector(cfg, db, nil)
	
	// 测试下降趋势
	scores := []models.HealthScore{
		{Score: 90},
		{Score: 85},
		{Score: 80},
		{Score: 75},
		{Score: 70},
	}
	
	alert, trend, slope := det.detectTrend(scores)
	if trend != "degrading" {
		t.Errorf("Expected degrading trend, got %s", trend)
	}
	if slope >= 0 {
		t.Errorf("Expected negative slope, got %.2f", slope)
	}
	if alert == nil {
		t.Error("Expected alert for strong degrading trend")
	}
	
	// 测试上升趋势
	scores2 := []models.HealthScore{
		{Score: 70},
		{Score: 75},
		{Score: 80},
		{Score: 85},
		{Score: 90},
	}
	
	_, trend2, slope2 := det.detectTrend(scores2)
	if trend2 != "improving" {
		t.Errorf("Expected improving trend, got %s", trend2)
	}
	if slope2 <= 0 {
		t.Errorf("Expected positive slope, got %.2f", slope2)
	}
	
	// 测试稳定趋势
	scores3 := []models.HealthScore{
		{Score: 85},
		{Score: 86},
		{Score: 85},
		{Score: 84},
		{Score: 85},
	}
	
	_, trend3, slope3 := det.detectTrend(scores3)
	if trend3 != "stable" {
		t.Errorf("Expected stable trend, got %s", trend3)
	}
	if slope3 < -0.5 || slope3 > 0.5 {
		t.Errorf("Expected near-zero slope, got %.2f", slope3)
	}
}

func TestDetectAnomaly(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.DetectorConfig{}
	det := NewDetector(cfg, db, nil)
	
	// 测试正常波动 (无异常)
	scores := []models.HealthScore{
		{Score: 85},
		{Score: 86},
		{Score: 84},
		{Score: 87},
		{Score: 85},
	}
	
	alert := det.detectAnomaly(scores)
	if alert != nil {
		t.Error("Expected no alert for normal fluctuation")
	}
	
	// 测试异常下降
	scores2 := []models.HealthScore{
		{Score: 90},
		{Score: 88},
		{Score: 92},
		{Score: 89},
		{Score: 50}, // 异常低
	}
	
	alert2 := det.detectAnomaly(scores2)
	if alert2 == nil {
		t.Error("Expected anomaly alert for outlier")
	}
}

func TestDetectSpike(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.DetectorConfig{}
	det := NewDetector(cfg, db, nil)
	
	// 测试正常变化
	scores := []models.HealthScore{
		{Score: 85},
		{Score: 83},
	}
	
	alert := det.detectSpike(scores)
	if alert != nil {
		t.Error("Expected no alert for normal change")
	}
	
	// 测试骤降 (>15 分)
	scores2 := []models.HealthScore{
		{Score: 90},
		{Score: 70}, // 下降 20 分
	}
	
	alert2 := det.detectSpike(scores2)
	if alert2 == nil {
		t.Error("Expected spike alert for sudden drop")
	} else if alert2.Level != "critical" {
		t.Errorf("Expected critical level, got %s", alert2.Level)
	}
}

func TestCalculateConfidence(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.DetectorConfig{}
	det := NewDetector(cfg, db, nil)
	
	// 测试无告警
	alerts := []AlertInfo{}
	confidence := det.calculateConfidence(alerts)
	if confidence != 0 {
		t.Errorf("Expected 0 confidence for no alerts, got %.2f", confidence)
	}
	
	// 测试单一警告
	alerts2 := []AlertInfo{
		{Level: "warning"},
	}
	confidence2 := det.calculateConfidence(alerts2)
	if confidence2 < 20 || confidence2 > 40 {
		t.Errorf("Expected ~20 confidence for single warning, got %.2f", confidence2)
	}
	
	// 测试严重告警
	alerts3 := []AlertInfo{
		{Level: "critical"},
	}
	confidence3 := det.calculateConfidence(alerts3)
	if confidence3 < 40 {
		t.Errorf("Expected >=40 confidence for critical alert, got %.2f", confidence3)
	}
	
	// 测试多重告警
	alerts4 := []AlertInfo{
		{Level: "critical"},
		{Level: "warning"},
		{Level: "warning"},
	}
	confidence4 := det.calculateConfidence(alerts4)
	if confidence4 < 80 {
		t.Errorf("Expected high confidence for multiple alerts, got %.2f", confidence4)
	}
}

func TestDetectorFull(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.DetectorConfig{
		CheckInterval: 300,
		WindowDays:    7,
		Threshold:     70.0,
	}
	
	det := NewDetector(cfg, db, nil)
	
	// 插入测试数据 (持续下降)
	now := time.Now()
	for i := 0; i < 10; i++ {
		score := 90 - float64(i)*5 // 90, 85, 80, ..., 45
		hs := models.HealthScore{
			AgentID:   "test-agent",
			Score:     score,
			CreatedAt: now.Add(-time.Duration(9-i) * time.Hour), // 最早的分数在最前面
		}
		db.Create(&hs)
	}
	
	result, err := det.Detect("test-agent")
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	
	if !result.IsDegraded {
		t.Error("Expected degraded status for declining scores")
	}
	
	if result.Trend != "degrading" {
		t.Errorf("Expected degrading trend, got %s", result.Trend)
	}
	
	if len(result.Alerts) == 0 {
		t.Error("Expected alerts for degraded agent")
	}
	
	t.Logf("Detection Result: Degraded=%v Type=%s Trend=%s Alerts=%d Confidence=%.1f",
		result.IsDegraded, result.DegradedType, result.Trend,
		len(result.Alerts), result.Confidence)
}
