package postgres

import (
	"encoding/json"
	"fmt"

	"urussu-be/internal/domain"
)

// decodeImages unmarshals the JSON array produced by the images subquery in
// the dots/objects/paths queries. A NULL result (entity has no images)
// arrives from pgx as a nil slice.
func decodeImages(data []byte) ([]domain.Image, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var images []domain.Image
	if err := json.Unmarshal(data, &images); err != nil {
		return nil, fmt.Errorf("decode images: %w", err)
	}

	return images, nil
}
