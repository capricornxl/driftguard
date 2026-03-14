# Security Policy - DriftGuard

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

We take the security of DriftGuard seriously. If you believe you have found a security vulnerability, please report it to us as described below.

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to [support@driftguard.dev](mailto:support@driftguard.dev) or create a draft security advisory on GitHub.

### Preferred Languages

We prefer all communications to be in English.

### Response Timeline

- **48 hours**: Initial acknowledgment of your report
- **7 days**: Preliminary assessment and timeline for fix
- **30 days**: Target for patch release (for critical issues)

### What to Include

Please include the following information in your report:

- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if any)
- Your contact information for follow-up

## Security Best Practices

When deploying DriftGuard in production:

1. **Use Secrets Management**: Never commit database passwords or API keys to version control
2. **Enable TLS**: Always use HTTPS in production
3. **Restrict Access**: Use network policies to limit access to the API
4. **Regular Updates**: Keep DriftGuard and dependencies up to date
5. **Monitor Logs**: Enable logging and monitor for suspicious activity
6. **Least Privilege**: Run with minimal required permissions

## Security Features

DriftGuard includes the following security features:

- ✅ Non-root container user
- ✅ Read-only root filesystem
- ✅ No privilege escalation
- ✅ Database credentials in Kubernetes Secrets
- ✅ Input validation on all API endpoints
- ✅ Rate limiting on sensitive endpoints
- ✅ Audit logging for security events

## Acknowledgments

We thank the following for their responsible disclosure of security issues:

- (To be updated as disclosures are received)
