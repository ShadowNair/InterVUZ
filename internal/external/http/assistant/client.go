package assistant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(baseURL string, apiKey string, timeout time.Duration, httpClient *http.Client) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://polza.ai/api/v1"
	}

	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     strings.TrimSpace(apiKey),
		httpClient: httpClient,
	}
}

func (c *Client) Complete(ctx context.Context, model string, messages []Message) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("assistant api key is not configured")
	}

	if strings.TrimSpace(model) == "" {
		return "", fmt.Errorf("assistant model is not configured")
	}

	reqBody := chatCompletionRequest{
		Model:       model,
		Temperature: 0.1,
		Messages:    messages,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("unexpected status=%d body=%s", resp.StatusCode, limitString(strings.TrimSpace(string(respBody)), 1000))
	}

	var payload chatCompletionResponse
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(payload.Choices) == 0 {
		return "", fmt.Errorf("empty choices in response")
	}

	content := strings.TrimSpace(payload.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("empty content in response")
	}

	return content, nil
}

type chatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func limitString(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max] + "..."
}
