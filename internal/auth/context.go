package auth

import (
	"context"

	"urussu-be/internal/domain"
)

// contextKey is an unexported type so the keys cannot collide with values
// set by other packages.
type contextKey int

const (
	userIDKey contextKey = iota
	roleKey
)

// ContextWithUser returns a context carrying the authenticated user's ID
// and role (populated by Middleware for every authorized request).
func ContextWithUser(ctx context.Context, userID, role string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	return context.WithValue(ctx, roleKey, role)
}

// UserIDFrom returns the authenticated user's ID, or "" when the request
// was not authenticated.
func UserIDFrom(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey).(string)
	return userID
}

// RoleFrom returns the authenticated user's role, or "" when the request
// was not authenticated.
func RoleFrom(ctx context.Context) string {
	role, _ := ctx.Value(roleKey).(string)
	return role
}

// IsAdmin reports whether the authenticated user has the admin role.
func IsAdmin(ctx context.Context) bool {
	return RoleFrom(ctx) == string(domain.RoleAdmin)
}
