// Package handler contains the gRPC service implementations.
package handler

import (
	"context"
	"log/slog"

	urussuv1 "urussu-be/gen/urussu-be/v1"
	"urussu-be/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DotsHandler implements urussuv1.DotsServiceServer.
type DotsHandler struct {
	urussuv1.UnimplementedDotsServiceServer

	svc *service.DotsService
	log *slog.Logger
}

func NewDotsHandler(svc *service.DotsService, log *slog.Logger) *DotsHandler {
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
