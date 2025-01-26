package auth_utils

import (
	"database/sql"
	"errors"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_models"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserDoesntExist    = errors.New("user doesn't exist")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

func RegisterUserTraditionalAuthToDatabase(db *sql.DB, req auth_models.CreateUserTraditionalAuthRequest) (*models.User, error) {
	userRepo := auth_repositories.NewUserRepository(db)

	exists, err := userRepo.EmailExists(*req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	if req.Password != nil {
		hashedPassword, err := HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		req.Password = &hashedPassword
	}

	user := &models.User{
		ID:          uuid.New(),
		Email:       *req.Email,
		Password:    req.Password,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsOnboarded: false,
	}

	err = userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	user.Password = nil // Clear password before returning
	return user, nil
}

func AuthenticateUser(db *sql.DB, req auth_models.LoginRequest) (*models.User, error) {
	userRepo := auth_repositories.NewUserRepository(db)

	user, err := userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}

	if err := ComparePasswords(*user.Password, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	user.Password = nil // Clear password before returning
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
