# Kubernetes 部署指南

**适用人群**: 运维工程师、SRE、生产环境部署  
**预计时间**: 30-60 分钟  
**前置条件**: Kubernetes 1.24+, Helm 3.10+

---

## 🎯 本指南将带你

- ✅ 在 Kubernetes 集群部署 DriftGuard
- ✅ 配置高可用和自动扩缩
- ✅ 集成监控告警
- ✅ 实现零停机发布

---

## 📦 部署架构

```
┌─────────────────────────────────────────────────────────┐
│              Kubernetes 生产部署架构                     │
├─────────────────────────────────────────────────────────┤
│                                                         │
│                    ┌─────────────┐                     │
│                    │   Ingress   │                     │
│                    │  Controller │                     │
│                    └──────┬──────┘                     │
│                           │                             │
│                    ┌──────▼──────┐                     │
│                    │   Service   │                     │
│                    │  (ClusterIP)│                     │
│                    └──────┬──────┘                     │
│                           │                             │
│         ┌─────────────────┼─────────────────┐          │
│         │                 │                 │          │
│    ┌────▼────┐      ┌────▼────┐      ┌────▼────┐     │
│    │   Pod   │      │   Pod   │      │   Pod   │     │
│    │  App 1  │      │  App 2  │      │  App 3  │     │
│    └────┬────┘      └────┬────┘      └────┬────┘     │
│         │                 │                 │          │
│         └─────────────────┼─────────────────┘          │
│                           │                             │
│                    ┌──────▼──────┐                     │
│                    │   Service   │                     │
│                    │  (Headless) │                     │
│                    └──────┬──────┘                     │
│                           │                             │
│         ┌─────────────────┼─────────────────┐          │
│         │                 │                 │          │
│    ┌────▼────┐      ┌────▼────┐      ┌────▼────┐     │
│    │   Pod   │      │   Pod   │      │   Pod   │     │
│    │  DB 1   │      │  DB 2   │      │  DB 3   │     │
│    │(Patroni)│      │(Patroni)│      │(Patroni)│     │
│    └─────────┘      └─────────┘      └─────────┘     │
│                                                         │
│    ┌─────────────┐         ┌─────────────┐            │
│    │ Prometheus  │         │   Grafana   │            │
│    │   (Stack)   │         │   (Stack)   │            │
│    └─────────────┘         └─────────────┘            │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## 🚀 快速部署

### 步骤 1: 准备集群

```bash
# 检查集群状态
kubectl cluster-info
kubectl get nodes

# 创建命名空间
kubectl create namespace production
kubectl create namespace monitoring
```

### 步骤 2: 安装 Helm Chart

```bash
cd /root/.openclaw/workspace/driftguard

# 添加依赖 Helm Repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# 安装 DriftGuard
helm install driftguard ./helm/driftguard \
  --namespace production \
  --create-namespace \
  --values helm/driftguard/values-production.yaml
```

### 步骤 3: 验证部署

```bash
# 查看 Pod 状态
kubectl get pods -n production

# 查看服务
kubectl get svc -n production

# 查看部署
kubectl get deployment -n production

# 查看日志
kubectl logs -f deployment/driftguard -n production
```

### 步骤 4: 访问应用

```bash
# 获取外部 IP
kubectl get svc driftguard -n production

# 或通过端口转发
kubectl port-forward svc/driftguard 8080:80 -n production

# 访问 API
curl http://localhost:8080/api/v1/health
```

---

## 📋 Helm Chart 结构

```
helm/driftguard/
├── Chart.yaml              # Chart 元数据
├── values.yaml             # 默认配置
├── values-production.yaml  # 生产环境配置
├── values-staging.yaml     # 测试环境配置
├── templates/
│   ├── _helpers.tpl        # 模板辅助函数
│   ├── deployment.yaml     # Deployment 配置
│   ├── service.yaml        # Service 配置
│   ├── configmap.yaml      # ConfigMap 配置
│   ├── secret.yaml         # Secret 配置
│   ├── ingress.yaml        # Ingress 配置
│   ├── hpa.yaml            # HorizontalPodAutoscaler
│   ├── pdb.yaml            # PodDisruptionBudget
│   ├── servicemonitor.yaml # Prometheus ServiceMonitor
│   └── NOTES.txt           # 安装后说明
└── charts/                 # 子 Chart (可选)
```

---

## 🔧 配置选项

### 生产环境配置 (values-production.yaml)

```yaml
# 副本数
replicaCount: 3

# 镜像配置
image:
  repository: ghcr.io/capricornxl/driftguard
  tag: "v0.1.1"
  pullPolicy: IfNotPresent

# 资源限制
resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

# 自动扩缩
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80

# 数据库配置
postgresql:
  enabled: true
  auth:
    postgresPassword: "YourSecurePassword123!"
    username: driftguard
    password: "YourSecurePassword123!"
    database: driftguard
  architecture: replication
  readReplicas:
    replicaCount: 2
  primary:
    persistence:
      size: 50Gi
  metrics:
    enabled: true

# 监控配置
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s

# Grafana Dashboard
grafana:
  enabled: true
  adminPassword: "YourGrafanaPassword123!"
  dashboards:
    provider:
      enabled: true

# Ingress 配置
ingress:
  enabled: true
  className: nginx
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: driftguard.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: driftguard-tls
      hosts:
        - driftguard.example.com

# 持久化存储
persistence:
  enabled: true
  storageClass: "standard"
  size: 10Gi

# 节点选择
nodeSelector: {}

# 容忍度
tolerations: []

# 亲和性
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchLabels:
              app.kubernetes.io/name: driftguard
          topologyKey: kubernetes.io/hostname

# 探针配置
livenessProbe:
  httpGet:
    path: /api/v1/live
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /api/v1/ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3

# 环境变量
env:
  LOG_LEVEL: info
  LOG_FORMAT: json
  ENVIRONMENT: production
  DEBUG: "false"

# Secret 配置
secrets:
  databasePassword: "YourSecurePassword123!"
  tavilyApiKey: "your-tavily-api-key"
  confluenceCookie: "your-confluence-cookie"
```

---

## 📊 监控配置

### 安装 Prometheus Stack

```bash
# 添加 Helm Repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# 安装 kube-prometheus-stack
helm install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --values monitoring-values.yaml
```

### 监控配置 (monitoring-values.yaml)

```yaml
prometheus:
  prometheusSpec:
    additionalScrapeConfigs:
      - job_name: 'driftguard'
        kubernetes_sd_configs:
          - role: endpoints
            namespaces:
              names:
                - production
        relabel_configs:
          - source_labels: [__meta_kubernetes_service_label_app]
            action: keep
            regex: driftguard
          - source_labels: [__meta_kubernetes_endpoint_port_name]
            action: keep
            regex: metrics

grafana:
  enabled: true
  adminPassword: "YourGrafanaPassword123!"
  dashboardProviders:
    dashboardproviders.yaml:
      providers:
        - name: 'driftguard'
          orgId: 1
          folder: 'DriftGuard'
          type: file
          disableDeletion: false
          editable: true
          options:
            path: /var/lib/grafana/dashboards/driftguard
  dashboards:
    driftguard:
      driftguard-overview:
        json: |
          {
            "dashboard": {
              "title": "DriftGuard Overview",
              "panels": [...]
            }
          }
```

---

## 🔄 滚动更新

### 更新镜像版本

```bash
# 更新镜像
helm upgrade driftguard ./helm/driftguard \
  --namespace production \
  --set image.tag=v0.1.2 \
  --set image.pullPolicy=Always

# 查看更新状态
kubectl rollout status deployment/driftguard -n production

# 查看 Pod 状态
kubectl get pods -n production -w
```

### 回滚

```bash
# 查看历史版本
helm history driftguard -n production

# 回滚到上一版本
helm rollback driftguard -n production

# 回滚到指定版本
helm rollback driftguard 1 -n production
```

### 零停机发布

```yaml
# Deployment 配置
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  minReadySeconds: 30
```

---

## 📈 自动扩缩

### 配置 HPA

```bash
# 查看 HPA 状态
kubectl get hpa -n production

# 查看扩缩事件
kubectl get events -n production --field-selector reason=SuccessfulRescale
```

### 手动扩缩

```bash
# 扩容到 5 副本
kubectl scale deployment/driftguard --replicas=5 -n production

# 缩容到 2 副本
kubectl scale deployment/driftguard --replicas=2 -n production
```

---

## 🛡️ 高可用配置

### Pod Disruption Budget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: driftguard-pdb
  namespace: production
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: driftguard
```

### 多可用区部署

```yaml
# 节点亲和性
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            - key: topology.kubernetes.io/zone
              operator: In
              values:
                - us-east-1a
                - us-east-1b
                - us-east-1c

# Pod 反亲和性
podAntiAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchLabels:
            app.kubernetes.io/name: driftguard
        topologyKey: topology.kubernetes.io/zone
```

---

## 🔐 安全配置

### Secret 管理

```bash
# 创建 Secret
kubectl create secret generic driftguard-secrets \
  --from-literal=database-password='YourSecurePassword123!' \
  --from-literal=tavily-api-key='your-api-key' \
  -n production

# 查看 Secret
kubectl get secrets -n production

# 更新 Secret
kubectl create secret generic driftguard-secrets \
  --from-literal=database-password='NewSecurePassword456!' \
  --dry-run=client -o yaml | kubectl apply -f -
```

### Network Policy

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: driftguard-network-policy
  namespace: production
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: driftguard
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - protocol: TCP
          port: 8080
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: production
      ports:
        - protocol: TCP
          port: 5432
```

---

## 🧹 运维操作

### 查看日志

```bash
# 查看所有 Pod 日志
kubectl logs -l app.kubernetes.io/name=driftguard -n production -f

# 查看特定 Pod 日志
kubectl logs driftguard-7d9f8b6c5-abc12 -n production -f

# 查看最近 100 行
kubectl logs driftguard-7d9f8b6c5-abc12 -n production --tail=100
```

### 进入 Pod

```bash
# 进入容器
kubectl exec -it deployment/driftguard -n production -- sh

# 执行命令
kubectl exec deployment/driftguard -n production -- \
  curl http://localhost:8080/api/v1/health
```

### 备份数据库

```bash
# 备份
kubectl exec -it driftguard-postgresql-0 -n production -- \
  pg_dump -U driftguard driftguard > backup.sql

# 恢复
cat backup.sql | kubectl exec -i driftguard-postgresql-0 -n production -- \
  psql -U driftguard -d driftguard
```

---

## ⚠️ 注意事项

### 生产层要求

- ✅ 至少 3 个 Worker 节点
- ✅ 每个节点至少 4GB 内存
- ✅ 持久化存储 (StorageClass)
- ✅ Ingress 控制器
- ✅ 备份策略
- ✅ 监控告警

### 最佳实践

- ✅ 使用 Secret 管理敏感信息
- ✅ 配置资源限制和请求
- ✅ 启用 Pod Disruption Budget
- ✅ 配置多可用区部署
- ✅ 启用自动扩缩
- ✅ 配置滚动更新策略
- ✅ 定期备份数据库

---

## 📖 相关文档

- **[Helm Chart 使用](helm-chart.md)** - Helm 详细配置
- **[生产最佳实践](../production/best-practices.md)** - 生产环境优化
- **[监控告警](../production/monitoring.md)** - 监控配置
- **[扩缩容指南](../production/scaling.md)** - 自动扩缩容

---

*文档维护：DriftGuard Team*  
*最后更新：2026-03-14 07:35 UTC*
