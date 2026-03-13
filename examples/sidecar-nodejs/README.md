# DriftGuard Sidecar Node.js 集成示例

## 快速开始

### 1. 安装依赖

```bash
npm install axios
# 或
yarn add axios
```

### 2. 基本用法

```javascript
const axios = require('axios');

const DRIFTGUARD_SIDECAR_URL = 'http://localhost:8081/collect';

async function trackInteraction(data) {
  try {
    await axios.post(DRIFTGUARD_SIDECAR_URL, {
      agent_id: data.agentId,
      session_id: data.sessionId,
      input: data.input,
      output: data.output,
      latency_ms: data.latencyMs,
      tokens_in: data.tokensIn,
      tokens_out: data.tokensOut
    }, {
      timeout: 5000
    });
    return true;
  } catch (error) {
    console.error('DriftGuard track failed:', error.message);
    return false;
  }
}

// 使用示例
async function main() {
  const startTime = Date.now();
  
  // 调用你的 LLM API
  const response = await callLLM('用户问题');
  
  const latencyMs = Date.now() - startTime;
  
  await trackInteraction({
    agentId: 'my-llm-agent',
    sessionId: 'session-123',
    input: '用户问题',
    output: response,
    latencyMs: latencyMs,
    tokensIn: 10,
    tokensOut: 50
  });
  
  console.log(`✅ Interaction tracked (latency: ${latencyMs}ms)`);
}
```

### 3. 完整类实现

```javascript
const axios = require('axios');

class DriftGuardTracker {
  constructor(options = {}) {
    this.agentId = options.agentId || 'default-agent';
    this.sessionId = options.sessionId || `session-${Date.now()}`;
    this.sidecarUrl = options.sidecarUrl || 'http://localhost:8081/collect';
    this.timeout = options.timeout || 5000;
    this.enabled = options.enabled !== false;
  }

  async track(input, output, latencyMs, tokensIn = null, tokensOut = null) {
    if (!this.enabled) return true;

    // 估算 tokens (如果没有提供)
    if (tokensIn === null) tokensIn = this.estimateTokens(input);
    if (tokensOut === null) tokensOut = this.estimateTokens(output);

    const payload = {
      agent_id: this.agentId,
      session_id: this.sessionId,
      input: input,
      output: output,
      latency_ms: latencyMs,
      tokens_in: tokensIn,
      tokens_out: tokensOut
    };

    try {
      await axios.post(this.sidecarUrl, payload, {
        timeout: this.timeout,
        headers: { 'Content-Type': 'application/json' }
      });
      return true;
    } catch (error) {
      console.error('DriftGuard track failed:', error.message);
      return false;
    }
  }

  estimateTokens(text) {
    // 简单估算：中文字符数/4 + 英文单词数
    const chineseChars = (text.match(/[\u4e00-\u9fff]/g) || []).length;
    const englishWords = text.split(/\s+/).filter(w => w.length > 0).length;
    return Math.max(1, Math.floor(chineseChars / 4) + englishWords);
  }

  newSession() {
    this.sessionId = `session-${Date.now()}`;
    return this.sessionId;
  }
}

// ============ 使用示例 ============

const tracker = new DriftGuardTracker({
  agentId: 'my-production-agent',
  sidecarUrl: 'http://localhost:8081/collect'
});

async function callLLMWithTracking(prompt) {
  const startTime = Date.now();
  
  // 调用 LLM
  const response = await callLLM(prompt);
  
  const latencyMs = Date.now() - startTime;
  
  // 上报到 DriftGuard
  await tracker.track(prompt, response, latencyMs);
  
  return response;
}
```

### 4. Express 中间件

```javascript
const express = require('express');
const axios = require('axios');

const app = express();
app.use(express.json());

const DRIFTGUARD_URL = 'http://localhost:8081/collect';
const AGENT_ID = 'express-llm-service';

// DriftGuard 追踪中间件
function driftGuardMiddleware(options = {}) {
  return async (req, res, next) => {
    // 只追踪 LLM 端点
    if (!req.path.startsWith('/api/llm')) {
      return next();
    }

    const startTime = Date.now();
    const sessionId = req.headers['x-session-id'] || `session-${Date.now()}`;
    const input = JSON.stringify(req.body).slice(0, 1000);

    // 拦截响应
    const originalJson = res.json;
    const originalSend = res.send;

    let responseBody = '';

    res.json = function(body) {
      responseBody = JSON.stringify(body);
      return originalJson.call(this, body);
    };

    res.send = function(body) {
      responseBody = String(body);
      return originalSend.call(this, body);
    };

    // 响应结束后上报
    res.on('finish', async () => {
      const latencyMs = Date.now() - startTime;
      
      try {
        await axios.post(DRIFTGUARD_URL, {
          agent_id: AGENT_ID,
          session_id: sessionId,
          input: input,
          output: responseBody.slice(0, 1000),
          latency_ms: latencyMs,
          tokens_in: input.length / 4,
          tokens_out: responseBody.length / 4
        }, { timeout: 5000 });
      } catch (error) {
        console.error('DriftGuard track failed:', error.message);
      }
    });

    next();
  };
}

// 使用中间件
app.use(driftGuardMiddleware());

// LLM 端点
app.post('/api/llm/chat', async (req, res) => {
  const { prompt } = req.body;
  
  // 调用 LLM...
  const response = { text: 'Hello!' };
  
  res.json(response);
});

app.listen(3000, () => {
  console.log('Server running on port 3000');
});
```

### 5. LangChain.js 集成

```javascript
const { ChatOpenAI } = require('@langchain/openai');
const { HumanMessage } = require('@langchain/core/messages');
const axios = require('axios');

class DriftGuardCallbackHandler {
  constructor(options = {}) {
    this.agentId = options.agentId || 'langchain-agent';
    this.sessionId = options.sessionId || `session-${Date.now()}`;
    this.sidecarUrl = options.sidecarUrl || 'http://localhost:8081/collect';
    this.startTime = null;
    this.promptText = '';
  }

  async handleLLMStart(llm, prompts) {
    this.startTime = Date.now();
    this.promptText = prompts[0] || '';
  }

  async handleLLMEnd(output) {
    const latencyMs = Date.now() - this.startTime;
    const completionText = output.generations[0][0].text;

    try {
      await axios.post(this.sidecarUrl, {
        agent_id: this.agentId,
        session_id: this.sessionId,
        input: this.promptText,
        output: completionText,
        latency_ms: latencyMs,
        tokens_in: Math.max(1, Math.floor(this.promptText.length / 4)),
        tokens_out: Math.max(1, Math.floor(completionText.length / 4))
      }, { timeout: 5000 });
    } catch (error) {
      console.error('DriftGuard callback failed:', error.message);
    }
  }

  async handleLLMError(error) {
    console.error('LLM error:', error.message);
  }
}

// ============ LangChain 使用示例 ============

const llm = new ChatOpenAI({
  modelName: 'gpt-3.5-turbo',
  callbacks: [new DriftGuardCallbackHandler({
    agentId: 'langchain-openai-agent'
  })]
});

async function chatWithTracking(message) {
  const response = await llm.invoke([new HumanMessage(message)]);
  return response.content;
}

// 使用
chatWithTracking('你好，请介绍一下自己').then(console.log);
```

### 6. NestJS 拦截器

```typescript
// driftguard.interceptor.ts
import {
  Injectable,
  NestInterceptor,
  ExecutionContext,
  CallHandler,
} from '@nestjs/common';
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import axios from 'axios';

@Injectable()
export class DriftGuardInterceptor implements NestInterceptor {
  private readonly sidecarUrl = 'http://localhost:8081/collect';
  private readonly agentId = 'nestjs-llm-service';

  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    const request = context.switchToHttp().getRequest();
    const response = context.switchToHttp().getResponse();

    // 只追踪 LLM 端点
    if (!request.url.startsWith('/api/llm')) {
      return next.handle();
    }

    const startTime = Date.now();
    const sessionId = request.headers['x-session-id'] || `session-${Date.now()}`;
    const input = JSON.stringify(request.body).slice(0, 1000);

    return next.handle().pipe(
      tap(async (responseBody) => {
        const latencyMs = Date.now() - startTime;
        const output = JSON.stringify(responseBody).slice(0, 1000);

        try {
          await axios.post(this.sidecarUrl, {
            agent_id: this.agentId,
            session_id: sessionId,
            input,
            output,
            latency_ms: latencyMs,
            tokens_in: Math.floor(input.length / 4),
            tokens_out: Math.floor(output.length / 4),
          }, { timeout: 5000 });
        } catch (error) {
          console.error('DriftGuard track failed:', error.message);
        }
      }),
    );
  }
}

// 使用
@UseInterceptors(DriftGuardInterceptor)
@Post('chat')
async chat(@Body() body: ChatDto) {
  // LLM 调用...
}
```

---

## 配置说明

### 环境变量

```bash
# DriftGuard Sidecar 地址
export DRIFTGUARD_SIDECAR_URL="http://localhost:8081/collect"

# Agent ID
export DRIFTGUARD_AGENT_ID="my-production-agent"

# 会话 ID (可选)
export DRIFTGUARD_SESSION_ID="session-123"

# 超时时间 (毫秒)
export DRIFTGUARD_TIMEOUT=5000
```

### 配置项

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `agentId` | string | 必填 | Agent 唯一标识 |
| `sessionId` | string | 自动生成 | 会话 ID |
| `sidecarUrl` | string | localhost:8081 | Sidecar 地址 |
| `timeout` | number | 5000 | 请求超时 (ms) |
| `enabled` | boolean | true | 是否启用 |

---

## 最佳实践

### 1. 异步上报

```typescript
// 使用 setImmediate 不阻塞主流程
setImmediate(async () => {
  try {
    await axios.post(DRIFTGUARD_URL, payload, { timeout: 5000 });
  } catch (error) {
    // 静默失败
  }
});
```

### 2. 批量上报

```javascript
class BatchTracker {
  constructor(batchSize = 10) {
    this.buffer = [];
    this.batchSize = batchSize;
  }

  track(data) {
    this.buffer.push(data);
    if (this.buffer.length >= this.batchSize) {
      this.flush();
    }
  }

  async flush() {
    if (this.buffer.length === 0) return;
    
    try {
      await axios.post(DRIFTGUARD_URL, { batch: this.buffer });
      this.buffer = [];
    } catch (error) {
      console.error('Batch track failed:', error.message);
    }
  }
}
```

### 3. 错误处理

```javascript
async function safeTrack(data) {
  try {
    await axios.post(DRIFTGUARD_URL, data, { timeout: 5000 });
    return true;
  } catch (error) {
    // 记录日志但不抛出
    console.warn('DriftGuard track failed:', error.message);
    return false;
  }
}
```

---

## 验证

```javascript
const axios = require('axios');

async function verify() {
  // 测试健康检查
  const health = await axios.get('http://localhost:8080/health');
  console.log('DriftGuard Health:', health.data);

  // 发送测试数据
  await axios.post('http://localhost:8081/collect', {
    agent_id: 'test-agent',
    session_id: 'test-session',
    input: 'test input',
    output: 'test output',
    latency_ms: 100,
    tokens_in: 5,
    tokens_out: 10
  });

  // 查看健康度
  const eval_result = await axios.post('http://localhost:8080/api/v1/agents/test-agent/evaluate');
  console.log('Health Score:', eval_result.data);
}

verify();
```
