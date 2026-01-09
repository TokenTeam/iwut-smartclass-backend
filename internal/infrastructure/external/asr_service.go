package external

import (
	"fmt"
	asr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/asr/v20190614"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"iwut-smartclass-backend/internal/domain/errors"
	"iwut-smartclass-backend/internal/infrastructure/logger"
	"regexp"
	"time"
)

// ASRService ASR服务
type ASRService struct {
	client *asr.Client
	logger logger.Logger
}

// NewASRService 创建ASR服务
func NewASRService(secretId, secretKey string, logger logger.Logger) (*ASRService, error) {
	credential := common.NewCredential(secretId, secretKey)
	cpf := profile.NewClientProfile()
	client, err := asr.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		return nil, errors.NewInternalError("failed to create ASR client", err)
	}
	return &ASRService{client: client, logger: logger}, nil
}

// Recognize 识别音频
func (s *ASRService) Recognize(audioFilePath string) (string, error) {
	// 配置识别参数
	request := asr.NewCreateRecTaskRequest()
	request.EngineModelType = common.StringPtr("16k_zh_dialect")
	request.ChannelNum = common.Uint64Ptr(1)
	request.ResTextFormat = common.Uint64Ptr(0)
	request.SourceType = common.Uint64Ptr(0)
	request.SpeakerDiarization = int64Ptr(1)
	request.SpeakerNumber = int64Ptr(0)
	request.ConvertNumMode = int64Ptr(3)
	request.Url = common.StringPtr(audioFilePath)

	s.logger.Info("creating ASR task", logger.String("file", audioFilePath))

	response, err := s.client.CreateRecTask(request)
	if err != nil {
		s.logger.Error("failed to create ASR task", logger.String("error", err.Error()))
		return "", errors.NewExternalError("asr", err)
	}

	taskId := response.Response.Data.TaskId
	s.logger.Info("ASR task created", logger.String("taskId", fmt.Sprintf("%d", *taskId)))

	// 查询识别结果
	for {
		resultRequest := asr.NewDescribeTaskStatusRequest()
		resultRequest.TaskId = taskId

		resultResponse, err := s.client.DescribeTaskStatus(resultRequest)
		if err != nil {
			s.logger.Error("failed to get task status", logger.String("error", err.Error()))
			return "", errors.NewExternalError("asr", err)
		}

		if *resultResponse.Response.Data.Status == 2 {
			s.logger.Info("ASR task finished", logger.String("taskId", fmt.Sprintf("%d", *taskId)))
			resultText := *resultResponse.Response.Data.Result

			// 使用正则表达式移除时间戳
			re := regexp.MustCompile(`\[\d{1,3}:\d{1,2}\.\d{3},\d{1,3}:\d{1,2}\.\d{3},\d]\s*`)
			cleanedText := re.ReplaceAllString(resultText, "")

			return cleanedText, nil
		} else if *resultResponse.Response.Data.Status == 3 {
			errorMsg := ""
			if resultResponse.Response.Data.ErrorMsg != nil {
				errorMsg = *resultResponse.Response.Data.ErrorMsg
			}
			s.logger.Error("ASR task failed", logger.String("error", errorMsg))
			return "", errors.NewExternalError("asr", fmt.Errorf("task failed: %s", errorMsg))
		}

		// 20 秒查询一次
		time.Sleep(20 * time.Second)
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}
