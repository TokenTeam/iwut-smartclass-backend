package middleware

import (
	"fmt"
	"time"

	loggerPkg "iwut-smartclass-backend/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware(logger loggerPkg.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		logger.Info("http request",
			loggerPkg.String("method", method),
			loggerPkg.String("path", path),
			loggerPkg.String("status", fmt.Sprintf("%d", status)),
			loggerPkg.String("latency", latency.String()),
			loggerPkg.String("client_ip", c.ClientIP()),
		)
	}
}
