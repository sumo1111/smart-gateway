package model

import "time"

type Log struct {
	ID              int64     `json:"id"`
	TokenID         int64     `json:"token_id"`
	ChannelID       int64     `json:"channel_id"`
	Model           string    `json:"model"`
	PromptTokens    int       `json:"prompt_tokens"`
	CompletionTokens int      `json:"completion_tokens"`
	Cost            float64   `json:"cost"`
	LatencyMs       int       `json:"latency_ms"`
	Status          string    `json:"status"`
	ErrorMsg        string    `json:"error_msg"`
	CreatedAt       time.Time `json:"created_at"`
}

func InsertLog(l *Log) error {
	_, err := DB.Exec(`INSERT INTO logs (token_id, channel_id, model, prompt_tokens, completion_tokens, cost, latency_ms, status, error_msg) VALUES (?,?,?,?,?,?,?,?,?)`,
		l.TokenID, l.ChannelID, l.Model, l.PromptTokens, l.CompletionTokens, l.Cost, l.LatencyMs, l.Status, l.ErrorMsg)
	return err
}

func GetLogs(limit, offset int, model, status string) ([]Log, error) {
	query := "SELECT id, token_id, channel_id, model, prompt_tokens, completion_tokens, cost, latency_ms, status, error_msg, created_at FROM logs WHERE 1=1"
	args := []interface{}{}
	if model != "" {
		query += " AND model=?"
		args = append(args, model)
	}
	if status != "" {
		query += " AND status=?"
		args = append(args, status)
	}
	query += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := DB.Query(query, args...)
	if err != nil { return nil, err }
	defer rows.Close()
	var logs []Log
	for rows.Next() {
		var l Log
		if err := rows.Scan(&l.ID, &l.TokenID, &l.ChannelID, &l.Model, &l.PromptTokens, &l.CompletionTokens, &l.Cost, &l.LatencyMs, &l.Status, &l.ErrorMsg, &l.CreatedAt); err != nil { continue }
		logs = append(logs, l)
	}
	return logs, nil
}

// GetDailyStats 获取每日统计
func GetDailyStats(days int) ([]map[string]interface{}, error) {
	rows, err := DB.Query(`SELECT DATE(created_at) as day, COUNT(*) as total,
		SUM(CASE WHEN status='success' THEN 1 ELSE 0 END) as success,
		AVG(CASE WHEN status='success' THEN latency_ms ELSE NULL END) as avg_latency
		FROM logs WHERE created_at >= datetime('now', '-'||?||' days') GROUP BY DATE(created_at) ORDER BY day`, days)
	if err != nil { return nil, err }
	defer rows.Close()
	var stats []map[string]interface{}
	for rows.Next() {
		var day string
		var total, success int
		var avgLatency float64
		if err := rows.Scan(&day, &total, &success, &avgLatency); err != nil { continue }
		stats = append(stats, map[string]interface{}{
			"day": day, "total": total, "success": success, "avg_latency": avgLatency,
		})
	}
	return stats, nil
}
