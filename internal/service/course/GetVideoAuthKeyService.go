package course

import (
	"fmt"
	"io"
	"iwut-smart-timetable-backend/internal/middleware"
	"net/http"
	"regexp"
	"time"
)

type GetVideoAuthKeyService struct {
	Token    string
	CourseId int
	SubId    int
	Logger   *middleware.Log
}

// NewGetVideoAuthKeyService 创建实例
func NewGetVideoAuthKeyService(token string, courseId, subId int, logger *middleware.Log) *GetVideoAuthKeyService {
	return &GetVideoAuthKeyService{
		Token:    token,
		CourseId: courseId,
		SubId:    subId,
		Logger:   logger,
	}
}

// GetVideoAuthKey 获取视频认证密钥
func (s *GetVideoAuthKeyService) GetVideoAuthKey() (string, error) {
	url := fmt.Sprintf("https://yjapi.lgzk.whut.edu.cn/courseapi/v2/course-live/search-live-course-list?all=1&course_id=%d&sub_id=%d&token=%s", s.CourseId, s.SubId, s.Token)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to create request: %v", err))
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to send request: %v", err))
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.Logger.Log("DEBUG", fmt.Sprintf("Received non-200 response code: %d", resp.StatusCode))
		return "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to read response body: %v", err))
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// 使用正则表达式提取 auth_key
	re := regexp.MustCompile(`auth_key=([0-9a-fA-F\-]+)`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		s.Logger.Log("DEBUG", "failed to extract auth_key")
		return "", fmt.Errorf("failed to extract auth_key")
	}

	authKey := string(matches[1])
	s.Logger.Log("DEBUG", fmt.Sprintf("auth_key found, CourseId: %s, SubId: %s, AuthKey: %s", s.CourseId, s.SubId, authKey))
	return authKey, nil
}
