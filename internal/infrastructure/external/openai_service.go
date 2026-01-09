package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"iwut-smartclass-backend/internal/domain/errors"
	"iwut-smartclass-backend/internal/infrastructure/config"
	"iwut-smartclass-backend/internal/infrastructure/logger"
)

// OpenAIRequest 通用 OpenAI 请求结构
type OpenAIRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Stream      bool    `json:"stream"`
	Temperature float32 `json:"temperature"`
}

// OpenAIResponse 通用 OpenAI 响应结构
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     uint32 `json:"prompt_tokens"`
		CompletionTokens uint32 `json:"completion_tokens"`
		TotalTokens      uint32 `json:"total_tokens"`
	} `json:"usage"`
}

// OpenAIService OpenAI服务
type OpenAIService struct {
	cfg    *config.Config
	logger logger.Logger
}

// NewOpenAIService 创建OpenAI服务
func NewOpenAIService(cfg *config.Config, logger logger.Logger) *OpenAIService {
	return &OpenAIService{
		cfg:    cfg,
		logger: logger,
	}
}

// CallOpenAI 调用 OpenAI API
func (s *OpenAIService) CallOpenAI(prompt, userInput string) (string, uint32, error) {
	s.logger.Info("creating OpenAI request")

	requestBody, err := json.Marshal(OpenAIRequest{
		Model: s.cfg.OpenaiModel,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "system", Content: prompt},
			{Role: "user", Content: userInput},
		},
		Stream:      false,
		Temperature: s.cfg.Temperature,
	})
	if err != nil {
		s.logger.Error("failed to marshal request body", logger.String("error", err.Error()))
		return "", 0, errors.NewInternalError("failed to marshal request", err)
	}

	req, err := http.NewRequest("POST", s.cfg.OpenaiEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		s.logger.Error("failed to create request", logger.String("error", err.Error()))
		return "", 0, errors.NewInternalError("failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.cfg.OpenaiKey))

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("failed to send request", logger.String("error", err.Error()))
		return "", 0, errors.NewExternalError("openai", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("received non-200 response", logger.String("status", fmt.Sprintf("%d", resp.StatusCode)))
		return "", 0, errors.NewExternalError("openai", fmt.Errorf("status code: %d", resp.StatusCode))
	}

	var openAIResponse OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResponse); err != nil {
		s.logger.Error("failed to decode response", logger.String("error", err.Error()))
		return "", 0, errors.NewExternalError("openai", err)
	}

	if len(openAIResponse.Choices) == 0 {
		s.logger.Error("no choices in response")
		return "", 0, errors.NewExternalError("openai", fmt.Errorf("no choices in response"))
	}

	content := openAIResponse.Choices[0].Message.Content
	usage := openAIResponse.Usage
	s.logger.Info("OpenAI call successful",
		logger.String("prompt_tokens", fmt.Sprintf("%d", usage.PromptTokens)),
		logger.String("completion_tokens", fmt.Sprintf("%d", usage.CompletionTokens)),
		logger.String("total_tokens", fmt.Sprintf("%d", usage.TotalTokens)),
	)
	return content, usage.TotalTokens, nil
}
