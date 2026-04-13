package structure

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
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
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &Client{
		url:        url,
		httpClient: httpClient,
	}
}

func (c *Client) FetchUnits(ctx context.Context) ([]domain.StructureUnit, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
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

	var payload domain.StructureResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode structure response: %w", err)
	}

	units := make([]domain.StructureUnit, 0, 256)
	walkNode(payload.Data, nil, &units)

	return units, nil
}

func walkNode(node domain.StructureNode, parentExternalID *string, out *[]domain.StructureUnit) {
	raw, _ := json.Marshal(node)

	externalUUID := strings.TrimSpace(node.UUID)
	if externalUUID == "" {
		externalUUID = syntheticID(node, parentExternalID)
	}

	code := strings.TrimSpace(node.Abbr)
	if code == "" {
		code = strings.TrimSpace(node.Name)
	}
	if code == "" {
		code = externalUUID
	}

	name := strings.TrimSpace(node.Name)
	if name == "" {
		name = code
	}

	nodeType := strings.TrimSpace(node.NodeType)
	if nodeType == "" {
		if len(node.Children) > 0 {
			nodeType = "container"
		} else {
			nodeType = "group"
		}
	}

	item := domain.StructureUnit{
		ExternalUUID:     externalUUID,
		ParentExternalID: parentExternalID,
		Code:             code,
		Name:             name,
		NodeType:         nodeType,
		RawPayload:       raw,
	}

	if node.Course > 0 {
		course := node.Course
		item.Course = &course
	}
	if node.Semester > 0 {
		semester := node.Semester
		item.Semester = &semester
	}

	*out = append(*out, item)

	for _, child := range node.Children {
		parentID := externalUUID
		walkNode(child, &parentID, out)
	}
}

func syntheticID(node domain.StructureNode, parentExternalID *string) string {
	parts := []string{
		strings.TrimSpace(node.Abbr),
		strings.TrimSpace(node.Name),
		fmt.Sprintf("%d", node.Course),
		fmt.Sprintf("%d", node.Semester),
		strings.TrimSpace(node.NodeType),
	}
	if parentExternalID != nil {
		parts = append(parts, *parentExternalID)
	}

	sum := sha1.Sum([]byte(strings.Join(parts, "|")))
	return "synthetic-" + hex.EncodeToString(sum[:])
}
