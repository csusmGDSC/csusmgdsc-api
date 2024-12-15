package auth

import (
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/golang-jwt/jwt/v5"
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

func GenerateJWT(userID string, role models.Role) (string, error) {
	claims := &Claims{
		UserID: userID,
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

func GenerateRefreshToken(userID string, role models.Role) (string, error) {
	claims := &Claims{
		UserID: userID,
		Role:   role.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtRefreshSecret, err := config.LoadJWTRefreshSecret()
	if err != nil {
		return "", err
	}

	signedString, err := token.SignedString([]byte(jwtRefreshSecret))
	return signedString, err
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
