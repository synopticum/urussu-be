package grpc

import (
	urussuv1 "urussu-be/gen/urussu/v1"
	"urussu-be/internal/domain"
)

// imagesToProto maps domain images shared by dots, objects and paths.
func imagesToProto(images []domain.Image) []*urussuv1.Image {
	out := make([]*urussuv1.Image, 0, len(images))
	for _, img := range images {
		out = append(out, &urussuv1.Image{Id: img.ID, Year: int32(img.Year)})
	}
	return out
}
