# Contributing to DriftGuard

首先，感谢你考虑为 DriftGuard 做出贡献！

## 📋 目录

- [行为准则](#行为准则)
- [如何贡献](#如何贡献)
- [开发环境设置](#开发环境设置)
- [提交指南](#提交指南)
- [Pull Request 流程](#pull-request-流程)

---

## 行为准则

本项目采用 [Contributor Covenant](https://www.contributor-covenant.org/) 行为准则。请尊重所有贡献者和用户。

---

## 如何贡献

### 报告 Bug

1. 搜索现有 [Issues](https://github.com/driftguard/driftguard/issues) 确认是否已报告
2. 如果没有，创建新 Issue 并包含：
   - 清晰的标题和描述
   - 复现步骤
   - 预期行为 vs 实际行为
   - 环境信息 (Go 版本、OS 等)
   - 日志/错误信息

### 提出新功能

1. 先创建 Issue 讨论功能设计
2. 说明使用场景和必要性
3. 等待维护者确认后再开始开发

### 提交代码

1. Fork 仓库
2. 创建特性分支
3. 提交变更
4. 推送到分支
5. 创建 Pull Request

---

## 开发环境设置

### 前置要求

- Go 1.21+
- Docker & Docker Compose
- Git

### 安装步骤

```bash
# 1. Fork 并克隆
git clone https://github.com/YOUR_USERNAME/driftguard.git
cd driftguard

# 2. 安装依赖
go mod download

# 3. 启动开发环境
docker compose up -d

# 4. 运行测试
go test ./... -v

# 5. 本地运行
go run cmd/main.go -config config.json
```

### 验证变更

```bash
# 运行所有测试
go test ./... -v -race

# 运行集成测试
./tests/integration-test.sh

# 代码格式化
go fmt ./...
go vet ./...
```

---

## 提交指南

### Commit Message 格式

遵循 [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Type 类型

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式 (不影响功能)
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具/配置

### 示例

```
feat(collector): add batch processing support

- Implement batch size configuration
- Add flush interval timer
- Update unit tests

Closes #123
```

```
fix(detector): correct threshold calculation

- Fix off-by-one error in trend detection
- Add edge case tests

Fixes #456
```

---

## Pull Request 流程

### 创建 PR

1. 确保代码通过所有测试
2. 更新文档 (如需要)
3. 添加/更新测试用例
4. 创建 PR 到 `main` 分支
5. 填写 PR 模板

### PR 模板

```markdown
## 描述
简要说明此 PR 的目的

## 相关 Issue
Closes #123

## 变更类型
- [ ] 新功能 (feat)
- [ ] Bug 修复 (fix)
- [ ] 文档更新 (docs)
- [ ] 重构 (refactor)
- [ ] 其他

## 测试
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 手动测试完成

## 截图 (如适用)
```

### 审核流程

1. 维护者会进行代码审查
2. 可能需要修改以满足项目标准
3. 审核通过后合并到 `main` 分支

---

## 代码标准

### Go 代码

- 遵循 [Effective Go](https://golang.org/doc/effective_go)
- 使用 `go fmt` 格式化
- 使用 `go vet` 检查
- 添加有意义的注释

### 测试

- 新功能必须包含测试
- 单元测试覆盖率 > 80%
- 集成测试验证端到端流程

### 文档

- 公共 API 必须有文档注释
- 更新 README (如影响用户使用)
- 添加示例代码 (如适用)

---

## 问题？

- 查看 [README](README.md)
- 阅读 [文档](docs/)
- 创建 [Issue](https://github.com/driftguard/driftguard/issues)

---

再次感谢你的贡献！🎉
