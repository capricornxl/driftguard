# Docker Compose 部署指南

**适用人群**: QA 测试、批量测试、CI/CD、客户环境预演  
**预计时间**: 15 分钟  
**前置条件**: Docker 20.10+, Docker Compose 2.0+

---

## 🎯 本指南将带你

- ✅ 部署完整功能栈 (App + PostgreSQL + Prometheus + Grafana)
- ✅ 配置监控告警
- ✅ 运行批量测试
- ✅ 集成 CI/CD 流水线

---

## 📦 部署架构

```
┌─────────────────────────────────────────────────────────┐
│              Docker Compose 完整栈                       │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌─────────────┐    ┌─────────────┐                   │
│  │   App       │───▶│  PostgreSQL │                   │
│  │  (Port 8080)│    │  (Port 5432)│                   │
│  └─────────────┘    └─────────────┘                   │
│         │                                               │
│         ▼                                               │
│  ┌─────────────┐    ┌─────────────┐                   │
│  │ Prometheus  │◀───│    App      │                   │
│  │  (Port 9090)│    │  (Metrics)  │                   │
│  └─────────────┘    └─────────────┘                   │
│         │                                               │
│         ▼                                               │
│  ┌─────────────┐                                        │
│  │  Grafana    │                                        │
│  │  (Port 3000)│                                        │
│  └─────────────┘                                        │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## 🚀 快速部署

### 方式一：简化版 (推荐快速测试)

**包含**: App + PostgreSQL

```bash
cd /root/.openclaw/workspace/driftguard

# 启动
docker-compose -f docker-compose.simple.yml up -d

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f app
```

**访问地址**:
- API: http://localhost:8080
- Database: localhost:5432

---

### 方式二：完整版 (推荐完整测试)

**包含**: App + PostgreSQL + Prometheus + Grafana

```bash
cd /root/.openclaw/workspace/driftguard

# 启动
docker-compose up -d

# 查看状态
docker-compose ps

# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f app
docker-compose logs -f prometheus
docker-compose logs -f grafana
```

**访问地址**:
- API: http://localhost:8080
- Database: localhost:5432
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin123)

---

## 📋 配置文件详解

### docker-compose.yml 结构

```yaml
version: '3.8'

services:
  # PostgreSQL 数据库
  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: driftguard
      POSTGRES_PASSWORD: driftguard123
      POSTGRES_DB: driftguard
    volumes:
      - postgres-data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U driftguard"]
      interval: 10s
      timeout: 5s
      retries: 5

  # DriftGuard 主应用
  app:
    build: .
    depends_on:
      db:
        condition: service_healthy
    environment:
      DB_DRIVER: postgres
      DB_HOST: db
      DB_PASSWORD: driftguard123
    ports:
      - "8080:8080"
    volumes:
      - app-data:/app/data

  # Prometheus 监控
  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    ports:
      - "9090:9090"

  # Grafana 可视化
  grafana:
    image: grafana/grafana:latest
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin123
    volumes:
      - grafana-data:/var/lib/grafana
    ports:
      - "3000:3000"

volumes:
  postgres-data:
  prometheus-data:
  grafana-data:
```

---

## 🔧 配置选项

### 环境变量配置

创建 `.env` 文件：

```bash
# .env 文件
# 服务器配置
PORT=8080
HOST=0.0.0.0

# 数据库配置
DB_DRIVER=postgres
DB_HOST=db
DB_PORT=5432
DB_NAME=driftguard
DB_USERNAME=driftguard
DB_PASSWORD=YourSecurePassword123!

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json

# 收集器配置
COLLECTOR_BATCH_SIZE=100
COLLECTOR_FLUSH_INTERVAL=60

# 检测器配置
DETECTOR_CHECK_INTERVAL=300
DETECTOR_WINDOW_DAYS=7
DETECTOR_THRESHOLD=70.0

# 健康权重
WEIGHT_LATENCY=0.15
WEIGHT_EFFICIENCY=0.10
WEIGHT_CONSISTENCY=0.30
WEIGHT_ACCURACY=0.35
WEIGHT_HALLUCINATION=0.10
```

### 自定义 docker-compose.override.yml

```yaml
version: '3.8'

services:
  app:
    environment:
      - DB_PASSWORD=YourSecurePassword123!
      - LOG_LEVEL=debug
    volumes:
      - ./config.yaml:/app/config/config.yaml
    ports:
      - "8080:8080"
      - "8081:8081"  # Agent endpoint

  db:
    environment:
      - POSTGRES_PASSWORD=YourSecurePassword123!
    ports:
      - "5432:5432"
```

---

## ✅ 验收测试

### 1. 基础健康检查

```bash
# 检查服务状态
docker-compose ps
# 所有服务应显示 "Up"

# 检查 API
curl http://localhost:8080/api/v1/health | jq

# 检查数据库连接
docker-compose exec db pg_isready -U driftguard -d driftguard
```

### 2. 运行自动化测试

```bash
./quick-acceptance-test.sh
```

### 3. 验证监控

```bash
# 访问 Prometheus
curl http://localhost:9090/api/v1/query?query=driftguard_health_score | jq

# 访问 Grafana API
curl -u admin:admin123 http://localhost:3000/api/health
```

### 4. 批量测试

```bash
# 发送 100 条交互数据
for i in {1..100}; do
  curl -X POST http://localhost:8080/api/v1/agents/batch-test/interactions \
    -H "Content-Type: application/json" \
    -d "{
      \"latency_ms\": $((400 + RANDOM % 200)),
      \"tokens_in\": $((80 + RANDOM % 50)),
      \"tokens_out\": $((200 + RANDOM % 100))
    }" > /dev/null
done

echo "✅ 已发送 100 条交互数据"

# 查看健康评分
curl http://localhost:8080/api/v1/agents/batch-test/scores/latest | jq
```

---

## 📊 监控配置

### Prometheus 配置

创建 `prometheus.yml`：

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'driftguard'
    static_configs:
      - targets: ['app:8080']
    metrics_path: '/api/v1/metrics'
```

### Grafana Dashboard

**导入预置 Dashboard**:

1. 访问 http://localhost:3000
2. 登录：admin / admin123
3. 导航到：Dashboards → Import
4. 输入 Dashboard ID 或上传 JSON 文件
5. 选择 Prometheus 数据源
6. 点击 Import

**预置面板**:
- 健康评分趋势
- 交互量统计
- 延迟分布
- 告警统计
- 系统资源

---

## 🔄 批量部署多环境

### 部署多个测试环境

```bash
# 创建测试环境 1-5
for i in {1..5}; do
  export COMPOSE_PROJECT_NAME=driftguard-test-$i
  docker-compose up -d
done

# 查看所有环境
docker ps --filter "name=driftguard-test"

# 停止特定环境
COMPOSE_PROJECT_NAME=driftguard-test-3 docker-compose down
```

### 环境端口映射

| 环境 | API 端口 | DB 端口 | Prometheus | Grafana |
|------|----------|---------|------------|---------|
| test-1 | 8081 | 5431 | 9091 | 3001 |
| test-2 | 8082 | 5432 | 9092 | 3002 |
| test-3 | 8083 | 5433 | 9093 | 3003 |

---

## 🧹 运维操作

### 查看日志

```bash
# 所有服务日志
docker-compose logs -f

# 特定服务日志
docker-compose logs -f app
docker-compose logs -f db

# 最近 100 行
docker-compose logs --tail=100 app

# 带时间戳
docker-compose logs -f --timestamps app
```

### 重启服务

```bash
# 重启所有
docker-compose restart

# 重启特定服务
docker-compose restart app
docker-compose restart db
```

### 进入容器

```bash
# 进入应用容器
docker-compose exec app sh

# 进入数据库容器
docker-compose exec db psql -U driftguard -d driftguard

# 查看数据库
docker-compose exec db psql -U driftguard -d driftguard -c "SELECT * FROM agents;"
```

### 备份数据

```bash
# 备份数据库
docker-compose exec db pg_dump -U driftguard driftguard > backup-$(date +%Y%m%d).sql

# 备份所有 volumes
docker run --rm \
  -v driftguard_postgres-data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/postgres-backup.tar.gz -C /data .
```

### 恢复数据

```bash
# 恢复数据库
cat backup-20260314.sql | docker-compose exec -T db psql -U driftguard -d driftguard
```

---

## 🛑 停止和清理

### 停止服务

```bash
# 停止所有服务
docker-compose down

# 停止并删除网络
docker-compose down --networks

# 停止并删除 volumes (删除数据!)
docker-compose down -v
```

### 完全清理

```bash
# 停止服务
docker-compose down -v

# 删除镜像
docker-compose rm -f

# 删除所有相关镜像
docker rmi driftguard-app postgres:15-alpine prom/prometheus grafana/grafana
```

---

## ⚠️ 注意事项

### 测试层限制

- ❌ 不支持高可用 (单点故障)
- ❌ 不支持自动扩缩
- ❌ 备份需手动执行
- ❌ 日志轮转需配置

### 适用场景

- ✅ 功能验收测试
- ✅ 性能/压力测试
- ✅ 集成测试
- ✅ 批量环境部署
- ✅ CI/CD 流水线
- ✅ 客户环境预演

### 不适用场景

- ❌ 正式生产环境
- ❌ 7x24 高可用要求
- ❌ 大规模并发

---

## 🔄 下一步

### 体验满意？升级到生产层

```bash
# Kubernetes + Helm
helm install driftguard ./helm/driftguard \
  --namespace production \
  --values values-production.yaml
```

**详细指南**: [kubernetes/README.md](kubernetes/README.md)

### 需要 CI/CD 集成？

**详细指南**: [examples/ci-cd.md](examples/ci-cd.md)

---

## 🆘 故障排查

### 服务无法启动

```bash
# 查看详细日志
docker-compose logs app

# 检查端口占用
lsof -i :8080
lsof -i :5432

# 重新构建
docker-compose build --no-cache
docker-compose up -d
```

### 数据库连接失败

```bash
# 检查数据库健康
docker-compose ps db

# 测试数据库连接
docker-compose exec db pg_isready -U driftguard

# 查看数据库日志
docker-compose logs db
```

### 监控无法访问

```bash
# 检查 Prometheus
curl http://localhost:9090/-/healthy

# 检查 Grafana
curl http://localhost:3000/api/health

# 重新创建网络
docker-compose down
docker network prune
docker-compose up -d
```

---

## 📖 相关文档

- **[快速入门](quick-start.md)** - 5 分钟体验
- **[本地开发](local-development.md)** - 开发环境配置
- **[Kubernetes 部署](kubernetes/README.md)** - 生产部署
- **[CI/CD 示例](examples/ci-cd.md)** - GitHub Actions 集成
- **[配置参考](../configs/staging-config.yaml)** - 测试环境配置

---

*文档维护：DriftGuard Team*  
*最后更新：2026-03-14 07:30 UTC*
