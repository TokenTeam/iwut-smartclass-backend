package summary

import (
	"encoding/json"
	"iwut-smartclass-backend/internal/config"
	"iwut-smartclass-backend/internal/database"
	"iwut-smartclass-backend/internal/middleware"
)

func init() {
	middleware.RegisterGlobalLoader("summary", func(data []byte, cfg *config.Config) (middleware.Job, error) {
		var job Job
		if err := json.Unmarshal(data, &job); err != nil {
			return nil, err
		}

		// 重新注入依赖
		db := database.GetDB()
		job.SummarySvc = NewGenerateSummaryService(db)
		job.ConvertSvc = NewConvertService(db)
		job.AsrSvc = NewAsrDBService(db)
		job.SummaryDbSvc = NewLlmDBService(db)
		job.Config = cfg

		return &job, nil
	})
}
