package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth"
	"github.com/csusmGDSC/csusmgdsc-api/internal/db"
	"github.com/csusmGDSC/csusmgdsc-api/internal/db/repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

func RegisterUser(c echo.Context) error {
	var req models.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}
	req.FullName = req.FirstName + " " + req.LastName

	if err := validate.Struct(req); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, err.Field()+" "+err.Tag())
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}

	dbConn, err := db.ConnectDB()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database connection failed"})
	}
	defer dbConn.Close()

	user, err := auth.RegisterUser(dbConn, req)
	if err != nil {
		if err == auth.ErrUserExists {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Email already registered"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Registration failed"})
	}

	return c.JSON(http.StatusCreated, user)
}

func LoginUser(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	if err := validate.Struct(req); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, err.Field()+" "+err.Tag())
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}

	dbConn, err := db.ConnectDB()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database connection failed"})
	}
	defer dbConn.Close()

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

func RefreshUser(c echo.Context) error {
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

	dbConn, err := db.ConnectDB()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database connection failed"})
	}
	defer dbConn.Close()

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

func UpdateUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	// Check if the authenticated user has permission to update this user
	authenticatedUserID := c.Get("user_id").(string)
	if authenticatedUserID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Not authorized to update this user"})
	}

	// Bind the update request
	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	// Validate the request
	if err := validate.Struct(req); err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, err.Field()+" "+err.Tag())
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"errors": validationErrors,
		})
	}

	dbConn, err := db.ConnectDB()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database connection failed"})
	}
	defer dbConn.Close()

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
