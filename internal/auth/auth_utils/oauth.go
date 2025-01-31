package auth_utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/auth"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_models"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

func RegisterUserOAuthToDatabase(db *sql.DB, req auth_models.CreateUserOAuthRequest) (*models.User, error) {
	userRepo := auth_repositories.NewUserRepository(db)

	if req.AuthID == nil {
		return nil, fmt.Errorf("auth_id is required")
	}

	exists, err := userRepo.AuthIDExists(*req.AuthID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	email := "not set"
	if req.Email != nil {
		email = *req.Email
	}

	user := &models.User{
		ID:          uuid.New(),
		Email:       email,
		Password:    nil,
		Provider:    req.Provider,
		AuthID:      req.AuthID,
		Image:       req.Image,
		FullName:    req.Name,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsOnboarded: false,
	}

	err = userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetOAuthURL generates an OAuth URL for the provider
func GetOAuthURL(provider string) (string, error) {
	githubConfig := auth.GetGitHubConfig()
	googleConfig := auth.GetGoogleConfig()
	switch provider {
	case "github":
		if githubConfig == nil {
			return "", fmt.Errorf("github OAuth not initialized")
		}
		return githubConfig.AuthCodeURL("state"), nil
	case "google":
		if googleConfig == nil {
			return "", fmt.Errorf("google OAuth not initialized")
		}
		return googleConfig.AuthCodeURL("state"), nil
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

// HandleOAuthCallback handles OAuth2 callback and returns user data
func HandleOAuthCallback(provider string, code string) (*auth.OAuthUserData, error) {
	githubConfig := auth.GetGitHubConfig()
	googleConfig := auth.GetGoogleConfig()
	var config *oauth2.Config
	var userInfoURL string

	switch provider {
	case "github":
		if githubConfig == nil {
			return nil, fmt.Errorf("github OAuth not initialized")
		}
		config = githubConfig
		userInfoURL = "https://api.github.com/user"
	case "google":
		if googleConfig == nil {
			return nil, fmt.Errorf("google OAuth not initialized")
		}
		config = googleConfig
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %v", err)
	}
	client := config.Client(context.Background(), token)
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	var rawData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %v", err)
	}

	userData := &auth.OAuthUserData{
		Provider: provider,
	}

	switch provider {
	case "github":
		// Handle required ID field
		if id, ok := rawData["id"].(float64); ok {
			userData.ID = fmt.Sprintf("%.0f", id)
		} else {
			return nil, fmt.Errorf("invalid or missing github user ID")
		}
		if email, ok := rawData["email"].(string); ok && email != "" {
			userData.Email = &email
		}
		if name, ok := rawData["name"].(string); ok && name != "" {
			userData.Name = name
		}
		if avatar, ok := rawData["avatar_url"].(string); ok {
			userData.AvatarURL = &avatar
		}
	case "google":
		if id, ok := rawData["id"].(string); ok && id != "" {
			userData.ID = id
		}
		if email, ok := rawData["email"].(string); ok && email != "" {
			userData.Email = &email
		}
		if name, ok := rawData["name"].(string); ok && name != "" {
			userData.Name = name
		}
		if picture, ok := rawData["picture"].(string); ok {
			userData.AvatarURL = &picture
		}
	}
	return userData, nil
}

func CreateLoginSession(dbConn *sql.DB, c echo.Context, user *models.User) (string, error) {

	accessToken, err := GenerateJWT(user.ID, user.Role)
	if err != nil {
		return "", c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate access token"})
	}

	refreshToken, issuedAt, expiresAt, err := GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return "", c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate refresh token"})
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

	err = CreateSession(dbConn, *sessionReq)
	if err != nil {
		return "", c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new session"})
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

	return accessToken, nil
}
