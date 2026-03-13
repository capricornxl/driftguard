# DriftGuard 🛡️

**AI Agent Behavior Degradation Monitoring System**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CI](https://github.com/capricornxl/driftguard/actions/workflows/ci.yml/badge.svg)](https://github.com/capricornxl/driftguard/actions)
[![Release](https://img.shields.io/github/v/release/capricornxl/driftguard)](https://github.com/capricornxl/driftguard/releases)

---

**🌐 Language**: [中文](README.md) | **English**

---

## 📖 Introduction

DriftGuard is a **non-intrusive** AI Agent behavior degradation monitoring system designed for real-time detection of AI Agent performance decline, behavioral anomalies, and hallucination issues.

### Core Features

- 📊 **5-Dimension Health Scoring**: Latency, Efficiency, Consistency, Accuracy, Hallucination
- 🔍 **4-Layer Degradation Detection**: Threshold, Trend Analysis, Anomaly Detection, Spike Detection
- 🚨 **Smart Alerting**: Multi-level alerts (Critical/Warning/Info) + Multi-channel notifications
- 📈 **Real-time Monitoring**: Prometheus Metrics + Grafana Dashboard
- 🔌 **Sidecar Mode**: Zero code intrusion, HTTP collection

### Use Cases

- LLM API Performance Monitoring
- AI Agent Production Deployment
- Model Degradation Early Warning
- Service Quality SLA Assurance

---

## 🚀 Quick Start

### One-Click Startup (Docker)

```bash
# Clone repository
git clone https://github.com/capricornxl/driftguard.git
cd driftguard

# Start all services
docker compose up -d

# Verify service
curl http://localhost:8080/health
```

### Access Dashboards

| Service | URL | Credentials |
|---------|-----|-------------|
| **Grafana** | http://localhost:3000 | admin / driftguard |
| **Prometheus** | http://localhost:9090 | - |
| **API** | http://localhost:8080 | - |

---

## 📦 Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     AI Agent Application                     │
└─────────────────────┬───────────────────────────────────────┘
                      │ (HTTP POST)
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                    DriftGuard Sidecar                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Collector  │→ │  Evaluator  │→ │    Detector         │  │
│  │  (Collect)  │  │  (Evaluate) │  │    (Detect)         │  │
│  └─────────────┘  └─────────────┘  └──────────┬──────────┘  │
│                                               │              │
│                                    ┌──────────▼──────────┐  │
│                                    │     Alerter         │  │
│                                    │     (Alert)         │  │
│                                    └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  PostgreSQL  ←→  Prometheus  ←→  Grafana                    │
│  (Storage)     (Metrics)     (Visualization)                │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔧 Core Components

### 1. Collector (Data Collection)

- **Ports**: 8081 (Sidecar), 8080 (API)
- **Function**: Receive Agent interaction data, batch write to database
- **Config**: BatchSize, FlushInterval

```bash
# Sidecar mode (recommended)
curl -X POST http://localhost:8081/collect \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "my-agent",
    "session_id": "session-123",
    "input": "User input",
    "output": "Agent output",
    "latency_ms": 250,
    "tokens_in": 10,
    "tokens_out": 20
  }'
```

### 2. Evaluator (Health Scoring)

- **5-Dimension Scoring**:
  - **Latency (15%)**: P95 < 500ms = 100 points
  - **Efficiency (10%)**: > 100 tokens/s = 100 points
  - **Consistency (30%)**: Output stability
  - **Accuracy (35%)**: Error pattern detection
  - **Hallucination (10%)**: Hallucination pattern detection

```bash
# Evaluate health
curl -X POST http://localhost:8080/api/v1/agents/my-agent/evaluate
```

### 3. Detector (Degradation Detection)

- **4-Layer Detection**:
  - **Threshold**: Health < 60 = Degraded
  - **Trend Analysis**: Linear regression slope < -2 = Declining trend
  - **Anomaly Detection**: Deviation > 2σ from mean
  - **Spike Detection**: Single drop > 15 points

```bash
# Detect degradation
curl http://localhost:8080/api/v1/agents/my-agent/detect
```

### 4. Alerter (Alert Processing)

- **Alert Levels**: Critical, Warning, Info
- **Notification Channels**: Webhook, Log, Database
- **Alert Management**: Acknowledge, Resolve

```bash
# Query alerts
curl http://localhost:8080/api/v1/alerts

# Acknowledge alert
curl -X POST http://localhost:8080/api/v1/alerts/1/ack

# Resolve alert
curl -X POST http://localhost:8080/api/v1/alerts/1/resolve
```

---

## 📊 Prometheus Metrics

DriftGuard exports 15+ core metrics:

### Health Scores

```promql
# Agent health score total
driftguard_health_score{agent_id="my-agent"}

# 5-dimension scores
driftguard_health_score_latency{agent_id="my-agent"}
driftguard_health_score_efficiency{agent_id="my-agent"}
driftguard_health_score_consistency{agent_id="my-agent"}
driftguard_health_score_accuracy{agent_id="my-agent"}
driftguard_health_score_hallucination{agent_id="my-agent"}
```

### Interaction Statistics

```promql
# P95 Latency
histogram_quantile(0.95, rate(driftguard_interaction_latency_ms_bucket[5m]))

# Interactions per minute
rate(driftguard_interactions_total[1m])
```

### Degradation Detection

```promql
# Number of degraded agents
sum(driftguard_degraded_agents)

# Detection confidence
driftguard_detection_confidence{agent_id="my-agent"}
```

### Alert Statistics

```promql
# Active alerts
sum(driftguard_alerts_active)

# Critical alerts
driftguard_alerts_by_level{level="critical"}
```

**Metrics Endpoint**: http://localhost:8080/metrics

---

## 📈 Grafana Dashboard

DriftGuard includes 10 pre-configured dashboards:

1. **DriftGuard Overview** - Overall statistics
2. **Average Health Score** - Average health (gauge)
3. **Degraded Agents** - Number of degraded agents
4. **Health Score Trend** - Health trend chart
5. **Health Score Dimensions** - 5-dimension score details
6. **Interaction Latency** - Latency distribution (P95)
7. **Active Alerts** - Active alerts
8. **Alerts by Level** - Alert level distribution
9. **Detection Trend** - Detection trend slope
10. **Agent Status Table** - Agent status table

**Access**: http://localhost:3000 (admin / driftguard)

---

## 🔌 Integration Examples

### Python Agent

```python
import requests

DRIFTGUARD_URL = "http://localhost:8081/collect"

def track_interaction(agent_id, session_id, input_text, output_text, latency_ms, tokens_in, tokens_out):
    """Report Agent interaction to DriftGuard"""
    requests.post(DRIFTGUARD_URL, json={
        "agent_id": agent_id,
        "session_id": session_id,
        "input": input_text,
        "output": output_text,
        "latency_ms": latency_ms,
        "tokens_in": tokens_in,
        "tokens_out": tokens_out
    })

# Usage example
import time
start = time.time()
response = call_llm("User question")
latency = int((time.time() - start) * 1000)

track_interaction(
    agent_id="my-llm-agent",
    session_id="session-123",
    input_text="User question",
    output_text=response,
    latency_ms=latency,
    tokens_in=10,
    tokens_out=50
)
```

### Node.js Agent

```javascript
const axios = require('axios');

const DRIFTGUARD_URL = 'http://localhost:8081/collect';

async function trackInteraction(data) {
  await axios.post(DRIFTGUARD_URL, {
    agent_id: 'my-agent',
    session_id: data.sessionId,
    input: data.input,
    output: data.output,
    latency_ms: data.latencyMs,
    tokens_in: data.tokensIn,
    tokens_out: data.tokensOut
  });
}

// Usage example
const start = Date.now();
const response = await callLLM('User question');
const latency = Date.now() - start;

await trackInteraction({
  sessionId: 'session-123',
  input: 'User question',
  output: response,
  latencyMs: latency,
  tokensIn: 10,
  tokensOut: 50
});
```

---

## ⚙️ Configuration

### config.json

```json
{
  "database": {
    "host": "postgres",
    "port": 5432,
    "user": "driftguard",
    "password": "driftguard123",
    "name": "driftguard"
  },
  "collector": {
    "batch_size": 100,
    "flush_interval": 60
  },
  "evaluator": {
    "weights": {
      "latency": 0.15,
      "efficiency": 0.10,
      "consistency": 0.30,
      "accuracy": 0.35,
      "hallucination": 0.10
    }
  },
  "detector": {
    "window_size": 100,
    "critical_threshold": 50,
    "warning_threshold": 70
  },
  "alerter": {
    "enabled": true,
    "channels": ["webhook", "log", "database"],
    "webhook_url": "http://your-webhook.com/alerts"
  },
  "api": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "sidecar": {
    "host": "0.0.0.0",
    "port": 8081
  }
}
```

---

## 🛠️ Development

### Requirements

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15+

### Local Development

```bash
# Install dependencies
go mod download

# Run tests
go test ./... -v

# Run locally
go run cmd/main.go -config config.json

# Build
go build -o driftguard ./cmd/main.go
```

### Run Tests

```bash
# Unit tests
go test ./internal/... -v

# Integration tests
./tests/integration-test.sh
```

---

## 📁 Project Structure

```
driftguard/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── alerter/             # Alert processing
│   ├── api/                 # API server
│   ├── collector/           # Data collection
│   ├── detector/            # Degradation detection
│   └── evaluator/           # Health evaluation
├── pkg/
│   ├── config/              # Configuration
│   ├── metrics/             # Prometheus Metrics
│   └── models/              # Data models
├── deploy/
│   ├── docker-compose.yml   # Docker orchestration
│   └── grafana/             # Grafana configuration
├── tests/
│   └── integration-test.sh  # Integration tests
├── docs/
│   └── metrics-guide.md     # Metrics guide
├── config.json              # Configuration file
├── go.mod
├── go.sum
└── README.md
```

---

## 📄 License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

---

## 🤝 Contributing

Welcome Issues and Pull Requests!

1. Fork the repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## 📞 Contact

- **GitHub**: https://github.com/capricornxl/driftguard
- **Issues**: https://github.com/capricornxl/driftguard/issues
- **Discussions**: https://github.com/capricornxl/driftguard/discussions
- **Security**: [SECURITY.md](SECURITY.md)

---

## 🙏 Acknowledgments

Thanks to the following open-source projects:

- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [GORM](https://github.com/go-gorm/gorm) - ORM library
- [Prometheus](https://github.com/prometheus/prometheus) - Monitoring system
- [Grafana](https://github.com/grafana/grafana) - Visualization platform

---

*Built with ❤️ for AI Agent Monitoring*
