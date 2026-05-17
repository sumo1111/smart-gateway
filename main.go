package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sumo1111/smart-gateway/auto"
	"github.com/sumo1111/smart-gateway/common"
	"github.com/sumo1111/smart-gateway/model"
	"github.com/sumo1111/smart-gateway/router"
)

//go:embed web/build/*
var buildFS embed.FS

func main() {
	// 加载配置
	common.LoadConfig()
	// 初始化数据库
	model.InitDB()

	// 启动健康检测 / 竞速引擎
	if common.AutoStrategy == "race" {
		go auto.GetRaceEngine().Start()
		log.Println("🏁 Race engine started — models compete every", common.RaceInterval, "seconds")
	} else {
		go common.StartHealthCheck()
	}

	// 设置路由
	r := router.SetupRouter(buildFS)

	// 优雅关闭
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")
		if common.AutoStrategy == "race" {
			auto.GetRaceEngine().Stop()
		}
		model.CloseDB()
		os.Exit(0)
	}()

	addr := fmt.Sprintf(":%d", common.Port)
	log.Printf("Smart Gateway v%s starting on %s (strategy=%s)", common.Version, addr, common.AutoStrategy)
	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
