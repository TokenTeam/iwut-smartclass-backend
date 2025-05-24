package asr

import (
	"fmt"
	asr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/asr/v20190614"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"iwut-smartclass-backend/internal/middleware"
	"regexp"
	"time"
)

type Service struct {
	client *asr.Client
}

func NewTencentASRService(secretId, secretKey string) (*Service, error) {
	credential := common.NewCredential(secretId, secretKey)
	cpf := profile.NewClientProfile()
	client, _ := asr.NewClient(credential, "ap-guangzhou", cpf)
	return &Service{client: client}, nil
}

func (s *Service) Recognize(audioFilePath string) (string, error) {
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

	middleware.Logger.Log("INFO", fmt.Sprintf("[ASR] Creating task with audio file: %s", audioFilePath))

	response, err := s.client.CreateRecTask(request)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[ASR] Failed to create task: %v", err))
		return "", err
	}

	taskId := response.Response.Data.TaskId
	middleware.Logger.Log("INFO", fmt.Sprintf("[ASR] Task created, TaskId: %d", *taskId))

	// 查询识别结果
	for {
		resultRequest := asr.NewDescribeTaskStatusRequest()
		resultRequest.TaskId = taskId

		resultResponse, err := s.client.DescribeTaskStatus(resultRequest)
		if err != nil {
			middleware.Logger.Log("ERROR", fmt.Sprintf("[ASR] Failed to get task status: %v", err))
			return "", err
		}

		if *resultResponse.Response.Data.Status == 2 {
			middleware.Logger.Log("INFO", fmt.Sprintf("[ASR] Task finished, TaskId: %d", *taskId))
			resultText := *resultResponse.Response.Data.Result

			// 使用正则表达式移除时间戳
			re := regexp.MustCompile(`\[\d{1,2}:\d{1,2}\.\d{3},\d{1,2}:\d{1,2}\.\d{3},\d]\s*`)
			cleanedText := re.ReplaceAllString(resultText, "")

			return cleanedText, nil
		} else if *resultResponse.Response.Data.Status == 3 {
			middleware.Logger.Log("ERROR", fmt.Sprintf("[ASR] Task failed: %v", *resultResponse.Response.Data.ErrorMsg))
			return "", nil
		}

		// 20 秒查询一次
		time.Sleep(20 * time.Second)
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}
