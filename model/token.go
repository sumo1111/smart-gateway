package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Key        string    `json:"key"`
	Models     string    `json:"models"`
	QuotaUsed  int64     `json:"quota_used"`
	QuotaLimit int64     `json:"quota_limit"` // 0=unlimited
	Status     int       `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

func GetAllTokens() ([]Token, error) {
	rows, err := DB.Query("SELECT id, name, key, models, quota_used, quota_limit, status, created_at FROM tokens ORDER BY id DESC")
	if err != nil { return nil, err }
	defer rows.Close()
	var tokens []Token
	for rows.Next() {
		var t Token
		if err := rows.Scan(&t.ID, &t.Name, &t.Key, &t.Models, &t.QuotaUsed, &t.QuotaLimit, &t.Status, &t.CreatedAt); err != nil { continue }
		tokens = append(tokens, t)
	}
	return tokens, nil
}

func GetTokenByKey(key string) (*Token, error) {
	var t Token
	err := DB.QueryRow("SELECT id, name, key, models, quota_used, quota_limit, status, created_at FROM tokens WHERE key=? AND status=1", key).
		Scan(&t.ID, &t.Name, &t.Key, &t.Models, &t.QuotaUsed, &t.QuotaLimit, &t.Status, &t.CreatedAt)
	if err != nil { return nil, errors.New("invalid token") }
	// 检查额度
	if t.QuotaLimit > 0 && t.QuotaUsed >= t.QuotaLimit {
		return nil, errors.New("token quota exhausted")
	}
	return &t, nil
}

func CreateToken(t *Token) error {
	if t.Key == "" {
		t.Key = "sk-" + uuid.New().String()[:32]
	}
	res, err := DB.Exec("INSERT INTO tokens (name, key, models, quota_limit, status) VALUES (?,?,?,?,?)",
		t.Name, t.Key, t.Models, t.QuotaLimit, t.Status)
	if err != nil { return err }
	t.ID, _ = res.LastInsertId()
	return nil
}

func UpdateToken(t *Token) error {
	_, err := DB.Exec("UPDATE tokens SET name=?, models=?, quota_limit=?, status=? WHERE id=?",
		t.Name, t.Models, t.QuotaLimit, t.Status, t.ID)
	return err
}

func DeleteToken(id int64) error {
	_, err := DB.Exec("DELETE FROM tokens WHERE id=?", id)
	return err
}

func (t *Token) AddQuotaUsed(amount int64) error {
	_, err := DB.Exec("UPDATE tokens SET quota_used = quota_used + ? WHERE id=?", amount, t.ID)
	return err
}
