package external

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tencentyun/cos-go-sdk-v5"
	"iwut-smartclass-backend/internal/domain/errors"
	"iwut-smartclass-backend/internal/infrastructure/logger"
)

// COSService COS服务
type COSService struct {
	client *cos.Client
	logger logger.Logger
}

// NewCOSService 创建COS服务
func NewCOSService(secretId, secretKey, bucketURL string, logger logger.Logger) (*COSService, error) {
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, errors.NewInternalError("failed to parse bucket URL", err)
	}
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretId,
			SecretKey: secretKey,
		},
	})
	return &COSService{client: client, logger: logger}, nil
}

// UploadFile 上传文件
func (s *COSService) UploadFile(localFilePath, remoteFilePath string) error {
	s.logger.Info("uploading file to COS",
		logger.String("local", localFilePath),
		logger.String("remote", remoteFilePath),
	)
	_, err := s.client.Object.PutFromFile(context.Background(), remoteFilePath, localFilePath, nil)
	if err != nil {
		s.logger.Error("failed to upload file", logger.String("error", err.Error()))
		return errors.NewExternalError("cos", fmt.Errorf("failed to upload file: %w", err))
	}
	s.logger.Info("file uploaded successfully", logger.String("remote", remoteFilePath))
	return nil
}

// DownloadFile 下载文件
func (s *COSService) DownloadFile(remoteFilePath, localFilePath string) error {
	s.logger.Info("downloading file from COS",
		logger.String("remote", remoteFilePath),
		logger.String("local", localFilePath),
	)
	_, err := s.client.Object.GetToFile(context.Background(), remoteFilePath, localFilePath, nil)
	if err != nil {
		s.logger.Error("failed to download file", logger.String("error", err.Error()))
		return errors.NewExternalError("cos", fmt.Errorf("failed to download file: %w", err))
	}
	s.logger.Info("file downloaded successfully", logger.String("local", localFilePath))
	return nil
}

// DeleteFile 删除文件
func (s *COSService) DeleteFile(remoteFilePath string) error {
	s.logger.Info("deleting file from COS", logger.String("remote", remoteFilePath))
	_, err := s.client.Object.Delete(context.Background(), remoteFilePath)
	if err != nil {
		s.logger.Error("failed to delete file", logger.String("error", err.Error()))
		return errors.NewExternalError("cos", fmt.Errorf("failed to delete file: %w", err))
	}
	s.logger.Info("file deleted successfully", logger.String("remote", remoteFilePath))
	return nil
}
