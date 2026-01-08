package summary

import (
	"context"
	"fmt"
	"iwut-smartclass-backend/assets"
	"iwut-smartclass-backend/internal/asr"
	"iwut-smartclass-backend/internal/config"
	"iwut-smartclass-backend/internal/cos"
	"iwut-smartclass-backend/internal/middleware"
	"iwut-smartclass-backend/internal/service/course"
	"iwut-smartclass-backend/internal/service/user"
	"iwut-smartclass-backend/internal/util"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type Job struct {
	Token        string
	SubID        int
	Task         string
	CourseID     int
	CourseName   string
	VideoURL     string
	Asr          string
	SummarySvc   *Service        `json:"-"`
	ConvertSvc   *ConvertService `json:"-"`
	AsrSvc       *AsrDBService   `json:"-"`
	SummaryDbSvc *LlmDBService   `json:"-"`
	Config       *config.Config  `json:"-"`
}

func (j *Job) GetID() string {
	return fmt.Sprintf("summary-%d-%d", j.SubID, time.Now().UnixNano())
}

func (j *Job) GetData() interface{} {
	return j
}

func (j *Job) GetType() string {
	return "summary"
}

func (j *Job) Execute() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	var status string
	var asrText string
	var timeStamp string

	userInfoService := user.NewGetUserInfoService(j.Token, middleware.Logger)

	// 获取用户信息
	userInfo, err := userInfoService.GetUserInfo()
	if err != nil {
		middleware.Logger.Log("DEBUG", fmt.Sprintf("Failed to get UserInfo: %v", err))
		return err
	}

	if j.Task == "new" && j.Asr == "" {
		_ = j.SummarySvc.WriteStatus(j.SubID, "generating")

		// 获取视频密钥
		videoAuthService := course.NewVideoAuthService(j.Token, j.CourseID, j.VideoURL, middleware.Logger)
		videoAuth, err := videoAuthService.VideoAuth(&userInfo)
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to get video auth_key: %v", err))
			return err
		}

		// 拼接带密钥的视频链接
		video := fmt.Sprintf("%s?%s", j.VideoURL, videoAuth)

		// 将视频转换为音频
		audioID, err := j.ConvertSvc.GetAudioId(ctx, j.SubID)
		if audioID == "" {
			convertCtx, convertCancel := context.WithTimeout(ctx, 5*time.Minute)
			audioID, err = j.ConvertSvc.Convert(convertCtx, j.SubID, video)
			convertCancel()
			if err != nil {
				_ = j.SummarySvc.WriteStatus(j.SubID, status)
				_ = j.ConvertSvc.SaveAudioId(ctx, j.SubID, audioID)
				middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to convert video to audio: %v", err))
				return err
			}
		}

		audioFileName := audioID + ".aac"
		audioFilePath := filepath.Join("temp", "audio", audioFileName)

		// 上传到 COS
		cosService, err := cos.NewCosService(j.Config.TencentSecretId[0], j.Config.TencentSecretKey[0], j.Config.BucketUrl)
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to create Tencent COS service: %v", err))
			return err
		}

		err = cosService.UploadFile(audioFilePath, audioFileName)
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to upload file: %v", err))
			return err
		}

		bucketFilePath := j.Config.BucketUrl + "/" + audioFileName

		// ASR 识别
		randIdx := rand.Intn(len(j.Config.TencentSecretId))
		tencentASRService, err := asr.NewTencentASRService(j.Config.TencentSecretId[randIdx], j.Config.TencentSecretKey[randIdx])
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to create Tencent ASR service: %v", err))
			return err
		}

		asrText, err = tencentASRService.Recognize(bucketFilePath)
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			return err
		}

		// 保存 ASR 结果
		_, err = j.AsrSvc.SaveAsr(j.SubID, asrText)
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			return err
		}

		// 清理文件
		middleware.Logger.Log("INFO", fmt.Sprintf("Delete file: %s", audioFileName))
		_ = cosService.DeleteFile(audioFileName)
		_ = os.Remove(audioFilePath)
	}

	if j.Task == "regenerate" {
		status = "finished"
		var err error

		// 初始化Summary行
		timeStamp, err = j.SummaryDbSvc.InitNewSummary(j.SubID, userInfo.Account)
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			return err
		}

		// 读取 ASR 结果
		asrText, err = j.AsrSvc.GetAsr(j.SubID)
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			return err
		}
		if asrText == "" {
			middleware.Logger.Log("ERROR", "ASR text is empty")
			return fmt.Errorf("ASR text is empty")
		}
	}

	// 读取提示词
	promptTemplate, err := assets.GetAssets("templates/course_summary_prompt.txt")
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to read prompt template: %v", err))
		_ = j.SummarySvc.WriteStatus(j.SubID, status)
		return err
	}
	prompt := fmt.Sprintf(string(promptTemplate), j.CourseName)

	// 生成摘要
	summaryText, token, err := util.CallOpenAI(j.Config, prompt, asrText)
	if err != nil {
		_ = j.SummarySvc.WriteStatus(j.SubID, status)
		return err
	}

	// 保存摘要
	if j.Task == "new" && j.Asr == "" {
		err = j.SummaryDbSvc.SaveSummary(j.SubID, userInfo.Account, summaryText, j.Config.OpenaiModel, token)
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			return err
		}
	}

	if j.Task == "regenerate" {
		err = j.SummaryDbSvc.SaveUserSummary(j.SubID, userInfo.Account, timeStamp, summaryText, j.Config.OpenaiModel, token)
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			return err
		}
	}

	return nil
}
