# 快速入门 - 5 分钟体验 DriftGuard

**适用人群**: 首次体验者、技术评估、POC 验证  
**预计时间**: 5 分钟  
**前置条件**: 无 (零依赖)

---

## 🎯 本指南将带你

- ✅ 5 分钟内启动 DriftGuard
- ✅ 运行自动化验收测试
- ✅ 体验核心功能
- ✅ 了解下一步方向

---

## 🚀 方式一：预编译二进制 (最快)

### 步骤 1: 下载二进制

```bash
# Linux
wget https://github.com/capricornxl/driftguard/releases/download/v0.1.1/driftguard-linux-amd64
chmod +x driftguard-linux-amd64

# macOS
wget https://github.com/capricornxl/driftguard/releases/download/v0.1.1/driftguard-darwin-amd64
chmod +x driftguard-darwin-amd64

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/capricornxl/driftguard/releases/download/v0.1.1/driftguard-windows-amd64.exe" -OutFile "driftguard.exe"
```

### 步骤 2: 启动服务

```bash
./driftguard-linux-amd64
```

### 步骤 3: 验证运行

```bash
curl http://localhost:8080/api/v1/health
```

**预计时间**: 2 分钟

---

## 🚀 方式二：从源码运行

### 步骤 1: 克隆仓库

```bash
git clone https://github.com/capricornxl/driftguard.git
cd driftguard
```

### 步骤 2: 检查 Go 版本

```bash
go version
# 需要 Go 1.18+
# 如未安装：https://golang.org/dl/
```

### 步骤 3: 一键启动

```bash
./run-local.sh
```

**或者手动启动**:

```bash
# 编译
go build -o bin/driftguard ./cmd/main.go

# 启动
./bin/driftguard
```

### 步骤 4: 验证运行

```bash
curl http://localhost:8080/api/v1/health
```

**预计时间**: 5 分钟

---

## ✅ 验收测试

### 运行自动化测试

```bash
./quick-acceptance-test.sh
```

**预期输出**:

```
========================================
🔍 DriftGuard 快速验收测试
========================================
API 地址：http://localhost:8080

📋 阶段 1: 健康检查
测试：健康检查 ... ✅ 通过 (HTTP 200)
测试：就绪检查 ... ✅ 通过 (HTTP 200)
测试：存活检查 ... ✅ 通过 (HTTP 200)

📋 阶段 2: Agent 管理
测试：创建 Agent ... ✅ 通过 (HTTP 201)
...

📊 测试结果汇总
========================================
通过：25
失败：0
总计：25

🎉 所有测试通过！验收成功！
```

---

## 🎮 手动体验

### 1. 健康检查

```bash
curl http://localhost:8080/api/v1/health | jq
```

**预期响应**:

```json
{
  "status": "healthy",
  "timestamp": "2026-03-14T07:30:00Z",
  "version": "0.1.1",
  "uptime": "1m30s",
  "checks": {
    "database": "ok",
    "configuration": "ok",
    "memory": "ok",
    "disk": "ok"
  },
  "environment": "development"
}
```

---

### 2. 创建测试 Agent

```bash
curl -X POST http://localhost:8080/api/v1/agents \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "my-first-agent",
    "name": "My Test Agent",
    "model": "gpt-4",
    "description": "我的第一个测试 Agent"
  }' | jq
```

**预期响应**:

```json
{
  "id": 1,
  "agent_id": "my-first-agent",
  "name": "My Test Agent",
  "model": "gpt-4",
  "description": "我的第一个测试 Agent",
  "created_at": "2026-03-14T07:31:00Z",
  "status": "active"
}
```

---

### 3. 发送交互数据

```bash
# 发送单次交互
curl -X POST http://localhost:8080/api/v1/agents/my-first-agent/interactions \
  -H "Content-Type: application/json" \
  -d '{
    "latency_ms": 450,
    "tokens_in": 100,
    "tokens_out": 250
  }' | jq

# 批量发送 20 次交互
for i in {1..20}; do
  curl -X POST http://localhost:8080/api/v1/agents/my-first-agent/interactions \
    -H "Content-Type: application/json" \
    -d "{
      \"latency_ms\": $((400 + RANDOM % 200)),
      \"tokens_in\": $((80 + RANDOM % 50)),
      \"tokens_out\": $((200 + RANDOM % 100))
    }" > /dev/null
done

echo "✅ 已发送 20 条交互数据"
```

---

### 4. 查看健康评分

```bash
curl http://localhost:8080/api/v1/agents/my-first-agent/scores/latest | jq
```

**预期响应**:

```json
{
  "agent_id": "my-first-agent",
  "score": 85.5,
  "latency_score": 90.0,
  "efficiency_score": 80.0,
  "consistency_score": 85.0,
  "accuracy_score": 90.0,
  "hallucination_score": 85.0,
  "is_degraded": false,
  "calculated_at": "2026-03-14T07:32:00Z"
}
```

---

### 5. 漂移检测

```bash
# KS 检验
curl http://localhost:8080/api/v1/agents/my-first-agent/drift/ks-test | jq

# PSI 计算
curl http://localhost:8080/api/v1/agents/my-first-agent/drift/psi | jq

# 综合报告
curl http://localhost:8080/api/v1/agents/my-first-agent/report/comprehensive | jq
```

---

## 🔧 配置选项

### 环境变量

```bash
# 修改端口
export PORT=9090
./bin/driftguard

# 修改日志级别
export LOG_LEVEL=debug
./bin/driftguard

# 使用 JSON 日志
export LOG_FORMAT=json
./bin/driftguard

# 启用调试模式
export DEBUG=true
./bin/driftguard
```

### 完整配置参考

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | 8080 | 服务端口 |
| `HOST` | 0.0.0.0 | 监听地址 |
| `DB_DRIVER` | sqlite | 数据库驱动 |
| `DB_NAME` | /tmp/driftguard.db | 数据库文件 |
| `LOG_LEVEL` | info | 日志级别 |
| `LOG_FORMAT` | text | 日志格式 |
| `DEBUG` | false | 调试模式 |

---

## 🧹 清理

### 停止服务

```bash
# Ctrl+C 停止
```

### 删除数据

```bash
# 删除 SQLite 数据库
rm /tmp/driftguard-*.db

# 删除日志
rm /tmp/driftguard-*.log
```

### 删除二进制

```bash
rm bin/driftguard
```

---

## 📊 性能指标

### 预期性能

| 指标 | 预期值 |
|------|--------|
| 启动时间 | < 3 秒 |
| 内存占用 | < 150MB |
| CPU 占用 | < 2% (空闲) |
| API 响应 | < 100ms (P95) |

### 实际测试

```bash
# 压力测试 (100 并发)
ab -n 1000 -c 100 http://localhost:8080/api/v1/health

# 预期输出:
# Requests per second: 5000+ [#/sec]
# Time per request: < 20ms
```

---

## ⚠️ 注意事项

### 体验层限制

- ❌ 数据持久化有限 (SQLite 文件)
- ❌ 不支持高可用
- ❌ 不支持自动扩缩
- ❌ 监控功能基础

### 适用场景

- ✅ 个人学习/体验
- ✅ 技术评估/POC
- ✅ 本地开发调试
- ✅ 功能演示

### 不适用场景

- ❌ 生产环境
- ❌ 多用户并发
- ❌ 7x24 运行
- ❌ 大规模数据

---

## 🔄 下一步

### 体验满意？升级到测试层

```bash
# Docker Compose 完整功能
cd /root/.openclaw/workspace/driftguard
docker-compose -f docker-compose.yml up -d
```

**详细指南**: [docker-compose.md](docker-compose.md)

### 需要生产部署？

```bash
# Kubernetes + Helm
helm install driftguard ./helm/driftguard \
  --namespace production \
  --values values-production.yaml
```

**详细指南**: [kubernetes/README.md](kubernetes/README.md)

---

## 🆘 故障排查

### 端口被占用

```bash
# 检查端口占用
lsof -i :8080

# 使用其他端口
export PORT=9090
./bin/driftguard
```

### 编译失败

```bash
# 检查 Go 版本
go version
# 需要 1.18+

# 清理缓存
go clean -cache
go build -o bin/driftguard ./cmd/main.go
```

### 数据库错误

```bash
# 删除旧数据库
rm /tmp/driftguard-*.db

# 重新启动
./bin/driftguard
```

---

## 📖 相关文档

- **[本地开发](local-development.md)** - 开发环境配置
- **[配置参考](../configs/local-config.yaml)** - 配置示例
- **[API 文档](../api/README.md)** - API 接口说明
- **[部署概览](README.md)** - 三层部署架构

---

## 🎉 恭喜！

你已经成功体验 DriftGuard！

**接下来**:
- 阅读 [Docker Compose 部署](docker-compose.md) 了解完整功能
- 查看 [Kubernetes 部署](kubernetes/README.md) 了解生产部署
- 提交 [GitHub Issue](https://github.com/capricornxl/driftguard/issues) 反馈问题

---

*文档维护：DriftGuard Team*  
*最后更新：2026-03-14 07:30 UTC*
