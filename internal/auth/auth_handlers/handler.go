package auth_handlers

import (
	"github.com/csusmGDSC/csusmgdsc-api/internal/db"
	"github.com/go-playground/validator"
)

type OAuthHandler struct {
	Validate *validator.Validate
	DB       db.DatabaseConnection
}

func NewOAuthHandler(dbConn db.DatabaseConnection) *OAuthHandler {
	return &OAuthHandler{
		Validate: validator.New(),
		DB:       dbConn,
	}
}
