package grpc

import (
	"context"
	"errors"
	"log/slog"

	urussuv1 "urussu-be/gen/urussu/v1"
	"urussu-be/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ObjectsService is the use-case contract required by ObjectsHandler.
// Defined on the consumer side so the handler can be tested with a stub.
type ObjectsService interface {
	GetObject(ctx context.Context, id string) (domain.Object, error)
	ListObjects(ctx context.Context, limit int32) ([]domain.Object, error)
}

// ObjectsHandler implements urussuv1.ObjectsServiceServer.
type ObjectsHandler struct {
	urussuv1.UnimplementedObjectsServiceServer

	svc ObjectsService
	log *slog.Logger
}

func NewObjectsHandler(svc ObjectsService, log *slog.Logger) *ObjectsHandler {
	return &ObjectsHandler{svc: svc, log: log}
}

func (h *ObjectsHandler) GetObject(ctx context.Context, req *urussuv1.GetObjectRequest) (*urussuv1.GetObjectResponse, error) {
	o, err := h.svc.GetObject(ctx, req.GetId())

	if errors.Is(err, domain.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "object %s not found", req.GetId())
	}

	if err != nil {
		h.log.ErrorContext(ctx, "failed to get object", slog.String("id", req.GetId()), slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to get object")
	}

	return &urussuv1.GetObjectResponse{Object: objectToProto(o)}, nil
}

func (h *ObjectsHandler) ListObjects(ctx context.Context, req *urussuv1.ListObjectsRequest) (*urussuv1.ListObjectsResponse, error) {
	objects, err := h.svc.ListObjects(ctx, req.GetLimit())

	if err != nil {
		h.log.ErrorContext(ctx, "failed to list objects", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to list objects")
	}

	resp := &urussuv1.ListObjectsResponse{Objects: make([]*urussuv1.Object, 0, len(objects))}
	for _, o := range objects {
		resp.Objects = append(resp.Objects, objectToProto(o))
	}

	return resp, nil
}

func objectToProto(o domain.Object) *urussuv1.Object {
	obj := &urussuv1.Object{
		Id:          o.ID,
		House:       o.House,
		Street:      o.Street,
		Description: o.Description,
		Radius:      o.Radius,
		Coordinates: make([]*urussuv1.Point, 0, len(o.Coordinates)),
	}

	for _, c := range o.Coordinates {
		if len(c) != 2 {
			continue
		}
		obj.Coordinates = append(obj.Coordinates, &urussuv1.Point{Latitude: c[0], Longitude: c[1]})
	}

	return obj
}
