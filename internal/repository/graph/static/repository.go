package static

import (
	"context"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Repository struct {
	graph domain.NavigationGraph
}

func New(graph domain.NavigationGraph) *Repository {
	return &Repository{graph: graph}
}

func (r *Repository) Get(context.Context) (domain.NavigationGraph, error) {
	graph := domain.NavigationGraph{
		Vertices: make([]domain.GraphVertex, len(r.graph.Vertices)),
		Edges:    make([]domain.GraphEdge, len(r.graph.Edges)),
	}

	copy(graph.Vertices, r.graph.Vertices)
	copy(graph.Edges, r.graph.Edges)

	return graph, nil
}
