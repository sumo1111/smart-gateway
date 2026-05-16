package common

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	Port            = 3000
	Version         = "1.0.0"
	AdminPassword   = "admin123"
	DBPath          = "data/smart-gateway.db"
	AutoStrategy    = "weighted" // round_robin / weighted / lowest_latency
	HealthInterval  = 60         // 秒
	FailBanCount    = 3          // 连续失败次数触发屏蔽
	FailBanDuration = 300        // 屏蔽时长秒
	RequestTimeout  = 120        // 请求超时秒
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
	log.Printf("Config: port=%d, strategy=%s, db=%s", Port, AutoStrategy, DBPath)
}

// HealthCheck 定期检测渠道健康
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
	// 在model包中实现具体检测逻辑
	// model.ProbeAllChannels()
}
