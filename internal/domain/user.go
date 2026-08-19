package domain

import "errors"

// ErrAlreadyExists is returned when an object conflicts with an existing
// one (e.g. a user with the same email).
var ErrAlreadyExists = errors.New("already exists")

// ErrInvalidCredentials is returned when authentication fails because the
// email is unknown or the password does not match. Both cases share one
// error so a login attempt cannot probe which emails are registered.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Role is a user's access level, stored in the `users.role` column.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// User is an account stored in the `users` table.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         Role
	FirstName    string
	LastName     string
}
