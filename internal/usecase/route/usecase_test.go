package usecase

import (
	"context"
	"testing"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type graphRepositoryStub struct {
	graph domain.NavigationGraph
}

func (stub graphRepositoryStub) Get(context.Context) (domain.NavigationGraph, error) {
	return stub.graph, nil
}

type placeRepositoryStub struct {
	places map[string]domain.Place
}

func (stub placeRepositoryStub) List(context.Context, domain.PlaceFilter) ([]domain.Place, error) {
	items := make([]domain.Place, 0, len(stub.places))
	for _, place := range stub.places {
		items = append(items, place)
	}

	return items, nil
}

func (stub placeRepositoryStub) GetByID(_ context.Context, placeID string) (domain.Place, error) {
	place, ok := stub.places[placeID]
	if ok {
		return place, nil
	}

	return domain.Place{}, domain.ErrNotFound
}

func TestBuildRoute(t *testing.T) {
	repository := placeRepositoryStub{
		places: map[string]domain.Place{
			"place-start": {
				ID:   "place-start",
				Name: "Start",
				Coordinates: domain.Coordinates{
					Building: "1",
					Floor:    1,
					X:        0,
					Y:        0,
				},
				IsAccessible: true,
			},
			"place-mid": {
				ID:   "place-mid",
				Name: "Mid",
				Coordinates: domain.Coordinates{
					Building: "1",
					Floor:    1,
					X:        3,
					Y:        0,
				},
				IsAccessible: true,
			},
			"place-destination": {
				ID:   "place-destination",
				Name: "Destination",
				Coordinates: domain.Coordinates{
					Building: "1",
					Floor:    1,
					X:        3,
					Y:        4,
				},
				IsAccessible: true,
			},
		},
	}

	useCase := New(
		graphRepositoryStub{
			graph: domain.NavigationGraph{
				Vertices: []domain.GraphVertex{
					{ID: "v1", PlaceID: "place-start", Building: "1", Floor: 1, X: 0, Y: 0, IsAccessible: true},
					{ID: "v2", PlaceID: "place-mid", Building: "1", Floor: 1, X: 3, Y: 0, IsAccessible: true},
					{ID: "v3", PlaceID: "place-destination", Building: "1", Floor: 1, X: 3, Y: 4, IsAccessible: true},
				},
				Edges: []domain.GraphEdge{
					{ID: "e1", From: "v1", To: "v2"},
					{ID: "e2", From: "v2", To: "v3"},
				},
			},
		},
		repository,
	)

	route, err := useCase.Build(context.Background(), domain.RouteRequest{
		FromPlaceID: "place-start",
		ToPlaceID:   "place-destination",
	})
	if err != nil {
		t.Fatalf("expected route, got error: %v", err)
	}

	if got, want := len(route.Steps), 3; got != want {
		t.Fatalf("expected %d steps, got %d", want, got)
	}

	if got, want := route.DistanceMeters, 7; got != want {
		t.Fatalf("expected distance %d, got %d", want, got)
	}

	if route.Steps[0].PlaceID != "place-start" || route.Steps[len(route.Steps)-1].PlaceID != "place-destination" {
		t.Fatalf("unexpected route endpoints: %#v", route.Steps)
	}
}
