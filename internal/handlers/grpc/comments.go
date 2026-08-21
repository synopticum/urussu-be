package grpc

import (
	"context"
	"log/slog"

	urussuv1 "urussu-be/gen/urussu/v1"
	"urussu-be/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CommentsService is the use-case contract required by CommentsHandler.
// Defined on the consumer side so the handler can be tested with a stub.
type CommentsService interface {
	ListComments(ctx context.Context, entityID string) ([]domain.Comment, error)
}

// CommentsHandler implements urussuv1.CommentsServiceServer.
type CommentsHandler struct {
	urussuv1.UnimplementedCommentsServiceServer

	svc CommentsService
	log *slog.Logger
}

func NewCommentsHandler(svc CommentsService, log *slog.Logger) *CommentsHandler {
	return &CommentsHandler{svc: svc, log: log}
}

func (h *CommentsHandler) ListComments(ctx context.Context, req *urussuv1.ListCommentsRequest) (*urussuv1.ListCommentsResponse, error) {
	if req.GetEntityId() == "" {
		return nil, status.Error(codes.InvalidArgument, "entity_id is required")
	}

	comments, err := h.svc.ListComments(ctx, req.GetEntityId())
	if err != nil {
		h.log.ErrorContext(ctx, "failed to list comments", slog.String("entity_id", req.GetEntityId()), slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to list comments")
	}

	resp := &urussuv1.ListCommentsResponse{Comments: make([]*urussuv1.Comment, 0, len(comments))}
	for _, c := range comments {
		resp.Comments = append(resp.Comments, commentToProto(c))
	}

	return resp, nil
}

func commentToProto(c domain.Comment) *urussuv1.Comment {
	return &urussuv1.Comment{
		Id:         c.ID,
		Name:       c.Name,
		Body:       c.Body,
		CreatedAt:  timestamppb.New(c.CreatedAt),
		ModifiedAt: timestamppb.New(c.ModifiedAt),
	}
}
