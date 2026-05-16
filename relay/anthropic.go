package relay

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// AnthropicAdaptor Anthropic Claude适配器
type AnthropicAdaptor struct {
	BaseURL string
	APIKey  string
}

func (a *AnthropicAdaptor) ConvertRequest(body []byte) ([]byte, error) {
	// 将OpenAI格式转换为Anthropic Messages API格式
	var openaiReq map[string]interface{}
	if err := json.Unmarshal(body, &openaiReq); err != nil {
		return nil, err
	}

	anthropicReq := map[string]interface{}{
		"model":      openaiReq["model"],
		"max_tokens": 4096,
	}

	if mt, ok := openaiReq["max_tokens"]; ok {
		anthropicReq["max_tokens"] = mt
	}

	// 转换messages格式
	if msgs, ok := openaiReq["messages"].([]interface{}); ok {
		var converted []interface{}
		var systemPrompt string
		for _, m := range msgs {
			msg, _ := m.(map[string]interface{})
			role, _ := msg["role"].(string)
			content, _ := msg["content"].(string)
			if role == "system" {
				systemPrompt = content
				continue
			}
			converted = append(converted, map[string]interface{}{
				"role":    role,
				"content": content,
			})
		}
		anthropicReq["messages"] = converted
		if systemPrompt != "" {
			anthropicReq["system"] = systemPrompt
		}
	}

	return json.Marshal(anthropicReq)
}

func (a *AnthropicAdaptor) ConvertResponse(resp *http.Response) ([]byte, error) {
	// 将Anthropic响应转换回OpenAI格式
	var anthropicResp map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil { return nil, err }

	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return body, nil // 非JSON直接返回
	}

	// 构造OpenAI兼容响应
	openaiResp := map[string]interface{}{
		"id":      anthropicResp["id"],
		"object":  "chat.completion",
		"model":   anthropicResp["model"],
		"created": 0,
	}

	if content, ok := anthropicResp["content"].([]interface{}); ok && len(content) > 0 {
		if first, ok := content[0].(map[string]interface{}); ok {
			openaiResp["choices"] = []interface{}{
				map[string]interface{}{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": first["text"],
					},
					"finish_reason": "stop",
				},
			}
		}
	}

	if usage, ok := anthropicResp["usage"].(map[string]interface{}); ok {
		openaiResp["usage"] = map[string]interface{}{
			"prompt_tokens":     usage["input_tokens"],
			"completion_tokens": usage["output_tokens"],
			"total_tokens":      0,
		}
	}

	return json.Marshal(openaiResp)
}

func (a *AnthropicAdaptor) GetBaseURL() string { return a.BaseURL }
func (a *AnthropicAdaptor) GetAPIKey() string  { return a.APIKey }

func (a *AnthropicAdaptor) DoRequest(url_, method string, headers map[string]string, body io.Reader) (*http.Response, error) {
	if headers == nil { headers = make(map[string]string) }
	headers["x-api-key"] = a.APIKey
	headers["anthropic-version"] = "2023-06-01"
	headers["Content-Type"] = "application/json"
	return DoHTTPRequest(url_, method, headers, body)
}

// RelayAnthropic 转发Anthropic请求
func RelayAnthropic(baseURL, apiKey string, originalBody []byte) (*http.Response, []byte, error) {
	adaptor := &AnthropicAdaptor{BaseURL: baseURL, APIKey: apiKey}
	converted, err := adaptor.ConvertRequest(originalBody)
	if err != nil { return nil, nil, err }

	url := strings.TrimRight(baseURL, "/") + "/v1/messages"
	resp, err := adaptor.DoRequest(url, "POST", nil, bytes.NewReader(converted))
	if err != nil { return nil, nil, err }

	respBody, err := adaptor.ConvertResponse(resp)
	return resp, respBody, err
}
