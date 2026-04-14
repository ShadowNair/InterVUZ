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
	url            string
	targetRootUUID string
	httpClient     *http.Client
}

func New(url string, targetRootUUID string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &Client{
		url:            url,
		targetRootUUID: strings.TrimSpace(targetRootUUID),
		httpClient:     httpClient,
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

	rootNode := payload.Data
	if c.targetRootUUID != "" {
		filteredRoot, ok := pruneTreeToTargetRoot(payload.Data, c.targetRootUUID)
		if !ok {
			return nil, fmt.Errorf("target root uuid %s not found in structure tree", c.targetRootUUID)
		}
		rootNode = filteredRoot
	}

	units := make([]domain.StructureUnit, 0, 2048)
	ctxWalk := walkContext{}
	walkNode(rootNode, nil, ctxWalk, &units)

	return units, nil
}

type nodeRef struct {
	ExternalUUID string
	Code         string
	Name         string
}

type walkContext struct {
	Faculty    *nodeRef
	Department *nodeRef
	CourseNode *nodeRef
}

func pruneTreeToTargetRoot(node domain.StructureNode, targetRootUUID string) (domain.StructureNode, bool) {
	targetRootUUID = strings.TrimSpace(targetRootUUID)
	if targetRootUUID == "" {
		return node, true
	}

	if strings.TrimSpace(node.UUID) == targetRootUUID {
		return node, true
	}

	filteredChildren := make([]domain.StructureNode, 0, len(node.Children))
	for _, child := range node.Children {
		filteredChild, ok := pruneTreeToTargetRoot(child, targetRootUUID)
		if ok {
			filteredChildren = append(filteredChildren, filteredChild)
		}
	}

	if len(filteredChildren) == 0 {
		return domain.StructureNode{}, false
	}

	node.Children = filteredChildren
	return node, true
}

func walkNode(node domain.StructureNode, parentExternalID *string, ctx walkContext, out *[]domain.StructureUnit) {
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

	nodeType := normalizeNodeType(node, len(node.Children) > 0)

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

	if nodeType == "group" {
		if ctx.Faculty != nil {
			item.FacultyExternalUUID = stringPtr(ctx.Faculty.ExternalUUID)
			item.FacultyCode = stringPtr(ctx.Faculty.Code)
			item.FacultyName = stringPtr(ctx.Faculty.Name)
		}
		if ctx.Department != nil {
			item.DepartmentExternalUUID = stringPtr(ctx.Department.ExternalUUID)
			item.DepartmentCode = stringPtr(ctx.Department.Code)
			item.DepartmentName = stringPtr(ctx.Department.Name)
		}
		if ctx.CourseNode != nil {
			item.CourseNodeExternalUUID = stringPtr(ctx.CourseNode.ExternalUUID)
			item.CourseNodeName = stringPtr(ctx.CourseNode.Name)
		}
	}

	*out = append(*out, item)

	nextCtx := ctx
	switch nodeType {
	case "faculty":
		nextCtx.Faculty = &nodeRef{ExternalUUID: externalUUID, Code: code, Name: name}
		nextCtx.Department = nil
		nextCtx.CourseNode = nil
	case "department":
		nextCtx.Department = &nodeRef{ExternalUUID: externalUUID, Code: code, Name: name}
		nextCtx.CourseNode = nil
	case "course":
		nextCtx.CourseNode = &nodeRef{ExternalUUID: externalUUID, Code: code, Name: name}
	}

	for _, child := range node.Children {
		parentID := externalUUID
		walkNode(child, &parentID, nextCtx, out)
	}
}

func normalizeNodeType(node domain.StructureNode, hasChildren bool) string {
	nodeType := strings.TrimSpace(node.NodeType)
	if nodeType != "" {
		return nodeType
	}

	abbr := strings.TrimSpace(node.Abbr)

	switch {
	case node.Course > 0 && hasChildren:
		return "course"
	case hasChildren && strings.Contains(abbr, "(") && strings.Contains(strings.ToLower(abbr), "курс"):
		return "course"
	case hasChildren:
		return "container"
	default:
		return "group"
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

func stringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
