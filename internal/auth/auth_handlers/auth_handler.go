package auth_handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_models"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_utils"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Register User from the registration page
func (h *OAuthHandler) RegisterUser(c echo.Context) error {
	var req auth_models.CreateUserTraditionalAuthRequest
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

	// encapsulate the registration logic, used in OAuthCallback
	user, err := auth_utils.RegisterUserTraditionalAuthToDatabase(dbConn, req)
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

	// Refactored code
	accessToken, cookie, err := auth_utils.CreateLoginSession(dbConn, c.RealIP(), c.Request().Header.Get("User-Agent"), user)

	if err == auth_utils.ErrAccessToken {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate access token"})
	}

	if err == auth_utils.ErrRefreshToken {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate refresh token"})
	}

	if err == auth_utils.ErrNewSession {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new session"})
	}

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

	dbConn := h.DB.GetDB()
	userRepo := auth_repositories.NewUserRepository(dbConn)

	var user *models.User
	// Check if user exists
	if userData.ID != "" {
		user, _ = userRepo.GetByAuthID(userData.ID)
	}

	if user == nil {
		// User doesn't exist - register the user
		req := &auth_models.CreateUserOAuthRequest{
			Email:    userData.Email,
			Provider: &userData.Provider,
			AuthID:   &userData.ID,
			Image:    userData.AvatarURL,
			Name:     &userData.Name,
		}

		user, err = auth_utils.RegisterUserOAuthToDatabase(dbConn, *req)
		if err != nil {
			if err == auth_utils.ErrUserExists {
				return c.JSON(http.StatusConflict, map[string]string{"error": "Email already registered"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Registration failed: " + err.Error()})
		}
	}

	//	Refactored code
	accessToken, cookie, err := auth_utils.CreateLoginSession(dbConn, c.RealIP(), c.Request().Header.Get("User-Agent"), user)
	if err == auth_utils.ErrAccessToken {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate access token"})
	}

	if err == auth_utils.ErrRefreshToken {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate refresh token"})
	}

	if err == auth_utils.ErrNewSession {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new session"})
	}

	c.SetCookie(cookie)

	frontendURL := "https://gdsc-csusm.com"

	if !user.IsOnboarded {
		frontendURL = frontendURL + "/onboarding"
	}

	redirectURL := fmt.Sprintf("%s?token=%s",
		frontendURL,
		url.QueryEscape(accessToken),
	)

	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *OAuthHandler) GetUserByIDHandler(c echo.Context) error {
	userIDStr, ok := c.Get("user_id").(string)
	if !ok || userIDStr == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	dbConn := h.DB.GetDB()
	userRepo := auth_repositories.NewUserRepository(dbConn)

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
