package grpc

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	urussuv1 "urussu-be/gen/urussu/v1"
	"urussu-be/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const minPasswordLength = 8

// AuthService is the use-case contract required by AuthHandler.
// Defined on the consumer side so the handler can be tested with a stub.
type AuthService interface {
	Register(ctx context.Context, email, password, firstName, lastName string) (domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

// AuthHandler implements urussuv1.AuthServiceServer.
type AuthHandler struct {
	urussuv1.UnimplementedAuthServiceServer

	svc AuthService
	log *slog.Logger
}

func NewAuthHandler(svc AuthService, log *slog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

func (h *AuthHandler) Register(ctx context.Context, req *urussuv1.RegisterRequest) (*urussuv1.RegisterResponse, error) {
	if strings.TrimSpace(req.GetEmail()) == "" || len(req.GetPassword()) < minPasswordLength {
		return nil, status.Errorf(codes.InvalidArgument, "email must not be empty and password must be at least %d characters", minPasswordLength)
	}

	user, err := h.svc.Register(ctx, req.GetEmail(), req.GetPassword(), req.GetFirstName(), req.GetLastName())

	if errors.Is(err, domain.ErrAlreadyExists) {
		return nil, status.Error(codes.AlreadyExists, "email already registered")
	}

	if err != nil {
		h.log.ErrorContext(ctx, "failed to register user", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &urussuv1.RegisterResponse{User: userToProto(user)}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *urussuv1.LoginRequest) (*urussuv1.LoginResponse, error) {
	if strings.TrimSpace(req.GetEmail()) == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password must not be empty")
	}

	token, err := h.svc.Login(ctx, req.GetEmail(), req.GetPassword())

	if errors.Is(err, domain.ErrInvalidCredentials) {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	if err != nil {
		h.log.ErrorContext(ctx, "failed to login", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &urussuv1.LoginResponse{Token: token}, nil
}

// userToProto deliberately drops the password hash.
func userToProto(u domain.User) *urussuv1.User {
	return &urussuv1.User{
		Id:        u.ID,
		Email:     u.Email,
		Role:      string(u.Role),
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}
