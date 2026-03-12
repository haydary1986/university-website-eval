package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"website-eval-system/config"
)

// AIProvider defines the interface for AI services
type AIProvider interface {
	Chat(prompt string) (string, error)
}

// AIService manages multiple AI providers
type AIService struct {
	Config   *config.Config
	deepseek *DeepSeekClient
	gemini   *GeminiClient
}

func NewAIService(cfg *config.Config) *AIService {
	return &AIService{
		Config:   cfg,
		deepseek: &DeepSeekClient{APIKey: cfg.DeepSeekKey, BaseURL: cfg.DeepSeekURL},
		gemini:   &GeminiClient{APIKey: cfg.GeminiKey, BaseURL: cfg.GeminiURL},
	}
}

func (s *AIService) GetProvider(name string) (AIProvider, error) {
	switch name {
	case "deepseek":
		if s.Config.DeepSeekKey == "" {
			return nil, fmt.Errorf("DeepSeek API key not configured")
		}
		return s.deepseek, nil
	case "gemini":
		if s.Config.GeminiKey == "" {
			return nil, fmt.Errorf("Gemini API key not configured")
		}
		return s.gemini, nil
	default:
		// Try deepseek first, then gemini
		if s.Config.DeepSeekKey != "" {
			return s.deepseek, nil
		}
		if s.Config.GeminiKey != "" {
			return s.gemini, nil
		}
		return nil, fmt.Errorf("no AI provider configured. Set DEEPSEEK_API_KEY or GEMINI_API_KEY")
	}
}

// === DeepSeek Client ===

type DeepSeekClient struct {
	APIKey  string
	BaseURL string
}

type deepSeekRequest struct {
	Model    string              `json:"model"`
	Messages []deepSeekMessage   `json:"messages"`
}

type deepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (d *DeepSeekClient) Chat(prompt string) (string, error) {
	reqBody := deepSeekRequest{
		Model: "deepseek-chat",
		Messages: []deepSeekMessage{
			{
				Role:    "system",
				Content: "You are an expert in evaluating university websites for Iraq's Ministry of Higher Education. Respond in Arabic when the input is in Arabic, otherwise respond in the same language as the input.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", d.BaseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result deepSeekResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return result.Choices[0].Message.Content, nil
}

// === Gemini Client ===

type GeminiClient struct {
	APIKey  string
	BaseURL string
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (g *GeminiClient) Chat(prompt string) (string, error) {
	fullPrompt := "You are an expert in evaluating university websites for Iraq's Ministry of Higher Education. Respond in Arabic when the input is in Arabic, otherwise respond in the same language as the input.\n\n" + prompt

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: fullPrompt},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := g.BaseURL + "?key=" + g.APIKey
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result geminiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}
