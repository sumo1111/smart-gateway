package auto

import (
	"strconv"
	"sync"
	"time"

	"github.com/sumo1111/smart-gateway/common"
	"github.com/sumo1111/smart-gateway/model"
)

// RaceResult 单次竞速测试结果
type RaceResult struct {
	Model     string  `json:"model"`
	ChannelID int64   `json:"channel_id"`
	Channel   string  `json:"channel_name"`
	TTFT      int     `json:"ttft_ms"`       // Time To First Token
	TTFTTotal int     `json:"ttft_total_ms"` // 完整响应耗时
	TPS       float64 `json:"tps"`           // tokens per second
	Success   bool    `json:"success"`
	ErrorMsg  string  `json:"error_msg,omitempty"`
	TestedAt  time.Time `json:"tested_at"`
}

// Racer 单个竞速选手的累积状态
type Racer struct {
	Model        string    `json:"model"`
	ChannelID    int64     `json:"channel_id"`
	ChannelName  string    `json:"channel_name"`
	Rank         int       `json:"rank"`           // 当前排名
	Role         string    `json:"role"`           // champion / hot_standby / cold_standby / benched
	Score        float64   `json:"score"`          // 综合评分 0-100
	AvgTTFT      int       `json:"avg_ttft_ms"`    // 近N轮平均TTFT
	AvgTotal     int       `json:"avg_total_ms"`   // 近N轮平均总耗时
	AvgTPS       float64   `json:"avg_tps"`        // 近N轮平均TPS
	SuccessRate  float64   `json:"success_rate"`   // 近N轮成功率
	ChampionSince *time.Time `json:"champion_since,omitempty"` // 当冠军多久了
	LastTestAt   time.Time `json:"last_test_at"`
	Streak       int       `json:"streak"`         // 连续成功次数
	FailStreak   int       `json:"fail_streak"`    // 连续失败次数
}

// RaceTrack 竞速赛道：一个model对应一条赛道
type RaceTrack struct {
	Model    string   `json:"model"`
	Racers   []*Racer `json:"racers"`
	Champion *Racer   `json:"champion,omitempty"`
	HotSpare *Racer   `json:"hot_spare,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RaceEngine 竞速引擎
type RaceEngine struct {
	mu       sync.RWMutex
	tracks   map[string]*RaceTrack // model -> track
	history  map[string][]RaceResult // "model:channel_id" -> recent results (keep last N)
	running  bool
	stopCh   chan struct{}
}

const (
	historyWindowSize = 10 // 保留最近10轮竞速结果
)

var (
	engine *RaceEngine
	once   sync.Once
)

// GetRaceEngine 单例
func GetRaceEngine() *RaceEngine {
	once.Do(func() {
		engine = &RaceEngine{
			tracks:  make(map[string]*RaceTrack),
			history: make(map[string][]RaceResult),
			stopCh:  make(chan struct{}),
		}
	})
	return engine
}

// Start 启动竞速引擎
func (e *RaceEngine) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.mu.Unlock()

	common.Info("🏁 Race engine started — interval=%ds, prompt=%q", common.RaceInterval, common.RaceTestPrompt)

	// 首次立即执行
	e.runRace()

	ticker := time.NewTicker(time.Duration(common.RaceInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.runRace()
		case <-e.stopCh:
			e.mu.Lock()
			e.running = false
			e.mu.Unlock()
			return
		}
	}
}

// Stop 停止引擎
func (e *RaceEngine) Stop() {
	close(e.stopCh)
}

// RunRace 公开方法，手动触发一轮竞速
func (e *RaceEngine) RunRace() {
	e.runRace()
}

// runRace 执行一轮竞速
func (e *RaceEngine) runRace() {
	channels, err := model.GetAllChannels()
	if err != nil || len(channels) == 0 {
		common.Warn("Race: no channels available")
		return
	}

	// 收集所有参赛 (model, channel) 对
	type entry struct {
		model   string
		channel model.Channel
	}
	var entries []entry
	for _, ch := range channels {
		if ch.Status != 1 {
			continue
		}
		for _, m := range ch.ModelList() {
			entries = append(entries, entry{model: m, channel: ch})
		}
	}

	if len(entries) == 0 {
		return
	}

	// 并发竞速测试（限制并发数）
	sem := make(chan struct{}, 5) // 最多5路并发
	var wg sync.WaitGroup
	var results []RaceResult
	var resMu sync.Mutex

	for _, ent := range entries {
		wg.Add(1)
		sem <- struct{}{}
		go func(m string, ch model.Channel) {
			defer wg.Done()
			defer func() { <-sem }()

			r := speedTest(m, ch)
			resMu.Lock()
			results = append(results, r)
			resMu.Unlock()
		}(ent.model, ent.channel)
	}
	wg.Wait()

	// 更新赛道排名
	e.updateRankings(results)
}

// updateRankings 根据本轮竞速结果更新排名和岗位
func (e *RaceEngine) updateRankings(results []RaceResult) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 按model分组
	byModel := make(map[string][]RaceResult)
	for _, r := range results {
		byModel[r.Model] = append(byModel[r.Model], r)
	}

	for modelName, raceResults := range byModel {
		track, exists := e.tracks[modelName]
		if !exists {
			track = &RaceTrack{Model: modelName}
			e.tracks[modelName] = track
		}

		// 更新历史记录
		for _, r := range raceResults {
			key := raceResultKey(r.Model, r.ChannelID)
			e.history[key] = append(e.history[key], r)
			// 只保留最近N轮
			if len(e.history[key]) > historyWindowSize {
				e.history[key] = e.history[key][len(e.history[key])-historyWindowSize:]
			}
		}

		// 构建或更新racer
		racerMap := make(map[int64]*Racer)
		for _, r := range track.Racers {
			racerMap[r.ChannelID] = r
		}

		for _, r := range raceResults {
			racer, ok := racerMap[r.ChannelID]
			if !ok {
				racer = &Racer{
					Model:       r.Model,
					ChannelID:   r.ChannelID,
					ChannelName: r.Channel,
					Role:        "cold_standby",
				}
				racerMap[r.ChannelID] = racer
			}
			racer.LastTestAt = r.TestedAt
			if r.Success {
				racer.Streak++
				racer.FailStreak = 0
			} else {
				racer.FailStreak++
				racer.Streak = 0
			}
		}

		// 计算每个racer的综合评分
		var racers []*Racer
		for _, racer := range racerMap {
			key := raceResultKey(racer.Model, racer.ChannelID)
			hist := e.history[key]
			racer = calcRacerScore(racer, hist)
			racers = append(racers, racer)
		}

		// 按评分排序
		sortRacers(racers)

		// 分配岗位
		assignRoles(racers, track)

		track.Racers = racers
		track.UpdatedAt = time.Now()

		// 持久化到DB
		for _, racer := range racers {
			model.UpdateRacerScore(racer.Model, racer.ChannelID, racer.Score, racer.AvgTTFT, racer.AvgTPS, racer.SuccessRate, racer.Rank, racer.Role)
		}

		// 日志
		if track.Champion != nil {
			common.Info("🏁 [%s] Champion: %s (score=%.1f, ttft=%dms, tps=%.1f)",
				modelName, track.Champion.ChannelName, track.Champion.Score, track.Champion.AvgTTFT, track.Champion.AvgTPS)
		}
		if track.HotSpare != nil {
			common.Info("🔥 [%s] Hot Spare: %s (score=%.1f)",
				modelName, track.HotSpare.ChannelName, track.HotSpare.Score)
		}
	}
}

// calcRacerScore 根据历史窗口计算综合评分
// score = TTFT分(35) + TPS分(30) + 成功率分(25) + 稳定性分(10)
func calcRacerScore(racer *Racer, history []RaceResult) *Racer {
	if len(history) == 0 {
		racer.Score = 10 // 无数据最低分
		return racer
	}

	var totalTTFT, totalElapsed int
	var totalTPS, successCount float64
	for _, h := range history {
		if h.Success {
			successCount++
			totalTTFT += h.TTFT
			totalElapsed += h.TTFTTotal
			totalTPS += h.TPS
		}
	}

	n := float64(len(history))
	racer.SuccessRate = successCount / n

	if successCount > 0 {
		racer.AvgTTFT = int(float64(totalTTFT) / successCount)
		racer.AvgTotal = int(float64(totalElapsed) / successCount)
		racer.AvgTPS = totalTPS / successCount
	}

	// TTFT评分：越低越好，0ms=35分，2000ms+=0分
	ttftScore := 35.0 * (1.0 - float64(racer.AvgTTFT)/2000.0)
	if ttftScore < 0 { ttftScore = 0 }
	if ttftScore > 35 { ttftScore = 35 }

	// TPS评分：越高越好，0tps=0分，100tps+=30分
	tpsScore := racer.AvgTPS / 100.0 * 30.0
	if tpsScore < 0 { tpsScore = 0 }
	if tpsScore > 30 { tpsScore = 30 }

	// 成功率评分
	srScore := racer.SuccessRate * 25.0

	// 稳定性评分：连续成功加分，连续失败扣分
	stabScore := 10.0
	if racer.Streak >= 5 { stabScore = 10 }
	if racer.Streak >= 3 { stabScore = 8 }
	if racer.FailStreak >= 2 { stabScore = 2 }
	if racer.FailStreak >= 3 { stabScore = 0 }

	racer.Score = ttftScore + tpsScore + srScore + stabScore
	return racer
}

// assignRoles 分配岗位：冠军、热备、冷备、坐板凳
func assignRoles(sorted []*Racer, track *RaceTrack) {
	oldChampion := track.Champion

	for i, r := range sorted {
		r.Rank = i + 1
		switch {
		case i == 0:
			r.Role = "champion"
			// 记录冠军时长
			if oldChampion != nil && oldChampion.ChannelID == r.ChannelID {
				// 保持冠军，不更新ChampionSince
			} else {
				now := time.Now()
				r.ChampionSince = &now
			}
			track.Champion = r
		case i == 1:
			r.Role = "hot_standby"
			track.HotSpare = r
		case r.FailStreak >= 3:
			r.Role = "benched"
		default:
			r.Role = "cold_standby"
		}
	}

	// 冠军换人日志
	if oldChampion != nil && track.Champion != nil && oldChampion.ChannelID != track.Champion.ChannelID {
		common.Info("👑 [%s] NEW CHAMPION: %s (%.1f) dethrones %s (%.1f)",
			track.Model, track.Champion.ChannelName, track.Champion.Score,
			oldChampion.ChannelName, oldChampion.Score)
	}
}

// GetTrack 获取某模型的赛道状态
func (e *RaceEngine) GetTrack(modelName string) *RaceTrack {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.tracks[modelName]
}

// GetAllTracks 获取所有赛道
func (e *RaceEngine) GetAllTracks() map[string]*RaceTrack {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// 返回副本
	cp := make(map[string]*RaceTrack, len(e.tracks))
	for k, v := range e.tracks {
		cp[k] = v
	}
	return cp
}

// GetChampion 获取某模型当前冠军
func (e *RaceEngine) GetChampion(modelName string) *Racer {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if track, ok := e.tracks[modelName]; ok && track.Champion != nil {
		return track.Champion
	}
	return nil
}

// GetChampionOrBest 获取冠军，若无冠军则返回最高分选手
func (e *RaceEngine) GetChampionOrBest(modelName string) *Racer {
	e.mu.RLock()
	defer e.mu.RUnlock()
	track, ok := e.tracks[modelName]
	if !ok || len(track.Racers) == 0 {
		return nil
	}
	if track.Champion != nil && track.Champion.FailStreak < 3 {
		return track.Champion
	}
	// 冠军挂了，看热备
	if track.HotSpare != nil && track.HotSpare.FailStreak < 3 {
		return track.HotSpare
	}
	// 找第一个没坐板凳的
	for _, r := range track.Racers {
		if r.Role != "benched" && r.FailStreak < 3 {
			return r
		}
	}
	// 全挂了，返回排名最高的
	return track.Racers[0]
}

// GetRecentResults 获取某(model, channel)的最近竞速结果
func (e *RaceEngine) GetRecentResults(modelName string, channelID int64) []RaceResult {
	e.mu.RLock()
	defer e.mu.RUnlock()
	key := raceResultKey(modelName, channelID)
	cp := make([]RaceResult, len(e.history[key]))
	copy(cp, e.history[key])
	return cp
}

func raceResultKey(model string, channelID int64) string {
	return model + ":" + formatInt64(channelID)
}

func formatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}
