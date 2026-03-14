# DriftGuard 高级功能配置指南

## 📊 KS/PSI 统计检验

### 概述

DriftGuard 现在支持两种业界标准的统计检验方法来检测 AI Agent 行为漂移：

1. **KS 检验 (Kolmogorov-Smirnov Test)** - 检测分布变化
2. **PSI (Population Stability Index)** - 评估群体稳定性

---

### KS 检验

**用途**: 检测两个样本是否来自同一分布

**假设**:
- H0 (原假设): 两个样本来自同一分布
- H1 (备择假设): 两个样本来自不同分布

**判断标准**:
- `p-value < 0.05`: 拒绝 H0，存在显著漂移
- `KS Statistic > 0.5`: 严重漂移

**配置示例**:
```yaml
driftguardConfig:
  detector:
    ksTest:
      enabled: true
      baselineWindow: 7    # 基线窗口 (天)
      currentWindow: 3     # 当前窗口 (天)
      significanceLevel: 0.05
```

**API 使用**:
```bash
# 执行 KS 检验
curl http://localhost:8080/api/v1/agents/agent-1/drift/ks-test?baseline=7&current=3

# 响应示例
{
  "statistic": 0.65,
  "pValue": 0.003,
  "significant": true,
  "interpretation": "Significant drift detected"
}
```

---

### PSI (群体稳定性指数)

**用途**: 衡量群体分布随时间的变化程度

**判断标准**:
- `PSI < 0.1`: 稳定 (Stable)
- `0.1 <= PSI < 0.2`: 中等变化 (Moderate)
- `PSI >= 0.2`: 显著变化 (Significant)

**配置示例**:
```yaml
driftguardConfig:
  detector:
    psi:
      enabled: true
      baselineDays: 7      # 基线期 (天)
      currentDays: 7       # 当前期 (天)
      numBuckets: 5        # 分桶数量
```

**API 使用**:
```bash
# 计算 PSI
curl http://localhost:8080/api/v1/agents/agent-1/drift/psi?baseline=7&current=7&buckets=5

# 响应示例
{
  "psi": 0.25,
  "stability": "significant",
  "details": [
    {
      "bucket": "0-20",
      "expectedPct": 20.0,
      "actualPct": 5.0,
      "contribution": 0.08
    }
  ]
}
```

---

## 📢 Slack/Discord 告警配置

### Slack 告警

#### 1. 创建 Incoming Webhook

1. 访问 https://slack.com/apps/new/A0F7XDUAZ-incoming-webhooks
2. 选择要发送告警的频道
3. 复制 Webhook URL

#### 2. 配置 DriftGuard

```yaml
driftguardConfig:
  alerter:
    enabled: true
    channels:
      - type: "slack"
        enabled: true
        webhook: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
```

#### 3. 告警消息格式

Slack 告警包含:
- 🚨 严重性标识 (颜色编码)
- Agent ID
- 告警类型 (drift/anomaly/threshold/mutation)
- 健康评分详情
- 时间戳

**示例**:
```
🚨 DriftGuard Alert: drift
━━━━━━━━━━━━━━━━━━━━━━
Health score dropped below threshold

Agent: agent-prod-1    Severity: critical
Status: active         Time: 2026-03-14T03:30:00Z
━━━━━━━━━━━━━━━━━━━━━━
DriftGuard
```

---

### Discord 告警

#### 1. 创建 Webhook

1. 进入服务器设置 → 频道
2. 选择频道 → 编辑频道 → 整合
3. 创建 Webhook 并复制 URL

#### 2. 配置 DriftGuard

```yaml
driftguardConfig:
  alerter:
    enabled: true
    channels:
      - type: "discord"
        enabled: true
        webhook: "https://discord.com/api/webhooks/WEBHOOK_ID/WEBHOOK_TOKEN"
```

#### 3. 告警消息格式

Discord 使用 Embed 格式:
- 颜色编码 (绿/橙/红)
- 结构化字段
- 时间戳

---

### 自定义 Webhook

支持任意 HTTP Webhook:

```yaml
driftguardConfig:
  alerter:
    channels:
      - type: "webhook"
        enabled: true
        url: "https://your-api.com/alerts"
```

**Payload 格式**:
```json
{
  "alert_id": "alert-123",
  "agent_id": "agent-1",
  "type": "drift",
  "severity": "critical",
  "status": "active",
  "message": "Health score dropped below threshold",
  "created_at": "2026-03-14T03:30:00Z"
}
```

---

## 🔍 综合漂移报告

### 生成报告

```bash
curl http://localhost:8080/api/v1/agents/agent-1/report/comprehensive
```

### 报告内容

```json
{
  "agentId": "agent-1",
  "analysisTime": "2026-03-14T03:30:00Z",
  "ksTest": {
    "statistic": 0.65,
    "pValue": 0.003,
    "significant": true
  },
  "psi": {
    "psi": 0.25,
    "stability": "significant"
  },
  "trend": {
    "slope": -0.5,
    "direction": "decreasing",
    "strength": 0.85,
    "significant": true
  },
  "spikes": {
    "hasSpikes": true,
    "spikeCount": 2,
    "mean": 75.5,
    "stdDev": 8.2
  },
  "overallRisk": "high",
  "recommendations": [
    "Significant distribution drift detected. Investigate recent changes.",
    "PSI of 0.250 indicates significant population shift.",
    "Health score declining trend detected.",
    "HIGH RISK: Immediate attention required."
  ]
}
```

### 风险等级

| 等级 | 条件 | 建议行动 |
|------|------|----------|
| **Low** | 风险分 < 3 | 继续常规监控 |
| **Medium** | 风险分 3-4 | 调查近期变更 |
| **High** | 风险分 ≥ 5 | 立即干预，考虑回滚 |

---

## 📈 趋势分析

### 配置

```yaml
driftguardConfig:
  detector:
    trendAnalysis:
      enabled: true
      windowDays: 14     # 分析窗口 (天)
      minDataPoints: 5   # 最小数据点
```

### API

```bash
curl http://localhost:8080/api/v1/agents/agent-1/trend?days=14
```

### 趋势方向

- `increasing`: 健康评分上升 (好)
- `decreasing`: 健康评分下降 (需关注)
- `stable`: 无明显趋势

---

## ⚡ 异常尖点检测

### 配置

```yaml
driftguardConfig:
  detector:
    spikeDetection:
      enabled: true
      windowDays: 7
      threshold: 2.5     # 标准差倍数
```

### API

```bash
curl http://localhost:8080/api/v1/agents/agent-1/spikes?days=7&threshold=2.5
```

---

## 🎯 最佳实践

### 1. 基线选择

- **KS 检验**: 基线 7 天，当前 3 天 (检测近期漂移)
- **PSI**: 基线 7 天，当前 7 天 (评估稳定性)

### 2. 告警阈值

| 严重性 | KS Statistic | PSI | 行动 |
|--------|--------------|-----|------|
| Info | > 0.3 | > 0.1 | 记录日志 |
| Warning | > 0.5 | > 0.2 | 发送 Slack/Discord |
| Critical | > 0.7 | > 0.3 | 立即告警 + 通知 |

### 3. 监控频率

- **生产环境**: 每 5 分钟检测一次
- **开发环境**: 每 30 分钟检测一次
- **报告生成**: 每日一次

### 4. 告警路由

```yaml
alerter:
  channels:
    - type: "log"        # 始终启用 (审计)
      enabled: true
    - type: "slack"      # Warning 级别
      enabled: true
      webhook: "..."
    - type: "discord"    # Critical 级别
      enabled: true
      webhook: "..."
    - type: "webhook"    # 自定义处理
      enabled: false
      url: "..."
```

---

## 📚 参考资料

- **KS 检验**: [Wikipedia](https://en.wikipedia.org/wiki/Kolmogorov%E2%80%93Smirnov_test)
- **PSI**: [Population Stability Index](https://en.wikipedia.org/wiki/Population_stability_index)
- **DriftGuard API**: `/docs/api/openapi.yaml`

---

*最后更新：2026-03-14*
