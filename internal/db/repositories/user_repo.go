package repositories

import (
	"database/sql"
	"strings"
	"time"

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
			id, full_name, first_name, last_name, email, password, 
			role, position, branch, graduation_date,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(query,
		user.ID,
		user.FullName,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
		user.Role,
		user.Position,
		user.Branch,
		user.GraduationDate,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// Fields that are specified as NOT NULL in the database must be assigned
// from current user data before updating
func (r *UserRepository) Update(userID string, req models.UpdateUserRequest) (*models.User, error) {
	userUuid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// Get the current user data
	currentUser, err := r.GetByID(userUuid)
	if err != nil {
		return nil, err
	}

	// Fullname has to be updated if firstName or lastName are provided
	firstName := currentUser.FirstName
	if req.FirstName != nil {
		firstName = *req.FirstName
	}

	lastName := currentUser.LastName
	if req.LastName != nil {
		lastName = *req.LastName
	}

	fullName := currentUser.FullName
	if req.FirstName != nil || req.LastName != nil {
		fullName = strings.TrimSpace(firstName + " " + lastName)
	}

	position := currentUser.Position
	if req.Position != nil {
		position = *req.Position
	}

	branch := currentUser.Branch
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
			graduation_date, created_at, updated_at
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
