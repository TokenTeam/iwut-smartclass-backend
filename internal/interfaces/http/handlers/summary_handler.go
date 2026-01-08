package handlers

import (
	"net/http"

	"iwut-smartclass-backend/internal/domain/errors"
	"iwut-smartclass-backend/internal/infrastructure/logger"
	"iwut-smartclass-backend/internal/middleware"
	"iwut-smartclass-backend/internal/interfaces/http/dto"

	"github.com/gin-gonic/gin"
)

// SummaryHandler 摘要处理器
type SummaryHandler struct {
	logger logger.Logger
	queue  *middleware.WorkQueue
}

// NewSummaryHandler 创建摘要处理器
func NewSummaryHandler(logger logger.Logger, queue *middleware.WorkQueue) *SummaryHandler {
	return &SummaryHandler{
		logger: logger,
		queue:  queue,
	}
}

// GenerateSummary 生成摘要
func (h *SummaryHandler) GenerateSummary(c *gin.Context) {
	var req dto.GenerateSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewValidationError("invalid request", err))
		return
	}

	// TODO: 实现摘要生成逻辑
	// 这里需要重构原有的Job系统

	c.JSON(http.StatusOK, dto.SuccessResponse(map[string]interface{}{
		"sub_id":         req.SubID,
		"summary_status": "generating",
	}))
}
