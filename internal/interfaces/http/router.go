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

	// 路由
	router.POST("/getCourse", courseHandler.GetCourse)
	router.POST("/generateSummary", summaryHandler.GenerateSummary)

	// 根路径
	router.GET("/", func(c *gin.Context) {
		c.JSON(403, gin.H{"code": 403, "msg": "Forbidden"})
	})

	// 未匹配路由返回 JSON 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": 404, "msg": "not found"})
	})

	return router
}
