package handler

import (
	"encoding/json"
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
		SubID int `json:"sub_id"`
	}

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
	getCourseDBService := course.NewGetCourseDBService(db)
	convertService := summary.NewConvertVideoToAudioService(db)

	// 尝试从数据库中获取课程数据
	subID, err := strconv.Atoi(strconv.Itoa(requestData.SubID))
	if err != nil {
		middleware.Logger.Log("DEBUG", "Invalid SubID")
		util.WriteResponse(w, http.StatusBadRequest, "Invalid SubID")
		return
	}
	courseData, err := getCourseDBService.GetCourseDataFromDB(subID)
	if err != nil {
		util.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	// 返回状态
	if courseData.Video != "" {
		util.WriteResponse(w, http.StatusOK,
			map[string]interface{}{
				"sub_id":         requestData.SubID,
				"summary_status": "generating",
			},
		)

		// (异步) 将视频转换为音频
		go func() {
			_, err := convertService.Convert(subID, courseData.Video)
			if err != nil {
				middleware.Logger.Log("ERROR", "Failed to convert video to audio: "+err.Error())
			}
		}()

		return
	}

	util.WriteResponse(w, http.StatusNotFound, "No video found")
}
