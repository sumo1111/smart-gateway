package model

import (
	"time"
)

type ModelStat struct {
	Model            string     `json:"model"`
	ChannelID        int64      `json:"channel_id"`
	TotalCalls       int        `json:"total_calls"`
	SuccessCalls     int        `json:"success_calls"`
	AvgLatencyMs     float64    `json:"avg_latency_ms"`
	LastSuccessAt    *time.Time `json:"last_success_at"`
	LastFailAt       *time.Time `json:"last_fail_at"`
	ConsecutiveFails int        `json:"consecutive_fails"`
	BannedUntil      *time.Time `json:"banned_until"`
	Score            float64    `json:"score"`
}

func GetModelStats() ([]ModelStat, error) {
	rows, err := DB.Query("SELECT model, channel_id, total_calls, success_calls, avg_latency_ms, last_success_at, last_fail_at, consecutive_fails, banned_until, score FROM model_stats ORDER BY model, score DESC")
	if err != nil { return nil, err }
	defer rows.Close()
	var stats []ModelStat
	for rows.Next() {
		var s ModelStat
		if err := rows.Scan(&s.Model, &s.ChannelID, &s.TotalCalls, &s.SuccessCalls, &s.AvgLatencyMs, &s.LastSuccessAt, &s.LastFailAt, &s.ConsecutiveFails, &s.BannedUntil, &s.Score); err != nil { continue }
		stats = append(stats, s)
	}
	return stats, nil
}

// RecordSuccess 记录成功调用并更新score
func RecordSuccess(model string, channelID int64, latencyMs int) {
	DB.Exec("INSERT INTO model_stats (model, channel_id, total_calls, success_calls, avg_latency_ms, last_success_at, consecutive_fails, score) VALUES (?, ?, 1, 1, ?, CURRENT_TIMESTAMP, 0, 50) ON CONFLICT(model, channel_id) DO UPDATE SET total_calls=total_calls+1, success_calls=success_calls+1, avg_latency_ms=(avg_latency_ms*success_calls+?)/(success_calls+1), last_success_at=CURRENT_TIMESTAMP, consecutive_fails=0, score=CASE WHEN score<100 THEN score+2 ELSE 100 END",
		model, channelID, float64(latencyMs), float64(latencyMs))
}

// RecordFailure 记录失败调用
func RecordFailure(model string, channelID int64) {
	DB.Exec("INSERT INTO model_stats (model, channel_id, total_calls, consecutive_fails, last_fail_at, score) VALUES (?, ?, 1, 1, CURRENT_TIMESTAMP, 45) ON CONFLICT(model, channel_id) DO UPDATE SET total_calls=total_calls+1, consecutive_fails=consecutive_fails+1, last_fail_at=CURRENT_TIMESTAMP, score=CASE WHEN score>5 THEN score-10 ELSE 0 END",
		model, channelID)
}

// IsBanned 检查渠道是否被临时屏蔽
func IsBanned(model string, channelID int64) bool {
	var bannedUntil *time.Time
	err := DB.QueryRow("SELECT banned_until FROM model_stats WHERE model=? AND channel_id=?", model, channelID).Scan(&bannedUntil)
	if err != nil || bannedUntil == nil { return false }
	return time.Now().Before(*bannedUntil)
}

// CheckAndBan 检查是否需要屏蔽
func CheckAndBan(model string, channelID int64, failThreshold int, banDuration int) {
	var fails int
	err := DB.QueryRow("SELECT consecutive_fails FROM model_stats WHERE model=? AND channel_id=?", model, channelID).Scan(&fails)
	if err != nil { return }
	if fails >= failThreshold {
		DB.Exec("UPDATE model_stats SET banned_until=datetime('now','+'||?||' seconds') WHERE model=? AND channel_id=?", banDuration, model, channelID)
	}
}
