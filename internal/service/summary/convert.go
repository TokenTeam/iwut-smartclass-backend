package summary

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"iwut-smartclass-backend/internal/middleware"
	"iwut-smartclass-backend/internal/util"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type ConvertService struct {
	Service
}

func NewConvertService(db *sql.DB) *ConvertService {
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
	_ = s.SaveAudioId(subId, audioID)

	return audioID, nil
}

// GetAudioId 从数据库中获取 audioID
func (s *ConvertService) GetAudioId(subId int) (string, error) {
	query := `SELECT audio_id FROM course WHERE sub_id = ?`
	row := s.Database.QueryRow(query, subId)
	var audioID sql.NullString
	err := row.Scan(&audioID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			middleware.Logger.Log("DEBUG", fmt.Sprintf("Could not find course data in database, subId: %d: %v", subId, err))
			return "", fmt.Errorf("sql: no rows in result set")
		}
		return "", err
	}
	if audioID.Valid {
		return audioID.String, nil
	}
	return "", nil
}

// SaveAudioId 将 audioID 写入数据库
func (s *ConvertService) SaveAudioId(subId int, audioID string) error {
	query := `UPDATE course SET audio_id = ? WHERE sub_id = ?`
	_, err := s.Database.Exec(query, audioID, subId)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database, subId: %d: %v", subId, err))
		return err
	}
	return nil
}
