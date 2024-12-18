package auth

import (
	"database/sql"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	"github.com/csusmGDSC/csusmgdsc-api/internal/db/repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	AccessTokenExpiry  = time.Minute * 15   // Short-lived access token
	RefreshTokenExpiry = time.Hour * 24 * 7 // Long-lived refresh token
)

// Custom claims to set on JWT https://pkg.go.dev/github.com/golang-jwt/jwt/v4#NewWithClaims
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID uuid.UUID, role models.Role) (string, error) {
	claims := &Claims{
		UserID: userID.String(),
		Role:   role.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtAccessSecret, err := config.LoadJWTAccessSecret()
	if err != nil {
		return "", err
	}

	signedString, err := token.SignedString([]byte(jwtAccessSecret))
	return signedString, err
}

func GenerateRefreshToken(userID uuid.UUID, role models.Role) (string, time.Time, time.Time, error) {
	issuedAt := jwt.NewNumericDate(time.Now())
	expiresAt := jwt.NewNumericDate(time.Now().Add(RefreshTokenExpiry))
	claims := &Claims{
		UserID: userID.String(),
		Role:   role.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  issuedAt,
			ExpiresAt: expiresAt,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtRefreshSecret, err := config.LoadJWTRefreshSecret()
	if err != nil {
		// zero value for a time.Time type specified in https://pkg.go.dev/time#Time
		zeroTime := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
		return "", zeroTime, zeroTime, err
	}

	signedString, err := token.SignedString([]byte(jwtRefreshSecret))
	return signedString, issuedAt.Time, expiresAt.Time, err
}

func CreateSession(db *sql.DB, req models.CreateSessionRequest) error {
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
	err := refreshTokenRepo.Create(&req)
	if err != nil {
		return err
	}

	return nil
}

func ValidateJWT(tokenString string, secret []byte) (*Claims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
