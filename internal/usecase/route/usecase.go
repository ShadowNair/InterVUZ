package usecase

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

var (
	ErrInvalidInput              = errors.New("invalid route input")
	ErrStartPlaceNotFound        = errors.New("start place not found")
	ErrDestinationPlaceNotFound  = errors.New("destination place not found")
	ErrStartVertexNotFound       = errors.New("start vertex not found")
	ErrDestinationVertexNotFound = errors.New("destination vertex not found")
	ErrDestinationWithoutVertex  = errors.New("destination is not linked to the graph")
	ErrRouteNotFound             = errors.New("route not found")
)

type GraphRepository interface {
	Get(ctx context.Context) (domain.NavigationGraph, error)
}

type PlaceRepository interface {
	List(ctx context.Context, filter domain.PlaceFilter) ([]domain.Place, error)
	GetByID(ctx context.Context, placeID string) (domain.Place, error)
}

type UseCase struct {
	graphRepository GraphRepository
	placeRepository PlaceRepository
}

func New(graphRepository GraphRepository, placeRepository PlaceRepository) *UseCase {
	return &UseCase{
		graphRepository: graphRepository,
		placeRepository: placeRepository,
	}
}

func (uc *UseCase) Build(ctx context.Context, request domain.RouteRequest) (*domain.Route, error) {
	if !hasRouteEndpoints(request) {
		return nil, ErrInvalidInput
	}

	graph, err := uc.graphRepository.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("load graph: %w", err)
	}

	placesByID := map[string]domain.Place{}
	if uc.placeRepository != nil {
		places, listErr := uc.placeRepository.List(ctx, domain.PlaceFilter{})
		if listErr != nil {
			return nil, fmt.Errorf("load places catalog: %w", listErr)
		}
		placesByID = make(map[string]domain.Place, len(places))
		for _, place := range places {
			placesByID[place.ID] = place
		}
	}

	startVertexID, err := uc.resolveVertexID(ctx, graph, request.FromVertexID, request.FromPlaceID, true, placesByID)
	if err != nil {
		return nil, err
	}

	destinationVertexID, err := uc.resolveVertexID(ctx, graph, request.ToVertexID, request.ToPlaceID, false, placesByID)
	if err != nil {
		return nil, err
	}

	if request.AccessibleOnly {
		graph = filterAccessibleGraph(graph)
	}

	if request.AccessibleOnly {
		if _, ok := findVertexByID(graph, startVertexID); !ok {
			return nil, ErrStartVertexNotFound
		}
		if _, ok := findVertexByID(graph, destinationVertexID); !ok {
			return nil, ErrDestinationVertexNotFound
		}
	}

	pathIDs, distance, err := shortestPath(graph, startVertexID, destinationVertexID)
	if err != nil {
		return nil, err
	}

	verticesByID := make(map[string]domain.GraphVertex, len(graph.Vertices))
	for _, vertex := range graph.Vertices {
		verticesByID[vertex.ID] = vertex
	}

	pathVertices := make([]domain.GraphVertex, 0, len(pathIDs))
	for _, vertexID := range pathIDs {
		vertex, ok := verticesByID[vertexID]
		if !ok {
			return nil, ErrRouteNotFound
		}
		pathVertices = append(pathVertices, vertex)
	}

	duration := distance / 45
	if duration == 0 {
		duration = 1
	}

	return &domain.Route{
		DistanceMeters:           distance,
		EstimatedDurationMinutes: duration,
		Steps:                    buildRouteSteps(pathVertices, placesByID),
	}, nil
}

func hasRouteEndpoints(request domain.RouteRequest) bool {
	return (request.FromVertexID != "" || request.FromPlaceID != "") &&
		(request.ToVertexID != "" || request.ToPlaceID != "")
}

func (uc *UseCase) resolveVertexID(
	ctx context.Context,
	graph domain.NavigationGraph,
	vertexID string,
	placeID string,
	isStart bool,
	placesByID map[string]domain.Place,
) (string, error) {
	if vertexID != "" {
		if _, ok := findVertexByID(graph, vertexID); ok {
			return vertexID, nil
		}
		if isStart {
			return "", ErrStartVertexNotFound
		}
		return "", ErrDestinationVertexNotFound
	}

	if placeID == "" {
		return "", ErrInvalidInput
	}

	place, ok := placesByID[placeID]
	if !ok {
		if uc.placeRepository == nil {
			if isStart {
				return "", ErrStartPlaceNotFound
			}
			return "", ErrDestinationPlaceNotFound
		}

		var err error
		place, err = uc.placeRepository.GetByID(ctx, placeID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				if isStart {
					return "", ErrStartPlaceNotFound
				}
				return "", ErrDestinationPlaceNotFound
			}
			if isStart {
				return "", fmt.Errorf("load start place: %w", err)
			}
			return "", fmt.Errorf("load destination place: %w", err)
		}
	}

	if _, ok := placesByID[place.ID]; !ok {
		placesByID[place.ID] = place
	}

	if vertexID, ok := findVertexIDByPlaceID(graph, place.ID); ok {
		return vertexID, nil
	}

	if isStart {
		return "", ErrStartVertexNotFound
	}

	return "", ErrDestinationWithoutVertex
}

type neighbor struct {
	id     string
	weight float64
}

type priorityNode struct {
	id       string
	distance float64
	index    int
}

type priorityQueue []*priorityNode

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool { return pq[i].distance < pq[j].distance }

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(item any) {
	node := item.(*priorityNode)
	node.index = len(*pq)
	*pq = append(*pq, node)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	length := len(old)
	node := old[length-1]
	*pq = old[:length-1]
	return node
}

func shortestPath(graph domain.NavigationGraph, startVertexID string, destinationVertexID string) ([]string, int, error) {
	verticesByID := make(map[string]domain.GraphVertex, len(graph.Vertices))
	for _, vertex := range graph.Vertices {
		verticesByID[vertex.ID] = vertex
	}

	if _, ok := verticesByID[startVertexID]; !ok {
		return nil, 0, ErrStartVertexNotFound
	}
	if _, ok := verticesByID[destinationVertexID]; !ok {
		return nil, 0, ErrDestinationWithoutVertex
	}
	if startVertexID == destinationVertexID {
		return []string{startVertexID}, 0, nil
	}

	adjacency := make(map[string][]neighbor, len(graph.Vertices))
	for _, edge := range graph.Edges {
		fromVertex, fromOK := verticesByID[edge.From]
		toVertex, toOK := verticesByID[edge.To]
		if !fromOK || !toOK {
			continue
		}

		weight := float64(domain.EstimateDistance(
			domain.Coordinates{
				Building: fromVertex.Building,
				Floor:    fromVertex.Floor,
				X:        fromVertex.X,
				Y:        fromVertex.Y,
			},
			domain.Coordinates{
				Building: toVertex.Building,
				Floor:    toVertex.Floor,
				X:        toVertex.X,
				Y:        toVertex.Y,
			},
		))

		adjacency[edge.From] = append(adjacency[edge.From], neighbor{id: edge.To, weight: weight})
		adjacency[edge.To] = append(adjacency[edge.To], neighbor{id: edge.From, weight: weight})
	}

	distances := make(map[string]float64, len(graph.Vertices))
	previous := make(map[string]string, len(graph.Vertices))
	for _, vertex := range graph.Vertices {
		distances[vertex.ID] = math.Inf(1)
	}
	distances[startVertexID] = 0

	queue := priorityQueue{&priorityNode{id: startVertexID, distance: 0}}
	heap.Init(&queue)

	for queue.Len() > 0 {
		current := heap.Pop(&queue).(*priorityNode)
		if current.distance > distances[current.id] {
			continue
		}
		if current.id == destinationVertexID {
			break
		}

		for _, next := range adjacency[current.id] {
			candidate := current.distance + next.weight
			if candidate >= distances[next.id] {
				continue
			}

			distances[next.id] = candidate
			previous[next.id] = current.id
			heap.Push(&queue, &priorityNode{id: next.id, distance: candidate})
		}
	}

	if math.IsInf(distances[destinationVertexID], 1) {
		return nil, 0, ErrRouteNotFound
	}

	reversed := []string{destinationVertexID}
	for current := destinationVertexID; current != startVertexID; {
		prev, ok := previous[current]
		if !ok {
			return nil, 0, ErrRouteNotFound
		}

		reversed = append(reversed, prev)
		current = prev
	}

	path := make([]string, 0, len(reversed))
	for index := len(reversed) - 1; index >= 0; index-- {
		path = append(path, reversed[index])
	}

	return path, int(math.Round(distances[destinationVertexID])), nil
}

func filterAccessibleGraph(graph domain.NavigationGraph) domain.NavigationGraph {
	vertices := make([]domain.GraphVertex, 0, len(graph.Vertices))
	allowed := make(map[string]struct{}, len(graph.Vertices))

	for _, vertex := range graph.Vertices {
		if !vertex.IsAccessible {
			continue
		}

		vertices = append(vertices, vertex)
		allowed[vertex.ID] = struct{}{}
	}

	edges := make([]domain.GraphEdge, 0, len(graph.Edges))
	for _, edge := range graph.Edges {
		if _, ok := allowed[edge.From]; !ok {
			continue
		}
		if _, ok := allowed[edge.To]; !ok {
			continue
		}

		edges = append(edges, edge)
	}

	return domain.NavigationGraph{
		Vertices: vertices,
		Edges:    edges,
	}
}

func findVertexIDByPlaceID(graph domain.NavigationGraph, placeID string) (string, bool) {
	for _, vertex := range graph.Vertices {
		if vertex.PlaceID == placeID {
			return vertex.ID, true
		}
	}

	return "", false
}

func findVertexByID(graph domain.NavigationGraph, vertexID string) (domain.GraphVertex, bool) {
	for _, vertex := range graph.Vertices {
		if vertex.ID == vertexID {
			return vertex, true
		}
	}

	return domain.GraphVertex{}, false
}

func buildRouteSteps(path []domain.GraphVertex, placesByID map[string]domain.Place) []domain.RouteStep {
	steps := make([]domain.RouteStep, 0, len(path))
	for index, vertex := range path {
		place, hasPlace := placesByID[vertex.PlaceID]
		instruction := buildStepInstruction(path, placesByID, index)
		step := domain.RouteStep{
			Order:       index + 1,
			VertexID:    vertex.ID,
			Instruction: instruction,
			Coordinates: domain.Coordinates{
				Building: vertex.Building,
				Floor:    vertex.Floor,
				X:        vertex.X,
				Y:        vertex.Y,
			},
		}

		if hasPlace {
			step.PlaceID = place.ID
		}

		steps = append(steps, step)
	}

	return steps
}

func buildStepInstruction(path []domain.GraphVertex, placesByID map[string]domain.Place, index int) string {
	vertex := path[index]
	place, hasPlace := placesByID[vertex.PlaceID]
	name := fmt.Sprintf("узел %s", vertex.ID)
	if hasPlace {
		name = place.Name
	}

	switch {
	case index == 0:
		return fmt.Sprintf("Начните маршрут от %s", name)
	case index == len(path)-1:
		return fmt.Sprintf("Следуйте к точке %s", name)
	default:
		return fmt.Sprintf("Продолжайте маршрут через %s", name)
	}
}
