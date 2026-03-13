# DriftGuard Sidecar Python 集成示例

## 快速开始

### 1. 安装依赖

```bash
pip install requests
```

### 2. 基本用法

```python
import requests
import time

DRIFTGUARD_SIDECAR_URL = "http://localhost:8081/collect"

def track_interaction(agent_id, session_id, input_text, output_text, latency_ms, tokens_in, tokens_out):
    """
    上报 Agent 交互到 DriftGuard
    
    Args:
        agent_id: Agent 唯一标识
        session_id: 会话 ID
        input_text: 用户输入
        output_text: Agent 输出
        latency_ms: 响应延迟 (毫秒)
        tokens_in: 输入 tokens 数
        tokens_out: 输出 tokens 数
    """
    try:
        response = requests.post(
            DRIFTGUARD_SIDECAR_URL,
            json={
                "agent_id": agent_id,
                "session_id": session_id,
                "input": input_text,
                "output": output_text,
                "latency_ms": latency_ms,
                "tokens_in": tokens_in,
                "tokens_out": tokens_out
            },
            timeout=5
        )
        response.raise_for_status()
        return True
    except Exception as e:
        print(f"Failed to track interaction: {e}")
        return False
```

### 3. 完整示例

```python
import requests
import time
import json

class DriftGuardTracker:
    def __init__(self, agent_id, sidecar_url="http://localhost:8081/collect"):
        self.agent_id = agent_id
        self.sidecar_url = sidecar_url
        self.session_id = f"session-{int(time.time())}"
    
    def track(self, input_text, output_text, latency_ms, tokens_in=None, tokens_out=None):
        """上报单次交互"""
        # 估算 tokens (简单规则：中文字符数/4 + 英文单词数)
        if tokens_in is None:
            tokens_in = self._estimate_tokens(input_text)
        if tokens_out is None:
            tokens_out = self._estimate_tokens(output_text)
        
        payload = {
            "agent_id": self.agent_id,
            "session_id": self.session_id,
            "input": input_text,
            "output": output_text,
            "latency_ms": latency_ms,
            "tokens_in": tokens_in,
            "tokens_out": tokens_out
        }
        
        try:
            requests.post(self.sidecar_url, json=payload, timeout=5)
            return True
        except Exception as e:
            print(f"DriftGuard track failed: {e}")
            return False
    
    def _estimate_tokens(self, text):
        """简单估算 tokens 数"""
        chinese_chars = sum(1 for c in text if '\u4e00' <= c <= '\u9fff')
        english_words = len(text.split())
        return max(1, chinese_chars // 4 + english_words)


# ============ 使用示例 ============

if __name__ == "__main__":
    # 初始化追踪器
    tracker = DriftGuardTracker(agent_id="my-llm-agent")
    
    # 模拟 LLM 调用
    start_time = time.time()
    
    # 调用你的 LLM API
    # response = call_llm_api("用户问题")
    response = "这是一个示例回复"
    
    latency_ms = int((time.time() - start_time) * 1000)
    
    # 上报到 DriftGuard
    tracker.track(
        input_text="用户问题",
        output_text=response,
        latency_ms=latency_ms
    )
    
    print(f"✅ Interaction tracked (latency: {latency_ms}ms)")
```

### 4. LangChain 集成

```python
from langchain.callbacks.base import BaseCallbackHandler
import requests
import time

class DriftGuardCallback(BaseCallbackHandler):
    """LangChain Callback Handler for DriftGuard"""
    
    def __init__(self, agent_id, session_id=None, sidecar_url="http://localhost:8081/collect"):
        self.agent_id = agent_id
        self.session_id = session_id or f"session-{int(time.time())}"
        self.sidecar_url = sidecar_url
        self.start_time = None
        self.prompt_text = ""
        self.completion_text = ""
    
    def on_llm_start(self, serialized, prompts, **kwargs):
        """LLM 开始调用"""
        self.start_time = time.time()
        self.prompt_text = prompts[0] if prompts else ""
    
    def on_llm_end(self, response, **kwargs):
        """LLM 结束调用"""
        latency_ms = int((time.time() - self.start_time) * 1000)
        self.completion_text = response.generations[0][0].text
        
        # 估算 tokens
        tokens_in = len(self.prompt_text) // 4
        tokens_out = len(self.completion_text) // 4
        
        # 上报到 DriftGuard
        try:
            requests.post(self.sidecar_url, json={
                "agent_id": self.agent_id,
                "session_id": self.session_id,
                "input": self.prompt_text,
                "output": self.completion_text,
                "latency_ms": latency_ms,
                "tokens_in": max(1, tokens_in),
                "tokens_out": max(1, tokens_out)
            }, timeout=5)
        except Exception as e:
            print(f"DriftGuard callback failed: {e}")
    
    def on_llm_error(self, error, **kwargs):
        """LLM 调用错误"""
        print(f"LLM error: {error}")


# ============ LangChain 使用示例 ============

from langchain.chat_models import ChatOpenAI
from langchain.chains import LLMChain
from langchain.prompts import ChatPromptTemplate

# 创建带 DriftGuard 监控的 LLM
llm = ChatOpenAI(
    model="gpt-3.5-turbo",
    callbacks=[DriftGuardCallback(agent_id="langchain-openai-agent")]
)

# 创建 Chain
prompt = ChatPromptTemplate.from_messages([
    ("human", "{input}")
])
chain = LLMChain(llm=llm, prompt=prompt)

# 运行 (自动监控)
result = chain.run("你好，请介绍一下自己")
print(result)
```

### 5. FastAPI 中间件

```python
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
import time
import requests

app = FastAPI()

DRIFTGUARD_URL = "http://localhost:8081/collect"
AGENT_ID = "fastapi-llm-service"

@app.middleware("http")
async def track_llm_requests(request: Request, call_next):
    """追踪所有 LLM 相关的 API 请求"""
    
    # 只追踪 LLM 端点
    if not request.url.path.startswith("/api/llm"):
        return await call_next(request)
    
    # 记录开始时间
    start_time = time.time()
    
    # 读取请求体
    body = await request.body()
    
    # 继续处理请求
    response = await call_next(request)
    
    # 计算延迟
    latency_ms = int((time.time() - start_time) * 1000)
    
    # 异步上报到 DriftGuard (不阻塞响应)
    try:
        import asyncio
        asyncio.create_task(
            requests.post(
                DRIFTGUARD_URL,
                json={
                    "agent_id": AGENT_ID,
                    "session_id": request.headers.get("X-Session-ID", "unknown"),
                    "input": body.decode()[:1000],  # 限制长度
                    "output": str(response.status_code),
                    "latency_ms": latency_ms,
                    "tokens_in": len(body) // 4,
                    "tokens_out": 0
                },
                timeout=5
            )
        )
    except:
        pass
    
    return response


@app.post("/api/llm/chat")
async def chat(request: Request):
    data = await request.json()
    # 调用 LLM...
    return {"response": "Hello"}
```

---

## 配置说明

### 环境变量

```bash
# DriftGuard Sidecar 地址
export DRIFTGUARD_SIDECAR_URL="http://localhost:8081/collect"

# Agent ID (唯一标识)
export DRIFTGUARD_AGENT_ID="my-production-agent"

# 会话 ID (可选，默认自动生成)
export DRIFTGUARD_SESSION_ID="session-123"
```

### 配置项

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `agent_id` | string | 必填 | Agent 唯一标识 |
| `session_id` | string | 自动生成 | 会话 ID |
| `sidecar_url` | string | localhost:8081 | Sidecar 地址 |
| `timeout` | int | 5 秒 | 请求超时 |

---

## 最佳实践

### 1. 异步上报

```python
import asyncio
import aiohttp

async def track_async(payload):
    """异步上报，不阻塞主流程"""
    async with aiohttp.ClientSession() as session:
        try:
            await session.post(DRIFTGUARD_URL, json=payload, timeout=5)
        except:
            pass  # 失败静默，不影响主流程
```

### 2. 批量上报

```python
class BatchTracker:
    def __init__(self, batch_size=10):
        self.buffer = []
        self.batch_size = batch_size
    
    def track(self, **kwargs):
        self.buffer.append(kwargs)
        if len(self.buffer) >= self.batch_size:
            self.flush()
    
    def flush(self):
        if self.buffer:
            requests.post(DRIFTGUARD_URL, json={"batch": self.buffer})
            self.buffer = []
```

### 3. 错误处理

```python
def safe_track(**kwargs):
    """安全的上报函数，失败不影响主流程"""
    try:
        requests.post(DRIFTGUARD_URL, json=kwargs, timeout=5)
    except Exception as e:
        # 记录日志但不抛出异常
        logging.warning(f"DriftGuard track failed: {e}")
```

---

## 验证

```python
# 测试连接
import requests

response = requests.get("http://localhost:8080/health")
print(f"DriftGuard Health: {response.json()}")

# 发送测试数据
requests.post("http://localhost:8081/collect", json={
    "agent_id": "test-agent",
    "session_id": "test-session",
    "input": "test",
    "output": "test output",
    "latency_ms": 100,
    "tokens_in": 5,
    "tokens_out": 10
})

# 查看健康度
response = requests.post("http://localhost:8080/api/v1/agents/test-agent/evaluate")
print(f"Health Score: {response.json()}")
```
