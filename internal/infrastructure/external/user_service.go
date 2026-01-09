package external

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"iwut-smartclass-backend/internal/domain/errors"
	"iwut-smartclass-backend/internal/domain/user"
	"iwut-smartclass-backend/internal/infrastructure/config"
	"iwut-smartclass-backend/internal/infrastructure/logger"
)

// UserService 用户外部服务实现
type UserService struct {
	cfg    *config.Config
	logger logger.Logger
}

// NewUserService 创建用户服务
func NewUserService(cfg *config.Config, logger logger.Logger) *UserService {
	return &UserService{
		cfg:    cfg,
		logger: logger,
	}
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(token string) (*user.User, error) {
	url := s.cfg.InfoSimple
	s.logger.Debug("sending request to get user info", logger.String("url", url))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.logger.Error("failed to create request", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("user service", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("failed to send request", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("user service", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("received non-200 response", logger.String("status", fmt.Sprintf("%d", resp.StatusCode)))
		return nil, errors.NewExternalError("user service", fmt.Errorf("status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("failed to read response", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("user service", err)
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Msg     string `json:"msg"`
		Params  struct {
			Account  string `json:"account"`
			ID       int    `json:"id"`
			Phone    string `json:"phone"`
			TenantID int    `json:"tenant_id"`
		} `json:"params"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		s.logger.Error("failed to unmarshal response", logger.String("error", err.Error()))
		return nil, errors.NewExternalError("user service", err)
	}

	// 接口有时返回 code=200/msg=查询成功，视为成功
	if response.Code != http.StatusOK {
		msg := response.Message
		if msg == "" {
			msg = response.Msg
		}
		if msg == "" {
			msg = fmt.Sprintf("api returned error code: %d", response.Code)
		}
		s.logger.Error("api error", logger.String("code", fmt.Sprintf("%d", response.Code)), logger.String("message", msg))
		return nil, errors.NewExternalError("user service", fmt.Errorf("api error: %s", msg))
	}

	payload := response.Params

	return &user.User{
		Account:  payload.Account,
		ID:       payload.ID,
		Phone:    payload.Phone,
		TenantID: payload.TenantID,
	}, nil
}
