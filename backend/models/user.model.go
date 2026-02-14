package models

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserId    int            `json:"user_id"`
	Username  string         `json:"username"`
	Password  string         `json:"password"` // This will store the hashed password
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
	DeletedAt sql.NullString `json:"deleted_at,omitempty"` // Soft delete timestamp
}

type UserCreate struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=100"` // Plain text password that will be hashed
}

type UserLogin struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword verifies a plain text password against a hashed password
func CheckPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
