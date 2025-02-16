package auth_repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/csusmGDSC/csusmgdsc-api/internal/auth/auth_models"
	"github.com/csusmGDSC/csusmgdsc-api/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (
			id, full_name, email, password, provider, auth_id, created_at, updated_at, is_onboarded, email_verified
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(query,
		user.ID,
		user.FullName,
		user.Email,
		user.Password,
		user.Provider,
		user.AuthID,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsOnboarded,
		user.EmailVerified,
	)

	return err
}

func (r *UserRepository) Update(userID string, req auth_models.UpdateUserRequest) error {
	var updates []string
	var args []interface{}
	argIdx := 1
	var firstName, lastName *string

	if req.FirstName != nil {
		updates = append(updates, fmt.Sprintf("first_name = $%d", argIdx))
		args = append(args, *req.FirstName)
		firstName = req.FirstName
		argIdx++
	}
	if req.LastName != nil {
		updates = append(updates, fmt.Sprintf("last_name = $%d", argIdx))
		args = append(args, *req.LastName)
		lastName = req.LastName
		argIdx++
	}
	if req.Image != nil {
		updates = append(updates, fmt.Sprintf("image = $%d", argIdx))
		args = append(args, *req.Image)
		argIdx++
	}
	if req.TotalPoints != nil {
		updates = append(updates, fmt.Sprintf("total_points = $%d", argIdx))
		args = append(args, *req.TotalPoints)
		argIdx++
	}
	if req.Role != nil {
		updates = append(updates, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, *req.Role)
		argIdx++
	}
	if req.Position != nil {
		updates = append(updates, fmt.Sprintf("position = $%d", argIdx))
		args = append(args, *req.Position)
		argIdx++
	}
	if req.Branch != nil {
		updates = append(updates, fmt.Sprintf("branch = $%d", argIdx))
		args = append(args, *req.Branch)
		argIdx++
	}
	if req.Github != nil {
		updates = append(updates, fmt.Sprintf("github = $%d", argIdx))
		args = append(args, *req.Github)
		argIdx++
	}
	if req.Linkedin != nil {
		updates = append(updates, fmt.Sprintf("linkedin = $%d", argIdx))
		args = append(args, *req.Linkedin)
		argIdx++
	}
	if req.Instagram != nil {
		updates = append(updates, fmt.Sprintf("instagram = $%d", argIdx))
		args = append(args, *req.Instagram)
		argIdx++
	}
	if req.Discord != nil {
		updates = append(updates, fmt.Sprintf("discord = $%d", argIdx))
		args = append(args, *req.Discord)
		argIdx++
	}
	if req.Bio != nil {
		updates = append(updates, fmt.Sprintf("bio = $%d", argIdx))
		args = append(args, *req.Bio)
		argIdx++
	}
	if req.Tags != nil {
		updates = append(updates, fmt.Sprintf("tags = $%d", argIdx))
		args = append(args, pq.Array(req.Tags))
		argIdx++
	}
	if req.Website != nil {
		updates = append(updates, fmt.Sprintf("website = $%d", argIdx))
		args = append(args, *req.Website)
		argIdx++
	}
	if req.GraduationDate != nil {
		updates = append(updates, fmt.Sprintf("graduation_date = $%d", argIdx))
		args = append(args, *req.GraduationDate)
		argIdx++
	}
	if req.IsOnboarded != nil {
		updates = append(updates, fmt.Sprintf("is_onboarded = $%d", argIdx))
		args = append(args, *req.IsOnboarded)
		argIdx++
	}
	if req.EmailVerified != nil {
		updates = append(updates, fmt.Sprintf("email_verified = $%d", argIdx))
		args = append(args, *req.EmailVerified)
		argIdx++
	}

	// Update full_name if first_name or last_name is updated
	if firstName != nil || lastName != nil {
		updates = append(updates, fmt.Sprintf("full_name = $%d", argIdx))

		// Use sql.NullString to handle potential NULL values
		var currentFirstName, currentLastName sql.NullString

		err := r.db.QueryRow("SELECT first_name, last_name FROM users WHERE id = $1", userID).Scan(&currentFirstName, &currentLastName)
		if err != nil {
			return fmt.Errorf("failed to fetch existing names: %w", err)
		}

		// Convert sql.NullString to string safely
		firstNameStr := ""
		lastNameStr := ""

		if currentFirstName.Valid {
			firstNameStr = currentFirstName.String
		}
		if currentLastName.Valid {
			lastNameStr = currentLastName.String
		}

		// Use updated values if available
		if firstName != nil {
			firstNameStr = *firstName
		}
		if lastName != nil {
			lastNameStr = *lastName
		}

		// Construct full_name
		fullName := strings.TrimSpace(fmt.Sprintf("%s %s", firstNameStr, lastNameStr))
		args = append(args, fullName)
		argIdx++
	}

	// If no fields to update, return early
	if len(updates) == 0 {
		return nil
	}

	// Construct the final query
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(updates, ", "), argIdx)
	args = append(args, userID)

	// Execute query
	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, full_name, first_name, last_name, email, password,
		 	role, position, branch, image, github,
			linkedin, instagram, discord, bio, tags, website,
			graduation_date, created_at, updated_at, provider, auth_id
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.FullName,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Position,
		&user.Branch,
		&user.Image,
		&user.Github,
		&user.Linkedin,
		&user.Instagram,
		&user.Discord,
		&user.Bio,
		&user.Tags,
		&user.Website,
		&user.GraduationDate,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Provider,
		&user.AuthID,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByAuthID(authID string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, full_name, first_name, last_name, email, password,
		 	role, position, branch, image, github,
			linkedin, instagram, discord, bio, tags, website,
			graduation_date, created_at, updated_at, provider, auth_id, is_onboarded
		FROM users
		WHERE auth_id = $1
	`

	err := r.db.QueryRow(query, authID).Scan(
		&user.ID,
		&user.FullName,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Position,
		&user.Branch,
		&user.Image,
		&user.Github,
		&user.Linkedin,
		&user.Instagram,
		&user.Discord,
		&user.Bio,
		&user.Tags,
		&user.Website,
		&user.GraduationDate,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Provider,
		&user.AuthID,
		&user.IsOnboarded,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRow(query, email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) AuthIDExists(authID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE auth_id = $1)`
	err := r.db.QueryRow(query, authID).Scan(&exists)
	return exists, err
}

func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, 
			full_name,
			first_name,
			last_name,
			email,
			password,
			image,
			total_points,
			role,
			position,
			branch,
			github,
			linkedin,
			instagram,
			discord,
			bio,
			tags,
			website,
			graduation_date,
			created_at,
			updated_at,
			provider,
			auth_id,
			is_onboarded,
			email_verified,
		FROM users
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.FullName,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.Image,
		&user.TotalPoints,
		&user.Role,
		&user.Position,
		&user.Branch,
		&user.Github,
		&user.Linkedin,
		&user.Instagram,
		&user.Discord,
		&user.Bio,
		&user.Tags,
		&user.Website,
		&user.GraduationDate,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Provider,
		&user.AuthID,
		&user.IsOnboarded,
		&user.EmailVerified,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) DeleteByID(userID string) error {
	query := `
		DELETE FROM users WHERE id = $1
	`

	result, err := r.db.Exec(query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("error checking rows affected")
	}

	return nil
}

func (r *UserRepository) GetAll(pageStr string, limitStr string) (*auth_models.AllUsersResponse, error) {
	// Convrt string parameters to integers with default values
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Calculate offset
	offset := (page - 1) * limit
	query := `
		SELECT 
			id, 
			full_name,
			first_name,
			last_name,
			email,
			password,
			image,
			total_points,
			role,
			position,
			branch,
			github,
			linkedin,
			instagram,
			discord,
			bio,
			tags,
			website,
			graduation_date,
			created_at,
			updated_at,
			provider,
			auth_id,
			is_onboarded,
			email_verified
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.FullName,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Password,
			&user.Image,
			&user.TotalPoints,
			&user.Role,
			&user.Position,
			&user.Branch,
			&user.Github,
			&user.Linkedin,
			&user.Instagram,
			&user.Discord,
			&user.Bio,
			&user.Tags,
			&user.Website,
			&user.GraduationDate,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Provider,
			&user.AuthID,
			&user.IsOnboarded,
			&user.EmailVerified,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	totalCount, err := r.GetTotalCount()
	if err != nil {
		return nil, err
	}

	return &auth_models.AllUsersResponse{
		Users:      users,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
	}, nil
}

func (r *UserRepository) GetTotalCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
