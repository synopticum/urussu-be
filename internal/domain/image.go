package domain

// Image is a photo of a map entity (dot, object or path) stored in the
// `images` table, linked by (entity_type, entity_id).
type Image struct {
	ID   string `json:"id"`
	Year int16  `json:"year"`
}
