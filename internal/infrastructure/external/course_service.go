package external

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"iwut-smartclass-backend/internal/domain/errors"
	"iwut-smartclass-backend/internal/infrastructure/config"
	"iwut-smartclass-backend/internal/infrastructure/logger"
)

// CourseExternalService 课程外部服务接口
type CourseExternalService interface {
	GetSchedule(token, date, courseName string) (*ScheduleResponse, error)
	SearchLiveCourse(token string, subID, courseID int) (map[string]interface{}, error)
	GetVideoAuthKey(token string, courseID, subID int) (string, error)
}

// ScheduleService 课程表服务
type ScheduleService struct {
	cfg    *config.Config
	logger logger.Logger
}

// NewScheduleService 创建课程表服务
func NewScheduleService(cfg *config.Config, logger logger.Logger) *ScheduleService {
	return &ScheduleService{
		cfg:    cfg,
		logger: logger,
	}
}

// ScheduleResponse 课程表响应
type ScheduleResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		List []struct {
			Day    string `json:"day"`
			Course []struct {
				ID          string `json:"id"`
				CourseID    string `json:"course_id"`
				CourseTitle string `json:"course_title"`
			} `json:"course"`
		} `json:"list"`
	} `json:"result"`
}

// GetSchedule 获取课程表
func (s *ScheduleService) GetSchedule(token, date, courseName string) (*ScheduleResponse, error) {
	url := fmt.Sprintf("%s?start_at=%s&end_at=%s&token=%s", s.cfg.GetWeekSchedules, date, date, token)
	s.logger.Debug("sending request to get schedule", logger.String("url", url))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.logger.Error("failed to create request", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("schedule service", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("failed to send request", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("schedule service", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("received non-200 response", logger.String("status", fmt.Sprintf("%d", resp.StatusCode)))
		return nil, errors.NewExternalError("schedule service", fmt.Errorf("status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("failed to read response", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("schedule service", err)
	}

	var scheduleResponse ScheduleResponse
	if err := json.Unmarshal(body, &scheduleResponse); err != nil {
		s.logger.Error("failed to unmarshal response", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("schedule service", err)
	}

	// 过滤课程
	var filteredCourses []struct {
		ID          string `json:"id"`
		CourseID    string `json:"course_id"`
		CourseTitle string `json:"course_title"`
	}
	for _, day := range scheduleResponse.Result.List {
		for _, course := range day.Course {
			if course.CourseTitle == courseName {
				filteredCourses = append(filteredCourses, struct {
					ID          string `json:"id"`
					CourseID    string `json:"course_id"`
					CourseTitle string `json:"course_title"`
				}{
					ID:          course.ID,
					CourseID:    course.CourseID,
					CourseTitle: course.CourseTitle,
				})
			}
		}
	}

	if len(filteredCourses) != 1 {
		s.logger.Error("course not found", logger.String("course_name", courseName))
		return nil, errors.NewNotFoundError("course")
	}

	return &ScheduleResponse{
		Success: scheduleResponse.Success,
		Result: struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			List []struct {
				Day    string `json:"day"`
				Course []struct {
					ID          string `json:"id"`
					CourseID    string `json:"course_id"`
					CourseTitle string `json:"course_title"`
				} `json:"course"`
			} `json:"list"`
		}{
			Code: scheduleResponse.Result.Code,
			Msg:  scheduleResponse.Result.Msg,
			List: []struct {
				Day    string `json:"day"`
				Course []struct {
					ID          string `json:"id"`
					CourseID    string `json:"course_id"`
					CourseTitle string `json:"course_title"`
				} `json:"course"`
			}{
				{
					Day:    scheduleResponse.Result.List[0].Day,
					Course: filteredCourses,
				},
			},
		},
	}, nil
}

// LiveCourseService 直播课程服务
type LiveCourseService struct {
	cfg    *config.Config
	logger logger.Logger
}

// NewLiveCourseService 创建直播课程服务
func NewLiveCourseService(cfg *config.Config, logger logger.Logger) *LiveCourseService {
	return &LiveCourseService{
		cfg:    cfg,
		logger: logger,
	}
}

// SearchLiveCourse 搜索直播课程
func (s *LiveCourseService) SearchLiveCourse(token string, subID, courseID int) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s?all=1&course_id=%d&sub_id=%d", s.cfg.SearchLiveCourseList, courseID, subID)
	s.logger.Debug("sending request to search live course", logger.String("url", url))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.logger.Error("failed to create request", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("live course service", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("failed to send request", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("live course service", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("received non-200 response", logger.String("status", fmt.Sprintf("%d", resp.StatusCode)))
		return nil, errors.NewExternalError("live course service", fmt.Errorf("status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("failed to read response", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("live course service", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		s.logger.Error("failed to unmarshal response", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("live course service", err)
	}

	// 检查code字段
	codeVal, codeExists := result["code"]
	if !codeExists {
		s.logger.Error("missing code field in response")
		return nil, errors.NewExternalError("live course service", fmt.Errorf("missing code field in response"))
	}
	code, ok := codeVal.(float64)
	if !ok {
		s.logger.Error("invalid code type in response")
		return nil, errors.NewExternalError("live course service", fmt.Errorf("invalid code type in response"))
	}
	if code != 0 {
		msg := ""
		if msgVal, ok := result["msg"].(string); ok {
			msg = msgVal
		}
		s.logger.Error("api error", logger.String("msg", msg))
		return nil, errors.NewExternalError("live course service", fmt.Errorf("api error: %s", msg))
	}

	// 检查list字段
	listVal, listExists := result["list"]
	if !listExists {
		s.logger.Error("missing list field in response")
		return nil, errors.NewExternalError("live course service", fmt.Errorf("missing list field in response"))
	}
	list, ok := listVal.([]interface{})
	if !ok {
		s.logger.Error("invalid list type in response")
		return nil, errors.NewExternalError("live course service", fmt.Errorf("invalid list type in response"))
	}
	if len(list) == 0 {
		s.logger.Error("no live courses found")
		return nil, errors.NewNotFoundError("live course")
	}

	// 检查courseData
	courseData, ok := list[0].(map[string]interface{})
	if !ok {
		s.logger.Error("invalid course data type in response")
		return nil, errors.NewExternalError("live course service", fmt.Errorf("invalid course data type in response"))
	}

	// 安全提取video
	video := ""
	if videoListVal, ok := courseData["video_list"]; ok {
		if videoList, ok := videoListVal.([]interface{}); ok && len(videoList) > 0 {
			if firstVideo, ok := videoList[0].(map[string]interface{}); ok {
				if previewURL, ok := firstVideo["preview_url"].(string); ok {
					video = previewURL
				}
			}
		}
	}

	// 安全提取course_begin
	courseBeginStr, ok := courseData["course_begin"].(string)
	if !ok {
		s.logger.Error("missing or invalid course_begin field")
		return nil, errors.NewExternalError("live course service", fmt.Errorf("missing or invalid course_begin field"))
	}
	courseBegin, err := strconv.ParseInt(courseBeginStr, 10, 64)
	if err != nil {
		s.logger.Error("failed to parse course_begin", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("live course service", err)
	}

	// 安全提取course_over
	courseOverStr, ok := courseData["course_over"].(string)
	if !ok {
		s.logger.Error("missing or invalid course_over field")
		return nil, errors.NewExternalError("live course service", fmt.Errorf("missing or invalid course_over field"))
	}
	courseOver, err := strconv.ParseInt(courseOverStr, 10, 64)
	if err != nil {
		s.logger.Error("failed to parse course_over", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("live course service", err)
	}

	courseTime := fmt.Sprintf("%s-%s", time.Unix(courseBegin, 0).Format("15:04"), time.Unix(courseOver, 0).Format("15:04"))

	// 安全提取其他字段
	var resultCourseID, resultSubID int
	if idVal, ok := courseData["id"]; ok {
		if idFloat, ok := idVal.(float64); ok {
			resultCourseID = int(idFloat)
		}
	}
	if subIDVal, ok := courseData["sub_id"]; ok {
		if subIDFloat, ok := subIDVal.(float64); ok {
			resultSubID = int(subIDFloat)
		}
	}
	name, _ := courseData["title"].(string)
	teacher, _ := courseData["realname"].(string)
	location, _ := courseData["room_name"].(string)
	date, _ := courseData["sub_title"].(string)

	return map[string]interface{}{
		"course_id": resultCourseID,
		"sub_id":    resultSubID,
		"name":      name,
		"teacher":   teacher,
		"location":  location,
		"date":      date,
		"time":      courseTime,
		"video":     video,
	}, nil
}

// VideoAuthService 视频认证服务
type VideoAuthService struct {
	cfg    *config.Config
	logger logger.Logger
}

// NewVideoAuthService 创建视频认证服务
func NewVideoAuthService(cfg *config.Config, logger logger.Logger) *VideoAuthService {
	return &VideoAuthService{
		cfg:    cfg,
		logger: logger,
	}
}

// GetVideoAuthKey 获取视频认证密钥
func (s *VideoAuthService) GetVideoAuthKey(token string, courseID, subID int) (string, error) {
	url := fmt.Sprintf("%s?all=1&course_id=%d&sub_id=%d&token=%s", s.cfg.SearchLiveCourseList, courseID, subID, token)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.logger.Error("failed to create request", logger.String("error", err.Error()))
		return "", errors.NewExternalError("video auth service", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("failed to send request", logger.String("error", err.Error()))
		return "", errors.NewExternalError("video auth service", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("received non-200 response", logger.String("status", fmt.Sprintf("%d", resp.StatusCode)))
		return "", errors.NewExternalError("video auth service", fmt.Errorf("status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("failed to read response", logger.String("error", err.Error()))
		return "", errors.NewExternalError("video auth service", err)
	}

	// 使用正则表达式提取 auth_key
	re := regexp.MustCompile(`auth_key=([0-9a-fA-F\-]+)`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		s.logger.Error("failed to extract auth_key")
		return "", errors.NewExternalError("video auth service", fmt.Errorf("failed to extract auth_key"))
	}

	return string(matches[1]), nil
}
