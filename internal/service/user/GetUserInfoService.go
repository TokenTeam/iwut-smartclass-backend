package user

import (
	"encoding/json"
	"fmt"
	"io"
	"iwut-smartclass-backend/internal/middleware"
	"net/http"
	"time"
)

type GetUserInfoService struct {
	Token  string
	Logger *middleware.Log
}

type UserInfo struct {
	Account  string
	Id       int
	Phone    string
	TenantId int
}

// NewGetUserInfoService 创建实例
func NewGetUserInfoService(token string, logger *middleware.Log) *GetUserInfoService {
	return &GetUserInfoService{
		Token:  token,
		Logger: logger,
	}
}

// GetUserInfo 获取用户信息
func (s *GetUserInfoService) GetUserInfo() (UserInfo, error) {
	userInfoUrl := "https://classroom.lgzk.whut.edu.cn/userapi/v1/infosimple"
	s.Logger.Log("DEBUG", fmt.Sprintf("Sending GET request to URL: %s", userInfoUrl))

	req, err := http.NewRequest("GET", userInfoUrl, nil)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to create request: %v", err))
		return UserInfo{}, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to send request: %v", err))
		return UserInfo{}, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.Logger.Log("DEBUG", fmt.Sprintf("Received non-200 response code: %d", resp.StatusCode))
		return UserInfo{}, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to read response body: %v", err))
		return UserInfo{}, fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Params  struct {
			Account  string `json:"account"`
			Id       int    `json:"id"`
			Phone    string `json:"phone"`
			TenantId int    `json:"tenant_id"`
		} `json:"params"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		s.Logger.Log("DEBUG", fmt.Sprintf("Failed to unmarshal JSON response: %v", err))
		return UserInfo{}, fmt.Errorf("failed to unmarshal JSON response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		s.Logger.Log("DEBUG", fmt.Sprintf("Received non-200 response code: %d", resp.StatusCode))
		return UserInfo{}, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	return UserInfo(response.Params), nil
}
