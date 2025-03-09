package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/marchelhutagalung/go-service/internal/database"
	"github.com/marchelhutagalung/go-service/internal/models"
	"time"
)

// Common errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserRepository handles database operations related to users
type UserRepository struct {
	db *database.PostgresDB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *database.PostgresDB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create creates a new user in the database
func (r *UserRepository) Create(ctx context.Context, input *models.CreateUserInput, passwordHash string) (*models.User, error) {
	// Check if email already exists
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", input.Email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Create the user
	query := `
        INSERT INTO users (email, password_hash, first_name, last_name, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $5)
        RETURNING id, email, password_hash, first_name, last_name, created_at, updated_at
    `

	now := time.Now()
	user := &models.User{}

	err = r.db.QueryRowContext(
		ctx, query, input.Email, passwordHash, input.FirstName, input.LastName, now,
	).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
        SELECT id, email, password_hash, first_name, last_name, created_at, updated_at
        FROM users
        WHERE id = $1
    `

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, email, password_hash, first_name, last_name, created_at, updated_at
        FROM users
        WHERE email = $1
    `

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// Authenticate verifies a user's credentials and returns the user if valid
func (r *UserRepository) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	user, err := r.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check password
	if !models.CheckPasswordHash(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
