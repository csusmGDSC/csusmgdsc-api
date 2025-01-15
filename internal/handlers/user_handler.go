package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth"
	"github.com/csusmGDSC/csusmgdsc-api/internal/db/repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

func (h *Handler) RegisterUser(c echo.Context) error {
	var req models.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}
	req.FullName = req.FirstName + " " + req.LastName

	if err := h.Validate.Struct(req); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, err.Field()+" "+err.Tag())
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}

	dbConn := h.DB.GetDB()

	user, err := auth.RegisterUser(dbConn, req)
	if err != nil {
		if err == auth.ErrUserExists {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Email already registered"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Registration failed"})
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *Handler) LoginUser(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	if err := h.Validate.Struct(req); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, err.Field()+" "+err.Tag())
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}

	dbConn := h.DB.GetDB()

	user, err := auth.AuthenticateUser(dbConn, req)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Authentication failed"})
	}

	accessToken, err := auth.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate access token"})
	}

	refreshToken, issuedAt, expiresAt, err := auth.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate refresh token"})
	}

	ipAddress := c.RealIP()
	userAgent := c.Request().Header.Get("User-Agent")
	sessionReq := &models.CreateSessionRequest{
		UserID:    user.ID,
		Token:     refreshToken,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	err = auth.CreateSession(dbConn, *sessionReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new session"})
	}

	cookie := new(http.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = refreshToken
	cookie.HttpOnly = true
	// cookie.Secure = true TODO: Once in production set enable this line
	cookie.SameSite = http.SameSiteStrictMode
	cookie.Path = "/"
	cookie.Expires = expiresAt
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"accessToken": accessToken,
		"user":        user,
	})
}

func (h *Handler) LogoutUser(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Cookie not found"})
	}

	refreshToken := cookie.Value
	if refreshToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Refresh token is required"})
	}

	dbConn := h.DB.GetDB()

	refreshTokensRepo := repositories.NewRefreshTokenRepository(dbConn)
	err = refreshTokensRepo.DeleteByToken(refreshToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete token"})
	}

	clearedCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0), // Expire immediately
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(clearedCookie)

	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (h *Handler) LogoutAll(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	dbConn := h.DB.GetDB()

	refreshTokensRepo := repositories.NewRefreshTokenRepository(dbConn)
	err := refreshTokensRepo.DeleteAllByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete tokens"})
	}

	clearedCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0), // Expire immediately
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(clearedCookie)

	return c.JSON(http.StatusOK, map[string]string{"message": "All sessions have been logged out successfully"})
}

func (h *Handler) RefreshUser(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Cookie not found"})
	}

	refreshToken := cookie.Value
	if refreshToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Refresh token is required"})
	}

	jwtRefreshSecret, err := config.LoadJWTRefreshSecret()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal loading error"})
	}

	claims, err := auth.ValidateJWT(refreshToken, []byte(jwtRefreshSecret))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}

	dbConn := h.DB.GetDB()

	refreshTokensRepo := repositories.NewRefreshTokenRepository(dbConn)
	storedToken, err := refreshTokensRepo.GetByToken(refreshToken)
	if err != nil {
		return c.JSON(http.StatusExpectationFailed, map[string]string{"error": err.Error()})
	}

	if claims.UserID != storedToken.UserID.String() {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token claims"})
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Refresh token expired"})
	}

	newAccessToken, err := auth.GenerateJWT(storedToken.UserID, models.Role(claims.Role))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create refresh token"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"accessToken": newAccessToken,
	})
}

// TODO: Add Role checking to allow Admin's to update non-Admin users
func (h *Handler) UpdateUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	// Check if the authenticated user has permission to update this user
	authenticatedUserID, ok := c.Get("user_id").(string)
	if !ok || authenticatedUserID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Not authorized to update this user"})
	}

	// Bind the update request
	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	// Validate the request
	if err := h.Validate.Struct(req); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, err.Field()+" "+err.Tag())
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}

	dbConn := h.DB.GetDB()

	// Update user
	userRepo := repositories.NewUserRepository(dbConn)
	updatedUser, err := userRepo.Update(userID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, updatedUser)
}

// TODO: Add Role checking to allow Admin's to delete non-Admin users
func (h *Handler) DeleteUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	authenticatedUserID, ok := c.Get("user_id").(string)
	if !ok || authenticatedUserID != userID {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Not authorized to delete this user"})
	}

	dbConn := h.DB.GetDB()

	userRepo := repositories.NewUserRepository(dbConn)
	err := userRepo.DeleteByID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted successfully"})
}
