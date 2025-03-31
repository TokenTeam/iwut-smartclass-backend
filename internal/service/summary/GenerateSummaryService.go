package summary

import (
	"database/sql"
	"fmt"
	"iwut-smart-timetable-backend/internal/middleware"
)

type Service struct {
	Database *sql.DB
}

// NewGenerateSummaryService 创建实例
func NewGenerateSummaryService(db *sql.DB) *Service {
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
