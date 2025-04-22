package course

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"iwut-smart-timetable-backend/internal/middleware"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

type VideoAuthService struct {
	Token    string
	CourseId int
	SubId    int
	Video    string
	Logger   *middleware.Log
}

// NewVideoAuthService 创建实例
func NewVideoAuthService(token string, courseId, subId int, video string, logger *middleware.Log) *VideoAuthService {
	return &VideoAuthService{
		Token:    token,
		CourseId: courseId,
		SubId:    subId,
		Video:    video,
		Logger:   logger,
	}
}

// VideoAuth 获取视频认证参数
func (s *VideoAuthService) VideoAuth() (string, error) {
	// 获取视频认证参数
	authKey, err := s.GetVideoAuthKey()
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to get auth_key: %v", err))
		return "", fmt.Errorf("failed to get auth_key: %v", err)
	}

	id, phone, err := s.GetUserInfo()
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to get phone: %v", err))
		return "", fmt.Errorf("failed to get phone: %v", err)
	}

	// 处理视频链接
	parsedURL, err := url.Parse(s.Video)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to parse video URL: %v", err))
		return "", fmt.Errorf("failed to parse video URL: %v", err)
	}

	// 反转 phone
	runes := []rune(phone)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	// 通过 md5 计算 k
	d := parsedURL.Path
	r := strconv.Itoa(id)
	o := strconv.Itoa(223)
	h := string(runes)
	f := strconv.FormatInt(time.Now().Unix(), 10)
	md5Input := d + r + o + h + f
	md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(md5Input)))
	s.Logger.Log("DEBUG", fmt.Sprintf("MD5 generated, CourseId: %d, SubId: %d, MD5: %s", s.CourseId, s.SubId, md5Hash))

	key := "auth_key=" + authKey + "&t=" + r + "-" + f + "-" + md5Hash

	return key, nil
}

// GetVideoAuthKey 获取视频认证参数
func (s *VideoAuthService) GetVideoAuthKey() (string, error) {
	authKeyUrl := fmt.Sprintf("https://yjapi.lgzk.whut.edu.cn/courseapi/v2/course-live/search-live-course-list?all=1&course_id=%d&sub_id=%d&token=%s", s.CourseId, s.SubId, s.Token)

	req, err := http.NewRequest("GET", authKeyUrl, nil)
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
	s.Logger.Log("DEBUG", fmt.Sprintf("auth_key found, CourseId: %d, SubId: %d, AuthKey: %s", s.CourseId, s.SubId, authKey))

	return authKey, nil
}

// GetUserInfo 获取用户信息
func (s *VideoAuthService) GetUserInfo() (int, string, error) {
	userInfoUrl := "https://classroom.lgzk.whut.edu.cn/userapi/v1/infosimple"
	s.Logger.Log("DEBUG", fmt.Sprintf("Sending GET request to URL: %s", userInfoUrl))

	req, err := http.NewRequest("GET", userInfoUrl, nil)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to create request: %v", err))
		return 0, "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to send request: %v", err))
		return 0, "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.Logger.Log("DEBUG", fmt.Sprintf("Received non-200 response code: %d", resp.StatusCode))
		return 0, "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to read response body: %v", err))
		return 0, "", fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Params  struct {
			Id    int    `json:"id"`
			Phone string `json:"phone"`
		} `json:"params"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to unmarshal JSON response: %v", err))
		return 0, "", fmt.Errorf("failed to unmarshal JSON response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.Logger.Log("DEBUG", fmt.Sprintf("Received non-200 response code: %d", resp.StatusCode))
		return 0, "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	id := response.Params.Id
	s.Logger.Log("DEBUG", fmt.Sprintf("id found, CourseId: %d, SubId: %d, Id: %d", s.CourseId, s.SubId, id))
	phone := response.Params.Phone
	s.Logger.Log("DEBUG", fmt.Sprintf("phone found, CourseId: %d, SubId: %d, Phone: %s", s.CourseId, s.SubId, phone))

	return id, phone, nil
}
