package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/driftguard/driftguard/internal/alerter"
	"github.com/driftguard/driftguard/internal/api"
	"github.com/driftguard/driftguard/internal/collector"
	"github.com/driftguard/driftguard/internal/detector"
	"github.com/driftguard/driftguard/internal/evaluator"
	"github.com/driftguard/driftguard/pkg/config"
	"github.com/driftguard/driftguard/pkg/metrics"
	"github.com/driftguard/driftguard/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	configPath = flag.String("config", "config.json", "配置文件路径")
	dbDSN      = flag.String("db", "", "数据库连接字符串 (覆盖配置文件)")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Printf("[Main] Config file not found, using defaults: %v", err)
		cfg = config.Default()
	}

	// 数据库连接覆盖
	if *dbDSN != "" {
		cfg.Database.Host = "" // 使用 DSN
	}

	log.Printf("[Main] DriftGuard starting...")
	log.Printf("[Main] Config: server=%s:%s, db=%s",
		cfg.Server.Host, cfg.Server.Port, cfg.Database.DBName)

	// 初始化数据库
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("[Main] Failed to initialize database: %v", err)
	}

	// 初始化组件
	collector := collector.NewCollector(&cfg.Collector, db)
	evaluator := evaluator.NewEvaluator(&cfg.Evaluator, db)
	detector := detector.NewDetector(&cfg.Detector, db)
	alerter := alerter.NewAlerter(&cfg.Alerter, db)
	metrics := metrics.NewMetrics()

	// 启动采集器
	collector.Start()
	defer collector.Stop()

	// 启动 API 服务器
	server := api.NewServer(collector, evaluator, detector, alerter, metrics)

	// 优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Printf("[Main] Shutting down...")
		collector.Stop()
		os.Exit(0)
	}()

	// 启动 HTTP 服务器 (用于采集器接收数据)
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/collect", collector.SidecarHandler)
		log.Printf("[Main] Sidecar collector listening on :8081")
		if err := http.ListenAndServe(":8081", mux); err != nil {
			log.Printf("[Main] Sidecar server error: %v", err)
		}
	}()

	// 运行 API 服务器
	if err := server.Run(cfg.Server.Host, cfg.Server.Port); err != nil {
		log.Fatalf("[Main] API server failed: %v", err)
	}
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	var dsn string
	if cfg.Database.Host != "" {
		dsn = "host=" + cfg.Database.Host +
			" user=" + cfg.Database.User +
			" password=" + cfg.Database.Password +
			" dbname=" + cfg.Database.DBName +
			" port=" + cfg.Database.Port +
			" sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	if err := db.AutoMigrate(
		&models.Interaction{},
		&models.HealthScore{},
		&models.Alert{},
		&models.AgentConfig{},
	); err != nil {
		return nil, err
	}

	log.Printf("[Main] Database initialized successfully")
	return db, nil
}
