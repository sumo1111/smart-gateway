package common

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	Port             = 3000
	Version          = "1.1.0"
	AdminPassword    = "admin123"
	DBPath           = "data/smart-gateway.db"
	AutoStrategy     = "race" // round_robin / weighted / lowest_latency / race
	HealthInterval   = 60    // 秒
	FailBanCount     = 3
	FailBanDuration  = 300  // 秒
	RequestTimeout   = 120  // 秒
	// 竞速引擎配置
	RaceInterval     = 60   // 秒 — 竞速测试间隔
	RaceTestPrompt   = "Say hi in one word" // 竞速测试提示词
	RaceTestMaxTokens = 5   // 竞速测试最大生成token数
	RaceTestTimeout  = 30   // 秒 — 单次竞速测试超时
)

func LoadConfig() {
	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil { Port = n }
	}
	if v := os.Getenv("ADMIN_PASSWORD"); v != "" { AdminPassword = v }
	if v := os.Getenv("DB_PATH"); v != "" { DBPath = v }
	if v := os.Getenv("AUTO_STRATEGY"); v != "" { AutoStrategy = v }
	if v := os.Getenv("HEALTH_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil { HealthInterval = n }
	}
	if v := os.Getenv("RACE_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil { RaceInterval = n }
	}
	if v := os.Getenv("RACE_TEST_PROMPT"); v != "" { RaceTestPrompt = v }
	if v := os.Getenv("RACE_TEST_MAX_TOKENS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil { RaceTestMaxTokens = n }
	}
	if v := os.Getenv("RACE_TEST_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil { RaceTestTimeout = n }
	}
	log.Printf("Config: port=%d, strategy=%s, race_interval=%ds", Port, AutoStrategy, RaceInterval)
}

// StartHealthCheck 定期检测渠道健康（仅非竞速模式时独立运行）
func StartHealthCheck() {
	ticker := time.NewTicker(time.Duration(HealthInterval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		runHealthProbe()
	}
}

var healthMu sync.Mutex

func runHealthProbe() {
	healthMu.Lock()
	defer healthMu.Unlock()
	// 竞速模式下由RaceEngine负责健康检测，此处跳过
	if AutoStrategy == "race" {
		return
	}
}
