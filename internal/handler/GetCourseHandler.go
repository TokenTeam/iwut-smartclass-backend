package handler

import (
	"encoding/json"
	"fmt"
	"iwut-smartclass-backend/internal/database"
	"iwut-smartclass-backend/internal/middleware"
	"iwut-smartclass-backend/internal/service/course"
	"iwut-smartclass-backend/internal/util"
	"net/http"
	"strconv"
)

// GetCourse 获取课程数据
func GetCourse(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		CourseName string `json:"course_name"`
		Date       string `json:"date"`
		Token      string `json:"token"`
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
	getScheduleService := course.NewGetScheduleService(requestData.Token, requestData.Date, requestData.CourseName, middleware.Logger)

	// 获取当天课程
	scheduleData, err := getScheduleService.GetSchedules()
	if err != nil {
		util.WriteResponse(w, http.StatusInternalServerError, nil)
		return
	}

	// 得到 subId 和 courseId
	var subId, courseId int
	if len(scheduleData.Result.List) > 0 && len(scheduleData.Result.List[0].Course) > 0 {
		subId, _ = strconv.Atoi(scheduleData.Result.List[0].Course[0].ID)
		courseId, _ = strconv.Atoi(scheduleData.Result.List[0].Course[0].CourseID)
	}

	// 创建实例
	courseDBService := course.NewCourseDbService(db)
	getLiveCourseService := course.NewGetLiveCourseService(requestData.Token, middleware.Logger)

	// 尝试从数据库中获取课程数据
	courseData, err := courseDBService.GetCourseDataFromDb(subId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// 如果数据库查询为空，则获取课程回放数据
			liveCourseData, err := getLiveCourseService.SearchLiveCourse(subId, courseId)
			if err != nil {
				util.WriteResponse(w, http.StatusInternalServerError, nil)
				return
			}

			// 将课程数据写入数据库
			courseData = course.Course{
				SubId:    subId,
				CourseId: courseId,
				Name:     liveCourseData["name"].(string),
				Teacher:  liveCourseData["teacher"].(string),
				Location: liveCourseData["location"].(string),
				Date:     liveCourseData["date"].(string),
				Time:     liveCourseData["time"].(string),
				Video:    liveCourseData["video"].(string),
				Summary:  map[string]string{"status": "", "data": ""},
			}

			err = courseDBService.SaveCourseDataToDb(courseData)
			if err != nil {
				util.WriteResponse(w, http.StatusInternalServerError, nil)
				return
			}
		} else {
			util.WriteResponse(w, http.StatusInternalServerError, nil)
			return
		}
	} else if courseData.Video == "" {
		// 如果视频为空，尝试再次获取
		liveCourseData, err := getLiveCourseService.SearchLiveCourse(subId, courseId)
		if err != nil {
			util.WriteResponse(w, http.StatusInternalServerError, nil)
			return
		}

		// 更新视频链接
		videoURL := liveCourseData["video"].(string)
		err = courseDBService.SaveVideo(subId, videoURL)
		if err != nil {
			util.WriteResponse(w, http.StatusInternalServerError, nil)
			return
		}
		courseData.Video = videoURL
	}

	// 将视频密钥拼接在视频链接后
	if courseData.Video != "" {
		// 创建实例
		videoAuthService := course.NewVideoAuthService(requestData.Token, courseId, courseData.Video, middleware.Logger)

		// 获取视频密钥
		videoAuth, err := videoAuthService.VideoAuth()
		if err != nil {
			util.WriteResponse(w, http.StatusInternalServerError, nil)
			return
		}
		courseData.Video = fmt.Sprintf("%s?%s", courseData.Video, videoAuth)
	}

	middleware.Logger.Log("INFO", fmt.Sprintf("GetCourse: CourseName=%s, CourseId=%d, SubId=%d", requestData.CourseName, courseId, subId))
	util.WriteResponse(w, http.StatusOK, courseData)
}
