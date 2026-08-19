package domain

// Path is a map path stored in the `paths` table.
type Path struct {
	ID          string
	Title       string
	Description string
	// Coordinates is the polyline as [latitude, longitude] pairs.
	Coordinates [][]float64
}
