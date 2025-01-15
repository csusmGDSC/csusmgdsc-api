package handlers

import (
	"github.com/csusmGDSC/csusmgdsc-api/internal/db"
	"github.com/go-playground/validator"
)

type Handler struct {
	Validate *validator.Validate
	DB       db.DatabaseConnection
}

func NewHandler(dbConn db.DatabaseConnection) *Handler {
	return &Handler{
		Validate: validator.New(),
		DB:       dbConn,
	}
}
