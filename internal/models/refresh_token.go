package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        int       `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	IssuedAt  time.Time `json:"issued_at" db:"issued_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IPAddress string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent string    `json:"user_agent,omitempty" db:"user_agent"`
}

type CreateSessionRequest struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	IssuedAt  time.Time `json:"issued_at" db:"issued_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IPAddress string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent string    `json:"user_agent,omitempty" db:"user_agent"`
}
