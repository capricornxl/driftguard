# DriftGuard GitHub 仓库设置指南

## 📋 前提条件

- GitHub 账号
- GitHub CLI (`gh`) 已安装并认证

---

## 🚀 快速创建仓库

### 方式一：使用脚本 (推荐)

```bash
cd /root/.openclaw/workspace/driftguard

# 运行创建脚本
./scripts/create-github-repo.sh
```

脚本会自动:
1. 检查 GitHub CLI 安装和认证
2. 创建仓库 `driftguard/driftguard`
3. 推送代码到 main 分支
4. 配置分支保护
5. 显示下一步操作指引

### 方式二：手动创建

#### 1. 在 GitHub 创建仓库

访问：https://github.com/new

- **Repository name**: `driftguard`
- **Description**: `AI Agent Behavior Degradation Monitoring System`
- **Visibility**: Public
- **Initialize**: ❌ 不要勾选 (我们已有代码)

点击 **Create repository**

#### 2. 推送代码

```bash
cd /root/.openclaw/workspace/driftguard

# 添加远程仓库 (替换 YOUR_USERNAME)
git remote add origin https://github.com/driftguard/driftguard.git

# 推送代码
git branch -M main
git push -u origin main
```

---

## 🔧 后续配置

### 1. 启用 GitHub Actions

访问：https://github.com/driftguard/driftguard/actions

- 点击 "I understand my workflows, go ahead and enable them"
- CI 会自动运行

### 2. 配置 Docker Hub (用于自动发布镜像)

1. 创建 Docker Hub 账号：https://hub.docker.com
2. 创建 Access Token:
   - Account Settings > Security > New Access Token
   - 权限：Read, Write, Delete
3. 在 GitHub 仓库添加 Secrets:
   - Settings > Secrets and variables > Actions > New repository secret
   - `DOCKERHUB_USERNAME`: 你的 Docker Hub 用户名
   - `DOCKERHUB_TOKEN`: 刚创建的 Token

### 3. 添加 Topics

访问：https://github.com/driftguard/driftguard

点击齿轮图标⚙️，添加以下 topics:
```
ai monitoring agent prometheus grafana go kubernetes llm observability mlops
```

### 4. 配置分支保护

```bash
# 使用 GitHub CLI
gh api \
  --method PUT \
  /repos/driftguard/driftguard/branches/main/protection \
  --input - <<EOF
{
  "required_status_checks": {"strict": true, "contexts": ["test", "lint", "build"]},
  "enforce_admins": false,
  "required_pull_request_reviews": {
    "required_approving_review_count": 1,
    "dismiss_stale_reviews": true
  },
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false
}
EOF
```

或手动配置:
- Settings > Branches > Add branch protection rule
- Branch name pattern: `main`
- 勾选:
  - Require a pull request before merging
  - Require status checks to pass before merging
  - Don't allow bypassing the above requirements

### 5. 启用 Dependabot

创建 `.github/dependabot.yml`:

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
  - package-ecosystem: "docker"
    directory: "/"
    schedule:
      interval: "weekly"
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

---

## 📦 发布第一个版本

### 方式一：使用 GitHub CLI

```bash
# 创建标签并发布
gh release create v0.1.0 \
  --title "v0.1.0 MVP Release" \
  --notes "First stable release of DriftGuard MVP" \
  --generate-notes
```

### 方式二：使用 GitHub UI

1. 访问：https://github.com/driftguard/driftguard/releases
2. 点击 "Draft a new release"
3. Tag version: `v0.1.0`
4. Release title: `v0.1.0 MVP Release`
5. 点击 "Generate release notes"
6. 点击 "Publish release"

### 方式三：使用 Makefile

```bash
# 创建标签
git tag -a v0.1.0 -m "v0.1.0 MVP Release"
git push origin v0.1.0

# GitHub Actions 会自动创建 Release
```

---

## 🎯 验证清单

- [ ] 仓库创建成功
- [ ] 代码已推送 (main 分支)
- [ ] GitHub Actions 已启用
- [ ] CI 工作流运行通过
- [ ] Docker Hub Secrets 已配置
- [ ] Topics 已添加
- [ ] 分支保护已配置
- [ ] 第一个 Release 已发布

---

## 📞 仓库链接

- **代码**: https://github.com/driftguard/driftguard
- **Issues**: https://github.com/driftguard/driftguard/issues
- **Actions**: https://github.com/driftguard/driftguard/actions
- **Releases**: https://github.com/driftguard/driftguard/releases

---

## 🙋 遇到问题？

1. 查看 [CONTRIBUTING.md](CONTRIBUTING.md)
2. 创建 Issue: https://github.com/driftguard/driftguard/issues/new
