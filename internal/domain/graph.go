package domain

type GraphNode struct {
	params     Place
	neighboors []GraphNode
}

func (n *GraphNode) FindShortestWay(idFrom, idTo string) ([]Place, error) {
	fromNode, err := n.findNodeById(idFrom)
	if err != nil {
		return nil, err
	}

	if _, err = n.findNodeById(idTo); err != nil {
		return nil, err
	}

	queue := []*GraphNode{fromNode}
	visited := map[string]bool{fromNode.params.ID: true}
	parents := make(map[string]string)
	placesByID := map[string]Place{
		fromNode.params.ID: fromNode.params,
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.params.ID == idTo {
			break
		}

		for i := range current.neighboors {
			neighbor := &current.neighboors[i]
			neighborID := neighbor.params.ID
			if visited[neighborID] {
				continue
			}

			visited[neighborID] = true
			parents[neighborID] = current.params.ID
			placesByID[neighborID] = neighbor.params
			queue = append(queue, neighbor)
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
		path = append(path, placesByID[pathIDs[i]])
	}

	return path, nil
}

func (n *GraphNode) findNodeById(id string) (*GraphNode, error) {
	queue := []*GraphNode{n}
	visited := map[string]bool{
		n.params.ID: true,
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.params.ID == id {
			return current, nil
		}

		for i := range current.neighboors {
			neighbor := &current.neighboors[i]
			neighborID := neighbor.params.ID
			if visited[neighborID] {
				continue
			}

			visited[neighborID] = true
			queue = append(queue, neighbor)
		}
	}

	return nil, ErrNotFound
}
