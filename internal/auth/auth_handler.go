package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	"github.com/csusmGDSC/csusmgdsc-api/internal/db/repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

// // OAuth2 login redirect
// func OAuthLogin(c echo.Context) error {
// 	provider := c.Param("provider")
// 	url, err := GetOAuthURL(provider)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid provider"})
// 	}
// 	return c.Redirect(http.StatusTemporaryRedirect, url)
// }

// // OAuth2 callback handler
// func OAuthCallback(c echo.Context) error {
// 	provider := c.Param("provider")
// 	code := c.QueryParam("code")

// 	userID, err := HandleOAuthCallback(provider, code)
// 	if err != nil {
// 		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Failed to authenticate"})
// 	}

// 	// Generate JWT
// 	token, err := GenerateJWT(userID)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
// 	}

// 	return c.JSON(http.StatusOK, map[string]string{"token": token})
// }

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
		Email:          req.Email,
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

		jwtAccessSecret, err := config.LoadJWTAccessSecret()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal loading error")
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidateJWT(tokenString, []byte(jwtAccessSecret))
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		// Add user info to context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		return next(c)
	}
}
