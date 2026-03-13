package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/metrics"
	"github.com/driftguard/driftguard/pkg/models"
	"gorm.io/gorm"
)

// Collector 数据采集器
type Collector struct {
	cfg     *config.CollectorConfig
	db      *gorm.DB
	metrics *metrics.Metrics
	buffer  []models.Interaction
	mu      sync.Mutex
	ticker  *time.Ticker
	done    chan struct{}
}

// NewCollector 创建采集器
func NewCollector(cfg *config.CollectorConfig, db *gorm.DB, m *metrics.Metrics) *Collector {
	return &Collector{
		cfg:     cfg,
		db:      db,
		metrics: m,
		buffer:  make([]models.Interaction, 0, cfg.BatchSize),
		done:    make(chan struct{}),
		ticker:  time.NewTicker(time.Duration(cfg.FlushInterval) * time.Second),
	}
}

// Start 启动采集器
func (c *Collector) Start() {
	go c.flushLoop()
	log.Printf("[Collector] Started with batch_size=%d, flush_interval=%ds", 
		c.cfg.BatchSize, c.cfg.FlushInterval)
}

// Stop 停止采集器
func (c *Collector) Stop() {
	close(c.done)
	c.ticker.Stop()
	c.flush() // 最后一次刷新
}

// Receive 接收交互数据
func (c *Collector) Receive(interaction models.Interaction) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	interaction.CreatedAt = time.Now()
	c.buffer = append(c.buffer, interaction)

	// 记录 Metrics
	if c.metrics != nil {
		c.metrics.RecordInteraction(
			interaction.AgentID,
			interaction.LatencyMs,
			interaction.TokensIn,
			interaction.TokensOut,
		)
	}

	if len(c.buffer) >= c.cfg.BatchSize {
		c.flushLocked()
	}
	return nil
}

// flushLoop 定时刷新
func (c *Collector) flushLoop() {
	for {
		select {
		case <-c.ticker.C:
			c.flush()
		case <-c.done:
			return
		}
	}
}

// flush 刷新数据到数据库
func (c *Collector) flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.flushLocked()
}

func (c *Collector) flushLocked() {
	if len(c.buffer) == 0 {
		return
	}

	// 批量插入
	if err := c.db.CreateInBatches(c.buffer, 100).Error; err != nil {
		log.Printf("[Collector] Flush error: %v", err)
		return
	}

	log.Printf("[Collector] Flushed %d interactions", len(c.buffer))
	c.buffer = c.buffer[:0]
}

// CollectFromAgent 从 Agent 主动采集数据
func (c *Collector) CollectFromAgent(agentURL string) error {
	resp, err := http.Get(agentURL)
	if err != nil {
		return fmt.Errorf("failed to fetch from agent: %w", err)
	}
	defer resp.Body.Close()

	var interactions []models.Interaction
	if err := json.NewDecoder(resp.Body).Decode(&interactions); err != nil {
		return fmt.Errorf("failed to decode interactions: %w", err)
	}

	for _, interaction := range interactions {
		c.Receive(interaction)
	}

	return nil
}

// SidecarHandler HTTP 处理器 (接收 Sidecar 推送)
func (c *Collector) SidecarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var interaction models.Interaction
	if err := json.NewDecoder(r.Body).Decode(&interaction); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err := c.Receive(interaction); err != nil {
		http.Error(w, fmt.Sprintf("Failed to receive: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
}

// Stats 返回采集器统计
func (c *Collector) Stats() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	var total int64
	c.db.Model(&models.Interaction{}).Count(&total)

	return map[string]interface{}{
		"buffer_size":   len(c.buffer),
		"batch_size":    c.cfg.BatchSize,
		"total_stored":  total,
		"flush_interval": c.cfg.FlushInterval,
	}
}
