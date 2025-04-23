package main

import (
	"fmt"
	"iwut-smart-timetable-backend/assets"
	"iwut-smart-timetable-backend/internal/config"
	"iwut-smart-timetable-backend/internal/database"
	"iwut-smart-timetable-backend/internal/middleware"
	"iwut-smart-timetable-backend/internal/router"
	"net/http"
	"strings"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()
	middleware.Logger = middleware.NewLogger(cfg)

	// 列出所有嵌入的静态资源
	assetsList, err := assets.ListAssets()
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to load packaged assets: %s", err))
		return
	} else {
		middleware.Logger.Log("DEBUG", fmt.Sprintf("Packaged assets: \n%s", strings.Join(assetsList, "\n")))
	}

	// 初始化路由
	r := router.NewRouter()
	loggedRouter := middleware.Logger.HttpMiddleware(r)

	// 初始化数据库连接
	err = database.NewDB(cfg)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to initialize database: %s", err))
		return
	}
	defer database.GetDB().Close()

	// 初始化工作队列
	middleware.InitQueues(cfg)

	// 启动服务
	middleware.Logger.Log("INFO", "Starting server on port "+cfg.Port)
	err = http.ListenAndServe(":"+cfg.Port, loggedRouter)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to start server: %s", err))
	}
}
