// Sidecar 集成示例 - 展示如何将 DriftGuard 集成到现有 Agent 系统

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// DriftGuardClient DriftGuard 客户端
type DriftGuardClient struct {
	endpoint string
	agentID  string
}

// NewDriftGuardClient 创建客户端
func NewDriftGuardClient(endpoint, agentID string) *DriftGuardClient {
	return &DriftGuardClient{
		endpoint: endpoint,
		agentID:  agentID,
	}
}

// Interaction 交互记录
type Interaction struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
	Input     string `json:"input"`
	Output    string `json:"output"`
	LatencyMs int    `json:"latency_ms"`
	TokensIn  int    `json:"tokens_in"`
	TokensOut int    `json:"tokens_out"`
}

// Report 上报交互数据
func (c *DriftGuardClient) Report(sessionID, input, output string, latencyMs, tokensIn, tokensOut int) error {
	interaction := Interaction{
		AgentID:   c.agentID,
		SessionID: sessionID,
		Input:     input,
		Output:    output,
		LatencyMs: latencyMs,
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
	}

	jsonData, err := json.Marshal(interaction)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.endpoint+"/api/v1/interactions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return &httpError{resp.StatusCode}
	}

	return nil
}

type httpError struct {
	code int
}

func (e *httpError) Error() string {
	return "http error: " + string(rune(e.code))
}

// 使用示例
func main() {
	client := NewDriftGuardClient("http://localhost:8080", "my-agent-001")

	// 模拟一次 Agent 交互
	startTime := time.Now()
	
	// ... 这里是你的 Agent 处理逻辑 ...
	input := "你好，请介绍一下自己"
	output := "你好！我是 AI 助手..."
	
	latencyMs := int(time.Since(startTime).Milliseconds())
	tokensIn := 10
	tokensOut := 50

	if err := client.Report("session-123", input, output, latencyMs, tokensIn, tokensOut); err != nil {
		log.Printf("Failed to report to DriftGuard: %v", err)
	} else {
		log.Println("Successfully reported to DriftGuard")
	}
}
