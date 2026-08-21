package domain

import "time"

// CommentEntityType identifies the kind of entity a comment belongs to.
// The values are DB-column-agnostic; the repository maps them to the
// matching FK column.
type CommentEntityType string

const (
	CommentEntityDot    CommentEntityType = "dot"
	CommentEntityObject CommentEntityType = "object"
	CommentEntityPath   CommentEntityType = "path"
)

// Comment is a user comment stored in the `comments` table, attached to
// exactly one entity (dot, object or path).
type Comment struct {
	ID         string
	Name       string
	Body       string
	CreatedAt  time.Time
	ModifiedAt time.Time
}
