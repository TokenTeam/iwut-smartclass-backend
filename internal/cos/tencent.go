package cos

import (
	"context"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"iwut-smart-timetable-backend/internal/middleware"
	"net/http"
	"net/url"
)

type Service struct {
	client *cos.Client
}

func NewCosService(secretId, secretKey, bucketURL string) (*Service, error) {
	u, _ := url.Parse(bucketURL)
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretId,
			SecretKey: secretKey,
		},
	})
	return &Service{client: client}, nil
}

// UploadFile 上传文件到 COS
func (s *Service) UploadFile(localFilePath, remoteFilePath string) error {
	middleware.Logger.Log("INFO", fmt.Sprintf("[COS] Creating upload task with file: %s", localFilePath))
	_, err := s.client.Object.PutFromFile(context.Background(), remoteFilePath, localFilePath, nil)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[COS] Failed to upload file: %v", err))
		return fmt.Errorf("failed to upload file: %v", err)
	}
	middleware.Logger.Log("INFO", fmt.Sprintf("[COS] File uploaded successfully: %s", remoteFilePath))
	return nil
}

// DownloadFile 从 COS 下载文件
func (s *Service) DownloadFile(remoteFilePath, localFilePath string) error {
	middleware.Logger.Log("INFO", fmt.Sprintf("[COS] Creating download with file: %s", localFilePath))
	_, err := s.client.Object.GetToFile(context.Background(), remoteFilePath, localFilePath, nil)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[COS] Failed to download file: %v", err))
		return fmt.Errorf("failed to download file: %v", err)
	}
	middleware.Logger.Log("INFO", fmt.Sprintf("[COS] File downloaded successfully: %s", localFilePath))
	return nil
}

// DeleteFile 删除 COS 中的文件
func (s *Service) DeleteFile(remoteFilePath string) error {
	middleware.Logger.Log("INFO", fmt.Sprintf("[COS] Creating delete task for file: %s", remoteFilePath))
	_, err := s.client.Object.Delete(context.Background(), remoteFilePath)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("[COS] Failed to delete file: %v", err))
		return fmt.Errorf("failed to delete file: %v", err)
	}
	middleware.Logger.Log("INFO", fmt.Sprintf("[COS] File deleted successfully: %s", remoteFilePath))
	return nil
}
