// Package domain contains the core domain models shared across layers.
package domain

// Dot is a map object stored in the `dots` table.
type Dot struct {
	ID               string
	Title            string
	ShortDescription string
	Layer            string
}
