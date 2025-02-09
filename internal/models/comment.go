package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID         uuid.UUID   `json:"id,omitempty" db:"id"`
	UserId     uuid.UUID   `json:"userId" db:"userid"`
	EventId    uuid.UUID   `json:"eventId" db:"eventid"`
	Content    string      `json:"content" db:"content"`
	PinnedBy   *uuid.UUID  `json:"pinnedBy,omitempty" db:"pinnedby"`
	CommentIds []uuid.UUID `json:"commentIds,omitempty" db:"commentids"`
	CreatedAt  time.Time   `json:"createdAt,omitempty" db:"createdat"`
	UpdatedAt  time.Time   `json:"updatedAt,omitempty" db:"updatedat"`
	Likes      []uuid.UUID `json:"likes,omitempty" db:"likes"`
	Dislikes   []uuid.UUID `json:"dislikes,omitempty" db:"dislikes"`
}

type UpdateCommentRquest struct {
	ID         *uuid.UUID  `json:"id,omitempty" db:"id"`
	Content    *string     `json:"content,omitempty" db:"content"`
	PinnedBy   *uuid.UUID  `json:"pinnedBy,omitempty" db:"pinnedby"`
	CommentIds []uuid.UUID `json:"commentIds,omitempty" db:"commentids"`
	Likes      []uuid.UUID `json:"likes,omitempty" db:"likes"`
	Dislikes   []uuid.UUID `json:"dislikes,omitempty" db:"dislikes"`
}
