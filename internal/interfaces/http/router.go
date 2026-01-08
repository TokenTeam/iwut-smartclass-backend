package http

import (
	"iwut-smartclass-backend/internal/interfaces/http/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(
	courseHandler *handlers.CourseHandler,
	summaryHandler *handlers.SummaryHandler,
	healthHandler *handlers.HealthHandler,
	errorHandler gin.HandlerFunc,
	loggerMiddleware gin.HandlerFunc,
) *gin.Engine {
	router := gin.New()

	// 中间件
	router.Use(gin.Recovery())
	router.Use(loggerMiddleware)
	router.Use(errorHandler)

	// 健康检查
	router.GET("/health", healthHandler.Health)

	// API路由
	api := router.Group("/api")
	{
		api.POST("/getCourse", courseHandler.GetCourse)
		api.POST("/generateSummary", summaryHandler.GenerateSummary)
	}

	// 根路径
	router.GET("/", func(c *gin.Context) {
		c.JSON(403, gin.H{"code": 403, "msg": "Forbidden"})
	})

	return router
}
