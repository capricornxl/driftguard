# DriftGuard API Documentation

## Overview

DriftGuard provides a RESTful API for monitoring AI agent behavior and detecting drift.

**Base URL**: `http://localhost:8080`

**API Version**: `v1`

## Quick Start

### Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2026-03-14T01:41:00Z",
  "version": "0.1.0",
  "uptime": "2h 15m 30s"
}
```

### List All Agents

```bash
curl http://localhost:8080/api/v1/agents
```

### Get Agent Details

```bash
curl http://localhost:8080/api/v1/agents/{agentId}
```

### Get Agent Health Score

```bash
curl http://localhost:8080/api/v1/agents/{agentId}/health
```

Response:
```json
{
  "agentId": "agent-1",
  "score": 85.5,
  "components": {
    "latency": 90.0,
    "efficiency": 85.0,
    "consistency": 88.0,
    "accuracy": 82.0,
    "hallucination": 5.0
  },
  "trend": "stable",
  "timestamp": "2026-03-14T01:41:00Z"
}
```

### List Alerts

```bash
curl http://localhost:8080/api/v1/alerts?severity=warning&status=active
```

### Prometheus Metrics

```bash
curl http://localhost:8080/metrics
```

## API Reference

Full API specification available in OpenAPI format:
- [openapi.yaml](./openapi.yaml)

### Interactive Documentation

When running with Swagger UI enabled:
```bash
curl http://localhost:8080/swagger/index.html
```

## Health Score Components

The overall health score (0-100) is calculated from:

| Component | Weight | Description |
|-----------|--------|-------------|
| Latency | 15% | Response time performance |
| Efficiency | 10% | Token usage efficiency |
| Consistency | 30% | Response variability |
| Accuracy | 35% | Output correctness |
| Hallucination | 10% | Hallucination rate (lower is better) |

## Alert Types

| Type | Description |
|------|-------------|
| `drift` | Significant behavior drift detected |
| `anomaly` | Statistical anomaly in metrics |
| `threshold` | Health score crossed threshold |
| `mutation` | Sudden behavioral mutation |

## Alert Severities

| Severity | Description | Action |
|----------|-------------|--------|
| `info` | Minor deviation | Monitor |
| `warning` | Moderate concern | Investigate |
| `critical` | Severe degradation | Immediate action |

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `OK` | 200 | Success |
| `BAD_REQUEST` | 400 | Invalid request |
| `NOT_FOUND` | 404 | Resource not found |
| `INTERNAL_ERROR` | 500 | Server error |
| `SERVICE_UNAVAILABLE` | 503 | Service unavailable |

## Rate Limiting

Rate limiting will be implemented in v0.2.0. Current limits:
- No limit (development)

## Authentication

Authentication will be implemented in v0.2.0. Planned methods:
- API Key
- JWT Token

## Examples

### Python Example

```python
import requests

BASE_URL = "http://localhost:8080"

# Get agent health
response = requests.get(f"{BASE_URL}/api/v1/agents/agent-1/health")
health = response.json()

print(f"Health Score: {health['score']}")
print(f"Trend: {health['trend']}")

# Check components
for component, score in health['components'].items():
    print(f"  {component}: {score}")
```

### JavaScript Example

```javascript
const BASE_URL = 'http://localhost:8080';

async function getAgentHealth(agentId) {
  const response = await fetch(`${BASE_URL}/api/v1/agents/${agentId}/health`);
  const health = await response.json();
  
  console.log(`Health Score: ${health.score}`);
  console.log(`Trend: ${health.trend}`);
  
  return health;
}

// Usage
getAgentHealth('agent-1');
```

## Support

- Documentation: https://github.com/capricornxl/driftguard
- Issues: https://github.com/capricornxl/driftguard/issues
- Email: support@driftguard.dev
