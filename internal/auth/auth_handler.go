package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	"github.com/csusmGDSC/csusmgdsc-api/internal/db"
	"github.com/csusmGDSC/csusmgdsc-api/internal/db/repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

// Handles the initial OAuth login request
func OAuthLogin(c echo.Context) error {
	provider := c.Param("provider")
	url, err := GetOAuthURL(provider)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func OAuthCallback(c echo.Context) error {
	provider := c.Param("provider")
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	// TODO: Implement state validation
	if state == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing state parameter"})
	}

	userData, err := HandleOAuthCallback(provider, code)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to authenticate"})
	}

	dbConn := db.GetInstance()
	defer dbConn.Close()
	userRepo := repositories.NewUserRepository(dbConn.GetDB())

	var user *models.User
	// Check if user exists
	if userData.Email != nil {
		user, _ = userRepo.GetByEmail(*userData.Email)
	}

	if userData.Email == nil || user == nil {
		// User doesn't exist - generate temporary registration token
		tempToken, err := GenerateTemporaryToken(userData)
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
	accessToken, err := GenerateJWT(user.ID, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate access token"})
	}

	refreshToken, issuedAt, expiresAt, err := GenerateRefreshToken(user.ID, user.Role)
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

	err = CreateSession(dbConn.GetDB(), *sessionReq)
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

func CompleteOAuthRegistration(c echo.Context) error {
	var req models.CompleteOAuthRegistrationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	cfg := config.LoadConfig()
	var tempClaims TempTokenClaims
	token, err := jwt.ParseWithClaims(req.TempToken, &tempClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTAccessSecret), nil
	})

	if err != nil || !token.Valid {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired registration token"})
	}

	// Create new user with OAuth and additional information
	user := &models.User{
		ID:             uuid.New(),
		Email:          tempClaims.OAuthUserData.Email,
		FullName:       tempClaims.OAuthUserData.Name,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Role:           models.UserRole,
		Position:       req.Position,
		Branch:         req.Branch,
		GraduationDate: &req.GraduationDate,
		Provider:       &tempClaims.OAuthUserData.Provider,
		AuthID:         &tempClaims.OAuthUserData.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Get database connection from context
	db := c.Get("db").(*sql.DB)
	userRepo := repositories.NewUserRepository(db)

	// Create the user
	err = userRepo.Create(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user account",
		})
	}

	// TODO: This logic is the same as in the LoginUser handler
	// Refactor this to a common function
	accessToken, err := GenerateJWT(user.ID, user.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate access token"})
	}

	refreshToken, issuedAt, expiresAt, err := GenerateRefreshToken(user.ID, user.Role)
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

	err = CreateSession(db, *sessionReq)
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

func RegisterUser(db *sql.DB, req models.CreateUserRequest) (*models.User, error) {
	userRepo := repositories.NewUserRepository(db)

	exists, err := userRepo.EmailExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:             uuid.New(),
		FullName:       req.FullName,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          &req.Email,
		Password:       hashedPassword,
		Role:           models.UserRole,
		Position:       req.Position,
		Branch:         req.Branch,
		GraduationDate: req.GraduationDate,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err = userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	user.Password = "" // Clear password before returning
	return user, nil
}

func AuthenticateUser(db *sql.DB, req models.LoginRequest) (*models.User, error) {
	userRepo := repositories.NewUserRepository(db)

	user, err := userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := ComparePasswords(user.Password, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	user.Password = "" // Clear password before returning
	return user, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func ComparePasswords(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
		}

		cfg := config.LoadConfig()
		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidateJWT(tokenString, []byte(cfg.JWTAccessSecret))
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		// Add user info to context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		return next(c)
	}
}
