package handlers

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"iwut-smartclass-backend/internal/application/course"
	"iwut-smartclass-backend/internal/domain/course"
	"iwut-smartclass-backend/internal/domain/errors"
	"iwut-smartclass-backend/internal/domain/summary"
	"iwut-smartclass-backend/internal/infrastructure/external"
	"iwut-smartclass-backend/internal/infrastructure/logger"
	"iwut-smartclass-backend/internal/interfaces/http/dto"

	"github.com/gin-gonic/gin"
)

// CourseHandler 课程处理器
type CourseHandler struct {
	courseService      *course.Service
	summaryRepo        summary.Repository
	userService        *external.UserService
	scheduleService    *external.ScheduleService
	liveCourseService  *external.LiveCourseService
	videoAuthService   *external.VideoAuthService
	logger             logger.Logger
}

// NewCourseHandler 创建课程处理器
func NewCourseHandler(
	courseService *course.Service,
	summaryRepo summary.Repository,
	userService *external.UserService,
	scheduleService *external.ScheduleService,
	liveCourseService *external.LiveCourseService,
	videoAuthService *external.VideoAuthService,
	logger logger.Logger,
) *CourseHandler {
	return &CourseHandler{
		courseService:     courseService,
		summaryRepo:       summaryRepo,
		userService:       userService,
		scheduleService:   scheduleService,
		liveCourseService: liveCourseService,
		videoAuthService:  videoAuthService,
		logger:            logger,
	}
}

// GetCourse 获取课程
func (h *CourseHandler) GetCourse(c *gin.Context) {
	var req dto.GetCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewValidationError("invalid request", err))
		return
	}

	ctx := c.Request.Context()

	// 获取课程表
	scheduleData, err := h.scheduleService.GetSchedule(req.Token, req.Date, req.CourseName)
	if err != nil {
		c.Error(err)
		return
	}

	// 解析subID和courseID
	if len(scheduleData.Result.List) == 0 || len(scheduleData.Result.List[0].Course) == 0 {
		c.Error(errors.NewNotFoundError("course"))
		return
	}

	subID, err := strconv.Atoi(scheduleData.Result.List[0].Course[0].ID)
	if err != nil {
		c.Error(errors.NewValidationError("invalid sub_id", err))
		return
	}

	courseID, err := strconv.Atoi(scheduleData.Result.List[0].Course[0].CourseID)
	if err != nil {
		c.Error(errors.NewValidationError("invalid course_id", err))
		return
	}

	// 获取用户信息
	userInfo, err := h.userService.GetUserInfo(req.Token)
	if err != nil {
		c.Error(err)
		return
	}

	// 尝试从数据库获取课程
	courseEntity, err := h.courseService.GetCourse(ctx, subID)
	if err != nil {
		// 如果不存在，从外部服务获取
		liveCourseData, err := h.liveCourseService.SearchLiveCourse(req.Token, subID, courseID)
		if err != nil {
			c.Error(err)
			return
		}

		// 创建新课程实体
		courseEntity = &course.Course{
			SubID:      subID,
			CourseID:   courseID,
			Name:       liveCourseData["name"].(string),
			Teacher:    liveCourseData["teacher"].(string),
			Location:   liveCourseData["location"].(string),
			Date:       liveCourseData["date"].(string),
			Time:       liveCourseData["time"].(string),
			Video:      liveCourseData["video"].(string),
			SummaryStatus: "",
			SummaryData:   "",
			Model:          "",
			Token:          0,
			SummaryUser:    "",
		}

		// 保存到数据库
		if err := h.courseService.SaveCourse(ctx, courseEntity); err != nil {
			c.Error(err)
			return
		}
	} else if !courseEntity.HasVideo() {
		// 如果视频为空，尝试再次获取
		liveCourseData, err := h.liveCourseService.SearchLiveCourse(req.Token, subID, courseID)
		if err != nil {
			c.Error(err)
			return
		}

		// 更新视频链接
		videoURL := liveCourseData["video"].(string)
		if err := h.courseService.UpdateVideo(ctx, subID, videoURL); err != nil {
			c.Error(err)
			return
		}
		courseEntity.Video = videoURL
	}

	// 获取用户摘要
	userSummaries, err := h.summaryRepo.FindBySubIDAndUser(ctx, subID, userInfo.Account)
	if err != nil {
		c.Error(err)
		return
	}

	// 构建响应
	response := map[string]interface{}{
		"course_id": courseEntity.CourseID,
		"sub_id":    courseEntity.SubID,
		"name":      courseEntity.Name,
		"teacher":   courseEntity.Teacher,
		"location":  courseEntity.Location,
		"date":      courseEntity.Date,
		"time":      courseEntity.Time,
		"video":     courseEntity.Video,
		"asr":       courseEntity.Asr,
		"summary": map[string]string{
			"status": courseEntity.SummaryStatus,
			"data":   courseEntity.SummaryData,
			"model":  courseEntity.Model,
			"token":  fmt.Sprintf("%d", courseEntity.Token),
		},
	}

	// 如果用户有摘要，使用用户的摘要
	if len(userSummaries) > 0 {
		status := "generating"
		if !userSummaries[0].IsEmpty() {
			status = "finished"
		}
		response["summary"] = map[string]string{
			"status": status,
			"data":   userSummaries[0].Summary,
			"model":  userSummaries[0].Model,
			"token":  fmt.Sprintf("%d", userSummaries[0].Token),
		}
	}

	// 添加视频认证
	if courseEntity.HasVideo() {
		authKey, err := h.videoAuthService.GetVideoAuthKey(req.Token, courseID, subID)
		if err != nil {
			c.Error(err)
			return
		}

		// 生成视频认证参数
		reversedPhone := userInfo.ReversePhone()
		timestamp := time.Now().Unix()
		
		// 计算MD5
		parsedURL, _ := url.Parse(courseEntity.Video)
		md5Input := fmt.Sprintf("%s%d%d%s%d", parsedURL.Path, userInfo.ID, userInfo.TenantID, reversedPhone, timestamp)
		md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(md5Input)))
		
		videoAuth := fmt.Sprintf("auth_key=%s&t=%d-%d-%s", authKey, userInfo.ID, timestamp, md5Hash)
		response["video"] = fmt.Sprintf("%s?%s", courseEntity.Video, videoAuth)
	}

	h.logger.Info("get course success",
		logger.String("course_name", req.CourseName),
		logger.String("course_id", fmt.Sprintf("%d", courseID)),
		logger.String("sub_id", fmt.Sprintf("%d", subID)),
	)

	c.JSON(http.StatusOK, dto.SuccessResponse(response))
}
