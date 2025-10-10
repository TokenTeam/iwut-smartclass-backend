package summary

import (
	"database/sql"
	"fmt"
	"iwut-smartclass-backend/internal/middleware"
	"time"
)

type LlmDBService struct {
	Service
}

func NewLlmDBService(db *sql.DB) *LlmDBService {
	return &LlmDBService{Service{Database: db}}
}

func (s *LlmDBService) SaveSummary(subId int, user, summary string, model string, token uint32) error {
	query := `UPDATE course SET summary_data = ?, model = ?, token = ?, summary_status = 'finished', summary_user = ? WHERE sub_id = ?`
	_, err := s.Database.Exec(query, summary, model, token, user, subId)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database, subId: %d: %v", subId, err))
		return err
	}
	return nil
}

func (s *LlmDBService) InitNewSummary(subId int, user string) (string, error) {
	query := `INSERT INTO summary (user, sub_id, create_at, summary, model, token) VALUES (?, ?, ?, ?, ?, ?)`
	timeStamp := time.Now().Format(time.RFC3339)
	_, err := s.Database.Exec(query, user, subId, timeStamp, "", "", 0)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to insert user summary into database, subId: %d: %v", subId, err))
		return timeStamp, err
	}
	return timeStamp, nil
}

func (s *LlmDBService) SaveUserSummary(subId int, user, creatAt, summary string, model string, token uint32) error {
	query := `UPDATE summary SET summary = ?, model = ?, token = ? WHERE sub_id = ? AND user = ? AND create_at = ?`
	_, err := s.Database.Exec(query, summary, model, token, subId, user, creatAt)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to insert user summary into database, subId: %d: %v", subId, err))
		return err
	}
	return nil
}
