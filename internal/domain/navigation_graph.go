package domain

type GraphVertex struct {
	ID           string  `json:"id"`
	PlaceID      string  `json:"placeId,omitempty"`
	Building     string  `json:"building"`
	Floor        int     `json:"floor"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	IsAccessible bool    `json:"isAccessible"`
}

type GraphEdge struct {
	ID   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
}

type NavigationGraph struct {
	Vertices []GraphVertex `json:"vertices"`
	Edges    []GraphEdge   `json:"edges"`
}
