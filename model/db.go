package model

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	dir := filepath.Dir("data/smart-gateway.db")
	os.MkdirAll(dir, 0755)

	var err error
	DB, err = sql.Open("sqlite3", "data/smart-gateway.db?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		panic(fmt.Sprintf("Failed to open database: %v", err))
	}
	DB.SetMaxOpenConns(1)
	migrate()
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

func migrate() {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS channels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'openai',
			base_url TEXT NOT NULL,
			api_key TEXT NOT NULL,
			models TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			priority INTEGER NOT NULL DEFAULT 5,
			weight INTEGER NOT NULL DEFAULT 50,
			max_qps INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			key TEXT NOT NULL UNIQUE,
			models TEXT NOT NULL DEFAULT '',
			quota_used INTEGER NOT NULL DEFAULT 0,
			quota_limit INTEGER NOT NULL DEFAULT 0,
			status INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token_id INTEGER,
			channel_id INTEGER,
			model TEXT NOT NULL,
			prompt_tokens INTEGER DEFAULT 0,
			completion_tokens INTEGER DEFAULT 0,
			cost REAL DEFAULT 0,
			latency_ms INTEGER DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'success',
			error_msg TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS model_stats (
			model TEXT NOT NULL,
			channel_id INTEGER NOT NULL,
			total_calls INTEGER DEFAULT 0,
			success_calls INTEGER DEFAULT 0,
			avg_latency_ms REAL DEFAULT 0,
			last_success_at DATETIME,
			last_fail_at DATETIME,
			consecutive_fails INTEGER DEFAULT 0,
			banned_until DATETIME,
			score REAL DEFAULT 50,
			PRIMARY KEY (model, channel_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_model ON logs(model)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_created ON logs(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_status ON logs(status)`,
	}
	for _, s := range statements {
		DB.Exec(s)
	}
}
