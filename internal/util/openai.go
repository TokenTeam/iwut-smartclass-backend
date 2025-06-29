package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"iwut-smartclass-backend/internal/config"
	"iwut-smartclass-backend/internal/middleware"
	"net/http"
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

// CallOpenAI 调用 OpenAI API
func CallOpenAI(cfg *config.Config, prompt, userInput string) (string, uint32, error) {

	middleware.Logger.Log("INFO", "[OpenAI] Creating a new request")

	requestBody, err := json.Marshal(OpenAIRequest{
		Model: cfg.OpenaiModel,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "system", Content: prompt},
			{Role: "user", Content: userInput},
		},
		Stream:      false,
		Temperature: cfg.Temperature,
	})
	if err != nil {
		middleware.Logger.Log("WARN", fmt.Sprintf("[OpenAI] Failed to marshal request body: %v", err))
		return "", 0, err
	}

	req, err := http.NewRequest("POST", cfg.OpenaiEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		middleware.Logger.Log("WARN", fmt.Sprintf("[OpenAI] Failed to create new request: %v", err))
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.OpenaiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		middleware.Logger.Log("WARN", fmt.Sprintf("[OpenAI] Failed to send request: %v", err))
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		middleware.Logger.Log("WARN", fmt.Sprintf("[OpenAI] Received non-200 response: %d", resp.StatusCode))
		return "", 0, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var openAIResponse OpenAIResponse
	err = json.NewDecoder(resp.Body).Decode(&openAIResponse)
	if err != nil {
		middleware.Logger.Log("WARN", fmt.Sprintf("[OpenAI] Failed to decode response body: %v", err))
		return "", 0, err
	}

	if len(openAIResponse.Choices) == 0 {
		middleware.Logger.Log("WARN", "[OpenAI] No choices in response")
		return "", 0, fmt.Errorf("no choices in response")
	}

	content := openAIResponse.Choices[0].Message.Content
	usage := openAIResponse.Usage
	middleware.Logger.Log("INFO", fmt.Sprintf("[OpenAI] Called successfully. Usage: Prompt %d, Completion %d, Total %d", usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens))
	return content, usage.TotalTokens, nil
}
