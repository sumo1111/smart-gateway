package model

import (
	"database/sql"
	"time"
)

type ModelStat struct {
	Model           string     `json:"model"`
	ChannelID       int64      `json:"channel_id"`
	TotalCalls      int        `json:"total_calls"`
	SuccessCalls    int        `json:"success_calls"`
	AvgLatencyMs    float64    `json:"avg_latency_ms"`
	LastSuccessAt   *time.Time `json:"last_success_at"`
	LastFailAt      *time.Time `json:"last_fail_at"`
	ConsecutiveFails int       `json:"consecutive_fails"`
	BannedUntil     *time.Time `json:"banned_until"`
	Score           float64    `json:"score"`
	// 竞速相关字段
	RaceRank    int     `json:"race_rank"`
	RaceRole    string  `json:"race_role"`     // champion / hot_standby / cold_standby / benched
	AvgTTFT     int     `json:"avg_ttft_ms"`
	AvgTPS      float64 `json:"avg_tps"`
	SuccessRate float64 `json:"race_success_rate"`
}

func GetModelStats() ([]ModelStat, error) {
	rows, err := DB.Query(`SELECT model, channel_id, total_calls, success_calls, avg_latency_ms,
		last_success_at, last_fail_at, consecutive_fails, banned_until, score,
		COALESCE(race_rank, 0), COALESCE(race_role, 'cold_standby'),
		COALESCE(avg_ttft_ms, 0), COALESCE(avg_tps, 0), COALESCE(race_success_rate, 0)
		FROM model_stats ORDER BY model, score DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats []ModelStat
	for rows.Next() {
		var s ModelStat
		if err := rows.Scan(&s.Model, &s.ChannelID, &s.TotalCalls, &s.SuccessCalls, &s.AvgLatencyMs,
			&s.LastSuccessAt, &s.LastFailAt, &s.ConsecutiveFails, &s.BannedUntil, &s.Score,
			&s.RaceRank, &s.RaceRole, &s.AvgTTFT, &s.AvgTPS, &s.SuccessRate); err != nil {
			continue
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// RecordSuccess 记录成功调用并更新score
func RecordSuccess(model string, channelID int64, latencyMs int) {
	DB.Exec(`INSERT INTO model_stats (model, channel_id, total_calls, success_calls, avg_latency_ms, last_success_at, consecutive_fails, score)
		VALUES (?, ?, 1, 1, ?, CURRENT_TIMESTAMP, 0, 50)
		ON CONFLICT(model, channel_id) DO UPDATE SET
		total_calls=total_calls+1, success_calls=success_calls+1,
		avg_latency_ms=(avg_latency_ms*success_calls+?)/(success_calls+1),
		last_success_at=CURRENT_TIMESTAMP, consecutive_fails=0,
		score=CASE WHEN score<100 THEN score+2 ELSE 100 END`,
		model, channelID, float64(latencyMs), float64(latencyMs))
}

// RecordFailure 记录失败调用
func RecordFailure(model string, channelID int64) {
	DB.Exec(`INSERT INTO model_stats (model, channel_id, total_calls, consecutive_fails, last_fail_at, score)
		VALUES (?, ?, 1, 1, CURRENT_TIMESTAMP, 45)
		ON CONFLICT(model, channel_id) DO UPDATE SET
		total_calls=total_calls+1, consecutive_fails=consecutive_fails+1,
		last_fail_at=CURRENT_TIMESTAMP,
		score=CASE WHEN score>5 THEN score-10 ELSE 0 END`,
		model, channelID)
}

// IsBanned 检查渠道是否被临时屏蔽
func IsBanned(model string, channelID int64) bool {
	var bannedUntil *time.Time
	err := DB.QueryRow("SELECT banned_until FROM model_stats WHERE model=? AND channel_id=?", model, channelID).Scan(&bannedUntil)
	if err != nil || bannedUntil == nil {
		return false
	}
	return time.Now().Before(*bannedUntil)
}

// CheckAndBan 检查是否需要屏蔽
func CheckAndBan(model string, channelID int64, failThreshold int, banDuration int) {
	var fails int
	err := DB.QueryRow("SELECT consecutive_fails FROM model_stats WHERE model=? AND channel_id=?", model, channelID).Scan(&fails)
	if err != nil {
		return
	}
	if fails >= failThreshold {
		DB.Exec("UPDATE model_stats SET banned_until=datetime('now','+'||?||' seconds') WHERE model=? AND channel_id=?", banDuration, model, channelID)
	}
}

// UpdateRacerScore 更新竞速排名和评分
func UpdateRacerScore(model string, channelID int64, score float64, avgTTFT int, avgTPS float64, successRate float64, rank int, role string) {
	DB.Exec(`INSERT INTO model_stats (model, channel_id, score, avg_ttft_ms, avg_tps, race_success_rate, race_rank, race_role, total_calls)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)
		ON CONFLICT(model, channel_id) DO UPDATE SET
		score=?, avg_ttft_ms=?, avg_tps=?, race_success_rate=?, race_rank=?, race_role=?`,
		model, channelID, score, avgTTFT, avgTPS, successRate, rank, role,
		score, avgTTFT, avgTPS, successRate, rank, role)
}

// GetChannelsByModel 获取支持某模型的所有渠道
func GetChannelsByModel(modelName string) ([]Channel, error) {
	rows, err := DB.Query(`SELECT id, name, type, base_url, api_key, models, status, priority, weight, max_qps, created_at, updated_at
		FROM channels WHERE status=1 AND models LIKE ? ORDER BY priority DESC, id`,
		"%"+modelName+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var channels []Channel
	for rows.Next() {
		var ch Channel
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Type, &ch.BaseURL, &ch.APIKey, &ch.Models, &ch.Status, &ch.Priority, &ch.Weight, &ch.MaxQPS, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			continue
		}
		// 精确匹配（防止"gpt-4"匹配到"gpt-4o"）
		matched := false
		for _, m := range ch.ModelList() {
			if m == modelName {
				matched = true
				break
			}
		}
		if matched {
			channels = append(channels, ch)
		}
	}
	return channels, nil
}

// GetChampionForModel 从DB获取某模型的冠军渠道
func GetChampionForModel(modelName string) (*Channel, error) {
	var channelID int64
	err := DB.QueryRow(`SELECT channel_id FROM model_stats
		WHERE model=? AND race_role='champion'
		ORDER BY score DESC LIMIT 1`, modelName).Scan(&channelID)
	if err == sql.ErrNoRows {
		// 没有冠军，取最高分的
		err = DB.QueryRow(`SELECT channel_id FROM model_stats
			WHERE model=? AND (race_role != 'benched' OR race_role IS NULL)
			ORDER BY score DESC LIMIT 1`, modelName).Scan(&channelID)
	}
	if err != nil {
		return nil, err
	}
	return GetChannelByID(channelID)
}
