package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sumo1111/smart-gateway/auto"
	"github.com/sumo1111/smart-gateway/common"
	"github.com/sumo1111/smart-gateway/model"
	"github.com/sumo1111/smart-gateway/relay"
)

// Relay 转发Chat Completions请求
func Relay(c *gin.Context) {
	startTime := time.Now()

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}
	c.Request.Body.Close()

	// 解析请求获取模型名
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	requestModel, _ := req["model"].(string)
	if requestModel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	// 获取认证token
	tokenObj, _ := c.Get("token")
	token := tokenObj.(*model.Token)

	// 确定实际模型和渠道
	var selectedModel string
	var selectedChannel *model.Channel

	if requestModel == "auto" {
		// Auto模式：智能选择模型+渠道
		m, ch, err := auto.SelectAutoModel()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no available channel for auto routing"})
			return
		}
		selectedModel = m
		selectedChannel = ch
		req["model"] = selectedModel
		body, _ = json.Marshal(req)
	} else {
		// 指定模型：选择该模型的最优渠道
		channels, err := model.GetChannelsByModel(requestModel)
		if err != nil || len(channels) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("no channel available for model: %s", requestModel)})
			return
		}
		ch, err := auto.SelectChannel(requestModel, channels)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "all channels for this model are temporarily unavailable"})
			return
		}
		selectedModel = requestModel
		selectedChannel = ch
	}

	log.Printf("Routing: model=%s, channel=%s(%d), strategy=%s", selectedModel, selectedChannel.Name, selectedChannel.ID, common.AutoStrategy)

	// 根据渠道类型选择适配器
	var resp *http.Response
	var respBody []byte

	switch selectedChannel.Type {
	case "anthropic":
		resp, respBody, err = relay.RelayAnthropic(selectedChannel.BaseURL, selectedChannel.APIKey, body)
	default: // openai, custom
		path := c.Request.URL.Path
		resp, err = relay.RelayOpenAI(selectedChannel.BaseURL, selectedChannel.APIKey, path, c.Request.Method, bytes.NewReader(body))
	}

	latency := int(time.Since(startTime).Milliseconds())

	if err != nil {
		// 记录失败
		model.RecordFailure(selectedModel, selectedChannel.ID)
		model.CheckAndBan(selectedModel, selectedChannel.ID, common.FailBanCount, common.FailBanDuration)
		model.InsertLog(&model.Log{
			TokenID: token.ID, ChannelID: selectedChannel.ID, Model: selectedModel,
			LatencyMs: latency, Status: "fail", ErrorMsg: err.Error(),
		})
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("upstream error: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errBody, _ := io.ReadAll(resp.Body)
		errMsg := string(errBody)
		if len(errMsg) > 500 { errMsg = errMsg[:500] }
		model.RecordFailure(selectedModel, selectedChannel.ID)
		model.CheckAndBan(selectedModel, selectedChannel.ID, common.FailBanCount, common.FailBanDuration)
		model.InsertLog(&model.Log{
			TokenID: token.ID, ChannelID: selectedChannel.ID, Model: selectedModel,
			LatencyMs: latency, Status: "fail", ErrorMsg: errMsg,
		})
		// 尝试fallback
		if shouldRetry(resp.StatusCode) {
			tryFallback(c, selectedModel, selectedChannel.ID, body, token, startTime)
			return
		}
		c.JSON(resp.StatusCode, gin.H{"error": errMsg})
		return
	}

	// 成功：记录统计
	model.RecordSuccess(selectedModel, selectedChannel.ID, latency)

	// 解析token使用量
	var promptTokens, completionTokens int
	var cost float64
	if selectedChannel.Type != "anthropic" && respBody == nil {
		// 流式或直接透传
		c.Header("Content-Type", resp.Header.Get("Content-Type"))
		// 透传响应头
		for k, vs := range resp.Header {
			for _, v := range vs {
				if strings.ToLower(k) != "content-length" {
					c.Header(k, v)
				}
			}
		}
		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)

		// 尝试从streaming响应提取token数（简化：不解析streaming）
	} else {
		// 非流式：解析响应
		if respBody != nil {
			var result map[string]interface{}
			json.Unmarshal(respBody, &result)
			if usage, ok := result["usage"].(map[string]interface{}); ok {
				if pt, ok := usage["prompt_tokens"].(float64); ok { promptTokens = int(pt) }
				if ct, ok := usage["completion_tokens"].(float64); ok { completionTokens = int(ct) }
			}
			c.Data(resp.StatusCode, "application/json", respBody)
		} else {
			respBody, _ = io.ReadAll(resp.Body)
			c.Data(resp.StatusCode, "application/json", respBody)
		}
	}

	token.AddQuotaUsed(int64(promptTokens + completionTokens))
	model.InsertLog(&model.Log{
		TokenID: token.ID, ChannelID: selectedChannel.ID, Model: selectedModel,
		PromptTokens: promptTokens, CompletionTokens: completionTokens,
		Cost: cost, LatencyMs: latency, Status: "success",
	})
}

// shouldRetry 判断是否应该重试
func shouldRetry(statusCode int) bool {
	return statusCode == 429 || statusCode == 500 || statusCode == 502 || statusCode == 503
}

// tryFallback 尝试fallback到其他渠道
func tryFallback(c *gin.Context, modelName string, excludeChannelID int64, body []byte, token *model.Token, startTime time.Time) {
	channels, _ := model.GetChannelsByModel(modelName)
	for _, ch := range channels {
		if ch.ID == excludeChannelID { continue }
		if model.IsBanned(modelName, ch.ID) { continue }

		var resp *http.Response
		var err error
		switch ch.Type {
		case "anthropic":
			resp, _, err = relay.RelayAnthropic(ch.BaseURL, ch.APIKey, body)
		default:
			resp, err = relay.RelayOpenAI(ch.BaseURL, ch.APIKey, "/v1/chat/completions", "POST", bytes.NewReader(body))
		}

		if err != nil || resp.StatusCode != 200 {
			model.RecordFailure(modelName, ch.ID)
			if resp != nil { resp.Body.Close() }
			continue
		}
		defer resp.Body.Close()

		latency := int(time.Since(startTime).Milliseconds())
		model.RecordSuccess(modelName, ch.ID, latency)

		respBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json", respBody)

		model.InsertLog(&model.Log{
			TokenID: token.ID, ChannelID: ch.ID, Model: modelName,
			LatencyMs: latency, Status: "success",
		})
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "all fallback channels failed"})
}

// ListModels 列出可用模型
func ListModels(c *gin.Context) {
	channels, _ := model.GetAllChannels()
	modelSet := make(map[string]bool)
	for _, ch := range channels {
		if ch.Status != 1 { continue }
		for _, m := range ch.ModelList() {
			modelSet[m] = true
		}
	}
	modelSet["auto"] = true // auto模型始终可用

	var models []map[string]interface{}
	for m := range modelSet {
		models = append(models, map[string]interface{}{
			"id":       m,
			"object":   "model",
			"owned_by": "smart-gateway",
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   models,
	})
}
