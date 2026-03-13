package collector

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
	
	if err := db.AutoMigrate(&models.Interaction{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
	
	return db
}

func TestCollectorReceive(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.CollectorConfig{
		BatchSize:     10,
		FlushInterval: 60,
	}
	
	collector := NewCollector(cfg, db)
	
	// 测试接收单个交互
	interaction := models.Interaction{
		AgentID:   "test-agent",
		SessionID: "session-1",
		Input:     "test input",
		Output:    "test output",
		LatencyMs: 250,
		TokensIn:  10,
		TokensOut: 20,
	}
	
	err := collector.Receive(interaction)
	if err != nil {
		t.Errorf("Receive failed: %v", err)
	}
	
	// 验证 buffer 中有数据
	collector.mu.Lock()
	bufferSize := len(collector.buffer)
	collector.mu.Unlock()
	
	if bufferSize != 1 {
		t.Errorf("Expected buffer size 1, got %d", bufferSize)
	}
}

func TestCollectorFlush(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.CollectorConfig{
		BatchSize:     5,
		FlushInterval: 1, // 1 秒刷新
	}
	
	collector := NewCollector(cfg, db)
	collector.Start()
	defer collector.Stop()
	
	// 发送超过 batch size 的数据
	for i := 0; i < 10; i++ {
		interaction := models.Interaction{
			AgentID:   "test-agent",
			SessionID: "session-1",
			Input:     "test input",
			Output:    "test output",
			LatencyMs: 250,
			TokensIn:  10,
			TokensOut: 20,
		}
		collector.Receive(interaction)
	}
	
	// 等待刷新
	time.Sleep(2 * time.Second)
	
	// 验证数据库中有数据
	var count int64
	db.Model(&models.Interaction{}).Count(&count)
	
	if count < 10 {
		t.Errorf("Expected at least 10 interactions in DB, got %d", count)
	}
}

func TestCollectorStats(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.CollectorConfig{
		BatchSize:     10,
		FlushInterval: 60,
	}
	
	collector := NewCollector(cfg, db)
	
	// 添加一些数据
	for i := 0; i < 5; i++ {
		interaction := models.Interaction{
			AgentID:   "test-agent",
			SessionID: "session-1",
			Input:     "test input",
			Output:    "test output",
			LatencyMs: 250,
			TokensIn:  10,
			TokensOut: 20,
		}
		collector.Receive(interaction)
	}
	
	stats := collector.Stats()
	
	if stats["buffer_size"].(int) != 5 {
		t.Errorf("Expected buffer_size 5, got %v", stats["buffer_size"])
	}
	
	if stats["batch_size"].(int) != 10 {
		t.Errorf("Expected batch_size 10, got %v", stats["batch_size"])
	}
	
	t.Logf("Stats: %v", stats)
}

func TestCollectorSidecarHandler(t *testing.T) {
	// 这个测试需要 HTTP 服务器，暂时跳过
	t.Skip("Sidecar handler test requires HTTP server")
}
