package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Role string

const (
	UserRole  Role = "USER"
	AdminRole Role = "ADMIN"
)

func (r Role) String() string {
	return string(r)
}

type User struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	FullName       *string        `json:"full_name" db:"full_name"`
	FirstName      *string        `json:"first_name" db:"first_name"`
	LastName       *string        `json:"last_name" db:"last_name"`
	Email          string         `json:"email" db:"email" validate:"required,email"`
	Password       *string        `json:"-" db:"password"`
	Image          *string        `json:"image,omitempty" db:"image"`
	Role           *Role          `json:"role" db:"role"`
	Position       *GDSCPosition  `json:"position" db:"position"`
	Branch         *GDSCBranch    `json:"branch" db:"branch"`
	Github         *string        `json:"github,omitempty" db:"github"`
	Linkedin       *string        `json:"linkedin,omitempty" db:"linkedin"`
	Instagram      *string        `json:"instagram,omitempty" db:"instagram"`
	Discord        *string        `json:"discord,omitempty" db:"discord"`
	Bio            *string        `json:"bio,omitempty" db:"bio"`
	Tags           pq.StringArray `json:"tags,omitempty" db:"tags" gorm:"type:text[]"`
	Website        *string        `json:"website,omitempty" db:"website"`
	GraduationDate *time.Time     `json:"graduation_date,omitempty" db:"graduation_date"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
	Provider       *string        `json:"provider" db:"provider"`
	AuthID         *string        `json:"auth_id" db:"auth_id"`
	IsOnboarded    bool           `json:"is_onboarded" db:"is_onboarded"`
}
