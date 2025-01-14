package handlers

import (
	"github.com/csusmGDSC/csusmgdsc-api/config"
	"github.com/go-playground/validator"
)

type Handler struct {
	Config   *config.Config
	Validate *validator.Validate
}

func NewHandler() *Handler {
	return &Handler{
		Config:   config.LoadConfig(),
		Validate: validator.New(),
	}
}
