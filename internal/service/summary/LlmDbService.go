package summary

import (
	"database/sql"
	"fmt"
	"iwut-smartclass-backend/internal/middleware"
)

type LlmDbService struct {
	Service
}

// NewLlmDbService 创建实例
func NewLlmDbService(db *sql.DB) *LlmDbService {
	return &LlmDbService{Service{Database: db}}
}

// SaveSummary 将摘要内容写入数据库
func (s *LlmDbService) SaveSummary(subId int, asrText string) (string, error) {
	query := `UPDATE course SET summary_data = ?, summary_status = 'finished' WHERE sub_id = ?`
	_, err := s.Database.Exec(query, asrText, subId)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database, subId: %d: %v", subId, err))
		return "", err
	}
	return asrText, nil
}
