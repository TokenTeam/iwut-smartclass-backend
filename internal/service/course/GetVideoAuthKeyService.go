package course

import (
	"crypto/md5"
	"fmt"
	"io"
	"iwut-smart-timetable-backend/internal/middleware"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

type GetVideoAuthKeyService struct {
	Token    string
	CourseId int
	SubId    int
	Video    string
	Logger   *middleware.Log
}

// NewGetVideoAuthKeyService 创建实例
func NewGetVideoAuthKeyService(token string, courseId, subId int, video string, logger *middleware.Log) *GetVideoAuthKeyService {
	return &GetVideoAuthKeyService{
		Token:    token,
		CourseId: courseId,
		SubId:    subId,
		Video:    video,
		Logger:   logger,
	}
}

// GetVideoAuthKey 获取视频认证参数
func (s *GetVideoAuthKeyService) GetVideoAuthKey() (string, error) {
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
	s.Logger.Log("DEBUG", fmt.Sprintf("auth_key found, CourseId: %s, SubId: %s, AuthKey: %s", s.CourseId, s.SubId, authKey))

	// 处理视频链接
	parsedURL, err := url.Parse(s.Video)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to parse video URL: %v", err))
		return "", fmt.Errorf("failed to parse video URL: %v", err)
	}

	// 通过 md5 计算 k
	d := parsedURL.Path
	r := strconv.Itoa(s.SubId)
	o := strconv.Itoa(223)
	h := "13110081151"
	f := strconv.FormatInt(time.Now().Unix(), 10)
	md5Input := d + r + o + h + f
	md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(md5Input)))

	key := "auth_key=" + authKey + "&k=" + r + "-" + f + md5Hash

	return key, nil
}
