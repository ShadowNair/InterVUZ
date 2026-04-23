package news

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Client struct {
	url        string
	httpClient *http.Client
}

func New(url string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 20 * time.Second,
		}
	}
	return &Client{
		url:        url,
		httpClient: httpClient,
	}
}

func (c *Client) LoadNews(ctx context.Context, limit int) ([]domain.NewsItem, error) {
	if limit <= 0 {
		limit = 100
	}

	endpoint := fmt.Sprintf("%s?limit=%d&offset=0", c.url, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload domain.NewsResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return payload.Items, nil
}