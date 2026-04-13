package bootstrap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

type parsedVertex struct {
	id            string
	placeID       string
	building      string
	hasBuilding   bool
	floor         int
	hasFloor      bool
	x             float64
	hasX          bool
	y             float64
	hasY          bool
	isAccessible  bool
	hasAccessible bool
}

type parsedEdge struct {
	id   string
	from string
	to   string
}

func LoadNavigationGraphFromSVG(path string, places []domain.Place) (domain.NavigationGraph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.NavigationGraph{}, fmt.Errorf("read svg graph %s: %w", filepath.Clean(path), err)
	}

	graph, err := ParseNavigationGraphSVG(data, places)
	if err != nil {
		return domain.NavigationGraph{}, fmt.Errorf("parse svg graph %s: %w", filepath.Clean(path), err)
	}

	return graph, nil
}

func ParseNavigationGraphSVG(data []byte, places []domain.Place) (domain.NavigationGraph, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	placesByID := make(map[string]domain.Place, len(places))
	for _, place := range places {
		placesByID[place.ID] = place
	}

	vertexByID := make(map[string]parsedVertex)
	vertexOrder := make([]string, 0)
	edges := make([]parsedEdge, 0)
	edgeCounter := 1

	for {
		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return domain.NavigationGraph{}, err
		}

		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		attributes := attrsByName(start.Attr)

		vertexID := firstNonEmpty(attributes["data-vertex-id"], attributes["vertex-id"])
		if vertexID != "" {
			vertex, err := parseVertex(attributes, placesByID)
			if err != nil {
				return domain.NavigationGraph{}, err
			}
			if _, exists := vertexByID[vertex.id]; exists {
				return domain.NavigationGraph{}, fmt.Errorf("duplicate vertex id %q", vertex.id)
			}

			vertexByID[vertex.id] = vertex
			vertexOrder = append(vertexOrder, vertex.id)
		}

		from := firstNonEmpty(attributes["data-from"], attributes["from"])
		to := firstNonEmpty(attributes["data-to"], attributes["to"])
		if from == "" || to == "" {
			continue
		}

		edgeID := firstNonEmpty(attributes["data-edge-id"], attributes["edge-id"], attributes["id"])
		if edgeID == "" {
			edgeID = fmt.Sprintf("edge-%d", edgeCounter)
		}
		edgeCounter++

		edges = append(edges, parsedEdge{
			id:   edgeID,
			from: strings.TrimSpace(from),
			to:   strings.TrimSpace(to),
		})
	}

	if len(vertexOrder) == 0 {
		return domain.NavigationGraph{}, fmt.Errorf("no vertices found in svg")
	}

	vertices := make([]domain.GraphVertex, 0, len(vertexOrder))
	for _, vertexID := range vertexOrder {
		vertex := vertexByID[vertexID]
		vertices = append(vertices, domain.GraphVertex{
			ID:           vertex.id,
			PlaceID:      vertex.placeID,
			Building:     vertex.building,
			Floor:        vertex.floor,
			X:            vertex.x,
			Y:            vertex.y,
			IsAccessible: vertex.isAccessible,
		})
	}

	graphEdges := make([]domain.GraphEdge, 0, len(edges))
	seenEdgeIDs := make(map[string]struct{}, len(edges))
	for _, edge := range edges {
		if edge.from == "" || edge.to == "" {
			return domain.NavigationGraph{}, fmt.Errorf("edge %q must have from and to vertices", edge.id)
		}
		if _, ok := vertexByID[edge.from]; !ok {
			return domain.NavigationGraph{}, fmt.Errorf("edge %q references unknown from vertex %q", edge.id, edge.from)
		}
		if _, ok := vertexByID[edge.to]; !ok {
			return domain.NavigationGraph{}, fmt.Errorf("edge %q references unknown to vertex %q", edge.id, edge.to)
		}
		if _, exists := seenEdgeIDs[edge.id]; exists {
			return domain.NavigationGraph{}, fmt.Errorf("duplicate edge id %q", edge.id)
		}
		seenEdgeIDs[edge.id] = struct{}{}

		graphEdges = append(graphEdges, domain.GraphEdge{
			ID:   edge.id,
			From: edge.from,
			To:   edge.to,
		})
	}

	return domain.NavigationGraph{
		Vertices: vertices,
		Edges:    graphEdges,
	}, nil
}

func parseVertex(attributes map[string]string, placesByID map[string]domain.Place) (parsedVertex, error) {
	vertex := parsedVertex{
		id: strings.TrimSpace(firstNonEmpty(attributes["data-vertex-id"], attributes["vertex-id"])),
	}
	if vertex.id == "" {
		return parsedVertex{}, fmt.Errorf("vertex without id")
	}

	vertex.placeID = strings.TrimSpace(firstNonEmpty(attributes["data-place-id"], attributes["place-id"]))
	if vertex.placeID == "" {
		if _, ok := placesByID[vertex.id]; ok {
			vertex.placeID = vertex.id
		}
	}

	if building := strings.TrimSpace(firstNonEmpty(attributes["data-building"], attributes["building"])); building != "" {
		vertex.building = building
		vertex.hasBuilding = true
	}

	if floorValue := strings.TrimSpace(firstNonEmpty(attributes["data-floor"], attributes["floor"])); floorValue != "" {
		floor, err := strconv.Atoi(floorValue)
		if err != nil {
			return parsedVertex{}, fmt.Errorf("vertex %q has invalid floor %q", vertex.id, floorValue)
		}
		vertex.floor = floor
		vertex.hasFloor = true
	}

	if x, ok, err := parseCoordinate(attributes, "cx", "x", "data-x"); err != nil {
		return parsedVertex{}, fmt.Errorf("vertex %q has invalid x coordinate: %w", vertex.id, err)
	} else if ok {
		vertex.x = x
		vertex.hasX = true
	}

	if y, ok, err := parseCoordinate(attributes, "cy", "y", "data-y"); err != nil {
		return parsedVertex{}, fmt.Errorf("vertex %q has invalid y coordinate: %w", vertex.id, err)
	} else if ok {
		vertex.y = y
		vertex.hasY = true
	}

	if rawAccessible := strings.TrimSpace(firstNonEmpty(attributes["data-accessible"], attributes["accessible"])); rawAccessible != "" {
		accessible, err := strconv.ParseBool(rawAccessible)
		if err != nil {
			return parsedVertex{}, fmt.Errorf("vertex %q has invalid accessible flag %q", vertex.id, rawAccessible)
		}
		vertex.isAccessible = accessible
		vertex.hasAccessible = true
	}

	if vertex.placeID != "" {
		place, ok := placesByID[vertex.placeID]
		if !ok {
			return parsedVertex{}, fmt.Errorf("vertex %q references unknown place %q", vertex.id, vertex.placeID)
		}

		if !vertex.hasBuilding {
			vertex.building = place.Coordinates.Building
			vertex.hasBuilding = true
		}
		if !vertex.hasFloor {
			vertex.floor = place.Coordinates.Floor
			vertex.hasFloor = true
		}
		if !vertex.hasX {
			vertex.x = place.Coordinates.X
			vertex.hasX = true
		}
		if !vertex.hasY {
			vertex.y = place.Coordinates.Y
			vertex.hasY = true
		}
		if !vertex.hasAccessible {
			vertex.isAccessible = place.IsAccessible
			vertex.hasAccessible = true
		}
	}

	if !vertex.hasBuilding {
		return parsedVertex{}, fmt.Errorf("vertex %q must have building or place binding", vertex.id)
	}
	if !vertex.hasFloor {
		return parsedVertex{}, fmt.Errorf("vertex %q must have floor or place binding", vertex.id)
	}
	if !vertex.hasX || !vertex.hasY {
		return parsedVertex{}, fmt.Errorf("vertex %q must have coordinates", vertex.id)
	}
	if !vertex.hasAccessible {
		vertex.isAccessible = true
		vertex.hasAccessible = true
	}

	return vertex, nil
}

func parseCoordinate(attributes map[string]string, keys ...string) (float64, bool, error) {
	for _, key := range keys {
		value := strings.TrimSpace(attributes[key])
		if value == "" {
			continue
		}

		coordinate, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, false, err
		}

		return coordinate, true, nil
	}

	return 0, false, nil
}

func attrsByName(attributes []xml.Attr) map[string]string {
	values := make(map[string]string, len(attributes))
	for _, attribute := range attributes {
		values[attribute.Name.Local] = attribute.Value
	}

	return values
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
}
