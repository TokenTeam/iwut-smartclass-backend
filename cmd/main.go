package main

import (
	"fmt"
	"iwut-smartclass-backend/assets"
	"iwut-smartclass-backend/internal/config"
	"iwut-smartclass-backend/internal/database"
	"iwut-smartclass-backend/internal/middleware"
	"iwut-smartclass-backend/internal/router"
	"net/http"
	"strings"
)

func main() {
	fmt.Println(`
	  _____                      _     _____ _
	 / ____|                    | |   / ____| |
	| (___  _ __ ___   __ _ _ __| |_ | |    | | __ _ ___ ___
	 \___ \| '_ ' _ \ / _' | '__| __|| |    | |/ _' / __/ __|
	 ____) | | | | | | (_| | |  | |_ | |____| | (_| \__ \__ \
	|_____/|_| |_| |_|\__,_|_|   \__| \_____|_|\__,_|___/___/
	`)

	// 加载配置
	cfg := config.LoadConfig()
	middleware.Logger = middleware.NewLogger(cfg)

	// 列出所有嵌入的静态资源
	assetsList, err := assets.ListAssets()
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to load packaged assets: %v", err))
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
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to initialize database: %v", err))
		return
	}
	defer database.GetDB().Close()

	// 初始化工作队列
	middleware.InitQueues(cfg)

	// 启动服务
	middleware.Logger.Log("INFO", fmt.Sprintf("Starting server on port %s", cfg.Port))
	err = http.ListenAndServe(":"+cfg.Port, loggedRouter)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to start server: %v", err))
	}
}
