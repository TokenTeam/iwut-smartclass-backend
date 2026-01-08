package summary

import (
	"context"
	"fmt"
	"iwut-smartclass-backend/internal/middleware"
	"iwut-smartclass-backend/internal/util"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConvertService struct {
	Service
}

func NewConvertService(db *gorm.DB) *ConvertService {
	return &ConvertService{Service{Database: db}}
}

// Convert 将视频转换为音频文件
func (s *ConvertService) Convert(ctx context.Context, subId int, videoFile string) (string, error) {
	// 生成音频文件名
	audioID := uuid.New().String()
	audioFileName := audioID + ".aac"
	audioFilePath := filepath.Join("temp", "audio", audioFileName)

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(audioFilePath), 0755); err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to create directory: %v", err))
		// 撤销生成状态
		_ = s.WriteStatus(subId, "")
		return "", err
	}

	// 将视频转换为音频
	err := util.ConvertVideoToAudio(ctx, videoFile, audioFilePath)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to convert video to audio: %v", err))
		// 撤销生成状态
		_ = s.WriteStatus(subId, "")
		return "", err
	}

	// 将 audioID 写入数据库
	_ = s.SaveAudioId(ctx, subId, audioID)

	return audioID, nil
}

// GetAudioId 从数据库中获取 audioID
func (s *ConvertService) GetAudioId(ctx context.Context, subId int) (string, error) {
	// 如果上下文没有超时，添加默认超时控制
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	var result struct {
		AudioID *string
	}

	dbResult := s.Database.WithContext(ctx).Raw(
		`SELECT audio_id FROM course WHERE sub_id = ?`,
		subId,
	).Scan(&result)

	if dbResult.Error != nil {
		return "", dbResult.Error
	}

	// 检查是否找到记录
	if dbResult.RowsAffected == 0 {
		middleware.Logger.Log("DEBUG", fmt.Sprintf("Could not find course data in database, subId: %d", subId))
		return "", fmt.Errorf("sql: no rows in result set")
	}

	if result.AudioID != nil {
		return *result.AudioID, nil
	}
	return "", nil
}

// SaveAudioId 将 audioID 写入数据库
func (s *ConvertService) SaveAudioId(ctx context.Context, subId int, audioID string) error {
	// 如果上下文没有超时，添加默认超时控制
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	err := s.Database.WithContext(ctx).Exec(
		`UPDATE course SET audio_id = ? WHERE sub_id = ?`,
		audioID, subId,
	).Error

	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database, subId: %d: %v", subId, err))
		return err
	}
	return nil
}
