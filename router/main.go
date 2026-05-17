package router

import (
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sumo1111/smart-gateway/controller"
	"github.com/sumo1111/smart-gateway/middleware"
)

func SetupRouter(buildFS embed.FS) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Admin-Password"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// === Relay路由（兼容OpenAI API） ===
	r.Use(middleware.RateLimit(100, time.Minute))
	r.GET("/v1/models", middleware.TokenAuth(), controller.ListModels)
	r.POST("/v1/chat/completions", middleware.TokenAuth(), controller.Relay)
	r.POST("/v1/completions", middleware.TokenAuth(), controller.Relay)
	r.POST("/v1/embeddings", middleware.TokenAuth(), controller.Relay)

	// === 管理API ===
	api := r.Group("/api")
	api.Use(gzip.Gzip(gzip.DefaultCompression))
	{
		ch := api.Group("/channel", middleware.AdminAuth())
		{
			ch.GET("", controller.GetAllChannels)
			ch.GET("/:id", controller.GetChannel)
			ch.POST("", controller.AddChannel)
			ch.PUT("", controller.UpdateChannel)
			ch.DELETE("/:id", controller.DeleteChannel)
			ch.GET("/test/:id", controller.TestChannel)
		}

		tk := api.Group("/token", middleware.AdminAuth())
		{
			tk.GET("", controller.GetAllTokens)
			tk.POST("", controller.AddToken)
			tk.PUT("", controller.UpdateToken)
			tk.DELETE("/:id", controller.DeleteToken)
			tk.POST("/refresh/:id", controller.RefreshTokenKey)
		}

		api.GET("/log", middleware.AdminAuth(), controller.GetLogs)
		api.GET("/stats", middleware.AdminAuth(), controller.GetStats)

		au := api.Group("/auto", middleware.AdminAuth())
		{
			au.GET("/status", controller.GetAutoStatus)
			au.POST("/strategy", controller.SetAutoStrategy)
			au.POST("/refresh", controller.RefreshAutoScores)
			au.POST("/probe", controller.ProbeAllChannels)
		}

		// === 竞速引擎 API ===
		race := api.Group("/race", middleware.AdminAuth())
		{
			race.GET("/leaderboard", controller.GetRaceLeaderboard)
			race.GET("/track/:model", controller.GetRaceTrack)
			race.GET("/history/:model/:channel_id", controller.GetRaceHistory)
			race.POST("/trigger", controller.TriggerRace)
			race.GET("/speedtest/:id", controller.SpeedTestChannel)
		}

		api.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "version": "1.1.0"})
		})
	}

	// === 前端静态文件 ===
	distFS, err := fs.Sub(buildFS, "web/build")
	if err == nil {
		r.StaticFS("/assets", http.FS(distFS))
		r.NoRoute(func(c *gin.Context) {
			data, err := fs.ReadFile(buildFS, "web/build/index.html")
			if err != nil {
				c.String(http.StatusNotFound, "404")
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		})
	}

	return r
}
