package summary

import (
	"context"
	"fmt"
	"iwut-smartclass-backend/internal/middleware"
	"time"

	"gorm.io/gorm"
)

type LlmDBService struct {
	Service
}

func NewLlmDBService(db *gorm.DB) *LlmDBService {
	return &LlmDBService{Service{Database: db}}
}

func (s *LlmDBService) SaveSummary(subId int, user, summary string, model string, token uint32) error {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.Database.WithContext(ctx).Exec(
		`UPDATE course SET summary_data = ?, model = ?, token = ?, summary_status = 'finished', summary_user = ? WHERE sub_id = ?`,
		summary, model, token, user, subId,
	).Error

	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database, subId: %d: %v", subId, err))
		return err
	}
	return nil
}

func (s *LlmDBService) InitNewSummary(subId int, user string) (string, error) {
	timeStamp := time.Now().Format(time.RFC3339)

	// 添加超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.Database.WithContext(ctx).Exec(
		`INSERT INTO summary (user, sub_id, create_at, summary, model, token) VALUES (?, ?, ?, ?, ?, ?)`,
		user, subId, timeStamp, "", "", 0,
	).Error

	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to insert user summary into database, subId: %d: %v", subId, err))
		return timeStamp, err
	}
	return timeStamp, nil
}

func (s *LlmDBService) SaveUserSummary(subId int, user, creatAt, summary string, model string, token uint32) error {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.Database.WithContext(ctx).Exec(
		`UPDATE summary SET summary = ?, model = ?, token = ? WHERE sub_id = ? AND user = ? AND create_at = ?`,
		summary, model, token, subId, user, creatAt,
	).Error

	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to insert user summary into database, subId: %d: %v", subId, err))
		return err
	}
	return nil
}
