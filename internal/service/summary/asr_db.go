package summary

import (
	"context"
	"fmt"
	"iwut-smartclass-backend/internal/middleware"
	"time"

	"gorm.io/gorm"
)

type AsrDBService struct {
	Service
}

func NewAsrDBService(db *gorm.DB) *AsrDBService {
	return &AsrDBService{Service{Database: db}}
}

// GetAsr 从数据库读取 ASR 内容
func (s *AsrDBService) GetAsr(subId int) (string, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result struct {
		Asr *string
	}

	dbResult := s.Database.WithContext(ctx).Raw(
		`SELECT asr FROM course WHERE sub_id = ?`,
		subId,
	).Scan(&result)

	if dbResult.Error != nil {
		return "", dbResult.Error
	}

	// 检查是否找到记录
	if dbResult.RowsAffected == 0 {
		middleware.Logger.Log("DEBUG", fmt.Sprintf("[DB] Could not find course data in database, subId: %d", subId))
		return "", fmt.Errorf("sql: no rows in result set")
	}

	if result.Asr != nil {
		return *result.Asr, nil
	}
	return "", nil
}

// SaveAsr 将 ASR 内容写入数据库
func (s *AsrDBService) SaveAsr(subId int, asrText string) (string, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.Database.WithContext(ctx).Exec(
		`UPDATE course SET asr = ? WHERE sub_id = ?`,
		asrText, subId,
	).Error

	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database, subId: %d: %v", subId, err))
		return "", err
	}
	return asrText, nil
}
