package domain

import "time"

// Comment is a user comment stored in the `comments` table, attached to
// exactly one entity (dot, object or path).
type Comment struct {
	ID         string
	Name       string
	Body       string
	CreatedAt  time.Time
	ModifiedAt time.Time
}
