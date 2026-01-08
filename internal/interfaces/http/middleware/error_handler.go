package middleware

import (
	"net/http"

	"iwut-smartclass-backend/internal/domain/errors"
	"iwut-smartclass-backend/internal/interfaces/http/dto"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// 处理领域错误
			if domainErr, ok := err.(*errors.DomainError); ok {
				c.JSON(domainErr.HTTPStatus(), dto.ErrorResponse(domainErr.HTTPStatus(), domainErr.Message))
				return
			}

			// 处理其他错误
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "internal server error"))
		}
	}
}
