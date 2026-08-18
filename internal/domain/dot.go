// Package domain contains the core domain models shared across layers.
package domain

import "errors"

// ErrNotFound is returned when the requested object does not exist.
var ErrNotFound = errors.New("not found")

// Dot is a map object stored in the `dots` table.
type Dot struct {
	ID               string
	Title            string
	ShortDescription string
	Layer            string
}
