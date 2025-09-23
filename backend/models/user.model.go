package models

type User struct {
	UserId    int    `json:"user_id"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserCreate struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=100,alphanum"`
}
