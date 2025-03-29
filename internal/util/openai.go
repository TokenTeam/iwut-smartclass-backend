package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"iwut-smart-timetable-backend/internal/config"
	"iwut-smart-timetable-backend/internal/middleware"
	"net/http"
)

// OpenAIRequest 通用 OpenAI 请求结构
type OpenAIRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

// OpenAIResponse 通用 OpenAI 响应结构
type OpenAIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

// CallOpenAI 调用 OpenAI API
func CallOpenAI(cfg *config.Config, userInput string) (string, error) {
	url := cfg.OpenaiEndpoint
	apiKey := cfg.OpenaiKey
	model := cfg.OpenaiModel

	requestBody, err := json.Marshal(OpenAIRequest{
		Model: model,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "system", Content: cfg.Prompt},
			{Role: "user", Content: userInput},
		},
	})
	if err != nil {
		middleware.Logger.Log("WARN", fmt.Sprintf("[OpenAI] Failed to marshal request body: %v", err))
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		middleware.Logger.Log("WARN", fmt.Sprintf("[OpenAI] Failed to create new request: %v", err))
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		middleware.Logger.Log("WARN", fmt.Sprintf("[OpenAI] Failed to send request: %v", err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		middleware.Logger.Log("WARN", fmt.Sprintf("[OpenAI] Received non-200 response: %d", resp.StatusCode))
		return "", fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var openAIResponse OpenAIResponse
	err = json.NewDecoder(resp.Body).Decode(&openAIResponse)
	if err != nil {
		middleware.Logger.Log("WARN", fmt.Sprintf("[OpenAI] Failed to decode response body: %v", err))
		return "", err
	}

	if len(openAIResponse.Choices) == 0 {
		middleware.Logger.Log("WARN", "[OpenAI] No choices in response")
		return "", fmt.Errorf("no choices in response")
	}

	middleware.Logger.Log("DEBUG", "[OpenAI] Called API successfully")
	return openAIResponse.Choices[0].Text, nil
}
