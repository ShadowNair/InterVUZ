package memory

import (
	"context"
	"fmt"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type Repository struct {
	graph domain.NavigationGraph
}

func New(places []domain.Place) *Repository {
	return &Repository{
		graph: buildGraph(places),
	}
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

func buildGraph(places []domain.Place) domain.NavigationGraph {
	vertices := make([]domain.GraphVertex, 0, len(places))
	for _, place := range places {
		vertices = append(vertices, domain.GraphVertex{
			ID:           place.ID,
			PlaceID:      place.ID,
			Building:     place.Coordinates.Building,
			Floor:        place.Coordinates.Floor,
			X:            place.Coordinates.X,
			Y:            place.Coordinates.Y,
			IsAccessible: place.IsAccessible,
		})
	}

	edges := make([]domain.GraphEdge, 0, len(vertices)*(len(vertices)-1)/2)
	for fromIndex := 0; fromIndex < len(vertices); fromIndex++ {
		for toIndex := fromIndex + 1; toIndex < len(vertices); toIndex++ {
			edges = append(edges, domain.GraphEdge{
				ID:   fmt.Sprintf("edge-%d", len(edges)+1),
				From: vertices[fromIndex].ID,
				To:   vertices[toIndex].ID,
			})
		}
	}

	return domain.NavigationGraph{
		Vertices: vertices,
		Edges:    edges,
	}
}
