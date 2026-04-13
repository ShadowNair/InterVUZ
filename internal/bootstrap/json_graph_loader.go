package bootstrap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type graphFile struct {
	Meta     map[string]any       `json:"meta"`
	Vertices []domain.GraphVertex `json:"vertices"`
	Edges    []domain.GraphEdge   `json:"edges"`
}

func LoadNavigationGraphFromJSON(path string) (domain.NavigationGraph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.NavigationGraph{}, fmt.Errorf("read graph json %s: %w", filepath.Clean(path), err)
	}

	var payload graphFile
	if err := json.Unmarshal(data, &payload); err != nil {
		return domain.NavigationGraph{}, fmt.Errorf("decode graph json %s: %w", filepath.Clean(path), err)
	}

	if len(payload.Vertices) == 0 {
		return domain.NavigationGraph{}, fmt.Errorf("graph json %s has no vertices", filepath.Clean(path))
	}

	if len(payload.Edges) == 0 {
		return domain.NavigationGraph{}, fmt.Errorf("graph json %s has no edges", filepath.Clean(path))
	}

	vertexIDs := make(map[string]struct{}, len(payload.Vertices))
	for _, vertex := range payload.Vertices {
		if vertex.ID == "" {
			return domain.NavigationGraph{}, fmt.Errorf("graph json %s contains vertex without id", filepath.Clean(path))
		}
		if _, exists := vertexIDs[vertex.ID]; exists {
			return domain.NavigationGraph{}, fmt.Errorf("graph json %s contains duplicate vertex id %q", filepath.Clean(path), vertex.ID)
		}
		vertexIDs[vertex.ID] = struct{}{}
	}

	edgeIDs := make(map[string]struct{}, len(payload.Edges))
	for _, edge := range payload.Edges {
		if edge.ID == "" {
			return domain.NavigationGraph{}, fmt.Errorf("graph json %s contains edge without id", filepath.Clean(path))
		}
		if edge.From == "" || edge.To == "" {
			return domain.NavigationGraph{}, fmt.Errorf("graph json %s contains edge %q without endpoints", filepath.Clean(path), edge.ID)
		}
		if _, ok := vertexIDs[edge.From]; !ok {
			return domain.NavigationGraph{}, fmt.Errorf("graph json %s edge %q references unknown vertex %q", filepath.Clean(path), edge.ID, edge.From)
		}
		if _, ok := vertexIDs[edge.To]; !ok {
			return domain.NavigationGraph{}, fmt.Errorf("graph json %s edge %q references unknown vertex %q", filepath.Clean(path), edge.ID, edge.To)
		}
		if _, exists := edgeIDs[edge.ID]; exists {
			return domain.NavigationGraph{}, fmt.Errorf("graph json %s contains duplicate edge id %q", filepath.Clean(path), edge.ID)
		}
		edgeIDs[edge.ID] = struct{}{}
	}

	return domain.NavigationGraph{
		Vertices: payload.Vertices,
		Edges:    payload.Edges,
	}, nil
}
