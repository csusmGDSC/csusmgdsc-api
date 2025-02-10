package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_repositories"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) GetUsersHandler(c echo.Context) error {
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
	userRepo := auth_repositories.NewUserRepository(dbConn)
	fmt.Println("Created repository")
	response, err := userRepo.GetAll(page, limit)
	fmt.Println("User response Gell All")
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "No users found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get users"})
	}

	// Clear sensitive information from response
	for _, user := range response.Users {
		user.Password = nil
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) GetUserByIDHandler(c echo.Context) error {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	dbConn := h.DB.GetDB()
	userRepo := auth_repositories.NewUserRepository(dbConn)

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	user, err := userRepo.GetByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user"})
	}

	user.Password = nil
	return c.JSON(http.StatusOK, user)
}
