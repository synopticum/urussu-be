package domain

// Object is a map object stored in the `objects` table.
type Object struct {
	ID          string
	House       string
	Street      string
	Description string
	// Radius is nil when the object has no radius.
	Radius *float32
	// Coordinates is the polygon outline as [latitude, longitude] pairs.
	Coordinates [][]float64
}
