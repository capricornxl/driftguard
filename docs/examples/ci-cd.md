# CI/CD 集成指南

**适用人群**: DevOps 工程师、开发团队  
**预计时间**: 30 分钟  
**前置条件**: GitHub 账号、Kubernetes 集群 (可选)

---

## 🎯 本指南将带你

- ✅ 配置 GitHub Actions CI/CD 流水线
- ✅ 自动化构建、测试和部署
- ✅ 集成 Docker 镜像推送
- ✅ 集成 Helm Chart 发布
- ✅ 实现多环境部署

---

## 📦 CI/CD 架构

```
┌─────────────────────────────────────────────────────────┐
│              GitHub Actions CI/CD Pipeline               │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Push/PR → Lint → Test → Build → Docker → Helm → Deploy│
│                                                         │
│     │        │       │       │        │       │        │
│     ▼        ▼       ▼       ▼        ▼       ▼        ▼
│  代码质量  单元测试  构建    镜像     Chart  部署到    │
│  检查             二进制  推送     推送    环境       │
│                                                         │
│  环境流程:                                               │
│  main 分支 → Staging 环境 (自动)                        │
│  v* Tag    → Production 环境 (自动)                      │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## 🚀 快速开始

### 步骤 1: 复制 CI/CD 配置文件

```bash
# 复制工作流文件
cp .github/workflows/ci-cd.yml.example .github/workflows/ci-cd.yml

# 提交到仓库
git add .github/workflows/ci-cd.yml
git commit -m "ci: add CI/CD pipeline"
git push
```

---

### 步骤 2: 配置 GitHub Secrets

在 GitHub 仓库设置中添加以下 Secrets：

**导航到**: Settings → Secrets and variables → Actions

#### 必需 Secrets

| Secret 名称 | 说明 | 示例 |
|------------|------|------|
| `KUBE_CONFIG_STAGING` | Staging 环境 Kubeconfig | `base64 kubeconfig` |
| `KUBE_CONFIG_PRODUCTION` | Production 环境 Kubeconfig | `base64 kubeconfig` |

#### 可选 Secrets

| Secret 名称 | 说明 |
|------------|------|
| `DOCKER_USERNAME` | Docker Hub 用户名 |
| `DOCKER_PASSWORD` | Docker Hub 密码 |
| `SLACK_WEBHOOK` | Slack 通知 Webhook |

---

### 步骤 3: 准备 Kubeconfig

```bash
# 创建 Kubeconfig 文件
kubectl config view --raw --minify > kubeconfig-staging

# Base64 编码
cat kubeconfig-staging | base64 -w 0

# 复制输出到 GitHub Secrets (KUBE_CONFIG_STAGING)
```

**生产环境同理**:

```bash
kubectl config view --raw --minify > kubeconfig-production
cat kubeconfig-production | base64 -w 0
```

---

## 📋 工作流详解

### 1. 代码质量检查 (Lint)

```yaml
name: Code Quality
runs-on: ubuntu-latest
steps:
  - golangci-lint        # 代码风格检查
  - go vet              # Go 内置检查
  - go mod tidy         # 依赖检查
```

**触发条件**: Push/PR  
**预计时间**: 2-3 分钟

---

### 2. 单元测试 (Test)

```yaml
name: Unit Tests
runs-on: ubuntu-latest
steps:
  - go test -v -race    # 运行测试
  - go tool cover       # 生成覆盖率报告
  - codecov             # 上传到 Codecov
```

**触发条件**: Push/PR  
**预计时间**: 5-10 分钟  
**输出**: 覆盖率报告、HTML 报告

---

### 3. 构建测试 (Build)

```yaml
name: Build
runs-on: ubuntu-latest
steps:
  - CGO_ENABLED=0       # 静态编译
  - GOOS=linux          # Linux 目标
  - go build            # 构建二进制
```

**触发条件**: Push/PR  
**预计时间**: 3-5 分钟  
**输出**: 可执行文件

---

### 4. Docker 镜像构建和推送

```yaml
name: Docker Build & Push
runs-on: ubuntu-latest
steps:
  - docker/setup-buildx     # 设置 Buildx
  - docker/login-action     # 登录 Registry
  - docker/build-push-action # 构建并推送
```

**触发条件**: Push 到 main/develop  
**预计时间**: 10-15 分钟  
**输出**: 多架构 Docker 镜像

**镜像标签策略**:

| 事件 | 标签示例 |
|------|----------|
| Push to main | `sha-abc1234`, `main` |
| Push to develop | `develop` |
| Tag v0.1.1 | `v0.1.1`, `0.1`, `latest` |

---

### 5. Helm Chart 验证和推送

```yaml
name: Helm Chart
runs-on: ubuntu-latest
steps:
  - ct lint             # Chart 语法检查
  - helm package        # 打包 Chart
  - helm push           # 推送到 GHCR
```

**触发条件**: Tag 推送 (v*)  
**预计时间**: 3-5 分钟  
**输出**: Helm Chart OCI 镜像

---

### 6. 部署到 Staging

```yaml
name: Deploy to Staging
runs-on: ubuntu-latest
steps:
  - helm upgrade --install  # 部署
  - kubectl wait           # 等待就绪
  - curl health            # 冒烟测试
```

**触发条件**: Push 到 main  
**环境**: Staging  
**预计时间**: 5-10 分钟

---

### 7. 部署到 Production

```yaml
name: Deploy to Production
runs-on: ubuntu-latest
steps:
  - helm upgrade --install  # 部署
  - kubectl wait           # 等待就绪
  - curl health            # 冒烟测试
  - gh release             # 创建 Release
```

**触发条件**: Tag 推送 (v*)  
**环境**: Production  
**预计时间**: 10-15 分钟

---

## 🔧 自定义配置

### 添加新环境

在 `.github/workflows/ci-cd.yml` 中添加：

```yaml
deploy-custom:
  name: Deploy to Custom
  runs-on: ubuntu-latest
  needs: deploy-staging
  if: github.ref == 'refs/heads/feature-x'
  environment:
    name: custom
    url: https://driftguard-custom.example.com
  steps:
    - name: Configure kubectl
      run: |
        echo "${{ secrets.KUBE_CONFIG_CUSTOM }}" | base64 -d > kubeconfig
        export KUBECONFIG=kubeconfig

    - name: Deploy
      run: |
        export KUBECONFIG=kubeconfig
        helm upgrade --install driftguard ./helm/driftguard \
          --namespace custom \
          --values values-custom.yaml
```

---

### 添加通知

#### Slack 通知

```yaml
- name: Notify Slack
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    text: |
      CI/CD Pipeline Status: ${{ job.status }}
      Branch: ${{ github.ref }}
      Commit: ${{ github.sha }}
  env:
    SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
  if: always()
```

#### 邮件通知

```yaml
- name: Send Email
  uses: dawidd6/action-send-mail@v3
  with:
    server_address: smtp.gmail.com
    server_port: 587
    username: ${{ secrets.EMAIL_USERNAME }}
    password: ${{ secrets.EMAIL_PASSWORD }}
    subject: CI/CD Pipeline - ${{ job.status }}
    body: |
      Pipeline completed with status: ${{ job.status }}
      Branch: ${{ github.ref }}
      Commit: ${{ github.sha }}
    to: team@example.com
    from: CI/CD Bot
  if: failure()
```

---

### 添加性能测试

```yaml
performance-test:
  name: Performance Tests
  runs-on: ubuntu-latest
  needs: deploy-staging
  steps:
    - name: Run k6 performance tests
      uses: grafana/k6-action@v0.3.0
      with:
        filename: tests/performance/load-test.js
        options: --out json=results.json

    - name: Upload results
      uses: actions/upload-artifact@v4
      with:
        name: performance-results
        path: results.json
```

---

## 📊 监控和告警

### 查看工作流状态

```bash
# GitHub CLI
gh run list
gh run view <run-id>
gh run watch <run-id>
```

### 设置告警规则

在 GitHub 仓库设置中：

1. Settings → Actions → General
2. Workflow permissions → Read and write permissions
3. 启用 "Allow GitHub Actions to create and approve pull requests"

---

## 🔄 手动触发部署

### 手动部署到 Staging

创建 `.github/workflows/manual-deploy-staging.yml`:

```yaml
name: Manual Deploy Staging

on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to deploy'
        required: true
        default: 'latest'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to staging
        run: |
          helm upgrade --install driftguard ./helm/driftguard \
            --namespace staging \
            --set image.tag=${{ github.event.inputs.version }}
```

**手动触发**: Actions → Manual Deploy Staging → Run workflow

---

## 🛡️ 安全最佳实践

### 1. 使用 OIDC 代替长期凭证

```yaml
- name: Configure AWS credentials
  uses: aws-actions/configure-aws-credentials@v4
  with:
    role-to-assume: arn:aws:iam::123456789:role/github-actions-role
    aws-region: us-east-1
```

### 2. 限制 Secret 访问

```yaml
# 只在特定环境使用 Secret
env:
  KUBE_CONFIG: ${{ secrets.KUBE_CONFIG_PRODUCTION }}
# 添加环境保护规则
environment: production
```

### 3. 启用分支保护

Settings → Branches → Branch protection rules:

- Require a pull request before merging
- Require status checks to pass before merging
- Require branches to be up to date before merging
- Require conversation resolution before merging

---

## 📖 相关文档

- **[GitHub Actions 文档](https://docs.github.com/en/actions)**
- **[Helm Chart 使用](../kubernetes/helm-chart.md)**
- **[Docker Compose 部署](../deployment/docker-compose.md)**
- **[生产最佳实践](../production/best-practices.md)**

---

## 🆘 故障排查

### 工作流失败

```bash
# 查看详细日志
gh run view <run-id> --log

# 重新运行失败的任务
gh run rerun <run-id> --failed
```

### 部署失败

```bash
# 查看 Kubernetes 事件
kubectl get events -n <namespace> --sort-by='.lastTimestamp'

# 查看 Pod 日志
kubectl logs -f deployment/driftguard -n <namespace>

# 回滚部署
helm rollback driftguard -n <namespace>
```

---

*文档维护：DriftGuard Team*  
*最后更新：2026-03-14 07:45 UTC*
