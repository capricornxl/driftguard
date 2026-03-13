# DriftGuard 🛡️

**AI Agent Behavior Degradation Monitoring System**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Build](https://img.shields.io/badge/build-passing-brightgreen)]()

---

## 📖 简介

DriftGuard 是一个**非侵入式**的 AI Agent 行为退化监测系统，用于实时检测 AI Agent 的性能下降、行为异常和幻觉问题。

### 核心功能

- 📊 **5 维度健康度评分**: 延迟、效率、一致性、准确性、幻觉
- 🔍 **4 层退化检测**: 阈值检测、趋势分析、异常检测、突变检测
- 🚨 **智能告警**: 分级告警 (Critical/Warning/Info) + 多渠道通知
- 📈 **实时监控**: Prometheus Metrics + Grafana Dashboard
- 🔌 **Sidecar 模式**: 零代码侵入，HTTP 采集

### 应用场景

- LLM API 性能监控
- AI Agent 生产环境部署
- 模型退化早期预警
- 服务质量 SLA 保障

---

## 🚀 快速开始

### 一键启动 (Docker)

```bash
# 克隆仓库
git clone https://github.com/driftguard/driftguard.git
cd driftguard

# 启动所有服务
docker compose up -d

# 验证服务
curl http://localhost:8080/health
```

### 访问面板

| 服务 | URL | 凭证 |
|------|-----|------|
| **Grafana** | http://localhost:3000 | admin / driftguard |
| **Prometheus** | http://localhost:9090 | - |
| **API** | http://localhost:8080 | - |

---

## 📦 架构

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
│  │  (采集)     │  │  (评估)     │  │    (检测)           │  │
│  └─────────────┘  └─────────────┘  └──────────┬──────────┘  │
│                                               │              │
│                                    ┌──────────▼──────────┐  │
│                                    │     Alerter         │  │
│                                    │     (告警)          │  │
│                                    └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│  PostgreSQL  ←→  Prometheus  ←→  Grafana                    │
│  (数据存储)     (指标抓取)     (可视化)                      │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔧 核心组件

### 1. Collector (数据采集)

- **端口**: 8081 (Sidecar), 8080 (API)
- **功能**: 接收 Agent 交互数据，批量写入数据库
- **配置**: BatchSize, FlushInterval

```bash
# Sidecar 模式 (推荐)
curl -X POST http://localhost:8081/collect \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "my-agent",
    "session_id": "session-123",
    "input": "用户输入",
    "output": "Agent 输出",
    "latency_ms": 250,
    "tokens_in": 10,
    "tokens_out": 20
  }'
```

### 2. Evaluator (健康度评估)

- **5 维度评分**:
  - **延迟 (15%)**: P95 < 500ms = 100 分
  - **效率 (10%)**: > 100 tokens/s = 100 分
  - **一致性 (30%)**: 输出稳定性
  - **准确性 (35%)**: 错误模式检测
  - **幻觉 (10%)**: 幻觉模式检测

```bash
# 评估健康度
curl -X POST http://localhost:8080/api/v1/agents/my-agent/evaluate
```

### 3. Detector (退化检测)

- **4 层检测**:
  - **阈值检测**: 健康度 < 60 = 退化
  - **趋势分析**: 线性回归斜率 < -2 = 下降趋势
  - **异常检测**: 偏离均值 > 2σ
  - **突变检测**: 单次下降 > 15 分

```bash
# 检测退化
curl http://localhost:8080/api/v1/agents/my-agent/detect
```

### 4. Alerter (告警处理)

- **告警级别**: Critical, Warning, Info
- **通知渠道**: Webhook, Log, Database
- **告警管理**: Acknowledge, Resolve

```bash
# 查询告警
curl http://localhost:8080/api/v1/alerts

# 确认告警
curl -X POST http://localhost:8080/api/v1/alerts/1/ack

# 解决告警
curl -X POST http://localhost:8080/api/v1/alerts/1/resolve
```

---

## 📊 Prometheus Metrics

DriftGuard 导出 15+ 个核心指标:

### 健康度评分

```promql
# Agent 健康度总分
driftguard_health_score{agent_id="my-agent"}

# 5 维度得分
driftguard_health_score_latency{agent_id="my-agent"}
driftguard_health_score_efficiency{agent_id="my-agent"}
driftguard_health_score_consistency{agent_id="my-agent"}
driftguard_health_score_accuracy{agent_id="my-agent"}
driftguard_health_score_hallucination{agent_id="my-agent"}
```

### 交互统计

```promql
# P95 延迟
histogram_quantile(0.95, rate(driftguard_interaction_latency_ms_bucket[5m]))

# 每分钟交互数
rate(driftguard_interactions_total[1m])
```

### 退化检测

```promql
# 退化 Agent 数量
sum(driftguard_degraded_agents)

# 检测置信度
driftguard_detection_confidence{agent_id="my-agent"}
```

### 告警统计

```promql
# 活跃告警
sum(driftguard_alerts_active)

# 严重告警
driftguard_alerts_by_level{level="critical"}
```

**Metrics 端点**: http://localhost:8080/metrics

---

## 📈 Grafana Dashboard

DriftGuard 预置 10 个监控面板:

1. **DriftGuard Overview** - 总体统计
2. **Average Health Score** - 平均健康度 (仪表盘)
3. **Degraded Agents** - 退化 Agent 数量
4. **Health Score Trend** - 健康度趋势图
5. **Health Score Dimensions** - 5 维度得分详情
6. **Interaction Latency** - 延迟分布 (P95)
7. **Active Alerts** - 活跃告警
8. **Alerts by Level** - 告警级别分布
9. **Detection Trend** - 检测趋势斜率
10. **Agent Status Table** - Agent 状态表格

**访问**: http://localhost:3000 (admin / driftguard)

---

## 🔌 集成示例

### Python Agent

```python
import requests

DRIFTGUARD_URL = "http://localhost:8081/collect"

def track_interaction(agent_id, session_id, input_text, output_text, latency_ms, tokens_in, tokens_out):
    """上报 Agent 交互到 DriftGuard"""
    requests.post(DRIFTGUARD_URL, json={
        "agent_id": agent_id,
        "session_id": session_id,
        "input": input_text,
        "output": output_text,
        "latency_ms": latency_ms,
        "tokens_in": tokens_in,
        "tokens_out": tokens_out
    })

# 使用示例
import time
start = time.time()
response = call_llm("用户问题")
latency = int((time.time() - start) * 1000)

track_interaction(
    agent_id="my-llm-agent",
    session_id="session-123",
    input_text="用户问题",
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

// 使用示例
const start = Date.now();
const response = await callLLM('用户问题');
const latency = Date.now() - start;

await trackInteraction({
  sessionId: 'session-123',
  input: '用户问题',
  output: response,
  latencyMs: latency,
  tokensIn: 10,
  tokensOut: 50
});
```

### LangChain 集成

```python
from langchain.callbacks.base import BaseCallbackHandler
import requests

class DriftGuardCallback(BaseCallbackHandler):
    def __init__(self, agent_id, session_id):
        self.agent_id = agent_id
        self.session_id = session_id
        self.start_time = None
        self.prompt_tokens = 0
        self.completion_tokens = 0
    
    def on_llm_start(self, serialized, prompts, **kwargs):
        self.start_time = time.time()
        self.prompt_tokens = len(prompts[0])
    
    def on_llm_end(self, response, **kwargs):
        latency = int((time.time() - self.start_time) * 1000)
        self.completion_tokens = len(response.generations[0][0].text)
        
        requests.post("http://localhost:8081/collect", json={
            "agent_id": self.agent_id,
            "session_id": self.session_id,
            "input": response.generations[0][0].text,
            "output": response.generations[0][0].text,
            "latency_ms": latency,
            "tokens_in": self.prompt_tokens,
            "tokens_out": self.completion_tokens
        })

# 使用示例
from langchain.chat_models import ChatOpenAI
from langchain.chains import LLMChain

llm = ChatOpenAI(callbacks=[DriftGuardCallback("langchain-agent", "session-1")])
chain = LLMChain(llm=llm, prompt=prompt)
chain.run("用户问题")
```

---

## ⚙️ 配置

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

## 🛠️ 开发

### 环境要求

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15+

### 本地开发

```bash
# 安装依赖
go mod download

# 运行测试
go test ./... -v

# 本地运行
go run cmd/main.go -config config.json

# 构建
go build -o driftguard ./cmd/main.go
```

### 运行测试

```bash
# 单元测试
go test ./internal/... -v

# 集成测试
./tests/integration-test.sh
```

---

## 📁 项目结构

```
driftguard/
├── cmd/
│   └── main.go              # 入口文件
├── internal/
│   ├── alerter/             # 告警处理
│   ├── api/                 # API 服务器
│   ├── collector/           # 数据采集
│   ├── detector/            # 退化检测
│   └── evaluator/           # 健康度评估
├── pkg/
│   ├── config/              # 配置管理
│   ├── metrics/             # Prometheus Metrics
│   └── models/              # 数据模型
├── deploy/
│   ├── docker-compose.yml   # Docker 编排
│   └── grafana/             # Grafana 配置
├── tests/
│   └── integration-test.sh  # 集成测试
├── docs/
│   └── metrics-guide.md     # Metrics 使用指南
├── config.json              # 配置文件
├── go.mod
├── go.sum
└── README.md
```

---

## 📄 License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request!

1. Fork 仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交变更 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

详细指南请查看 [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 📞 联系方式

- **GitHub**: https://github.com/driftguard/driftguard
- **Issues**: https://github.com/driftguard/driftguard/issues
- **Discussions**: https://github.com/driftguard/driftguard/discussions
- **Security**: [SECURITY.md](SECURITY.md)

---

## 🙏 致谢

感谢以下开源项目:

- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [GORM](https://github.com/go-gorm/gorm) - ORM 库
- [Prometheus](https://github.com/prometheus/prometheus) - 监控系统
- [Grafana](https://github.com/grafana/grafana) - 可视化平台

---

*Built with ❤️ for AI Agent Monitoring*
