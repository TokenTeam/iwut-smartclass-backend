package summary

import (
	"database/sql"
	"errors"
	"fmt"
	"iwut-smartclass-backend/internal/middleware"
)

type AsrDBService struct {
	Service
}

func NewAsrDBService(db *sql.DB) *AsrDBService {
	return &AsrDBService{Service{Database: db}}
}

// GetAsr 从数据库读取 ASR 内容
func (s *AsrDBService) GetAsr(subId int) (string, error) {
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
func (s *AsrDBService) SaveAsr(subId int, asrText string) (string, error) {
	query := `UPDATE course SET asr = ? WHERE sub_id = ?`
	_, err := s.Database.Exec(query, asrText, subId)
	if err != nil {
		middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to update database, subId: %d: %v", subId, err))
		return "", err
	}
	return asrText, nil
}
