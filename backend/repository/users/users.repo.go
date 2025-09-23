package users

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type UsersRepo struct{}

func NewUsersRepo() *UsersRepo {
	return &UsersRepo{}
}

// TODO: Implement FindById [Users]
func (r *UsersRepo) FindById(tx *sql.Tx, userId int) (models.User, error) {

	return models.User{}, nil
}

// TODO: implement Create [Users]
func (r *UsersRepo) Create(tx *sql.Tx, userData models.UserCreate) error {

	return nil
}
