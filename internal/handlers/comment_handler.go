package handlers

import (
	"net/http"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/db/repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// InsertCommentHandler inserts a new comment into the database.
//
// This handler is used to insert a new comment into the database.
// The request body should contain a JSON object with the following
// fields:
//
// - user_id: the ID of the user who posted the comment
// - event_id: the ID of the event the comment refers to
// - content: the text of the comment
// - parent_id: the ID of the comment's parent (optional)
//
// The handler returns a JSON object with a single field, "message",
// which contains a success message if the comment is inserted
// successfully, or an error message if there is a failure.
//
// The handler also returns a JSON object with a single field, "id",
// which contains the ID of the newly inserted comment.
func (h *Handler) InsertCommentHandler(c echo.Context) error {
	var comment models.Comment

	// Bind JSON request body to the Comment struct
	if err := c.Bind(&comment); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	// Check if user_id is provided
	if comment.UserId == uuid.Nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "user_id is required"})
	}

	// Check if event_id is provided
	if comment.EventId == uuid.Nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "event_id is required"})
	}

	// Check if content is provided
	if comment.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "content is required"})
	}

	comment.ID = uuid.New()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	dbConn := h.DB.GetDB()
	commentRepo := repositories.NewCommentRepository(dbConn)
	utilsRepo := repositories.NewUtilsRepository(dbConn)

	// Check if event exists
	if exists, err := utilsRepo.CheckIfUUIDExists("events", "id", comment.EventId); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event not found"})
	}

	// If given the parent comment ID, check if it exists
	if comment.ParentId != nil {
		if exists, err := utilsRepo.CheckIfUUIDExists("comments", "id", *comment.ParentId); !exists || err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Parent comment not found"})
		}
	}

	if err := commentRepo.CreateComment(dbConn, comment); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to create comment"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Comment created successfully", "id": comment.ID.String()})
}

// DeleteCommentHandler deletes a comment from the database.
//
// This handler is used to delete a comment from the database.
// The comment ID is passed as a parameter in the URL.
// The handler returns a JSON object with a single field, "message",
// which contains a success message if the comment is deleted
// successfully, or an error message if there is a failure.
func (h *Handler) DeleteCommentHandler(c echo.Context) error {
	commentID := c.Param("id")

	// Convert string ID to UUID
	commentUUID, err := uuid.Parse(commentID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid comment ID"})
	}

	dbConn := h.DB.GetDB()
	commentRepo := repositories.NewCommentRepository(dbConn)
	utilsRepo := repositories.NewUtilsRepository(dbConn)

	// Check if comment exists
	if exists, err := utilsRepo.CheckIfUUIDExists("comments", "id", commentUUID); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Comment not found"})
	}

	err = commentRepo.DeleteCommentById(commentUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comment"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Comment deleted successfully"})
}

// GetCommentsHandler returns a list of comments that match the query parameters.
//
// The handler takes two query parameters, user_id and event_id, which are used to filter the results.
// If both parameters are provided, the handler returns comments that belong to the specified user and event.
// If only user_id is provided, the handler returns comments that belong to the specified user.
// If only event_id is provided, the handler returns comments that belong to the specified event.
// If neither parameter is provided, the handler returns all comments.
//
// The handler returns a JSON object with a single field, "comments", which contains the list of comments.
// Each comment is represented as a JSON object with the following fields:
//
// - id: the ID of the comment
// - user_id: the ID of the user who posted the comment
// - event_id: the ID of the event the comment refers to
// - content: the text of the comment
// - parent_id: the ID of the comment's parent (optional)
// - created_at: the timestamp when the comment was created
// - updated_at: the timestamp when the comment was last updated
func (h *Handler) GetCommentsHandler(c echo.Context) error {
	userID := c.QueryParam("user_id")
	eventID := c.QueryParam("event_id")
	dbConn := h.DB.GetDB()
	commentRepo := repositories.NewCommentRepository(dbConn)
	utilsRepo := repositories.NewUtilsRepository(dbConn)

	var userUUID, eventUUID uuid.UUID
	var err error

	// Convert userID and eventID to UUID if they exist
	if userID != "" {
		userUUID, err = uuid.Parse(userID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
		}
	}

	if eventID != "" {
		eventUUID, err = uuid.Parse(eventID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid event ID"})
		}
	}

	// Check if event exists
	if eventID != "" {
		if exists, err := utilsRepo.CheckIfUUIDExists("events", "id", eventUUID); !exists || err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event not found"})
		}
	}

	var comments []*models.Comment

	switch {
	case userID != "" && eventID != "":
		comments, err = commentRepo.GetCommentsByUserIdAndEventId(userUUID, eventUUID)

	case userID != "":
		comments, err = commentRepo.GetCommentsByUserId(userUUID)

	case eventID != "":
		comments, err = commentRepo.GetCommentsWithRepliesByEventId(eventUUID)

	default:
		comments, err = commentRepo.GetAllComments()
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get comments"})
	}
	return c.JSON(http.StatusOK, comments)
}

// GetCommentByIdHandler returns a comment by its ID.
//
// The handler takes a single path parameter, "id", which is the ID of the comment to be returned.
// The handler returns a JSON object with a single field, "comment", which contains the comment.
// The comment is represented as a JSON object with the following fields:
//
// - id: the ID of the comment
// - user_id: the ID of the user who posted the comment
// - event_id: the ID of the event the comment refers to
// - content: the text of the comment
// - parent_id: the ID of the comment's parent (optional)
// - created_at: the timestamp when the comment was created
// - updated_at: the timestamp when the comment was last updated
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
	utilsRepo := repositories.NewUtilsRepository(dbConn)

	// Check if comment exists
	if exists, err := utilsRepo.CheckIfUUIDExists("comments", "id", commentUUID); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Comment not found"})
	}

	comment, err := commentRepo.GetCommentByCommentId(commentUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get comment"})
	}

	return c.JSON(http.StatusOK, comment)
}

// UpdateCommentHandler updates a comment in the database.
//
// This handler is used to update an existing comment in the database.
// The comment ID is passed as a parameter in the URL, and the updated
// fields are provided in the request body as a JSON object.
//
// The JSON object in the request body should contain one or more of the
// following fields:
//
// - content: the updated text of the comment
// - pinned_by: the ID of the user who pinned the comment
// - parent_id: the ID of the comment's parent (optional)
//
// The handler performs validation on the request data, and if successful,
// updates the comment in the database. If the update is successful, it
// returns a JSON object with a single field, "message", containing a
// success message. If there is an error, it returns a JSON object with
// an "error" field containing an error message.
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
	var comment models.UpdateCommentRequest
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
	utilsRepo := repositories.NewUtilsRepository(dbConn)

	// Check if comment exists
	if exists, err := utilsRepo.CheckIfUUIDExists("comments", "id", commentUUID); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Comment not found"})
	}

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid comment ID"})
	}

	err = commentRepo.UpdateCommentByCommentId(commentUUID, comment)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update comment"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Comment updated successfully"})
}

// GetCommentRepliesHandler returns the replies to a specific comment.
//
// This handler takes a single path parameter, "id", which is the ID of the comment
// for which replies are being requested. The handler returns a JSON array of replies,
// each represented as a JSON object with the following fields:
//
// - id: the ID of the reply
// - user_id: the ID of the user who posted the reply
// - event_id: the ID of the event the reply refers to
// - content: the text of the reply
// - parent_id: the ID of the comment's parent (optional)
// - created_at: the timestamp when the reply was created
// - updated_at: the timestamp when the reply was last updated
//
// If the comment ID is invalid or there is an error retrieving the replies,
// an appropriate error message is returned.
func (h *Handler) GetCommentRepliesHandler(c echo.Context) error {
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
	utilsRepo := repositories.NewUtilsRepository(dbConn)

	// Check if comment exists
	if exists, err := utilsRepo.CheckIfUUIDExists("comments", "id", commentUUID); !exists || err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Comment not found"})
	}

	replies, err := commentRepo.GetRepliesByCommentId(commentUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get replies"})
	}

	return c.JSON(http.StatusOK, replies)
}
