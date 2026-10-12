// Package meedoo provides the MeeDoo AI Go SDK client.
//
// Usage:
//
//	client := meedoo.NewClient("md-your-api-key")
//	resp, err := client.Chat.Completions.Create(&meedoo.ChatCompletionParams{
//	    Model:    "meedoo:latest",
//	    Messages: []meedoo.Message{{Role: "user", Content: "Hello!"}},
//	})
//	fmt.Println(resp.Content())
//
// Install: go get github.com/MeeDoo-AI/SDK_Go
package meedoo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the main MeeDoo AI SDK client.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	MaxRetries int

	Chat   *ChatService
	Models *ModelsService
	Usage  *UsageService
	Judge  *JudgeService
}

// NewClient creates a new MeeDoo AI client.
func NewClient(apiKey string) *Client {
	return NewClientWithBaseURL(apiKey, "https://api.meedoo.co/api")
}

// NewClientWithBaseURL creates a client with a custom base URL.
func NewClientWithBaseURL(apiKey, baseURL string) *Client {
	c := &Client{
		APIKey:     apiKey,
		BaseURL:    trimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
		MaxRetries: 3,
	}
	c.Chat = &ChatService{client: c}
	c.Models = &ModelsService{client: c}
	c.Usage = &UsageService{client: c}
	c.Judge = &JudgeService{client: c}
	return c
}

func (c *Client) do(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	var lastErr error
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = &APIError{Message: fmt.Sprintf("connection error: %v", err), StatusCode: 0}
			if attempt < c.MaxRetries {
				time.Sleep(time.Duration(1<<attempt) * time.Second)
				continue
			}
			return nil, lastErr
		}

		respBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 200 {
			return respBytes, nil
		}

		var errBody struct {
			Detail string `json:"detail"`
		}
		_ = json.Unmarshal(respBytes, &errBody)
		detail := errBody.Detail
		if detail == "" {
			detail = string(respBytes)
		}

		switch {
		case resp.StatusCode == 401:
			return nil, &AuthenticationError{Message: detail}
		case resp.StatusCode == 429:
			return nil, &RateLimitError{Message: detail}
		case contains(detail, "Insufficient API credits"):
			return nil, &InsufficientCreditsError{Message: detail}
		case contains(detail, "Model not found"):
			return nil, &ModelNotFoundError{Message: detail}
		case resp.StatusCode >= 500:
			lastErr = &APIError{Message: detail, StatusCode: resp.StatusCode}
			if attempt < c.MaxRetries {
				time.Sleep(time.Duration(1<<attempt) * time.Second)
				continue
			}
			return nil, lastErr
		default:
			return nil, &APIError{Message: detail, StatusCode: resp.StatusCode}
		}
	}
	return nil, lastErr
}

// GetConfig returns the API configuration.
func (c *Client) GetConfig() (map[string]interface{}, error) {
	data, err := c.do("GET", "/config", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	return result, json.Unmarshal(data, &result)
}

// GetHealth returns the health status.
func (c *Client) GetHealth() (map[string]interface{}, error) {
	healthURL := trimSuffix(c.BaseURL, "/api") + "/health"
	req, _ := http.NewRequest("GET", healthURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &APIError{Message: err.Error()}
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	return result, json.Unmarshal(data, &result)
}

// --- Services ---

type ChatService struct{ client *Client }

func (s *ChatService) Completions() *CompletionsService { return &CompletionsService{client: s.client} }

type CompletionsService struct{ client *Client }

func (s *CompletionsService) Create(params *ChatCompletionParams) (*ChatCompletion, error) {
	data, err := s.client.do("POST", "/chat/completions", params)
	if err != nil {
		return nil, err
	}
	var resp ChatCompletion
	return &resp, json.Unmarshal(data, &resp)
}

type ModelsService struct{ client *Client }

func (s *ModelsService) List() (*ModelList, error) {
	data, err := s.client.do("GET", "/models", nil)
	if err != nil {
		return nil, err
	}
	var result ModelList
	return &result, json.Unmarshal(data, &result)
}

type UsageService struct{ client *Client }

func (s *UsageService) Credits() (*CreditsResponse, error) {
	data, err := s.client.do("GET", "/v1/usage/credits", nil)
	if err != nil {
		return nil, err
	}
	var result CreditsResponse
	return &result, json.Unmarshal(data, &result)
}

func (s *UsageService) Stats() (*UsageStats, error) {
	data, err := s.client.do("GET", "/v1/usage/stats", nil)
	if err != nil {
		return nil, err
	}
	var result UsageStats
	return &result, json.Unmarshal(data, &result)
}

type JudgeService struct{ client *Client }

func (s *JudgeService) Status() (map[string]interface{}, error) {
	data, err := s.client.do("GET", "/v1/judge/status", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	return result, json.Unmarshal(data, &result)
}

func (s *JudgeService) ShouldSearchWeb(userMessage string) (map[string]interface{}, error) {
	return s.post("/v1/judge/should-search-web", map[string]string{"user_message": userMessage})
}

func (s *JudgeService) DetectInjection(userMessage string) (map[string]interface{}, error) {
	return s.post("/v1/judge/detect-injection", map[string]string{"user_message": userMessage})
}

func (s *JudgeService) Moderate(text string) (map[string]interface{}, error) {
	return s.post("/v1/judge/moderate", map[string]string{"text": text})
}

func (s *JudgeService) DetectSpam(message, userHistory string) (map[string]interface{}, error) {
	return s.post("/v1/judge/detect-spam", map[string]string{"message": message, "user_history": userHistory})
}

func (s *JudgeService) ValidateResponse(userMessage, llmResponse, sources string) (map[string]interface{}, error) {
	return s.post("/v1/judge/validate-response", map[string]string{
		"user_message": userMessage, "llm_response": llmResponse, "sources": sources,
	})
}

func (s *JudgeService) FilterMemory(information, conversationContext string) (map[string]interface{}, error) {
	return s.post("/v1/judge/filter-memory", map[string]string{
		"information": information, "conversation_context": conversationContext,
	})
}

func (s *JudgeService) IsPrayerTimes(userMessage string) (map[string]interface{}, error) {
	return s.post("/v1/judge/is-prayer-times", map[string]string{"user_message": userMessage})
}

func (s *JudgeService) ClassifyTicket(subject, body string) (map[string]interface{}, error) {
	return s.post("/v1/judge/classify-ticket", map[string]string{"subject": subject, "body": body})
}

func (s *JudgeService) Sentiment() (map[string]interface{}, error) {
	data, err := s.client.do("GET", "/v1/judge/sentiment", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	return result, json.Unmarshal(data, &result)
}

func (s *JudgeService) post(path string, body interface{}) (map[string]interface{}, error) {
	data, err := s.client.do("POST", path, body)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	return result, json.Unmarshal(data, &result)
}

// --- Types ---

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionParams struct {
	Model             string                 `json:"model"`
	Messages          []Message              `json:"messages"`
	Stream            bool                   `json:"stream,omitempty"`
	Temperature       *float64               `json:"temperature,omitempty"`
	TopP              *float64               `json:"top_p,omitempty"`
	MinP              *float64               `json:"min_p,omitempty"`
	MaxTokens         *int                   `json:"max_tokens,omitempty"`
	FrequencyPenalty  *float64               `json:"frequency_penalty,omitempty"`
	PresencePenalty   *float64               `json:"presence_penalty,omitempty"`
	ReasoningEffort   string                 `json:"reasoning_effort,omitempty"`
	Seed              interface{}            `json:"seed,omitempty"`
	Stop              []string               `json:"stop,omitempty"`
	ResponseFormat    map[string]string      `json:"response_format,omitempty"`
	Think             *bool                  `json:"think,omitempty"`
	Features          map[string]bool        `json:"features,omitempty"`
	ToolIDs           []string               `json:"tool_ids,omitempty"`
	Tools             []map[string]interface{} `json:"tools,omitempty"`
	ToolChoice        string                 `json:"tool_choice,omitempty"`
	Params            map[string]interface{} `json:"params,omitempty"`
	ChatID            string                 `json:"chat_id,omitempty"`
	ID                string                 `json:"id,omitempty"`
	ParentID          string                 `json:"parent_id,omitempty"`
	SessionID         string                 `json:"session_id,omitempty"`
	Files             []map[string]interface{} `json:"files,omitempty"`
	SkillIDs          []string               `json:"skill_ids,omitempty"`
	Variables         map[string]interface{} `json:"variables,omitempty"`
	FilterIDs         []string               `json:"filter_ids,omitempty"`
	BackgroundTasks   map[string]bool        `json:"background_tasks,omitempty"`
	UseCredits        *bool                  `json:"use_credits,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

type ChatCompletion struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index        int     `json:"index"`
		FinishReason string  `json:"finish_reason"`
		Message      struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
		InputTokens      int `json:"input_tokens"`
		OutputTokens     int `json:"output_tokens"`
	} `json:"usage"`
}

func (c *ChatCompletion) Content() string {
	if len(c.Choices) > 0 {
		return c.Choices[0].Message.Content
	}
	return ""
}

func (c *ChatCompletion) TotalTokens() int {
	return c.Usage.TotalTokens
}

type ModelList struct {
	Data []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"data"`
}

type CreditsResponse struct {
	CreditBalance float64 `json:"credit_balance"`
	TokenBalance  int64   `json:"token_balance"`
}

type UsageStats struct {
	CurrentRequests int    `json:"current_requests"`
	MaxRequests     int    `json:"max_requests"`
	Remaining       int    `json:"remaining"`
	Plan            string `json:"plan"`
}

// --- Errors ---

type MeeDooError interface {
	Error() string
}

type BaseError struct {
	Message string
}

func (e *BaseError) Error() string { return e.Message }

type AuthenticationError struct{ Message string }
func (e *AuthenticationError) Error() string { return "authentication error: " + e.Message }

type RateLimitError struct{ Message string }
func (e *RateLimitError) Error() string { return "rate limit: " + e.Message }

type InsufficientCreditsError struct{ Message string }
func (e *InsufficientCreditsError) Error() string { return "insufficient credits: " + e.Message }

type ModelNotFoundError struct{ Message string }
func (e *ModelNotFoundError) Error() string { return "model not found: " + e.Message }

type APIError struct {
	Message    string
	StatusCode int
}
func (e *APIError) Error() string { return fmt.Sprintf("api error ( %d): %s", e.StatusCode, e.Message) }

// --- Helpers ---

func trimSuffix(s, suffix string) string {
	for len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}