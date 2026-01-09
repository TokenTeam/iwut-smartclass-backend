package summary

import (
	"context"
	"crypto/md5"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"iwut-smartclass-backend/assets"
	"iwut-smartclass-backend/internal/application/course"
	"iwut-smartclass-backend/internal/domain/errors"
	"iwut-smartclass-backend/internal/domain/summary"
	"iwut-smartclass-backend/internal/infrastructure/config"
	"iwut-smartclass-backend/internal/infrastructure/external"
	"iwut-smartclass-backend/internal/infrastructure/logger"
)

// SummaryJob 摘要生成任务
type SummaryJob struct {
	Token      string
	SubID      int
	Task       string
	CourseID   int
	CourseName string
	VideoURL   string
	Asr        string

	// 依赖注入
	courseService    *course.Service
	summaryRepo      summary.Repository
	userService      *external.UserService
	videoAuthService *external.VideoAuthService
	ffmpegService    *external.FFmpegService
	cosService       *external.COSService
	asrService       *external.ASRService
	openaiService    *external.OpenAIService
	config           *config.Config
	logger           logger.Logger
}

// NewSummaryJob 创建摘要任务
func NewSummaryJob(
	token string,
	subID int,
	task string,
	courseID int,
	courseName string,
	videoURL string,
	asr string,
	courseService *course.Service,
	summaryRepo summary.Repository,
	userService *external.UserService,
	videoAuthService *external.VideoAuthService,
	ffmpegService *external.FFmpegService,
	cosService *external.COSService,
	asrService *external.ASRService,
	openaiService *external.OpenAIService,
	cfg *config.Config,
	logger logger.Logger,
) *SummaryJob {
	return &SummaryJob{
		Token:            token,
		SubID:            subID,
		Task:             task,
		CourseID:         courseID,
		CourseName:       courseName,
		VideoURL:         videoURL,
		Asr:              asr,
		courseService:    courseService,
		summaryRepo:      summaryRepo,
		userService:      userService,
		videoAuthService: videoAuthService,
		ffmpegService:    ffmpegService,
		cosService:       cosService,
		asrService:       asrService,
		openaiService:    openaiService,
		config:           cfg,
		logger:           logger,
	}
}

// GetID 获取任务ID
func (j *SummaryJob) GetID() string {
	return fmt.Sprintf("summary-%d", j.SubID)
}

// GetData 获取任务数据（用于序列化）
func (j *SummaryJob) GetData() interface{} {
	return map[string]interface{}{
		"token":       j.Token,
		"sub_id":      j.SubID,
		"task":        j.Task,
		"course_id":   j.CourseID,
		"course_name": j.CourseName,
		"video_url":   j.VideoURL,
		"asr":         j.Asr,
	}
}

// GetType 获取任务类型
func (j *SummaryJob) GetType() string {
	return "summary"
}

// Execute 执行任务
func (j *SummaryJob) Execute() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// 获取用户信息
	userInfo, err := j.userService.GetUserInfo(j.Token)
	if err != nil {
		j.logger.Error("failed to get user info", logger.String("error", err.Error()))
		return err
	}

	var asrText string

	if j.Task == "new" && j.Asr == "" {
		// 更新状态为生成中
		if err := j.courseService.UpdateSummaryStatus(ctx, j.SubID, "generating"); err != nil {
			j.logger.Error("failed to update summary status", logger.String("error", err.Error()))
			return err
		}

		// 获取视频密钥
		authKey, err := j.videoAuthService.GetVideoAuthKey(j.Token, j.CourseID, j.SubID)
		if err != nil {
			j.logger.Error("failed to get video auth key", logger.String("error", err.Error()))
			_ = j.courseService.UpdateSummaryStatus(ctx, j.SubID, "")
			return err
		}

		// 拼接带密钥的视频链接
		parsedURL, err := url.Parse(j.VideoURL)
		if err != nil {
			j.logger.Error("failed to parse video URL", logger.String("error", err.Error()))
			_ = j.courseService.UpdateSummaryStatus(ctx, j.SubID, "")
			return errors.NewInternalError("failed to parse video URL", err)
		}
		timestamp := time.Now().Unix()
		md5Input := fmt.Sprintf("%s%d%d%s%d", parsedURL.Path, userInfo.ID, userInfo.TenantID, userInfo.ReversePhone(), timestamp)
		md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(md5Input)))
		videoAuth := fmt.Sprintf("auth_key=%s&t=%d-%d-%s", authKey, userInfo.ID, timestamp, md5Hash)
		video := fmt.Sprintf("%s?%s", j.VideoURL, videoAuth)

		// 检查是否已有音频ID
		courseEntity, err := j.courseService.GetCourse(ctx, j.SubID)
		if err != nil {
			j.logger.Error("failed to get course", logger.String("error", err.Error()))
			_ = j.courseService.UpdateSummaryStatus(ctx, j.SubID, "")
			return err
		}

		audioID := courseEntity.AudioID
		if audioID == "" {
			// 将视频转换为音频
			audioID = fmt.Sprintf("%d-%d", j.SubID, time.Now().Unix())
			audioFileName := audioID + ".aac"
			audioFilePath := filepath.Join("temp", "audio", audioFileName)

			// 创建目录
			if err := os.MkdirAll(filepath.Dir(audioFilePath), 0755); err != nil {
				j.logger.Error("failed to create directory", logger.String("error", err.Error()))
				_ = j.courseService.UpdateSummaryStatus(ctx, j.SubID, "")
				return errors.NewInternalError("failed to create directory", err)
			}

			// 转换视频为音频
			convertCtx, convertCancel := context.WithTimeout(ctx, 5*time.Minute)
			err = j.ffmpegService.ConvertVideoToAudio(convertCtx, video, audioFilePath)
			convertCancel()
			if err != nil {
				j.logger.Error("failed to convert video to audio", logger.String("error", err.Error()))
				_ = j.courseService.UpdateSummaryStatus(ctx, j.SubID, "")
				return err
			}

			// 保存音频ID
			if err := j.courseService.UpdateAudioID(ctx, j.SubID, audioID); err != nil {
				j.logger.Error("failed to save audio id", logger.String("error", err.Error()))
			}
		}

		audioFileName := audioID + ".aac"
		audioFilePath := filepath.Join("temp", "audio", audioFileName)

		// 上传到 COS
		err = j.cosService.UploadFile(audioFilePath, audioFileName)
		if err != nil {
			j.logger.Error("failed to upload file", logger.String("error", err.Error()))
			_ = j.courseService.UpdateSummaryStatus(ctx, j.SubID, "")
			return err
		}

		bucketFilePath := j.config.BucketUrl + "/" + audioFileName

		// ASR 识别
		randIdx := rand.Intn(len(j.config.TencentSecretId))
		asrSvc, err := external.NewASRService(j.config.TencentSecretId[randIdx], j.config.TencentSecretKey[randIdx], j.logger)
		if err != nil {
			j.logger.Error("failed to create ASR service", logger.String("error", err.Error()))
			_ = j.courseService.UpdateSummaryStatus(ctx, j.SubID, "")
			return err
		}

		asrText, err = asrSvc.Recognize(bucketFilePath)
		if err != nil {
			j.logger.Error("failed to recognize audio", logger.String("error", err.Error()))
			_ = j.courseService.UpdateSummaryStatus(ctx, j.SubID, "")
			return err
		}

		// 保存 ASR 结果
		if err := j.courseService.UpdateAsr(ctx, j.SubID, asrText); err != nil {
			j.logger.Error("failed to save ASR", logger.String("error", err.Error()))
			_ = j.courseService.UpdateSummaryStatus(ctx, j.SubID, "")
			return err
		}

		// 清理文件
		j.logger.Info("deleting temporary files", logger.String("file", audioFileName))
		_ = j.cosService.DeleteFile(audioFileName)
		_ = os.Remove(audioFilePath)
	}

	if j.Task == "new" && j.Asr != "" {
		// 使用已有的ASR文本
		asrText = j.Asr
		if asrText == "" {
			j.logger.Error("ASR text is empty")
			return errors.NewInternalError("ASR text is empty", fmt.Errorf("ASR text is empty"))
		}
	}

	if j.Task == "regenerate" {
		// 初始化Summary行
		_, err := j.summaryRepo.InitNewSummary(ctx, j.SubID, userInfo.Account)
		if err != nil {
			j.logger.Error("failed to init new summary", logger.String("error", err.Error()))
			return err
		}

		// 读取 ASR 结果
		courseEntity, err := j.courseService.GetCourse(ctx, j.SubID)
		if err != nil {
			j.logger.Error("failed to get course", logger.String("error", err.Error()))
			return err
		}
		asrText = courseEntity.Asr
		if asrText == "" {
			j.logger.Error("ASR text is empty")
			return errors.NewInternalError("ASR text is empty", fmt.Errorf("ASR text is empty"))
		}
	}

	// 读取提示词
	promptTemplate, err := assets.GetAssets("templates/course_summary_prompt.txt")
	if err != nil {
		j.logger.Error("failed to read prompt template", logger.String("error", err.Error()))
		return errors.NewInternalError("failed to read prompt template", err)
	}
	prompt := fmt.Sprintf(string(promptTemplate), j.CourseName)

	// 生成摘要
	summaryText, token, err := j.openaiService.CallOpenAI(prompt, asrText)
	if err != nil {
		j.logger.Error("failed to call OpenAI", logger.String("error", err.Error()))
		return err
	}

	// 保存摘要
	if j.Task == "new" {
		err = j.courseService.UpdateSummary(ctx, j.SubID, summaryText, j.config.OpenaiModel, token, userInfo.Account)
		if err != nil {
			j.logger.Error("failed to save summary", logger.String("error", err.Error()))
			return err
		}
	}

	if j.Task == "regenerate" {
		// 查找已存在的摘要
		summaries, err := j.summaryRepo.FindBySubIDAndUser(ctx, j.SubID, userInfo.Account)
		if err != nil || len(summaries) == 0 {
			j.logger.Error("failed to find summary", logger.String("error", err.Error()))
			return errors.NewInternalError("failed to find summary", err)
		}

		// 更新最新的摘要
		summaryEntity := summaries[0]
		summaryEntity.Summary = summaryText
		summaryEntity.Model = j.config.OpenaiModel
		summaryEntity.Token = token
		if err := j.summaryRepo.Update(ctx, summaryEntity); err != nil {
			j.logger.Error("failed to update summary", logger.String("error", err.Error()))
			return err
		}
	}

	return nil
}
