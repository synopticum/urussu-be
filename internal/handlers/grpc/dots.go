// Package grpc contains the gRPC service implementations.
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

// DotsService is the use-case contract required by DotsHandler.
// Defined on the consumer side so the handler can be tested with a stub.
type DotsService interface {
	GetDot(ctx context.Context, id string) (domain.Dot, error)
	ListDots(ctx context.Context, limit int32, layer *int32) ([]domain.Dot, error)
}

// DotsHandler implements urussuv1.DotsServiceServer.
type DotsHandler struct {
	urussuv1.UnimplementedDotsServiceServer

	svc DotsService
	log *slog.Logger
}

func NewDotsHandler(svc DotsService, log *slog.Logger) *DotsHandler {
	return &DotsHandler{svc: svc, log: log}
}

func (h *DotsHandler) GetDot(ctx context.Context, req *urussuv1.GetDotRequest) (*urussuv1.GetDotResponse, error) {
	d, err := h.svc.GetDot(ctx, req.GetId())

	if errors.Is(err, domain.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "dot %s not found", req.GetId())
	}

	if err != nil {
		h.log.ErrorContext(ctx, "failed to get dot", slog.String("id", req.GetId()), slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to get dot")
	}

	return &urussuv1.GetDotResponse{Dot: dotToProto(d)}, nil
}

func (h *DotsHandler) ListDots(ctx context.Context, req *urussuv1.ListDotsRequest) (*urussuv1.ListDotsResponse, error) {
	dots, err := h.svc.ListDots(ctx, req.GetLimit(), req.Layer)

	if err != nil {
		h.log.ErrorContext(ctx, "failed to list dots", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to list dots")
	}

	resp := &urussuv1.ListDotsResponse{Dots: make([]*urussuv1.Dot, 0, len(dots))}
	for _, d := range dots {
		resp.Dots = append(resp.Dots, dotToProto(d))
	}

	return resp, nil
}

func dotToProto(d domain.Dot) *urussuv1.Dot {
	return &urussuv1.Dot{
		Id:               d.ID,
		Title:            d.Title,
		ShortDescription: d.ShortDescription,
		Layer:            d.Layer,
	}
}
