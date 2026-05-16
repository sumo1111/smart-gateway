package auto

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sumo1111/smart-gateway/common"
	"github.com/sumo1111/smart-gateway/model"
)

// ProbeChannel 测试渠道连通性
func ProbeChannel(ch *model.Channel) (bool, string, int) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	url := strings.TrimRight(ch.BaseURL, "/") + "/v1/models"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+ch.APIKey)

	start := time.Now()
	resp, err := client.Do(req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		return false, fmt.Sprintf("connection failed: %v", err), latency
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		return true, "ok", latency
	}

	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)[:200]), latency
}

// ProbeAllChannels 检测所有渠道
func ProbeAllChannels() map[int64]map[string]interface{} {
	channels, _ := model.GetAllChannels()
	results := make(map[int64]map[string]interface{})

	for _, ch := range channels {
		if ch.Status != 1 { continue }
		ok, msg, latency := ProbeChannel(&ch)
		results[ch.ID] = map[string]interface{}{
			"ok":      ok,
			"message": msg,
			"latency": latency,
		}
		if ok {
			common.Info("Health check channel %d (%s): OK (%dms)", ch.ID, ch.Name, latency)
		} else {
			common.Warn("Health check channel %d (%s): FAIL - %s", ch.ID, ch.Name, msg)
		}
	}
	return results
}
