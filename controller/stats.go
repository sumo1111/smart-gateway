package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sumo1111/smart-gateway/auto"
	"github.com/sumo1111/smart-gateway/common"
	"github.com/sumo1111/smart-gateway/model"
)

func GetAutoStatus(c *gin.Context) {
	stats, _ := model.GetModelStats()
	c.JSON(http.StatusOK, gin.H{
		"strategy":      common.AutoStrategy,
		"model_stats":   stats,
		"fail_ban_count": common.FailBanCount,
		"fail_ban_duration": common.FailBanDuration,
	})
}

func SetAutoStrategy(c *gin.Context) {
	var body struct {
		Strategy string `json:"strategy"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Strategy != "round_robin" && body.Strategy != "weighted" && body.Strategy != "lowest_latency" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid strategy, must be: round_robin, weighted, lowest_latency"})
		return
	}
	common.AutoStrategy = body.Strategy
	c.JSON(http.StatusOK, gin.H{"strategy": common.AutoStrategy})
}

func RefreshAutoScores(c *gin.Context) {
	auto.RefreshScores()
	c.JSON(http.StatusOK, gin.H{"message": "scores refreshed"})
}

func ProbeAllChannels(c *gin.Context) {
	results := auto.ProbeAllChannels()
	c.JSON(http.StatusOK, gin.H{"results": results})
}
