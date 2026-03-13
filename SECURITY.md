# Security Policy

## 支持版本

| 版本 | 支持状态 |
|------|----------|
| v0.1.x | 🟢 支持中 |

## 报告漏洞

我们非常重视 DriftGuard 的安全性。如果您发现任何安全漏洞，请负责任地披露。

### 如何报告

**请不要**直接创建公开 Issue 报告安全问题。

请通过以下方式之一报告:

1. **GitHub Private Vulnerability Reporting** (推荐)
   - 访问：https://github.com/driftguard/driftguard/security/advisories
   - 点击 "Report a vulnerability"
   - 填写详细信息

2. **邮件联系**
   - 发送邮件到：security@driftguard.dev (待配置)

### 报告内容

请尽可能提供详细信息:

- 漏洞类型和描述
- 复现步骤
- 影响范围
- 建议的修复方案 (如有)

### 响应时间

- **确认收到**: 48 小时内
- **状态更新**: 每周
- **修复目标**: 根据严重程度
  - Critical: 7 天
  - High: 14 天
  - Medium: 30 天
  - Low: 60 天

## 安全最佳实践

### 部署建议

1. **使用最新稳定版本**
   - 定期更新到最新版本
   - 订阅 Release 通知

2. **配置认证**
   - 修改默认密码 (Grafana: admin/driftguard)
   - 使用强密码策略
   - 启用双因素认证 (如支持)

3. **网络安全**
   - 仅暴露必要端口
   - 使用防火墙限制访问
   - 生产环境使用 HTTPS

4. **数据保护**
   - 定期备份数据库
   - 加密敏感配置
   - 使用环境变量存储密钥

5. **监控日志**
   - 启用审计日志
   - 定期检查异常访问
   - 配置告警通知

## 已知限制

- v0.1.x: 初始版本，功能有限
- 不支持多租户隔离
- Webhook 通知未加密

## 致谢

感谢为 DriftGuard 安全做出贡献的所有人!
