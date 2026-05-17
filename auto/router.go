package auto

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/sumo1111/smart-gateway/common"
	"github.com/sumo1111/smart-gateway/model"
)

var mu sync.RWMutex

// ChannelScore 带评分的渠道
type ChannelScore struct {
	Channel model.Channel
	Score   float64
}

// SelectChannel 为指定模型选择最优渠道
func SelectChannel(model_ string, channels []model.Channel) (*model.Channel, error) {
	if len(channels) == 0 {
		return nil, ErrNoAvailableChannel
	}

	mu.RLock()
	defer mu.RUnlock()

	strategy := common.AutoStrategy

	// 竞速模式：直接用竞速引擎的排名
	if strategy == "race" {
		return selectByRace(model_, channels)
	}

	// 传统模式
	var available []ChannelScore
	for _, ch := range channels {
		if model.IsBanned(model_, ch.ID) {
			continue
		}
		score := calcScore(model_, ch)
		available = append(available, ChannelScore{Channel: ch, Score: score})
	}

	if len(available) == 0 {
		return nil, ErrNoAvailableChannel
	}

	switch strategy {
	case "round_robin":
		return roundRobin(available), nil
	case "lowest_latency":
		return lowestLatency(available), nil
	case "weighted":
		return weighted(available), nil
	default:
		return weighted(available), nil
	}
}

// selectByRace 竞速模式选路：冠军优先 → 热备 → 冷备中最高分
func selectByRace(model_ string, channels []model.Channel) (*model.Channel, error) {
	engine := GetRaceEngine()
	champion := engine.GetChampionOrBest(model_)

	if champion != nil {
		// 在可用channels中找到冠军对应的channel
		for _, ch := range channels {
			if ch.ID == champion.ChannelID && ch.Status == 1 {
				return &ch, nil
			}
		}
	}

	// 竞速引擎无数据或冠军不可用，回退到DB评分
	ch, err := model.GetChampionForModel(model_)
	if err == nil && ch != nil {
		// 验证在可用列表中
		for _, c := range channels {
			if c.ID == ch.ID {
				return &c, nil
			}
		}
	}

	// 最终回退：加权随机
	var available []ChannelScore
	for _, ch := range channels {
		if model.IsBanned(model_, ch.ID) {
			continue
		}
		score := calcScore(model_, ch)
		available = append(available, ChannelScore{Channel: ch, Score: score})
	}
	if len(available) == 0 {
		return nil, ErrNoAvailableChannel
	}
	return weighted(available), nil
}

// calcScore 计算渠道综合评分
func calcScore(model_ string, ch model.Channel) float64 {
	var totalCalls, successCalls int
	var avgLatency float64
	var dbScore float64

	err := model.DB.QueryRow(
		"SELECT total_calls, success_calls, avg_latency_ms, score FROM model_stats WHERE model=? AND channel_id=?",
		model_, ch.ID,
	).Scan(&totalCalls, &successCalls, &avgLatency, &dbScore)

	if err != nil {
		return float64(ch.Weight)/100.0*50 + float64(ch.Priority)*5
	}

	var successRate float64
	if totalCalls > 0 {
		successRate = float64(successCalls) / float64(totalCalls)
	} else {
		successRate = 0.5
	}

	var latencyScore float64
	if avgLatency > 0 {
		latencyScore = 1.0 - avgLatency/float64(common.RequestTimeout*1000)
		if latencyScore < 0 { latencyScore = 0 }
	} else {
		latencyScore = 0.5
	}

	priorityScore := float64(ch.Weight) / 100.0

	return successRate*50 + latencyScore*30 + priorityScore*20
}

// weighted 加权随机选择
func weighted(candidates []ChannelScore) *model.Channel {
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	totalScore := 0.0
	for _, c := range candidates {
		totalScore += c.Score
	}
	if totalScore == 0 {
		return &candidates[0].Channel
	}

	r := rand.Float64() * totalScore
	cumsum := 0.0
	for _, c := range candidates {
		cumsum += c.Score
		if r <= cumsum {
			return &c.Channel
		}
	}
	return &candidates[0].Channel
}

// roundRobin 简单轮询
var rrCounter int

func roundRobin(candidates []ChannelScore) *model.Channel {
	idx := rrCounter % len(candidates)
	rrCounter++
	return &candidates[idx].Channel
}

// lowestLatency 最低延迟优先
func lowestLatency(candidates []ChannelScore) *model.Channel {
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})
	return &candidates[0].Channel
}

// SelectAutoModel "auto"模型选择：竞速模式下找全局冠军
func SelectAutoModel() (string, *model.Channel, error) {
	engine := GetRaceEngine()

	// 竞速模式：找所有赛道中评分最高的冠军
	if common.AutoStrategy == "race" {
		tracks := engine.GetAllTracks()
		var bestModel string
		var bestRacer *Racer
		for modelName, track := range tracks {
			if track.Champion != nil {
				if bestRacer == nil || track.Champion.Score > bestRacer.Score {
					bestModel = modelName
					bestRacer = track.Champion
				}
			}
		}
		if bestRacer != nil {
			ch, err := model.GetChannelByID(bestRacer.ChannelID)
			if err == nil {
				return bestModel, ch, nil
			}
		}
	}

	// 回退：传统逻辑
	channels, err := model.GetAllChannels()
	if err != nil || len(channels) == 0 {
		return "", nil, ErrNoAvailableChannel
	}

	type MC struct {
		Model   string
		Channel model.Channel
		Score   float64
	}
	var candidates []MC

	for _, ch := range channels {
		if ch.Status != 1 { continue }
		for _, m := range ch.ModelList() {
			if model.IsBanned(m, ch.ID) { continue }
			score := calcScore(m, ch)
			candidates = append(candidates, MC{Model: m, Channel: ch, Score: score})
		}
	}

	if len(candidates) == 0 {
		return "", nil, ErrNoAvailableChannel
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	best := candidates[0]
	return best.Model, &best.Channel, nil
}

// RefreshScores 刷新所有评分
func RefreshScores() {
	stats, _ := model.GetModelStats()
	for _, s := range stats {
		newScore := calcScore(s.Model, model.Channel{ID: s.ChannelID, Weight: 50, Priority: 5})
		model.DB.Exec("UPDATE model_stats SET score=? WHERE model=? AND channel_id=?", newScore, s.Model, s.ChannelID)
	}
}

// sortRacers 按评分降序排列
func sortRacers(racers []*Racer) {
	sort.Slice(racers, func(i, j int) bool {
		return racers[i].Score > racers[j].Score
	})
}

var ErrNoAvailableChannel = &noChannelError{}

type noChannelError struct{}

func (e *noChannelError) Error() string { return "no available channel" }

func init() {
	rand.Seed(time.Now().UnixNano())
}
