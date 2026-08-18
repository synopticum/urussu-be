// Package handler contains the gRPC service implementations.
package handler

import (
	"context"
	"log/slog"

	urussuv1 "urussu-be/gen/urussu-be/v1"
	"urussu-be/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DotsService is the use-case contract required by DotsHandler.
// Defined on the consumer side so the handler can be tested with a stub.
type DotsService interface {
	ListDots(ctx context.Context, limit int32) ([]domain.Dot, error)
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

func (h *DotsHandler) ListDots(ctx context.Context, req *urussuv1.ListDotsRequest) (*urussuv1.ListDotsResponse, error) {
	dots, err := h.svc.ListDots(ctx, req.GetLimit())
	if err != nil {
		h.log.ErrorContext(ctx, "failed to list dots", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to list dots")
	}

	resp := &urussuv1.ListDotsResponse{Dots: make([]*urussuv1.Dot, 0, len(dots))}
	for _, d := range dots {
		resp.Dots = append(resp.Dots, &urussuv1.Dot{
			Id:               d.ID,
			Title:            d.Title,
			ShortDescription: d.ShortDescription,
			Layer:            d.Layer,
		})
	}

	return resp, nil
}
