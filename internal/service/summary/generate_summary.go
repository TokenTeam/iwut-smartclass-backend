package summary

import (
	"context"
	"fmt"
	"iwut-smartclass-backend/internal/middleware"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	Database *gorm.DB
}

func NewGenerateSummaryService(db *gorm.DB) *Service {
	return &Service{Database: db}
}

// WriteStatus 将状态写入数据库
func (s *Service) WriteStatus(subId int, status string) error {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.Database.WithContext(ctx).Exec(
		`UPDATE course SET summary_status = ? WHERE sub_id = ?`,
		status, subId,
	).Error

	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database: %v", err))
		return err
	}
	return nil
}
