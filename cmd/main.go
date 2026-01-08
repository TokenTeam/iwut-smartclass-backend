package main

import (
	"fmt"
	"iwut-smartclass-backend/assets"
	"iwut-smartclass-backend/internal/application/course"
	"iwut-smartclass-backend/internal/database"
	"iwut-smartclass-backend/internal/infrastructure/config"
	"iwut-smartclass-backend/internal/infrastructure/external"
	"iwut-smartclass-backend/internal/infrastructure/logger"
	"iwut-smartclass-backend/internal/infrastructure/persistence"
	"iwut-smartclass-backend/internal/interfaces/http"
	httpHandlers "iwut-smartclass-backend/internal/interfaces/http/handlers"
	httpMiddleware "iwut-smartclass-backend/internal/interfaces/http/middleware"
	"iwut-smartclass-backend/internal/middleware"
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
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 初始化日志
	appLogger, err := logger.NewLogger(&logger.Config{
		Debug:   cfg.Debug,
		LogSave: cfg.LogSave,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	// 列出所有嵌入的静态资源
	assetsList, err := assets.ListAssets()
	if err != nil {
		appLogger.Error("Failed to load packaged assets", logger.String("error", err.Error()))
	} else {
		appLogger.Debug("Packaged assets loaded", logger.String("assets", strings.Join(assetsList, "\n")))
	}

	// 初始化数据库连接
	err = database.NewDB(cfg)
	if err != nil {
		appLogger.Error("Failed to initialize database", logger.String("error", err.Error()))
		return
	}
	defer func() {
		if db := database.GetDB(); db != nil {
			sqlDB, err := db.DB()
			if err == nil {
				sqlDB.Close()
			}
		}
	}()

	db := database.GetDB()
	if db == nil {
		panic("Database not initialized")
	}

	// 初始化仓储
	courseRepo := persistence.NewCourseRepository(db, appLogger)
	summaryRepo := persistence.NewSummaryRepository(db, appLogger)

	// 初始化外部服务
	userService := external.NewUserService(cfg, appLogger)
	scheduleService := external.NewScheduleService(cfg, appLogger)
	liveCourseService := external.NewLiveCourseService(cfg, appLogger)
	videoAuthService := external.NewVideoAuthService(cfg, appLogger)

	// 初始化应用服务
	courseService := course.NewService(courseRepo, appLogger)

	// 初始化工作队列
	middleware.InitQueues(cfg)
	summaryQueue := middleware.GetQueue("SummaryServiceQueue")

	// 初始化处理器
	courseHandler := httpHandlers.NewCourseHandler(
		courseService,
		summaryRepo,
		userService,
		scheduleService,
		liveCourseService,
		videoAuthService,
		appLogger,
	)
	summaryHandler := httpHandlers.NewSummaryHandler(appLogger, summaryQueue)
	healthHandler := httpHandlers.NewHealthHandler()

	// 设置路由
	router := http.SetupRouter(
		courseHandler,
		summaryHandler,
		healthHandler,
		httpMiddleware.ErrorHandler(),
		httpMiddleware.LoggerMiddleware(appLogger),
	)

	// 启动服务
	appLogger.Info("Starting server", logger.String("port", cfg.Port))
	if err := router.Run(":" + cfg.Port); err != nil {
		appLogger.Error("Failed to start server", logger.String("error", err.Error()))
	}
}
