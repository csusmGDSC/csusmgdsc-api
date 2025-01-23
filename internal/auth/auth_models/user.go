package auth_models

import (
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
)

type CreateUserRequest struct {
	Email    string  `json:"email" validate:"required,email"`
	Password *string `json:"password,omitempty"`
	Provider *string `json:"provider,omitempty"`
	AuthID   *string `json:"auth_id,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type UpdateUserRequest struct {
	FirstName      *string              `json:"first_name,omitempty"`
	LastName       *string              `json:"last_name,omitempty"`
	Image          *string              `json:"image,omitempty"`
	Position       *models.GDSCPosition `json:"position,omitempty"`
	Branch         *models.GDSCBranch   `json:"branch,omitempty"`
	Github         *string              `json:"github,omitempty"`
	Linkedin       *string              `json:"linkedin,omitempty"`
	Instagram      *string              `json:"instagram,omitempty"`
	Discord        *string              `json:"discord,omitempty"`
	Bio            *string              `json:"bio,omitempty"`
	Tags           []string             `json:"tags,omitempty"`
	Website        *string              `json:"website,omitempty"`
	GraduationDate *time.Time           `json:"graduation_date,omitempty"`
}
