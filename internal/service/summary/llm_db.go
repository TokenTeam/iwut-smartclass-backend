package summary

import (
	"database/sql"
	"fmt"
	"iwut-smartclass-backend/internal/middleware"
)

type LlmDBService struct {
	Service
}

func NewLlmDBService(db *sql.DB) *LlmDBService {
	return &LlmDBService{Service{Database: db}}
}

// SaveSummary 将摘要内容写入数据库
func (s *LlmDBService) SaveSummary(subId int, user, asrText string) (string, error) {
	query := `UPDATE course SET summary_data = ?, summary_status = 'finished', summary_user = ? WHERE sub_id = ?`
	_, err := s.Database.Exec(query, asrText, user, subId)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database, subId: %d: %v", subId, err))
		return "", err
	}
	return asrText, nil
}
