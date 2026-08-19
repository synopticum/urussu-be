package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"urussu-be/internal/auth"
	"urussu-be/internal/domain"
)

// tokenTTL is how long a login token stays valid. Kept as a constant to
// keep the config surface minimal.
const tokenTTL = time.Hour

// dummyPasswordHash is a valid bcrypt hash (cost bcrypt.DefaultCost) of a
// throwaway password. Login compares against it when the email is unknown so
// that case takes as long as a real password check.
var dummyPasswordHash = []byte("$2a$10$ZIwih.KgAX6Aa9jIl.Lvg.bv4zYXYYHmgW85OSuniE4MJnSG3b1G.")

// UsersRepository is the storage contract required by AuthService.
// Defined on the consumer side so the service can be tested with a stub.
type UsersRepository interface {
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Create(ctx context.Context, user domain.User) (domain.User, error)
}

// AuthService implements the registration and login use-cases.
type AuthService struct {
	repo   UsersRepository
	secret string
}

func NewAuthService(repo UsersRepository, secret string) *AuthService {
	return &AuthService{repo: repo, secret: secret}
}

func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName string) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{
		Email:        normalizeEmail(email),
		PasswordHash: string(hash),
		Role:         domain.RoleUser, // registration always creates plain users
		FirstName:    firstName,
		LastName:     lastName,
	}

	return s.repo.Create(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, domain.ErrNotFound) {
		// Unknown email and wrong password must look identical, otherwise a
		// caller could enumerate registered emails. Run bcrypt anyway so the
		// response time is the same in both cases: skipping the comparison
		// would leak registered emails through a timing side channel.
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
		return "", domain.ErrInvalidCredentials
	}
	if err != nil {
		return "", fmt.Errorf("get user %s: %w", email, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	return auth.Generate(s.secret, user.ID, string(user.Role), tokenTTL)
}

// normalizeEmail lowercases and trims the address because the UNIQUE
// constraint on users.email is case-sensitive.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
