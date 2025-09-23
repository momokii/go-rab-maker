package models

type Project struct {
	ProjectId   int    `json:"project_id"`
	UserId      int    `json:"user_id"`
	ProjectName string `json:"project_name"`
	Location    string `json:"location"`
	ClientName  string `json:"client_name"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ProjectCreate struct {
	ProjectName string `json:"project_name" validate:"required,min=3,max=100"`
	Location    string `json:"location" validate:"required,min=3,max=100"`
	ClientName  string `json:"client_name" validate:"required,min=3,max=100"`
	UserId      int    `json:"user_id"`
}
