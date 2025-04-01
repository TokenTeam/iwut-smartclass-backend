package handler

import (
	"encoding/json"
	"iwut-smart-timetable-backend/internal/config"
	"iwut-smart-timetable-backend/internal/database"
	"iwut-smart-timetable-backend/internal/middleware"
	"iwut-smart-timetable-backend/internal/service/course"
	"iwut-smart-timetable-backend/internal/service/summary"
	"iwut-smart-timetable-backend/internal/util"
	"net/http"
	"strconv"
)

// GenerateSummary 创建 AI 课程总结
func GenerateSummary(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		SubId int `json:"sub_id"`
	}

	cfg := config.LoadConfig()

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		util.WriteResponse(w, http.StatusBadRequest, nil)
		return
	}

	db := database.GetDB()
	if db == nil {
		middleware.Logger.Log("ERROR", "Database not initialized")
		util.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	// 创建实例
	generateSummaryService := summary.NewGenerateSummaryService(db)
	getCourseDBService := course.NewGetCourseDbService(db)
	convertService := summary.NewConvertVideoToAudioService(db)

	// 尝试从数据库中获取课程数据
	subId, err := strconv.Atoi(strconv.Itoa(requestData.SubId))
	if err != nil {
		middleware.Logger.Log("DEBUG", "Invalid subId")
		util.WriteResponse(w, http.StatusBadRequest, "Invalid subId")
		return
	}
	courseData, err := getCourseDBService.GetCourseDataFromDb(subId)
	if err != nil {
		util.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	// 返回状态
	if courseData.Video != "" {
		util.WriteResponse(w, http.StatusOK,
			map[string]interface{}{
				"sub_id":         requestData.SubId,
				"summary_status": "generating",
			},
		)

		// 写入生成状态
		err := generateSummaryService.WriteStatus(subId, "generating")
		if err != nil {
			util.WriteResponse(w, http.StatusInternalServerError, nil)
			return
		}

		// 创建队列实例
		summaryQueue := middleware.GetQueue("SummaryServiceQueue")
		if summaryQueue == nil {
			middleware.Logger.Log("ERROR", "Summary queue not initialized")
			util.WriteResponse(w, http.StatusInternalServerError, nil)
			return
		}

		// 创建任务
		job := &summary.Job{
			SubID:        subId,
			CourseName:   courseData.Name,
			VideoURL:     courseData.Video,
			SummarySvc:   generateSummaryService,
			ConvertSvc:   convertService,
			AsrSvc:       summary.NewAsrDbService(db),
			SummaryDbSvc: summary.NewLlmDbService(db),
			Config:       cfg,
		}
		summaryQueue.AddJob(job)

		return
	}

	util.WriteResponse(w, http.StatusNotFound, "No video found")
}
