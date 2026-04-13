package bootstrap

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNavigationGraphFromJSON(t *testing.T) {
	tempDir := t.TempDir()
	graphPath := filepath.Join(tempDir, "graph.json")

	if err := os.WriteFile(graphPath, []byte(`{
  "meta": {"version": 1},
  "vertices": [
    {"id": "v1", "placeId": "place_100", "building": "1", "floor": 1, "x": 10, "y": 20, "isAccessible": true},
    {"id": "v2", "building": "1", "floor": 1, "x": 40, "y": 20, "isAccessible": true}
  ],
  "edges": [
    {"id": "e1", "from": "v1", "to": "v2"}
  ]
}`), 0o644); err != nil {
		t.Fatalf("write temp graph: %v", err)
	}

	graph, err := LoadNavigationGraphFromJSON(graphPath)
	if err != nil {
		t.Fatalf("load graph: %v", err)
	}

	if got, want := len(graph.Vertices), 2; got != want {
		t.Fatalf("expected %d vertices, got %d", want, got)
	}

	if got, want := len(graph.Edges), 1; got != want {
		t.Fatalf("expected %d edges, got %d", want, got)
	}
}

func TestLoadNavigationGraphFromJSONFailsWithoutEdges(t *testing.T) {
	tempDir := t.TempDir()
	graphPath := filepath.Join(tempDir, "graph.json")

	if err := os.WriteFile(graphPath, []byte(`{
  "meta": {"version": 1},
  "vertices": [
    {"id": "v1", "building": "1", "floor": 1, "x": 10, "y": 10, "isAccessible": true},
    {"id": "v2", "building": "1", "floor": 1, "x": 30, "y": 10, "isAccessible": true},
    {"id": "v3", "building": "1", "floor": 1, "x": 30, "y": 30, "isAccessible": true}
  ],
  "edges": []
}`), 0o644); err != nil {
		t.Fatalf("write temp graph: %v", err)
	}

	graph, err := LoadNavigationGraphFromJSON(graphPath)
	if err == nil {
		t.Fatalf("expected error, got graph: %#v", graph)
	}

	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		t.Fatalf("unexpected path error: %v", err)
	}
}
