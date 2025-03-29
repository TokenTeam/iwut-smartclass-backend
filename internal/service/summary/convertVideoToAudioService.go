package summary

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"iwut-smart-timetable-backend/internal/middleware"
	"iwut-smart-timetable-backend/internal/util"
	"os"
	"path/filepath"
)

type Service struct {
	Database *sql.DB
}

type ConvertVideoToAudioService struct {
	Service *Service
}

// NewConvertVideoToAudioService 创建实例
func NewConvertVideoToAudioService(db *sql.DB) *Service {
	return &Service{Database: db}
}

// WriteStatus 将状态写入数据库
func (s *Service) WriteStatus(subId int, status string) error {
	query := `UPDATE course SET summary_status = ? WHERE sub_id = ?`
	_, err := s.Database.Exec(query, status, subId)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database: %v", err))
		return err
	}
	return nil
}

// Convert 将视频转换为音频文件
func (s *Service) Convert(subId int, videoFile string) (string, error) {
	// 写入生成状态
	_ = s.WriteStatus(subId, "generating")

	// 生成音频文件名
	audioID := uuid.New()
	audioFileName := audioID.String() + ".aac"
	audioFilePath := filepath.Join("data", "audio", audioFileName)

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(audioFilePath), 0755); err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to create directory: %v", err))
		// 撤销生成状态
		_ = s.WriteStatus(subId, "")
		return "", err
	}

	// 将视频转换为音频
	err := util.ConvertVideoToAudio(videoFile, audioFilePath)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to convert video to audio: %v", err))
		// 撤销生成状态
		_ = s.WriteStatus(subId, "")
		return "", err
	}

	// 将 audioID 写入数据库
	query := `UPDATE course SET audio_id = ? WHERE sub_id = ?`
	_, err = s.Database.Exec(query, audioID.String(), subId)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database: %v", err))
		return "", err
	}

	return audioFilePath, nil
}
