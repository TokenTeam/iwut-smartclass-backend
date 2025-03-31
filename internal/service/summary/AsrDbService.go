package summary

import (
	"database/sql"
	"errors"
	"fmt"
	"iwut-smart-timetable-backend/internal/middleware"
)

type AsrDbService struct {
	Service
}

// NewAsrDbService 创建实例
func NewAsrDbService(db *sql.DB) *AsrDbService {
	return &AsrDbService{Service{Database: db}}
}

// GetAsr 从数据库读取 ASR 内容
func (s *AsrDbService) GetAsr(subId int) (string, error) {
	query := `SELECT asr FROM course WHERE sub_id = ?`
	row := s.Database.QueryRow(query, subId)
	var asr sql.NullString
	err := row.Scan(&asr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			middleware.Logger.Log("DEBUG", fmt.Sprintf("[DB] Could not find course data in database, subId: %d: %v", subId, err))
			return "", fmt.Errorf("sql: no rows in result set")
		}
		return "", err
	}
	if asr.Valid {
		return asr.String, nil
	}
	return "", nil
}

// SaveAsr 将 ASR 内容写入数据库
func (s *AsrDbService) SaveAsr(subId int, asrText string) (string, error) {
	query := `UPDATE course SET asr = ? WHERE sub_id = ?`
	_, err := s.Database.Exec(query, asrText, subId)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database, subId: %d: %v", subId, err))
		return "", err
	}
	return asrText, nil
}
