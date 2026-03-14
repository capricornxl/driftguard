package database

import (
	"fmt"
	"time"

	"github.com/driftguard/driftguard/pkg/config"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase creates a new database connection with connection pool configuration
func NewDatabase(cfg *config.DatabaseConfig, log *logrus.Logger) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port)
		dialector = postgres.New(dsn)
	case "sqlite":
		dialector = sqlite.Open(cfg.Database)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	// Configure GORM logger
	gormLogLevel := logger.Silent
	if cfg.Debug {
		gormLogLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying database: %w", err)
	}

	// Configure connection pool (P0 fix: 数据库连接池配置)
	// These settings prevent connection exhaustion under high concurrency
	sqlDB.SetMaxOpenConns(25)           // 最大打开连接数
	sqlDB.SetMaxIdleConns(5)            // 空闲连接数
	sqlDB.SetConnMaxLifetime(5 * time.Minute)  // 连接最大生命周期
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)  // 连接最大空闲时间

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.WithFields(logrus.Fields{
		"driver":          cfg.Driver,
		"max_open_conns":  25,
		"max_idle_conns":  5,
		"conn_max_lifetime": "5m",
		"conn_max_idle_time": "2m",
	}).Info("Database connection established with connection pool")

	return db, nil
}

// CloseDatabase gracefully closes the database connection
func CloseDatabase(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
