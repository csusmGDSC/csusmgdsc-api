package handlers

import (
	"database/sql"
	"net/http"
	"net/url"

	"github.com/csusmGDSC/csusmgdsc-api/internal/db/repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// InsertEventHandler creates a new event and inserts it into the database.
//
// It first checks if the requesting user has the "ADMIN" role and if not, returns a 401 status code.
// It then binds the JSON request body to an Event struct and validates the struct.
// If the validation fails, it returns a 400 status code with the validation errors.
// If the validation is successful, it calls the InsertEvent method of the EventRepository to insert the event into the database.
// If the insertion fails, it returns a 500 status code with an error message.
// If the insertion is successful, it returns a 201 status code with a message saying that the event was created successfully and the event ID.
func (h *Handler) InsertEventHandler(c echo.Context) error {
	userRole, ok := c.Get("user_role").(string)
	if !ok || userRole != "ADMIN" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Insufficient permissions"})
	}

	var event models.Event
	// Bind JSON request body to the Event struct
	if err := c.Bind(&event); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request structure"})
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

	// Check if the image URL is valid, if its given.
	if event.ImageSrc != nil {
		_, err := url.ParseRequestURI(*event.ImageSrc)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid image URL"})
		}
	}

	// Set default for user id if not set
	if event.CreatedBy == nil {
		userId, _ := c.Get("user_id").(string)
		userUUID, err := uuid.Parse(userId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "User id does not exist in context."})
		}
		event.CreatedBy = &userUUID
	}

	dbConn := h.DB.GetDB()
	eventRepo := repositories.NewEventRepository(dbConn)

	eventId, err := eventRepo.InsertEvent(dbConn, event)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to insert event"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Event created successfully", "eventID": eventId.String()})
}

// GetEventsHandler retrieves a paginated list of events from the database.
//
// It takes pagination parameters from the query, and returns a list of Event objects, total count of events, current page, and limit.
//
// The function returns an error if the query fails.
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
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get events"})
	}

	return c.JSON(http.StatusOK, response)
}

// GetEventByIDHandler retrieves an event by its ID.
//
// It first checks if the event ID is valid and if the event exists, and if not, returns a 400 status code.
// If the retrieval of the event from the events table fails, it returns a 500 status code.
// If the retrieval is successful, it returns a 200 status code with the event.
func (h *Handler) GetEventByIDHandler(c echo.Context) error {
	eventId := c.Param("id")
	if eventId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event ID is required"})
	}

	eventID, err := uuid.Parse(eventId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid event ID"})
	}

	dbConn := h.DB.GetDB()
	eventRepo := repositories.NewEventRepository(dbConn)

	event, err := eventRepo.GetByID(eventID)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Event not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get event"})
	}

	return c.JSON(http.StatusOK, event)
}

// UpdateEventByID updates an event given its ID.
//
// It takes an EventID and an UpdateEventRequest object as parameters. The function
// updates the event record in the database with the fields that are not empty in the
// UpdateEventRequest object, and returns an error if the query fails.
//
// If no fields are changed, the function returns nil.
func (h *Handler) UpdateEventByID(c echo.Context) error {
	userRole, ok := c.Get("user_role").(string)
	if !ok || userRole != "ADMIN" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Insufficient permissions"})
	}

	eventId := c.Param("id")

	if eventId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event ID is required"})
	}

	eventUUID, err := uuid.Parse(eventId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid event ID"})
	}

	var event models.UpdateEventRequest
	if err := c.Bind(&event); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request structure"})
	}

	if err := h.Validate.Struct(event); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, err.Field()+" "+err.Tag())
		}

		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}

	dbConn := h.DB.GetDB()
	eventRepo := repositories.NewEventRepository(dbConn)
	utilsRepo := repositories.NewUtilsRepository(dbConn)

	if exists, err := utilsRepo.CheckIfUUIDExists("events", "id", eventUUID); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event not found"})
	}

	oldEvent, err := eventRepo.GetByID(eventUUID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Existing event exists but could not fetch it."})
	}

	// If a new image URL is provided and it's different from the current one, remove the old image
	if event.ImageSrc != nil && oldEvent.ImageSrc != nil && event.ImageSrc != oldEvent.ImageSrc {
		_, err = url.ParseRequestURI(*event.ImageSrc)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid image URL"})
		}

		req := c.Request()
		q := req.URL.Query()
		q.Set("url", *oldEvent.ImageSrc)
		req.URL.RawQuery = q.Encode()

		if err := h.RemoveImage(c); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete old event image"})
		}
	}

	err = eventRepo.UpdateEventById(eventUUID, event)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update event", "message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Event updated successfully"})
}

// DeleteEventByID deletes an event given its ID.
//
// It first checks if the event ID is valid and if the event exists, and if not, returns a 400 status code.
// If the deletion of the event from the events table fails, it returns a 500 status code.
// If the deletion is successful, it returns a 200 status code with a message saying that the event was deleted successfully.
func (h *Handler) DeleteEventByID(c echo.Context) error {
	userRole, ok := c.Get("user_role").(string)
	if !ok || userRole != "ADMIN" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Insufficient permissions"})
	}

	eventId := c.Param("id")

	if eventId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event id is required"})
	}

	eventUUID, err := uuid.Parse(eventId)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid event id"})
	}

	dbConn := h.DB.GetDB()
	eventRepo := repositories.NewEventRepository(dbConn)
	utilsRepo := repositories.NewUtilsRepository(dbConn)

	if exists, err := utilsRepo.CheckIfUUIDExists("events", "id", eventUUID); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event not found"})
	}

	event, err := eventRepo.GetByID(eventUUID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Event exists but could not fetch it."})
	}

	if event.ImageSrc != nil {
		req := c.Request()
		q := req.URL.Query()
		q.Set("url", *event.ImageSrc)
		req.URL.RawQuery = q.Encode()

		if err := h.RemoveImage(c); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete event image"})
		}
	}

	err = eventRepo.DeleteEventById(eventUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete event"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Event successfully deleted."})
}
