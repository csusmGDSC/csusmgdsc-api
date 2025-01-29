package auth_repositories

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

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
			id, full_name, email, password, provider, auth_id, created_at, updated_at, is_onboarded
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
	)

	return err
}

// Fields that are specified as NOT NULL in the database must be assigned
// from current user data before updating
func (r *UserRepository) Update(userID string, req auth_models.UpdateUserRequest) (*models.User, error) {
	userUuid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// Get the current user data
	currentUser, err := r.GetByID(userUuid)
	if err != nil {
		return nil, err
	}

	// If currentUser.FirstName, currentUser.LastName, or currentUser.FullName are nil,
	// set to empty strings, this may occur if the user has not been onboarded yet
	// or onboarded incorrectly
	firstName := ""
	lastName := ""
	fullName := ""

	// User has already been onboarded, so use current values
	if currentUser.FirstName != nil {
		firstName = *currentUser.FirstName
	}
	if currentUser.LastName != nil {
		lastName = *currentUser.LastName
	}
	if currentUser.FullName != nil {
		fullName = *currentUser.FullName
	}

	if req.FirstName != nil {
		firstName = *req.FirstName
	}
	if req.LastName != nil {
		lastName = *req.LastName
	}
	if req.FirstName != nil || req.LastName != nil {
		fullName = strings.TrimSpace(firstName + " " + lastName)
	}

	// If position and branch have not been set, default to student and project branch
	position := models.Student
	branch := models.Projects
	if currentUser.Position != nil {
		position = *currentUser.Position
	}
	if currentUser.Branch != nil {
		branch = *currentUser.Branch
	}
	if req.Position != nil {
		position = *req.Position
	}
	if req.Branch != nil {
		branch = *req.Branch
	}

	query := `
        UPDATE users
        SET
            full_name = $1,
            first_name = $2,
            last_name = $3,
            position = $4,
            branch = $5,
            github = $6,
            linkedin = $7,
            instagram = $8,
            discord = $9,
            bio = $10,
            tags = $11,
            website = $12,
            graduation_date = $13,
            updated_at = $14
        WHERE id = $15
        RETURNING id, full_name, first_name, last_name, email, image,
				role, position, branch, github, linkedin,
                instagram, discord, bio, tags, website, graduation_date,
                created_at, updated_at
    `
	now := time.Now()

	// if tags is nil, use the current tags
	tags := currentUser.Tags
	if req.Tags != nil {
		tags = req.Tags
	}

	user := &models.User{}
	err = r.db.QueryRow(
		query,
		fullName,
		firstName,
		lastName,
		position,
		branch,
		req.Github,
		req.Linkedin,
		req.Instagram,
		req.Discord,
		req.Bio,
		tags,
		req.Website,
		req.GraduationDate,
		now,
		userID,
	).Scan(
		&user.ID,
		&user.FullName,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Image,
		&user.Role,
		&user.Position,
		&user.Branch,
		&user.Github,
		&user.Linkedin,
		&user.Instagram,
		&user.Discord,
		&user.Bio,
		pq.Array(&user.Tags),
		&user.Website,
		&user.GraduationDate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
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
		SELECT id, full_name, first_name, last_name, email, 
			role, position, branch, image, github,
			linkedin, instagram, discord, bio, tags, website,
			graduation_date, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.FullName,
		&user.FirstName,
		&user.LastName,
		&user.Email,
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
			email, 
			password,
			full_name,
			first_name,
			last_name,
			image,
			is_onboarded,
			role,
			provider,
			auth_id,
			created_at,
			updated_at
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
			&user.Email,
			&user.Password,
			&user.FullName,
			&user.FirstName,
			&user.LastName,
			&user.Image,
			&user.IsOnboarded,
			&user.Role,
			&user.Provider,
			&user.AuthID,
			&user.CreatedAt,
			&user.UpdatedAt,
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
