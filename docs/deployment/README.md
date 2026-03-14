# DriftGuard 部署指南

**版本**: MVP v0.1.1  
**最后更新**: 2026-03-14  
**状态**: 生产就绪 ✅

---

## 📖 概述

DriftGuard 采用**三层部署策略**，满足不同场景需求：

```
┌─────────────────────────────────────────────────────────┐
│              DriftGuard 三层部署架构                     │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  🏆 生产层 (Production)                                 │
│     Kubernetes + Helm Chart                            │
│     高可用 | 自动扩缩 | 完整监控                        │
│                        ▲                                │
│                       / \                               │
│                      /   \                              │
│                     /     \                             │
│                    /       \                            │
│  🧪 测试层 (Staging)                                    │
│     Docker Compose                                      │
│     完整功能 | 接近生产 | 批量测试                      │
│                  ▲                                      │
│                 / \                                     │
│                /   \                                    │
│               /     \                                   │
│              /       \                                  │
│  🚀 体验层 (Quick Start)                                │
│     本地运行 (SQLite)                                   │
│     零依赖 | 5 分钟启动 | 快速体验                       │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## 🎯 选择部署方式

### 我应该选择哪种部署方式？

| 场景 | 推荐部署 | 预计时间 |
|------|----------|----------|
| 首次体验/评估 | 体验层 (本地运行) | 5 分钟 |
| 功能验证/POC | 体验层 (本地运行) | 10 分钟 |
| 本地开发调试 | 体验层 (本地运行) | 5 分钟 |
| 集成测试/CI/CD | 测试层 (Docker Compose) | 15 分钟 |
| 客户环境预演 | 测试层 (Docker Compose) | 30 分钟 |
| 生产环境部署 | 生产层 (Kubernetes) | 1 小时 |
| 大规模 SaaS | 生产层 (Kubernetes) | 2 小时 |

---

## 📚 文档导航

### 体验层 (Quick Start)

- **[快速入门](quick-start.md)** - 5 分钟快速体验
- **[本地开发](local-development.md)** - 本地开发环境配置
- **[配置参考](configs/local-config.yaml)** - 本地配置示例

**适用场景**:
- ✅ 首次体验 DriftGuard
- ✅ 技术评估/POC 验证
- ✅ 本地开发调试
- ✅ 功能演示/Demo

**快速启动**:
```bash
# 一键启动 (无需 Docker)
./run-local.sh

# 访问 API
curl http://localhost:8080/api/v1/health
```

---

### 测试层 (Staging)

- **[Docker Compose 部署](docker-compose.md)** - 完整功能部署
- **[批量测试](examples/batch-testing.md)** - 多实例测试
- **[CI/CD 集成](examples/ci-cd.md)** - GitHub Actions 示例
- **[配置参考](configs/staging-config.yaml)** - 测试环境配置

**适用场景**:
- ✅ 功能验收测试
- ✅ 性能/压力测试
- ✅ 集成测试
- ✅ 批量环境部署
- ✅ CI/CD 流水线

**快速部署**:
```bash
# 启动完整栈 (App + PostgreSQL + Prometheus + Grafana)
docker-compose -f docker-compose.yml up -d

# 查看状态
docker-compose ps
```

---

### 生产层 (Production)

- **[Kubernetes 部署](kubernetes/README.md)** - K8s 部署指南
- **[Helm Chart 使用](kubernetes/helm-chart.md)** - Helm 安装配置
- **[生产最佳实践](production/best-practices.md)** - 生产环境优化
- **[扩缩容指南](production/scaling.md)** - 自动扩缩容配置
- **[监控告警](production/monitoring.md)** - 监控告警配置
- **[配置参考](configs/production-config.yaml)** - 生产环境配置

**适用场景**:
- ✅ 正式生产环境
- ✅ 多租户 SaaS 服务
- ✅ 大规模部署
- ✅ 企业级客户
- ✅ 7x24 高可用

**快速部署**:
```bash
# Helm 安装
helm install driftguard ./helm/driftguard \
  --namespace production \
  --values values-production.yaml
```

---

## 🔄 升级路径

### 从体验层到测试层

```bash
# 1. 导出 SQLite 数据 (可选)
./bin/driftguard export --output backup.json

# 2. 启动 Docker Compose
docker-compose -f docker-compose.yml up -d

# 3. 导入数据 (可选)
curl -X POST http://localhost:8080/api/v1/import \
  -d @backup.json
```

**详细指南**: [migration/sqlite-to-postgres.md](migration/sqlite-to-postgres.md)

---

### 从测试层到生产层

```bash
# 1. 备份数据
docker-compose exec db pg_dump -U driftguard driftguard > backup.sql

# 2. 创建 K8s 集群 (如需要)
kubectl create namespace production

# 3. 部署 Helm Chart
helm install driftguard ./helm/driftguard \
  --namespace production \
  --values values-production.yaml

# 4. 迁移数据
kubectl exec -it driftguard-postgresql-0 -- psql -U driftguard -d driftguard < backup.sql
```

**详细指南**: [migration/compose-to-k8s.md](migration/compose-to-k8s.md)

---

## 📊 部署对比

| 特性 | 体验层 | 测试层 | 生产层 |
|------|--------|--------|--------|
| **部署方式** | 本地运行 | Docker Compose | Kubernetes + Helm |
| **数据库** | SQLite | PostgreSQL | PostgreSQL (HA) |
| **部署时间** | < 1 分钟 | < 5 分钟 | < 15 分钟 |
| **资源占用** | < 200MB | < 1GB | 按需扩展 |
| **依赖** | 无 | Docker | K8s 集群 |
| **监控** | 基础 | Prometheus | 完整监控栈 |
| **日志** | 控制台 | 文件 | ELK/Loki |
| **备份** | 手动 | 手动/脚本 | 自动备份 |
| **高可用** | ❌ | ❌ | ✅ |
| **自动扩缩** | ❌ | ❌ | ✅ |
| **负载均衡** | ❌ | ❌ | ✅ |
| **配置中心** | ❌ | ❌ | ✅ |

---

## 🛠️ 前置条件

### 体验层

- Go 1.18+ (用于从源码运行)
- 或下载预编译二进制
- 无需其他依赖

### 测试层

- Docker 20.10+
- Docker Compose 2.0+
- 至少 2GB 可用内存
- 至少 10GB 可用磁盘

### 生产层

- Kubernetes 1.24+ 集群
- Helm 3.10+
- 至少 3 个 Worker 节点
- 每个节点至少 4GB 内存
- 持久化存储 (StorageClass)
- Ingress 控制器 (可选)

---

## 🚀 快速开始

### 5 分钟体验 DriftGuard

```bash
# 1. 克隆仓库
git clone https://github.com/capricornxl/driftguard.git
cd driftguard

# 2. 启动本地服务
./run-local.sh

# 3. 测试健康检查
curl http://localhost:8080/api/v1/health

# 4. 运行验收测试
./quick-acceptance-test.sh
```

**预期输出**:
```
🎉 所有测试通过！验收成功！
```

---

## 📖 相关文档

- **[架构设计](../architecture/README.md)** - 系统架构说明
- **[API 文档](../api/README.md)** - API 接口文档
- **[配置参考](../configuration/README.md)** - 配置项详解
- **[故障排查](troubleshooting.md)** - 常见问题解决
- **[性能调优](performance-tuning.md)** - 性能优化指南

---

## 🆘 获取帮助

### 文档资源

- 本地文档：`docs/` 目录
- 在线文档：https://github.com/capricornxl/driftguard/docs

### 社区支持

- GitHub Issues: https://github.com/capricornxl/driftguard/issues
- Discord: https://discord.gg/driftguard
- 邮件：support@driftguard.io

### 商业支持

如需企业级支持，请联系：
- 邮箱：enterprise@driftguard.io
- 网站：https://driftguard.io/enterprise

---

## 📝 版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| v0.1.1 | 2026-03-14 | 完善三层部署策略 |
| v0.1.0 | 2026-03-13 | 初始部署文档 |

---

*文档维护：DriftGuard Team*  
*最后更新：2026-03-14 07:30 UTC*
