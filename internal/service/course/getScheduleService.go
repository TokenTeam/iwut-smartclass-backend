package course

import (
	"encoding/json"
	"fmt"
	"io"
	"iwut-smart-timetable-backend/internal/middleware"
	"net/http"
	"time"
)

type ScheduleService struct {
	Token      string
	Date       string
	CourseName string
	Logger     *middleware.Log
}

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

// NewScheduleService 创建实例
func NewScheduleService(token, date, courseName string, logger *middleware.Log) *ScheduleService {
	return &ScheduleService{
		Token:      token,
		Date:       date,
		CourseName: courseName,
		Logger:     logger,
	}
}

// GetSchedules 获取当天课程信息
func (s *ScheduleService) GetSchedules() (*ScheduleResponse, error) {
	url := fmt.Sprintf("https://yjapi.lgzk.whut.edu.cn/courseapi/v2/schedule/get-week-schedules?start_at=%s&end_at=%s&token=%s", s.Date, s.Date, s.Token)
	s.Logger.Log("DEBUG", fmt.Sprintf("Sending GET request to URL: %s", url))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to create request: %v", err))
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to send request: %v", err))
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.Logger.Log("DEBUG", fmt.Sprintf("Received non-200 response code: %d", resp.StatusCode))
		return nil, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to read response body: %v", err))
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var scheduleResponse ScheduleResponse
	err = json.Unmarshal(body, &scheduleResponse)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to unmarshal response body: %v", err))
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	var filteredCourses []struct {
		ID          string `json:"id"`
		CourseID    string `json:"course_id"`
		CourseTitle string `json:"course_title"`
	}
	for _, day := range scheduleResponse.Result.List {
		for _, course := range day.Course {
			if course.CourseTitle == s.CourseName {
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
		s.Logger.Log("DEBUG", fmt.Sprintf("%s not found", s.CourseName))
		return nil, fmt.Errorf("%s not found", s.CourseName)
	}

	s.Logger.Log("DEBUG", fmt.Sprintf("%s found, CourseId: %s, SubId: %s", s.CourseName, filteredCourses[0].CourseID, filteredCourses[0].ID))
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
