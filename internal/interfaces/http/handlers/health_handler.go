package handlers

import (
	"net/http"

	"iwut-smartclass-backend/internal/interfaces/http/dto"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	// 可以添加数据库、队列等依赖用于健康检查
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health 健康检查
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(map[string]string{
		"status": "ok",
	}))
}
