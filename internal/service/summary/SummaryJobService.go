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
	// 获取视频密钥
	videoAuthService := course.NewVideoAuthService(j.Token, j.CourseID, j.VideoURL, middleware.Logger)
	videoAuth, err := videoAuthService.VideoAuth()
	if err != nil {
		_ = j.SummarySvc.WriteStatus(j.SubID, "")
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
			_ = j.SummarySvc.WriteStatus(j.SubID, "")
			middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to convert video to audio: %s", err))
			return err
		}
	}

	audioFileName := audioID + ".aac"
	audioFilePath := filepath.Join("data", "audio", audioFileName)

	// 上传到 COS
	cosService, err := cos.NewCosService(j.Config.TencentSecretId, j.Config.TencentSecretKey, j.Config.BucketUrl)
	if err != nil {
		_ = j.SummarySvc.WriteStatus(j.SubID, "")
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to create Tencent COS service: %s", err))
		return err
	}

	err = cosService.UploadFile(audioFilePath, audioFileName)
	if err != nil {
		_ = j.SummarySvc.WriteStatus(j.SubID, "")
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to upload file: %s", err))
		return err
	}

	bucketFilePath := j.Config.BucketUrl + "/" + audioFileName

	// ASR 识别
	tencentASRService, err := asr.NewTencentASRService(j.Config.TencentSecretId, j.Config.TencentSecretKey)
	if err != nil {
		_ = j.SummarySvc.WriteStatus(j.SubID, "")
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to create Tencent ASR service: %s", err))
		return err
	}

	asrText, err := tencentASRService.Recognize(bucketFilePath)
	if err != nil {
		_ = j.SummarySvc.WriteStatus(j.SubID, "")
		return err
	}

	// 保存 ASR 结果
	_, err = j.AsrSvc.SaveAsr(j.SubID, asrText)
	if err != nil {
		_ = j.SummarySvc.WriteStatus(j.SubID, "")
		return err
	}

	// 清理文件
	middleware.Logger.Log("INFO", fmt.Sprintf("Delete file: %s", audioFileName))
	_ = cosService.DeleteFile(audioFileName)
	_ = os.Remove(audioFilePath)

	// 读取提示词
	promptTemplate, err := assets.GetAssets("assets/templates/course_summary_prompt.txt")
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to read prompt template: %s", err))
		_ = j.SummarySvc.WriteStatus(j.SubID, "")
		return err
	}
	prompt := fmt.Sprintf(string(promptTemplate), j.CourseName)

	// 生成摘要
	summaryText, err := util.CallOpenAI(j.Config, prompt, asrText)
	if err != nil {
		_ = j.SummarySvc.WriteStatus(j.SubID, "")
		return err
	}

	// 保存摘要
	_, err = j.SummaryDbSvc.SaveSummary(j.SubID, summaryText)
	if err != nil {
		_ = j.SummarySvc.WriteStatus(j.SubID, "")
		return err
	}

	return nil
}
