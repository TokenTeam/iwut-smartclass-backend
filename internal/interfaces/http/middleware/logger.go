package middleware

import (
	"time"

	"iwut-smartclass-backend/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("http request",
			logger.String("method", method),
			logger.String("path", path),
			logger.String("status", string(rune(status))),
			logger.String("latency", latency.String()),
			logger.String("client_ip", c.ClientIP()),
		)
	}
}
