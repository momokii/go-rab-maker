package users

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type UsersRepo struct{}

func NewUsersRepo() *UsersRepo {
	return &UsersRepo{}
}

// FindById retrieves a user by their ID
func (r *UsersRepo) FindById(tx *sql.Tx, userId int) (models.User, error) {
	var user models.User

	query := "SELECT user_id, username, password, created_at, updated_at FROM users WHERE user_id = ?"
	if err := tx.QueryRow(
		query,
		userId,
	).Scan(
		&user.UserId,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return user, err
	}

	return user, nil
}

// FindByUsername retrieves a user by their username
func (r *UsersRepo) FindByUsername(tx *sql.Tx, username string) (models.User, error) {
	var user models.User

	query := "SELECT user_id, username, password, created_at, updated_at FROM users WHERE username = ?"
	if err := tx.QueryRow(
		query,
		username,
	).Scan(
		&user.UserId,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return user, err
	}

	return user, nil
}

// Create creates a new user in the database
func (r *UsersRepo) Create(tx *sql.Tx, userData models.UserCreate) error {

	query := "INSERT INTO users (username, password) VALUES (?, ?)"
	if _, err := tx.Exec(
		query,
		userData.Username,
		userData.Password,
	); err != nil {
		return err
	}

	return nil
}

// Update updates an existing user's information
func (r *UsersRepo) Update(tx *sql.Tx, userData models.User) error {

	query := "UPDATE users SET username = ?, password = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?"
	if _, err := tx.Exec(
		query,
		userData.Username,
		userData.Password,
		userData.UserId,
	); err != nil {
		return err
	}

	return nil
}
