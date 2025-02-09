package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/db/repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) InsertCommentHandler(c echo.Context) error {
	var comment models.Comment

	// Bind JSON request body to the Comment struct
	if err := c.Bind(&comment); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	// Validate the struct
	if err := h.Validate.Struct(comment); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, err.Field()+" "+err.Tag())
		}

		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}

	comment.ID = uuid.New()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	dbConn := h.DB.GetDB()
	commentRepo := repositories.NewCommentRepository(dbConn)

	if err := commentRepo.CreateComment(dbConn, comment); err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to create comment"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Comment created successfully", "id": comment.ID.String()})
}

func (h *Handler) DeleteCommentHandler(c echo.Context) error {
	userID := c.Param("userId")
	commentID := c.Param("commentId")

	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	// Convert string ID to UUID
	commentUUID, err := uuid.Parse(commentID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid comment ID"})
	}

	// Check if the authenticated user has permission to delete this comment
	authenticatedUserID, ok := c.Get("user_id").(string)
	if !ok || authenticatedUserID != userID {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Not authorized to delete this comment"})
	}

	dbConn := h.DB.GetDB()
	commentRepo := repositories.NewCommentRepository(dbConn)

	err = commentRepo.DeleteCommentById(commentUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comment"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Comment deleted successfully"})
}

func (h *Handler) GetCommentsByUserIdHandler(c echo.Context) error {
	userID := c.Param("userId")

	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	// Convert string ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	dbConn := h.DB.GetDB()
	commentRepo := repositories.NewCommentRepository(dbConn)

	comments, err := commentRepo.GetCommentsByUserId(userUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get comments"})
	}

	return c.JSON(http.StatusOK, comments)
}

func (h *Handler) GetCommentsByEventIdHandler(c echo.Context) error {
	eventID := c.Param("eventId")

	if eventID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event ID is required"})
	}

	// Convert string ID to UUID
	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid event ID"})
	}

	dbConn := h.DB.GetDB()
	commentRepo := repositories.NewCommentRepository(dbConn)

	comments, err := commentRepo.GetCommentsByEventId(eventUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get comments"})
	}

	return c.JSON(http.StatusOK, comments)
}

func (h *Handler) GetCommentByIdHandler(c echo.Context) error {
	commentID := c.Param("id")

	if commentID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Comment ID is required"})
	}

	// Convert string ID to UUID
	commentUUID, err := uuid.Parse(commentID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid comment ID"})
	}

	dbConn := h.DB.GetDB()
	commentRepo := repositories.NewCommentRepository(dbConn)

	comment, err := commentRepo.GetCommentByCommentId(commentUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get comment"})
	}

	return c.JSON(http.StatusOK, comment)
}

func (h *Handler) UpdateCommentHandler(c echo.Context) error {
	commentID := c.Param("id")

	if commentID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Comment ID is required"})
	}

	// Convert string ID to UUID
	commentUUID, err := uuid.Parse(commentID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid comment ID"})
	}

	// Bind JSON request body to the Comment struct
	var comment models.UpdateCommentRquest
	if err := c.Bind(&comment); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	// Validate the struct
	if err := h.Validate.Struct(comment); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, err.Field()+" "+err.Tag())
		}

		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}

	dbConn := h.DB.GetDB()
	commentRepo := repositories.NewCommentRepository(dbConn)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid comment ID"})
	}

	err = commentRepo.UpdateCommentByCommentId(commentUUID, comment)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update comment"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Comment updated successfully"})
}
