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
	FullName       string         `json:"full_name" db:"full_name" validate:"required"`
	FirstName      string         `json:"first_name" db:"first_name" validate:"required"`
	LastName       string         `json:"last_name" db:"last_name" validate:"required"`
	Email          string         `json:"email" db:"email" validate:"required,email"`
	Password       string         `json:"-" db:"password"`
	Image          *string        `json:"image,omitempty" db:"image"`
	Role           Role           `json:"role" db:"role" validate:"required,oneof=USER ADMIN"`
	Position       GDSCPosition   `json:"position" db:"position" validate:"required"`
	Branch         GDSCBranch     `json:"branch" db:"branch" validate:"required"`
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
}

type CreateUserRequest struct {
	FullName       string       `json:"full_name,omitempty"`
	FirstName      string       `json:"first_name" validate:"required"`
	LastName       string       `json:"last_name" validate:"required"`
	Email          string       `json:"email" validate:"required,email"`
	Password       string       `json:"password" validate:"required,min=8"`
	Role           Role         `json:"role,omitempty" validate:"required,oneof=USER ADMIN"`
	Image          *string      `json:"image,omitempty"`
	Position       GDSCPosition `json:"position" validate:"required"`
	Branch         GDSCBranch   `json:"branch" validate:"required"`
	Github         *string      `json:"github,omitempty"`
	Linkedin       *string      `json:"linkedin,omitempty"`
	Instagram      *string      `json:"instagram,omitempty"`
	Discord        *string      `json:"discord,omitempty"`
	Bio            *string      `json:"bio,omitempty"`
	Tags           []string     `json:"tags,omitempty"`
	Website        *string      `json:"website,omitempty"`
	GraduationDate *time.Time   `json:"graduation_date,omitempty" validate:"required"`
}

type UpdateUserRequest struct {
	FirstName      *string       `json:"first_name,omitempty"`
	LastName       *string       `json:"last_name,omitempty"`
	Image          *string       `json:"image,omitempty"`
	Position       *GDSCPosition `json:"position,omitempty"`
	Branch         *GDSCBranch   `json:"branch,omitempty"`
	Github         *string       `json:"github,omitempty"`
	Linkedin       *string       `json:"linkedin,omitempty"`
	Instagram      *string       `json:"instagram,omitempty"`
	Discord        *string       `json:"discord,omitempty"`
	Bio            *string       `json:"bio,omitempty"`
	Tags           []string      `json:"tags,omitempty"`
	Website        *string       `json:"website,omitempty"`
	GraduationDate *time.Time    `json:"graduation_date,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
