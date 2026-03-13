package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Collector CollectorConfig `json:"collector"`
	Evaluator EvaluatorConfig `json:"evaluator"`
	Detector  DetectorConfig  `json:"detector"`
	Alerter   AlerterConfig   `json:"alerter"`
}

type ServerConfig struct {
	Port string `json:"port"`
	Host string `json:"host"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

type CollectorConfig struct {
	AgentEndpoint string `json:"agent_endpoint"`
	BatchSize     int    `json:"batch_size"`
	FlushInterval int    `json:"flush_interval"`
}

type EvaluatorConfig struct {
	Weights HealthWeights `json:"weights"`
}

type HealthWeights struct {
	Latency     float64 `json:"latency"`
	Efficiency  float64 `json:"efficiency"`
	Consistency float64 `json:"consistency"`
	Accuracy    float64 `json:"accuracy"`
	Hallucination float64 `json:"hallucination"`
}

type DetectorConfig struct {
	CheckInterval int     `json:"check_interval"`
	WindowDays    int     `json:"window_days"`
	Threshold     float64 `json:"threshold"`
}

type AlerterConfig struct {
	Enabled    bool     `json:"enabled"`
	Channels   []string `json:"channels"`
	WebhookURL string   `json:"webhook_url"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	return &cfg, err
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "driftguard",
			Password: "driftguard",
			DBName:   "driftguard",
		},
		Collector: CollectorConfig{
			AgentEndpoint: "http://localhost:8081",
			BatchSize:     100,
			FlushInterval: 60,
		},
		Evaluator: EvaluatorConfig{
			Weights: HealthWeights{
				Latency:       0.15,
				Efficiency:    0.10,
				Consistency:   0.30,
				Accuracy:      0.35,
				Hallucination: 0.10,
			},
		},
		Detector: DetectorConfig{
			CheckInterval: 300,
			WindowDays:    7,
			Threshold:     70.0,
		},
		Alerter: AlerterConfig{
			Enabled:    true,
			Channels:   []string{"webhook"},
			WebhookURL: "",
		},
	}
}
