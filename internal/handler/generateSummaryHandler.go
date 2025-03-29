package handler

import (
	"encoding/json"
	"fmt"
	"iwut-smart-timetable-backend/internal/asr"
	"iwut-smart-timetable-backend/internal/config"
	"iwut-smart-timetable-backend/internal/cos"
	"iwut-smart-timetable-backend/internal/database"
	"iwut-smart-timetable-backend/internal/middleware"
	"iwut-smart-timetable-backend/internal/service/course"
	"iwut-smart-timetable-backend/internal/service/summary"
	"iwut-smart-timetable-backend/internal/util"
	"net/http"
	"os"
	"path/filepath"
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

		// (异步) 将视频转换为音频并调用ASR识别
		go func() {
			// 查找 audioID 是否存在
			audioID, err := convertService.GetAudioId(subId)
			if audioID == "" {
				audioID, err = convertService.Convert(subId, courseData.Video)
				if err != nil {
					// 撤销生成状态
					_ = generateSummaryService.WriteStatus(subId, "")
					middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to convert video to audio: %s", err))
					return
				}
			}

			audioFileName := audioID + ".aac"
			audioFilePath := filepath.Join("data", "audio", audioFileName)

			// 创建 COS 任务实例
			cosService, err := cos.NewCosService(cfg.TencentSecretId, cfg.TencentSecretKey, cfg.BucketUrl)
			if err != nil {
				// 撤销生成状态
				_ = generateSummaryService.WriteStatus(subId, "")
				middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to create Tencent COS service: %s", err))
				return
			}
			err = cosService.UploadFile(audioFilePath, audioFileName)
			if err != nil {
				// 撤销生成状态
				_ = generateSummaryService.WriteStatus(subId, "")
				middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to upload file: %s", err))
				return
			}

			bucketFilePath := cfg.BucketUrl + "/" + audioFileName

			// 创建 ASR 任务实例
			tencentASRService, err := asr.NewTencentASRService(cfg.TencentSecretId, cfg.TencentSecretKey)
			if err != nil {
				// 撤销生成状态
				_ = generateSummaryService.WriteStatus(subId, "")
				middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to create Tencent ASR service: %s", err))
				return
			}
			asrText, err := tencentASRService.Recognize(bucketFilePath)
			if err != nil {
				// 撤销生成状态
				_ = generateSummaryService.WriteStatus(subId, "")
				return
			}

			// 将 ASR 结果写入数据库
			saveAsrService := summary.NewAsrDbService(db)
			_, err = saveAsrService.SaveAsr(subId, asrText)

			// 释放文件
			err = cosService.DeleteFile(audioFileName)
			if err != nil {
				middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to delete file: %s", err))
				return
			}
			err = os.Remove(audioFilePath)
			if err != nil {
				middleware.Logger.Log("ERROR", fmt.Sprintf("Failed to delete file: %s", err))
			}
		}()

		return
	}

	util.WriteResponse(w, http.StatusNotFound, "No video found")
}
