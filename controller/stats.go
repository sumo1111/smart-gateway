package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sumo1111/smart-gateway/auto"
	"github.com/sumo1111/smart-gateway/common"
	"github.com/sumo1111/smart-gateway/model"
)

// GetAutoStatus 获取智能路由状态
func GetAutoStatus(c *gin.Context) {
	stats, _ := model.GetModelStats()
	c.JSON(http.StatusOK, gin.H{
		"strategy":         common.AutoStrategy,
		"model_stats":      stats,
		"fail_ban_count":   common.FailBanCount,
		"fail_ban_duration": common.FailBanDuration,
	})
}

// SetAutoStrategy 切换策略
func SetAutoStrategy(c *gin.Context) {
	var body struct {
		Strategy string `json:"strategy"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	valid := map[string]bool{
		"round_robin": true, "weighted": true,
		"lowest_latency": true, "race": true,
	}
	if !valid[body.Strategy] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid strategy, must be: round_robin, weighted, lowest_latency, race"})
		return
	}
	common.AutoStrategy = body.Strategy
	c.JSON(http.StatusOK, gin.H{"strategy": common.AutoStrategy})
}

// RefreshAutoScores 手动刷新评分
func RefreshAutoScores(c *gin.Context) {
	auto.RefreshScores()
	c.JSON(http.StatusOK, gin.H{"message": "scores refreshed"})
}

// ProbeAllChannels 手动触发健康检测
func ProbeAllChannels(c *gin.Context) {
	results := auto.ProbeAllChannels()
	c.JSON(http.StatusOK, gin.H{"results": results})
}

// ===== 竞速引擎 API =====

// GetRaceLeaderboard 竞速排行榜
func GetRaceLeaderboard(c *gin.Context) {
	engine := auto.GetRaceEngine()
	tracks := engine.GetAllTracks()

	// 按model组织排行榜
	type TrackView struct {
		Model     string         `json:"model"`
		Champion  *auto.Racer    `json:"champion,omitempty"`
		HotSpare  *auto.Racer    `json:"hot_spare,omitempty"`
		Racers    []*auto.Racer  `json:"racers"`
		UpdatedAt string         `json:"updated_at"`
	}

	var views []TrackView
	for modelName, track := range tracks {
		tv := TrackView{
			Model:   modelName,
			Racers:  track.Racers,
		}
		if track.Champion != nil {
			tv.Champion = track.Champion
		}
		if track.HotSpare != nil {
			tv.HotSpare = track.HotSpare
		}
		tv.UpdatedAt = track.UpdatedAt.Format("2006-01-02 15:04:05")
		views = append(views, tv)
	}

	c.JSON(http.StatusOK, gin.H{
		"strategy":      common.AutoStrategy,
		"race_interval": common.RaceInterval,
		"tracks":        views,
	})
}

// GetRaceTrack 某模型的竞速详情
func GetRaceTrack(c *gin.Context) {
	modelName := c.Param("model")
	engine := auto.GetRaceEngine()
	track := engine.GetTrack(modelName)
	if track == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no race data for model: " + modelName})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"model":    track.Model,
		"champion": track.Champion,
		"hot_spare": track.HotSpare,
		"racers":   track.Racers,
	})
}

// TriggerRace 手动触发一轮竞速
func TriggerRace(c *gin.Context) {
	engine := auto.GetRaceEngine()
	go engine.RunRace()
	c.JSON(http.StatusOK, gin.H{"message": "race triggered"})
}

// GetRaceHistory 某个(model, channel)的竞速历史
func GetRaceHistory(c *gin.Context) {
	modelName := c.Param("model")
	channelID, _ := strconv.ParseInt(c.Param("channel_id"), 10, 64)
	engine := auto.GetRaceEngine()
	results := engine.GetRecentResults(modelName, channelID)
	c.JSON(http.StatusOK, gin.H{"results": results})
}

// SpeedTestChannel 手动测试单个渠道
func SpeedTestChannel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ch, err := model.GetChannelByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	modelName := c.Query("model")
	if modelName == "" {
		// 用该渠道的第一个模型
		ml := ch.ModelList()
		if len(ml) > 0 {
			modelName = ml[0]
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no model specified and channel has no models"})
			return
		}
	}

	result := auto.SpeedTestSingle(modelName, ch)
	c.JSON(http.StatusOK, gin.H{"result": result})
}
