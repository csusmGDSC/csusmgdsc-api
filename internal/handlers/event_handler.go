package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/db"
	"github.com/csusmGDSC/csusmgdsc-api/internal/db/repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func InsertEventHandler(c echo.Context) error {
	var event models.Event
	// Bind JSON request body to the Event struct
	if err := c.Bind(&event); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	// Validate the struct
	if err := validate.Struct(event); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, err.Field()+" "+err.Tag())
		}

		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}

	event.ID = uuid.New()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	dbConnection, err := db.ConnectDB()
	if err != nil {
		log.Println("Database connection error: ", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to the database"})
	}
	defer dbConnection.Close()

	if err := repositories.InsertEvent(dbConnection, event); err != nil {
		log.Println("Insert event error:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to insert event"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Event created successfully", "eventID": event.ID.String()})
}
