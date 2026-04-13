package bootstrap

import (
	"testing"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

func TestParseNavigationGraphSVG(t *testing.T) {
	svg := []byte(`
<svg xmlns="http://www.w3.org/2000/svg">
  <circle data-vertex-id="v-start" data-place-id="place-start" cx="10" cy="20" />
  <circle data-vertex-id="v-hall" data-building="1" data-floor="1" cx="40" cy="20" data-accessible="true" />
  <circle data-vertex-id="v-end" data-place-id="place-end" />
  <line data-edge-id="e1" data-from="v-start" data-to="v-hall" />
  <line data-from="v-hall" data-to="v-end" />
</svg>
`)

	graph, err := ParseNavigationGraphSVG(svg, []domain.Place{
		{
			ID: "place-start",
			Coordinates: domain.Coordinates{
				Building: "1",
				Floor:    1,
				X:        10,
				Y:        20,
			},
			IsAccessible: true,
		},
		{
			ID: "place-end",
			Coordinates: domain.Coordinates{
				Building: "1",
				Floor:    1,
				X:        80,
				Y:        20,
			},
			IsAccessible: false,
		},
	})
	if err != nil {
		t.Fatalf("expected graph, got error: %v", err)
	}

	if got, want := len(graph.Vertices), 3; got != want {
		t.Fatalf("expected %d vertices, got %d", want, got)
	}

	if got, want := len(graph.Edges), 2; got != want {
		t.Fatalf("expected %d edges, got %d", want, got)
	}

	if graph.Vertices[0].PlaceID != "place-start" {
		t.Fatalf("expected first vertex to keep place binding, got %#v", graph.Vertices[0])
	}

	if graph.Vertices[2].X != 80 || graph.Vertices[2].Y != 20 {
		t.Fatalf("expected place-linked fallback coordinates, got %#v", graph.Vertices[2])
	}

	if graph.Vertices[2].IsAccessible {
		t.Fatalf("expected place accessibility to propagate, got %#v", graph.Vertices[2])
	}
}

func TestParseNavigationGraphSVGRejectsUnknownEdgeVertex(t *testing.T) {
	svg := []byte(`
<svg xmlns="http://www.w3.org/2000/svg">
  <circle data-vertex-id="v1" data-building="1" data-floor="1" cx="10" cy="20" />
  <line data-from="v1" data-to="missing" />
</svg>
`)

	_, err := ParseNavigationGraphSVG(svg, nil)
	if err == nil {
		t.Fatal("expected parser error for unknown vertex reference")
	}
}
