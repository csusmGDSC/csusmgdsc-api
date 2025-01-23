package auth_handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_models"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_utils"
	"github.com/csusmGDSC/csusmgdsc-api/internal/db"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

func (h *OAuthHandler) RegisterUser(c echo.Context) error {
	var req auth_models.CreateUserRequest
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

	if req.Password == nil && req.Provider == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password or provider are required"})
	}

	dbConn := h.DB.GetDB()

	user, err := auth_utils.RegisterUserToDatabase(dbConn, req)
	if err != nil {
		if err == auth_utils.ErrUserExists {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Email already registered"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Registration failed: " + err.Error()})
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *OAuthHandler) LoginUser(c echo.Context) error {
	var req auth_models.LoginRequest
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

	user, err := auth_utils.AuthenticateUser(dbConn, req)
	if err != nil {
		if err == auth_utils.ErrInvalidCredentials {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials: "})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Authentication failed:" + err.Error()})
	}

	accessToken, err := auth_utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate access token"})
	}

	refreshToken, issuedAt, expiresAt, err := auth_utils.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate refresh token"})
	}

	ipAddress := c.RealIP()
	userAgent := c.Request().Header.Get("User-Agent")
	sessionReq := &auth_models.CreateSessionRequest{
		UserID:    user.ID,
		Token:     refreshToken,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	err = auth_utils.CreateSession(dbConn, *sessionReq)
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

func (h *OAuthHandler) LogoutUser(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Cookie not found"})
	}

	refreshToken := cookie.Value
	if refreshToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Refresh token is required"})
	}

	dbConn := h.DB.GetDB()

	refreshTokensRepo := auth_repositories.NewRefreshTokenRepository(dbConn)
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

func (h *OAuthHandler) LogoutAll(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	dbConn := h.DB.GetDB()

	refreshTokensRepo := auth_repositories.NewRefreshTokenRepository(dbConn)
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

func (h *OAuthHandler) RefreshUser(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Cookie not found"})
	}

	refreshToken := cookie.Value
	if refreshToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Refresh token is required"})
	}

	cfg := config.LoadConfig()
	claims, err := auth_utils.ValidateJWT(refreshToken, []byte(cfg.JWTRefreshSecret))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}

	dbConn := h.DB.GetDB()

	refreshTokensRepo := auth_repositories.NewRefreshTokenRepository(dbConn)
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

	role := models.Role(claims.Role)
	newAccessToken, err := auth_utils.GenerateJWT(storedToken.UserID, &role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create refresh token"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"accessToken": newAccessToken,
	})
}

func (h *OAuthHandler) UpdateUser(c echo.Context) error {
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
	var req auth_models.UpdateUserRequest
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
	userRepo := auth_repositories.NewUserRepository(dbConn)
	updatedUser, err := userRepo.Update(userID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, updatedUser)
}

// // TODO: Add Role checking to allow Admin's to delete non-Admin users
func (h *OAuthHandler) DeleteUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	authenticatedUserID, ok := c.Get("user_id").(string)
	if !ok || authenticatedUserID != userID {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Not authorized to delete this user"})
	}

	dbConn := h.DB.GetDB()

	userRepo := auth_repositories.NewUserRepository(dbConn)
	err := userRepo.DeleteByID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// Handles the initial OAuth login request
func (h *OAuthHandler) OAuthLogin(c echo.Context) error {
	provider := c.Param("provider")
	url, err := auth_utils.GetOAuthURL(provider)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *OAuthHandler) OAuthCallback(c echo.Context) error {
	provider := c.Param("provider")
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	// TODO: Implement state validation
	if state == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing state parameter"})
	}

	userData, err := auth_utils.HandleOAuthCallback(provider, code)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to authenticate"})
	}

	dbConn := db.GetInstance()
	defer dbConn.Close()
	userRepo := auth_repositories.NewUserRepository(dbConn.GetDB())

	var user *models.User
	// Check if user exists
	if userData.Email != nil {
		user, _ = userRepo.GetByEmail(*userData.Email)
	}

	if userData.Email == nil || user == nil {
		// User doesn't exist - generate temporary registration token
		tempToken, err := auth_utils.GenerateTemporaryToken(userData)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate registration token"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":     "registration_required",
			"temp_token": tempToken,
			"user_data": map[string]interface{}{
				"email":      userData.Email,
				"name":       userData.Name,
				"avatar_url": userData.AvatarURL,
				"provider":   userData.Provider,
			},
			"message": "Additional information required to complete registration",
		})
	}

	// User exists - generate tokens
	// TODO: This logic is the same as in the LoginUser handler
	// Refactor this to a common function
	accessToken, err := auth_utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate access token"})
	}

	refreshToken, issuedAt, expiresAt, err := auth_utils.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate refresh token"})
	}

	ipAddress := c.RealIP()
	userAgent := c.Request().Header.Get("User-Agent")
	sessionReq := &auth_models.CreateSessionRequest{
		UserID:    user.ID,
		Token:     refreshToken,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	err = auth_utils.CreateSession(dbConn.GetDB(), *sessionReq)
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
