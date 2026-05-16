package model

import (
	"strings"
	"time"
)

type Channel struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`   // openai, anthropic, custom
	BaseURL   string    `json:"base_url"`
	APIKey    string    `json:"api_key"`
	Models    string    `json:"models"` // 逗号分隔
	Status    int       `json:"status"` // 1=enabled, 0=disabled
	Priority  int       `json:"priority"`
	Weight    int       `json:"weight"`
	MaxQPS    int       `json:"max_qps"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *Channel) ModelList() []string {
	if c.Models == "" { return nil }
	return strings.Split(c.Models, ",")
}

func GetAllChannels() ([]Channel, error) {
	rows, err := DB.Query("SELECT id, name, type, base_url, api_key, models, status, priority, weight, max_qps, created_at, updated_at FROM channels ORDER BY priority DESC, id")
	if err != nil { return nil, err }
	defer rows.Close()
	var channels []Channel
	for rows.Next() {
		var ch Channel
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Type, &ch.BaseURL, &ch.APIKey, &ch.Models, &ch.Status, &ch.Priority, &ch.Weight, &ch.MaxQPS, &ch.CreatedAt, &ch.UpdatedAt); err != nil { continue }
		channels = append(channels, ch)
	}
	return channels, nil
}

func GetChannelByID(id int64) (*Channel, error) {
	var ch Channel
	err := DB.QueryRow("SELECT id, name, type, base_url, api_key, models, status, priority, weight, max_qps, created_at, updated_at FROM channels WHERE id=?", id).
		Scan(&ch.ID, &ch.Name, &ch.Type, &ch.BaseURL, &ch.APIKey, &ch.Models, &ch.Status, &ch.Priority, &ch.Weight, &ch.MaxQPS, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil { return nil, err }
	return &ch, nil
}

func CreateChannel(ch *Channel) error {
	res, err := DB.Exec("INSERT INTO channels (name, type, base_url, api_key, models, status, priority, weight, max_qps) VALUES (?,?,?,?,?,?,?,?,?)",
		ch.Name, ch.Type, ch.BaseURL, ch.APIKey, ch.Models, ch.Status, ch.Priority, ch.Weight, ch.MaxQPS)
	if err != nil { return err }
	ch.ID, _ = res.LastInsertId()
	return nil
}

func UpdateChannel(ch *Channel) error {
	_, err := DB.Exec("UPDATE channels SET name=?, type=?, base_url=?, api_key=?, models=?, status=?, priority=?, weight=?, max_qps=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		ch.Name, ch.Type, ch.BaseURL, ch.APIKey, ch.Models, ch.Status, ch.Priority, ch.Weight, ch.MaxQPS, ch.ID)
	return err
}

func DeleteChannel(id int64) error {
	_, err := DB.Exec("DELETE FROM channels WHERE id=?", id)
	if err != nil { return err }
	DB.Exec("DELETE FROM model_stats WHERE channel_id=?", id)
	return nil
}

func GetChannelsByModel(model string) ([]Channel, error) {
	rows, err := DB.Query("SELECT id, name, type, base_url, api_key, models, status, priority, weight, max_qps, created_at, updated_at FROM channels WHERE status=1 AND (models LIKE ? OR models LIKE ? OR models LIKE ?) ORDER BY priority DESC",
		model+",%", "%,"+model+",%", "%,"+model)
	if err != nil { return nil, err }
	defer rows.Close()
	var channels []Channel
	for rows.Next() {
		var ch Channel
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Type, &ch.BaseURL, &ch.APIKey, &ch.Models, &ch.Status, &ch.Priority, &ch.Weight, &ch.MaxQPS, &ch.CreatedAt, &ch.UpdatedAt); err != nil { continue }
		channels = append(channels, ch)
	}
	return channels, nil
}
