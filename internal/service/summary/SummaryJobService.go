package summary

import (
	"fmt"
	"iwut-smartclass-backend/assets"
	"iwut-smartclass-backend/internal/asr"
	"iwut-smartclass-backend/internal/config"
	"iwut-smartclass-backend/internal/cos"
	"iwut-smartclass-backend/internal/middleware"
	"iwut-smartclass-backend/internal/service/course"
	"iwut-smartclass-backend/internal/util"
	"os"
	"path/filepath"
)

type Job struct {
	Token        string
	SubID        int
	Task         string
	CourseID     int
	CourseName   string
	VideoURL     string
	SummarySvc   *Service
	ConvertSvc   *ConvertVideoToAudioService
	AsrSvc       *AsrDbService
	SummaryDbSvc *LlmDbService
	Config       *config.Config
}

// Execute 实现 Job 接口的方法
func (j *Job) Execute() error {
	// 声明必要的变量
	var status string
	var asrText string

	if j.Task == "new" {
		// 获取视频密钥
		videoAuthService := course.NewVideoAuthService(j.Token, j.CourseID, j.VideoURL, middleware.Logger)
		videoAuth, err := videoAuthService.VideoAuth()
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to get video auth_key: %s", err))
			return err
		}

		// 拼接带密钥的视频链接
		video := fmt.Sprintf("%s?%s", j.VideoURL, videoAuth)

		// 将视频转换为音频
		audioID, err := j.ConvertSvc.GetAudioId(j.SubID)
		if audioID == "" {
			audioID, err = j.ConvertSvc.Convert(j.SubID, video)
			if err != nil {
				_ = j.SummarySvc.WriteStatus(j.SubID, status)
				middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to convert video to audio: %s", err))
				return err
			}
		}

		audioFileName := audioID + ".aac"
		audioFilePath := filepath.Join("data", "audio", audioFileName)

		// 上传到 COS
		cosService, err := cos.NewCosService(j.Config.TencentSecretId, j.Config.TencentSecretKey, j.Config.BucketUrl)
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to create Tencent COS service: %s", err))
			return err
		}

		err = cosService.UploadFile(audioFilePath, audioFileName)
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to upload file: %s", err))
			return err
		}

		bucketFilePath := j.Config.BucketUrl + "/" + audioFileName

		// ASR 识别
		tencentASRService, err := asr.NewTencentASRService(j.Config.TencentSecretId, j.Config.TencentSecretKey)
		if err != nil {
			_ = j.SummarySvc.WriteStatus(j.SubID, status)
			middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to create Tencent ASR service: %s", err))
			return err
		}

		asrText, err := tencentASRService.Recognize(bucketFilePath)
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

		// 读取 ASR 结果
		asrText, err := j.AsrSvc.GetAsr(j.SubID)
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
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to read prompt template: %s", err))
		_ = j.SummarySvc.WriteStatus(j.SubID, status)
		return err
	}
	prompt := fmt.Sprintf(string(promptTemplate), j.CourseName)

	// 生成摘要
	summaryText, err := util.CallOpenAI(j.Config, prompt, asrText)
	if err != nil {
		_ = j.SummarySvc.WriteStatus(j.SubID, status)
		return err
	}

	// 保存摘要
	_, err = j.SummaryDbSvc.SaveSummary(j.SubID, summaryText)
	if err != nil {
		_ = j.SummarySvc.WriteStatus(j.SubID, status)
		return err
	}

	return nil
}
