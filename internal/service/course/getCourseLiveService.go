package course

import (
	"encoding/json"
	"fmt"
	"io"
	"iwut-smart-timetable-backend/internal/middleware"
	"net/http"
	"strconv"
	"time"
)

type LiveCourseService struct {
	Token  string
	Logger *middleware.Log
}

// NewGetLiveCourseService 创建实例
func NewGetLiveCourseService(token string, logger *middleware.Log) *LiveCourseService {
	return &LiveCourseService{
		Token:  token,
		Logger: logger,
	}
}

// SearchLiveCourse 查询课程回放
func (s *LiveCourseService) SearchLiveCourse(subId, courseId int) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://yjapi.lgzk.whut.edu.cn/courseapi/v2/course-live/search-live-course-list?all=1&course_id=%d&sub_id=%d", courseId, subId)
	s.Logger.Log("DEBUG", fmt.Sprintf("Sending GET request to URL: %s", url))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to create request: %v", err))
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.Token)

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

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to unmarshal response body: %v", err))
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	if result["code"].(float64) != 0 {
		s.Logger.Log("DEBUG", fmt.Sprintf("API error: %s", result["msg"].(string)))
		return nil, fmt.Errorf("API error: %s", result["msg"].(string))
	}

	list := result["list"].([]interface{})
	if len(list) == 0 {
		s.Logger.Log("DEBUG", "No live courses found")
		return nil, fmt.Errorf("no live courses found")
	}

	course := list[0].(map[string]interface{})
	video := ""
	if videoList, ok := course["video_list"].([]interface{}); ok && len(videoList) > 0 {
		video = videoList[0].(map[string]interface{})["preview_url"].(string)
	}

	courseBegin, err := strconv.ParseInt(course["course_begin"].(string), 10, 64)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to parse course_begin: %v", err))
		return nil, fmt.Errorf("failed to parse course_begin: %v", err)
	}
	courseOver, err := strconv.ParseInt(course["course_over"].(string), 10, 64)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to parse course_over: %v", err))
		return nil, fmt.Errorf("failed to parse course_over: %v", err)
	}

	courseTime := fmt.Sprintf("%s-%s", time.Unix(courseBegin, 0).Format("15:04"), time.Unix(courseOver, 0).Format("15:04"))

	courseData := map[string]interface{}{
		"course_id": course["id"],
		"sub_id":    course["sub_id"],
		"name":      course["title"],
		"teacher":   course["realname"],
		"location":  course["room_name"],
		"date":      course["sub_title"],
		"time":      courseTime,
		"video":     video,
	}

	s.Logger.Log("DEBUG", fmt.Sprintf("%s found, CourseId: %s, SubId: %s", courseData["name"], courseData["course_id"], courseData["sub_id"]))
	return courseData, nil
}
