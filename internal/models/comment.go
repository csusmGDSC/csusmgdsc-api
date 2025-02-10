package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID  `json:"id,omitempty" db:"id"`
	UserId    uuid.UUID  `json:"user_id" db:"user_id"`
	EventId   uuid.UUID  `json:"event_id" db:"event_id"`
	Content   string     `json:"content" db:"content"`
	PinnedBy  *uuid.UUID `json:"pinned_by,omitempty" db:"pinned_by"`
	CreatedAt time.Time  `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at,omitempty" db:"updated_at"`
	ParentId  *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Replies   []*Comment `json:"replies,omitempty"`
}

type UpdateCommentRquest struct {
	ID       *uuid.UUID `json:"id,omitempty" db:"id"`
	Content  *string    `json:"content,omitempty" db:"content"`
	PinnedBy *uuid.UUID `json:"pinned_by,omitempty" db:"pinned_by"`
	ParentId *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
}
