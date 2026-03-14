# Scripts

本目录存放项目相关的工具脚本。

## 分类

### 内部脚本 (不在 Git 中)

以下脚本存储在 `/root/.openclaw/workspace/learning/agent-drift-detection/scripts/`：

- `quick-acceptance-test.sh` - 快速验收测试 (25 项)
- `run-local.sh` - 本地运行脚本
- `start-test-server.sh` - 启动测试服务

这些脚本包含本地路径和配置，不应提交到 Git。

### 公共脚本 (在 Git 中)

可以存放在此目录的脚本类型：

- 构建辅助脚本
- 代码生成工具
- CI/CD 相关脚本
- 文档生成工具

## 规则

✅ 可以提交：通用的、不包含敏感信息的脚本
❌ 禁止提交：包含本地路径、API Key、密码的脚本

