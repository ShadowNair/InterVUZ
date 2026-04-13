package domain

import "math"

func EstimateDistance(from Coordinates, to Coordinates) int {
	if from.Building != to.Building {
		return 300
	}

	dx := from.X - to.X
	dy := from.Y - to.Y
	floorPenalty := math.Abs(float64(from.Floor-to.Floor)) * 25

	return int(math.Round(math.Sqrt(dx*dx+dy*dy) + floorPenalty))
}
