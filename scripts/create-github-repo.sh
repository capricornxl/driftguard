#!/bin/bash

# DriftGuard GitHub 仓库创建脚本
# 用法：./scripts/create-github-repo.sh

set -e

REPO_OWNER="driftguard"
REPO_NAME="driftguard"
REPO_DESC="AI Agent Behavior Degradation Monitoring System - 非侵入式 AI Agent 行为退化监测系统"
REPO_VISIBILITY="public"

echo "========================================"
echo "DriftGuard GitHub 仓库创建"
echo "========================================"
echo ""

# 检查 GitHub CLI
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) 未安装"
    echo ""
    echo "请安装:"
    echo "  Ubuntu/Debian: sudo apt install gh"
    echo "  macOS: brew install gh"
    echo "  或访问：https://cli.github.com/"
    echo ""
    echo "然后认证:"
    echo "  gh auth login"
    echo ""
    exit 1
fi

# 检查认证状态
if ! gh auth status &> /dev/null; then
    echo "❌ GitHub 未认证"
    echo ""
    echo "请运行：gh auth login"
    echo ""
    exit 1
fi

echo "✅ GitHub CLI 已安装并认证"
echo ""

# 检查仓库是否已存在
echo "🔍 检查仓库是否存在..."
if gh repo view ${REPO_OWNER}/${REPO_NAME} &> /dev/null; then
    echo "⚠️  仓库 ${REPO_OWNER}/${REPO_NAME} 已存在"
    echo ""
    read -p "是否继续推送代码？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "已取消"
        exit 0
    fi
else
    echo "✅ 仓库不存在，将创建新仓库"
    echo ""
    
    # 创建仓库
    echo "📦 创建 GitHub 仓库..."
    gh repo create ${REPO_NAME} \
      --owner ${REPO_OWNER} \
      --description "${REPO_DESC}" \
      --visibility ${REPO_VISIBILITY} \
      --source=. \
      --remote=origin \
      --push
    
    echo ""
    echo "✅ 仓库创建成功!"
fi

echo ""
echo "========================================"
echo "仓库信息"
echo "========================================"
echo ""
echo "📁 仓库地址：https://github.com/${REPO_OWNER}/${REPO_NAME}"
echo "📊 克隆命令：git clone https://github.com/${REPO_OWNER}/${REPO_NAME}.git"
echo ""

# 创建主要分支保护 (需要 admin 权限)
echo "🔒 配置分支保护..."
gh api \
  --method PUT \
  /repos/${REPO_OWNER}/${REPO_NAME}/branches/main/protection \
  --input - <<EOF || echo "⚠️  分支保护配置失败 (可能需要 admin 权限)"
{
  "required_status_checks": null,
  "enforce_admins": false,
  "required_pull_request_reviews": null,
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false
}
EOF

echo ""
echo "========================================"
echo "下一步操作"
echo "========================================"
echo ""
echo "1. 启用 GitHub Actions:"
echo "   https://github.com/${REPO_OWNER}/${REPO_NAME}/actions"
echo ""
echo "2. 配置 GitHub Pages (可选):"
echo "   Settings > Pages > Source: main branch /docs"
echo ""
echo "3. 添加 Topics:"
echo "   ai monitoring agent prometheus grafana go kubernetes llm"
echo ""
echo "4. 创建第一个 Release:"
echo "   gh release create v0.1.0 --title 'v0.1.0 MVP' --generate-notes"
echo ""
echo "5. 邀请协作者 (可选):"
echo "   gh repo collaborator add USERNAME"
echo ""
echo "========================================"
