# DriftGuard Prometheus Metrics 指南

## 📊 可用指标

### 健康度评分

| 指标名 | 类型 | 标签 | 说明 |
|--------|------|------|------|
| `driftguard_health_score` | Gauge | agent_id, level | Agent 健康度总分 (0-100) |
| `driftguard_health_score_latency` | Gauge | agent_id | 延迟维度得分 |
| `driftguard_health_score_efficiency` | Gauge | agent_id | 效率维度得分 |
| `driftguard_health_score_consistency` | Gauge | agent_id | 一致性维度得分 |
| `driftguard_health_score_accuracy` | Gauge | agent_id | 准确性维度得分 |
| `driftguard_health_score_hallucination` | Gauge | agent_id | 幻觉维度得分 |

### 交互统计

| 指标名 | 类型 | 标签 | 说明 |
|--------|------|------|------|
| `driftguard_interactions_total` | Counter | agent_id | 交互总数 |
| `driftguard_interaction_latency_ms` | Histogram | agent_id | 交互延迟分布 |
| `driftguard_interaction_tokens` | Histogram | agent_id, type | Tokens 分布 (input/output) |

### 退化检测

| 指标名 | 类型 | 标签 | 说明 |
|--------|------|------|------|
| `driftguard_degraded_agents` | Gauge | agent_id, degraded_type | 退化 Agent 数量 |
| `driftguard_detection_trend` | Gauge | agent_id | 检测趋势斜率 |
| `driftguard_detection_confidence` | Gauge | agent_id | 检测置信度 |

### 告警统计

| 指标名 | 类型 | 标签 | 说明 |
|--------|------|------|------|
| `driftguard_alerts_total` | Counter | agent_id, level, type | 告警总数 |
| `driftguard_alerts_active` | Gauge | agent_id | 活跃告警数 |
| `driftguard_alerts_by_level` | Gauge | level | 按级别分组的告警 |

---

## 🔍 查询示例

### Prometheus Query Examples

```promql
# 所有 Agent 的平均健康度
avg(driftguard_health_score)

# 特定 Agent 的健康度
driftguard_health_score{agent_id="my-agent"}

# 退化 Agent 数量
sum(driftguard_degraded_agents)

# P95 延迟
histogram_quantile(0.95, rate(driftguard_interaction_latency_ms_bucket[5m]))

# 每分钟交互数
rate(driftguard_interactions_total[1m])

# 活跃告警数
sum(driftguard_alerts_active)

# 严重告警数量
driftguard_alerts_by_level{level="critical"}
```

---

## 📈 Grafana Dashboard

### 访问方式

- **URL**: http://localhost:3000
- **用户名**: admin
- **密码**: driftguard

### 预置面板

1. **DriftGuard Overview** - 总体统计
2. **Average Health Score** - 平均健康度 (仪表盘)
3. **Degraded Agents** - 退化 Agent 数量
4. **Health Score Trend** - 健康度趋势图
5. **Health Score Dimensions** - 5 维度得分详情
6. **Interaction Latency** - 延迟分布
7. **Active Alerts** - 活跃告警
8. **Alerts by Level** - 告警级别统计
9. **Detection Trend** - 检测趋势
10. **Agent Status Table** - Agent 状态表格

### 导入 Dashboard

Dashboard 已预配置在 `deploy/grafana/dashboards/`，启动 Docker 后自动加载。

---

## 🚀 快速开始

```bash
# 1. 启动服务
docker compose up -d

# 2. 验证 Metrics 端点
curl http://localhost:8080/metrics

# 3. 访问 Grafana
open http://localhost:3000

# 4. 访问 Prometheus
open http://localhost:9090
```

---

## 🔧 告警规则 (Alerting Rules)

可在 Prometheus 中配置告警规则:

```yaml
groups:
  - name: driftguard
    rules:
      - alert: AgentHealthCritical
        expr: driftguard_health_score < 50
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Agent {{ $labels.agent_id }} health critical"
          
      - alert: AgentDegraded
        expr: driftguard_degraded_agents == 1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Agent {{ $labels.agent_id }} is degraded"
```
