package auth_utils

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/csusmGDSC/csusmgdsc-api/config"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_models"
	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_repositories"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/google/uuid"
	"github.com/resend/resend-go/v2"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserDoesntExist    = errors.New("user doesn't exist")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrAccessToken        = errors.New("failed to generate access token")
	ErrRefreshToken       = errors.New("failed to generate refresh token")
	ErrVerificationToken  = errors.New("failed to generate a verification token")
	ErrNewSession         = errors.New("failed to create new session")
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
		ID:            uuid.New(),
		Email:         *req.Email,
		Password:      req.Password,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsOnboarded:   false,
		EmailVerified: false,
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

func SendVerificationEmail(userEmail string, verificationToken string) error {

	// Load the conigration keys
	cfg := config.LoadConfig()
	apiKey := cfg.ResendAPIKey

	client := resend.NewClient(apiKey)

	emailDomain := "gdsc-csusm.com" // Needs to be verified to be able to send email or use verified Domain
	URL := "https://gdsc-csusm.com/verify"

	params := &resend.SendEmailRequest{
		From: fmt.Sprintf("CSUSM_GDSC <CSUSM_GDSC@%s>", emailDomain),
		To:   []string{userEmail},
		Html: fmt.Sprintf(`
			<p>Hello %s,</p>
			<p>Welcome to GDSC-CSUSM! Please verify your email by clicking the button below:</p>
			<p>
				<a href="%s?token=%s" style="
					display: inline-block;
					padding: 10px 20px;
					font-size: 16px;
					color: #fff;
					background-color: #007bff;
					text-decoration: none;
					border-radius: 5px;">
					Verify Email
				</a>
			</p>
			<p>If you didn’t request this, please ignore this email.</p>
			<p>Best,<br>GDSC-CSUSM Team</p>
		`, userEmail, URL, verificationToken),
		Subject: "Verify Your Email for GDSC-CSUSM",
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		return err
	}
	return nil
}
