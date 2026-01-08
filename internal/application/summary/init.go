package summary

import (
	"encoding/json"
	"fmt"
	"iwut-smartclass-backend/internal/application/course"
	"iwut-smartclass-backend/internal/database"
	"iwut-smartclass-backend/internal/infrastructure/config"
	"iwut-smartclass-backend/internal/infrastructure/external"
	"iwut-smartclass-backend/internal/infrastructure/logger"
	"iwut-smartclass-backend/internal/infrastructure/persistence"
	"iwut-smartclass-backend/internal/middleware"
)

func init() {
	middleware.RegisterGlobalLoader("summary", func(data []byte, cfg *config.Config, logger logger.Logger) (middleware.Job, error) {
		var jobData struct {
			Token       string `json:"token"`
			SubID       int    `json:"sub_id"`
			Task        string `json:"task"`
			CourseID    int    `json:"course_id"`
			CourseName  string `json:"course_name"`
			VideoURL    string `json:"video_url"`
			Asr         string `json:"asr"`
		}
		if err := json.Unmarshal(data, &jobData); err != nil {
			return nil, err
		}

		// 重新注入依赖
		db := database.GetDB()
		if db == nil {
			return nil, fmt.Errorf("database not initialized")
		}

		// 使用传入的logger
		appLogger := logger

		// 创建仓储
		courseRepo := persistence.NewCourseRepository(db, appLogger)
		summaryRepo := persistence.NewSummaryRepository(db, appLogger)

		// 创建应用服务
		courseService := course.NewService(courseRepo, appLogger)

		// 创建外部服务
		userService := external.NewUserService(cfg, appLogger)
		videoAuthService := external.NewVideoAuthService(cfg, appLogger)
		ffmpegService := external.NewFFmpegService(appLogger)
		
		// COS和ASR服务需要根据配置创建
		cosService, _ := external.NewCOSService(cfg.TencentSecretId[0], cfg.TencentSecretKey[0], cfg.BucketUrl, appLogger)
		asrService, _ := external.NewASRService(cfg.TencentSecretId[0], cfg.TencentSecretKey[0], appLogger)
		openaiService := external.NewOpenAIService(cfg, appLogger)

		return NewSummaryJob(
			jobData.Token,
			jobData.SubID,
			jobData.Task,
			jobData.CourseID,
			jobData.CourseName,
			jobData.VideoURL,
			jobData.Asr,
			courseService,
			summaryRepo,
			userService,
			videoAuthService,
			ffmpegService,
			cosService,
			asrService,
			openaiService,
			cfg,
			appLogger,
		), nil
	})
}
