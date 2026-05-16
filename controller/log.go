package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sumo1111/smart-gateway/model"
)

func GetLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	modelName := c.Query("model")
	status := c.Query("status")

	logs, err := model.GetLogs(limit, offset, modelName, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func GetStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	stats, err := model.GetDailyStats(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 汇总
	var totalReqs, totalSuccess int
	var totalLatency float64
	for _, s := range stats {
		if t, ok := s["total"].(int); ok { totalReqs += t }
		if s, ok := s["success"].(int); ok { totalSuccess += s }
		if l, ok := s["avg_latency"].(float64); ok { totalLatency += l }
	}

	successRate := 0.0
	if totalReqs > 0 { successRate = float64(totalSuccess) / float64(totalReqs) * 100 }

	avgLat := 0.0
	if len(stats) > 0 { avgLat = totalLatency / float64(len(stats)) }

	channels, _ := model.GetAllChannels()
	activeChannels := 0
	modelSet := map[string]bool{}
	for _, ch := range channels {
		if ch.Status == 1 { activeChannels++ }
		for _, m := range ch.ModelList() { modelSet[m] = true }
	}

	c.JSON(http.StatusOK, gin.H{
		"total_requests":  totalReqs,
		"success_rate":    successRate,
		"avg_latency_ms":  avgLat,
		"active_channels": activeChannels,
		"total_models":    len(modelSet),
		"daily":           stats,
	})
}
