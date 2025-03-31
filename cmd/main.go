package main

import (
	"iwut-smart-timetable-backend/internal/config"
	"iwut-smart-timetable-backend/internal/database"
	"iwut-smart-timetable-backend/internal/middleware"
	"iwut-smart-timetable-backend/internal/router"
	"net/http"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()
	middleware.Logger = middleware.NewLogger(cfg)

	// 初始化路由
	r := router.NewRouter()
	loggedRouter := middleware.Logger.HttpMiddleware(r)

	// 初始化数据库连接
	err := database.NewDB(cfg)
	if err != nil {
		middleware.Logger.Log("ERROR", "Failed to initialize database: "+err.Error())
		return
	}
	defer database.GetDB().Close()

	// 初始化工作队列
	middleware.InitQueues(cfg)

	// 启动服务
	middleware.Logger.Log("INFO", "Starting server on port "+cfg.Port)
	err = http.ListenAndServe(":"+cfg.Port, loggedRouter)
	if err != nil {
		middleware.Logger.Log("ERROR", "Failed to start server: "+err.Error())
	}
}
