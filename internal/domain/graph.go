package domain

import "math"

type Graph struct {
	nodes map[string]*GraphNode
}

type GraphNode struct {
	params     Place
	neighboors map[string]*GraphNode
}

func NewPlaceGraph(places []Place) *Graph {
	graph := &Graph{
		nodes: make(map[string]*GraphNode, len(places)),
	}

	for _, place := range places {
		graph.nodes[place.ID] = &GraphNode{
			params:     place,
			neighboors: make(map[string]*GraphNode),
		}
	}

	return graph
}

func (g *Graph) ConnectAllNodes() error {
	if g == nil {
		return ErrNotFound
	}

	nodeIDs := make([]string, 0, len(g.nodes))
	for nodeID := range g.nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}

	for i := 0; i < len(nodeIDs); i++ {
		for j := i + 1; j < len(nodeIDs); j++ {
			if err := g.AddEdgeByID(nodeIDs[i], nodeIDs[j]); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Graph) AddEdgeByID(fromID, toID string) error {
	if g == nil {
		return ErrNotFound
	}

	fromNode, err := g.FindNodeById(fromID)
	if err != nil {
		return err
	}

	toNode, err := g.FindNodeById(toID)
	if err != nil {
		return err
	}

	fromNode.neighboors[toID] = toNode
	toNode.neighboors[fromID] = fromNode

	return nil
}

func (g *Graph) FindShortestWay(idFrom, idTo string) ([]Place, error) {
	if g == nil {
		return nil, ErrNotFound
	}

	fromNode, err := g.FindNodeById(idFrom)
	if err != nil {
		return nil, err
	}

	if _, err := g.FindNodeById(idTo); err != nil {
		return nil, err
	}

	if idFrom == idTo {
		return []Place{fromNode.params}, nil
	}

	queue := []string{idFrom}
	visited := map[string]bool{idFrom: true}
	parents := make(map[string]string, len(g.nodes))

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if currentID == idTo {
			break
		}

		currentNode := g.nodes[currentID]
		for neighborID := range currentNode.neighboors {
			if visited[neighborID] {
				continue
			}

			visited[neighborID] = true
			parents[neighborID] = currentID
			queue = append(queue, neighborID)
		}
	}

	if !visited[idTo] {
		return nil, ErrNotFound
	}

	pathIDs := []string{idTo}
	for currentID := idTo; currentID != idFrom; {
		parentID, ok := parents[currentID]
		if !ok {
			return nil, ErrNotFound
		}

		pathIDs = append(pathIDs, parentID)
		currentID = parentID
	}

	path := make([]Place, 0, len(pathIDs))
	for i := len(pathIDs) - 1; i >= 0; i-- {
		path = append(path, g.nodes[pathIDs[i]].params)
	}

	return path, nil
}

func (g *Graph) FindNodeById(id string) (*GraphNode, error) {
	if g == nil {
		return nil, ErrNotFound
	}

	node, ok := g.nodes[id]
	if !ok {
		return nil, ErrNotFound
	}

	return node, nil
}

func EstimateDistance(from Coordinates, to Coordinates) int {
	if from.Building != to.Building {
		return 300
	}

	dx := from.X - to.X
	dy := from.Y - to.Y
	floorPenalty := math.Abs(float64(from.Floor-to.Floor)) * 25

	return int(math.Round(math.Sqrt(dx*dx+dy*dy) + floorPenalty))
}
