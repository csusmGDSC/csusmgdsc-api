package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/db/repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) InsertEventHandler(c echo.Context) error {
	userRole, ok := c.Get("user_role").(string)
	if !ok || userRole != "ADMIN" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Insufficient permissions"})
	}

	var event models.Event
	// Bind JSON request body to the Event struct
	if err := c.Bind(&event); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	// Validate the struct
	if err := h.Validate.Struct(event); err != nil {
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

	dbConn := h.DB.GetDB()
	eventRepo := repositories.NewEventRepository(dbConn)

	if err := eventRepo.InsertEvent(dbConn, event); err != nil {
		log.Println("Insert event error:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to insert event"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Event created successfully", "eventID": event.ID.String()})
}

func (h *Handler) GetEventsHandler(c echo.Context) error {
	// Get pagination parameters from query
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	if page == "" {
		page = "1"
	}
	if limit == "" {
		limit = "10"
	}

	dbConn := h.DB.GetDB()
	eventRepo := repositories.NewEventRepository(dbConn)

	response, err := eventRepo.GetAll(page, limit)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "No events found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get events: " + err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) GetEventByIDHandler(c echo.Context) error {
	eventIDStr := c.Param("id")
	if eventIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event ID is required"})
	}

	dbConn := h.DB.GetDB()
	eventRepo := repositories.NewEventRepository(dbConn)

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid event ID"})
	}

	event, err := eventRepo.GetByID(eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Event not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get event: " + err.Error()})
	}

	return c.JSON(http.StatusOK, event)
}
