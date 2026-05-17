package auto

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sumo1111/smart-gateway/common"
	"github.com/sumo1111/smart-gateway/model"
)

// speedTest 对指定(model, channel)发送真实Chat请求测速
// 返回：TTFT(首token延迟ms)、总耗时ms、TPS(tokens/s)、是否成功
func speedTest(modelName string, ch model.Channel) RaceResult {
	result := RaceResult{
		Model:     modelName,
		ChannelID: ch.ID,
		Channel:   ch.Name,
		TestedAt:  time.Now(),
	}

	// 构造测试请求 — 最小token的Chat请求
	payload := map[string]interface{}{
		"model": modelName,
		"messages": []map[string]string{
			{"role": "user", "content": common.RaceTestPrompt},
		},
		"max_tokens":  common.RaceTestMaxTokens,
		"temperature": 0.1, // 低温度保证一致性
		"stream":      false,
	}
	body, _ := json.Marshal(payload)

	url := strings.TrimRight(ch.BaseURL, "/") + "/v1/chat/completions"

	client := &http.Client{
		Timeout: time.Duration(common.RaceTestTimeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ch.APIKey)

	// Anthropic适配
	if ch.Type == "anthropic" {
		req.Header.Set("x-api-key", ch.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Del("Authorization")
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		result.Success = false
		result.ErrorMsg = fmt.Sprintf("request failed: %v", err)
		result.TTFT = int(elapsed.Milliseconds())
		result.TTFTTotal = int(elapsed.Milliseconds())
		return result
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		result.Success = false
		result.ErrorMsg = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 200))
		result.TTFT = int(elapsed.Milliseconds())
		result.TTFTTotal = int(elapsed.Milliseconds())
		return result
	}

	result.Success = true
	result.TTFTTotal = int(elapsed.Milliseconds())

	// 解析OpenAI格式响应，提取usage
	var chatResp struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		// 流式响应可能没有usage
	}

	if err := json.Unmarshal(respBody, &chatResp); err == nil && chatResp.Usage.CompletionTokens > 0 {
		// TPS = completion_tokens / 总耗时(秒)
		result.TPS = float64(chatResp.Usage.CompletionTokens) / (float64(elapsed.Milliseconds()) / 1000.0)
		// TTFT估算：总耗时 - 生成耗时，如果无法区分则用总耗时的60%估算
		// (非流式无法精确测TTFT，用经验公式：TTFT ≈ 总耗时 * 0.3 + 网络延迟)
		result.TTFT = int(float64(elapsed.Milliseconds()) * 0.3)
	} else {
		// 无法解析usage，用总耗时粗估
		result.TPS = float64(common.RaceTestMaxTokens) / (float64(elapsed.Milliseconds()) / 1000.0)
		result.TTFT = int(float64(elapsed.Milliseconds()) * 0.3)
	}

	return result
}

// SpeedTestSingle 手动触发单个竞速测试（管理API用）
func SpeedTestSingle(modelName string, ch *model.Channel) RaceResult {
	return speedTest(modelName, *ch)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
