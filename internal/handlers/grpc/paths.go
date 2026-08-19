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

// PathsService is the use-case contract required by PathsHandler.
// Defined on the consumer side so the handler can be tested with a stub.
type PathsService interface {
	GetPath(ctx context.Context, id string) (domain.Path, error)
	ListPaths(ctx context.Context, limit int32) ([]domain.Path, error)
}

// PathsHandler implements urussuv1.PathsServiceServer.
type PathsHandler struct {
	urussuv1.UnimplementedPathsServiceServer

	svc PathsService
	log *slog.Logger
}

func NewPathsHandler(svc PathsService, log *slog.Logger) *PathsHandler {
	return &PathsHandler{svc: svc, log: log}
}

func (h *PathsHandler) GetPath(ctx context.Context, req *urussuv1.GetPathRequest) (*urussuv1.GetPathResponse, error) {
	p, err := h.svc.GetPath(ctx, req.GetId())

	if errors.Is(err, domain.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "path %s not found", req.GetId())
	}

	if err != nil {
		h.log.ErrorContext(ctx, "failed to get path", slog.String("id", req.GetId()), slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to get path")
	}

	return &urussuv1.GetPathResponse{Path: pathToProto(p)}, nil
}

func (h *PathsHandler) ListPaths(ctx context.Context, req *urussuv1.ListPathsRequest) (*urussuv1.ListPathsResponse, error) {
	paths, err := h.svc.ListPaths(ctx, req.GetLimit())

	if err != nil {
		h.log.ErrorContext(ctx, "failed to list paths", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to list paths")
	}

	resp := &urussuv1.ListPathsResponse{Paths: make([]*urussuv1.Path, 0, len(paths))}
	for _, p := range paths {
		resp.Paths = append(resp.Paths, pathToProto(p))
	}

	return resp, nil
}

func pathToProto(p domain.Path) *urussuv1.Path {
	path := &urussuv1.Path{
		Id:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Coordinates: make([]*urussuv1.Point, 0, len(p.Coordinates)),
	}

	for _, c := range p.Coordinates {
		if len(c) != 2 {
			continue
		}
		path.Coordinates = append(path.Coordinates, &urussuv1.Point{Latitude: c[0], Longitude: c[1]})
	}

	return path
}
