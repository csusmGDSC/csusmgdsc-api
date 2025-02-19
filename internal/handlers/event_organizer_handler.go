package handlers

import (
	"net/http"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/db/repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AddEventOrganizer adds a user as an organizer to an event.
//
// It first checks if the event and user exist, and if not, returns a 400 status code.
// If the insertion of the event organizer into the event_organizers table fails, it returns a 500 status code.
// If the insertion is successful, it returns a 200 status code with a message saying that the organizer was added successfully to the event.
func (h *Handler) AddEventOrganizer(c echo.Context) error {
	eventId := c.Param("id")
	userId := c.Param("userId")

	if eventId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event ID is required"})
	}

	if userId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	eventUUID, err := uuid.Parse(eventId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid event ID"})
	}

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	dbConn := h.DB.GetDB()
	eventOrganizerRepo := repositories.NewEventOrganizerRepository(dbConn)
	utilsRepo := repositories.NewUtilsRepository(dbConn)

	if exists, err := utilsRepo.CheckIfUUIDExists("events", "id", eventUUID); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event not found"})
	}

	if exists, err := utilsRepo.CheckIfUUIDExists("users", "id", userUUID); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User not found"})
	}

	eventOrganizer := models.EventOrganizer{
		EventID:   eventUUID,
		UserID:    userUUID,
		CreatedAt: time.Now(),
	}

	err = eventOrganizerRepo.InsertEventOrganizer(eventOrganizer)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to insert event organizer"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Organizer added successfully to event."})
}

// GetEventOrganizers retrieves a list of users who are organizers of a given event ID.
//
// It takes an event ID as a parameter and returns a list of User objects that correspond to the organizers of the event.
//
// The function returns an error if the query fails, or if the event ID is invalid or does not correspond to an existing event.
func (h *Handler) GetEventOrganizers(c echo.Context) error {
	eventId := c.Param("id")

	if eventId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event ID is required"})
	}

	eventUUID, err := uuid.Parse(eventId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid event ID"})
	}

	dbConn := h.DB.GetDB()
	eventOrganizerRepo := repositories.NewEventOrganizerRepository(dbConn)
	utilsRepo := repositories.NewUtilsRepository(dbConn)

	if exists, err := utilsRepo.CheckIfUUIDExists("events", "id", eventUUID); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event not found"})
	}

	organizers, err := eventOrganizerRepo.GetEventOrganizers(eventUUID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error: ": "Unable to get event organizers"})
	}

	return c.JSON(http.StatusOK, organizers)
}

// GetUserAssignedEvents retrieves a list of events that a given user is an organizer of.
//
// It takes a user ID as a parameter and returns a list of Event objects that the user is organizing.
//
// The function returns an error if the query fails, or if the user ID is invalid or does not correspond to an existing user.
func (h *Handler) GetUserAssignedEvents(c echo.Context) error {
	userId := c.Param("id")

	if userId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user id"})
	}

	dbConn := h.DB.GetDB()
	eventOrganizerRepo := repositories.NewEventOrganizerRepository(dbConn)
	utilsRepo := repositories.NewUtilsRepository(dbConn)

	if exists, err := utilsRepo.CheckIfUUIDExists("users", "id", userUUID); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User not found"})
	}

	events, err := eventOrganizerRepo.GetEventsByUserID(userUUID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Unable to get user assigned events"})
	}

	return c.JSON(http.StatusOK, events)
}

// DeleteOrganizerFromEvent deletes an event organizer from the database. It takes two parameters, an event ID and a user ID, and deletes the row from the event_organizers table that matches these IDs.
//
// The function first checks if the event and user exist, and if not, returns a 400 status code.
// If the deletion of the event organizer from the event_organizers table fails, it returns a 500 status code.
// If the deletion is successful, it returns a 200 status code with a message saying that the organizer was removed successfully from the event.
func (h *Handler) DeleteOrganizerFromEvent(c echo.Context) error {
	eventId := c.Param("id")
	userId := c.Param("userId")

	if eventId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event ID is required"})
	}

	if userId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	eventUUID, err := uuid.Parse(eventId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid event ID"})
	}

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	dbConn := h.DB.GetDB()
	eventOrganizerRepo := repositories.NewEventOrganizerRepository(dbConn)
	utilsRepo := repositories.NewUtilsRepository(dbConn)

	if exists, err := utilsRepo.CheckIfUUIDExists("events", "id", eventUUID); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event not found"})
	}

	if exists, err := utilsRepo.CheckIfUUIDExists("users", "id", userUUID); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User not found"})
	}

	err = eventOrganizerRepo.DeleteEventOrganizer(eventUUID, userUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete organizer from event"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Organizer removed succesfully from event."})
}
