package handlers

import (
	"net/http"

	appCourse "iwut-smartclass-backend/internal/application/course"
	appSummary "iwut-smartclass-backend/internal/application/summary"
	"iwut-smartclass-backend/internal/domain/errors"
	domainSummary "iwut-smartclass-backend/internal/domain/summary"
	"iwut-smartclass-backend/internal/infrastructure/config"
	"iwut-smartclass-backend/internal/infrastructure/external"
	"iwut-smartclass-backend/internal/infrastructure/logger"
	"iwut-smartclass-backend/internal/middleware"
	"iwut-smartclass-backend/internal/interfaces/http/dto"

	"github.com/gin-gonic/gin"
)

// SummaryHandler 摘要处理器
type SummaryHandler struct {
	logger           logger.Logger
	queue            *middleware.WorkQueue
	courseService    *appCourse.Service
	summaryRepo      domainSummary.Repository
	userService      *external.UserService
	videoAuthService *external.VideoAuthService
	ffmpegService    *external.FFmpegService
	cosService       *external.COSService
	asrService       *external.ASRService
	openaiService    *external.OpenAIService
	config           *config.Config
}

// NewSummaryHandler 创建摘要处理器
func NewSummaryHandler(
	logger logger.Logger,
	queue *middleware.WorkQueue,
	courseService *appCourse.Service,
	summaryRepo domainSummary.Repository,
	userService *external.UserService,
	videoAuthService *external.VideoAuthService,
	ffmpegService *external.FFmpegService,
	cosService *external.COSService,
	asrService *external.ASRService,
	openaiService *external.OpenAIService,
	cfg *config.Config,
) *SummaryHandler {
	return &SummaryHandler{
		logger:           logger,
		queue:            queue,
		courseService:    courseService,
		summaryRepo:      summaryRepo,
		userService:      userService,
		videoAuthService: videoAuthService,
		ffmpegService:    ffmpegService,
		cosService:       cosService,
		asrService:       asrService,
		openaiService:    openaiService,
		config:           cfg,
	}
}

// GenerateSummary 生成摘要
func (h *SummaryHandler) GenerateSummary(c *gin.Context) {
	var req dto.GenerateSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewValidationError("invalid request", err))
		return
	}

	ctx := c.Request.Context()

	// 获取课程信息
	courseEntity, err := h.courseService.GetCourse(ctx, req.SubID)
	if err != nil {
		c.Error(err)
		return
	}

	if !courseEntity.HasVideo() {
		c.Error(errors.NewNotFoundError("video"))
		return
	}

	// 创建摘要任务
	job := appSummary.NewSummaryJob(
		req.Token,
		req.SubID,
		req.Task,
		courseEntity.CourseID,
		courseEntity.Name,
		courseEntity.Video,
		courseEntity.Asr,
		h.courseService,
		h.summaryRepo,
		h.userService,
		h.videoAuthService,
		h.ffmpegService,
		h.cosService,
		h.asrService,
		h.openaiService,
		h.config,
		h.logger,
	)

	// 添加到队列
	h.queue.AddJob(job)

	c.JSON(http.StatusOK, dto.SuccessResponse(map[string]interface{}{
		"sub_id":         req.SubID,
		"summary_status": "generating",
	}))
}
