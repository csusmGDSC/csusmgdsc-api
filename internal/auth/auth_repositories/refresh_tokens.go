package auth_repositories

import (
	"database/sql"

	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_models"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(refresh_token *auth_models.CreateSessionRequest) error {
	query := `
	    INSERT INTO refresh_tokens(
		user_id, token, issued_at, expires_at, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query,
		refresh_token.UserID,
		refresh_token.Token,
		refresh_token.IssuedAt,
		refresh_token.ExpiresAt,
		refresh_token.IPAddress,
		refresh_token.UserAgent,
	)
	return err
}

func (r *RefreshTokenRepository) GetByToken(cookieToken string) (*auth_models.RefreshToken, error) {
	refreshToken := &auth_models.RefreshToken{}
	query := `
	    SELECT token, user_id, expires_at FROM refresh_tokens WHERE token = $1
	`
	err := r.db.QueryRow(query, cookieToken).Scan(
		&refreshToken.Token,
		&refreshToken.UserID,
		&refreshToken.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	return refreshToken, nil
}

func (r *RefreshTokenRepository) DeleteByToken(cookieToken string) error {
	query := `
		DELETE FROM refresh_tokens WHERE token = $1
	`
	_, err := r.db.Exec(query, cookieToken)
	if err != nil {
		return err
	}
	return nil
}

func (r *RefreshTokenRepository) DeleteAllByUserID(userID string) error {
	query := `
		DELETE FROM refresh_tokens WHERE user_id = $1
	`
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return err
	}
	return nil
}
