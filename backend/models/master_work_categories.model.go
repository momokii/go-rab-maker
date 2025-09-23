package models

type MasterWorkCategory struct {
	CategoryId   int    `json:"category_id"`
	UserId       int    `json:"user_id"`
	CategoryName string `json:"category_name"`
	DisplayOrder int    `json:"display_order"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type MasterWorkCategoryCreate struct {
	CategoryName string `json:"category_name" validate:"required,min=1,max=100"`
	UserId       int    `json:"user_id"`
	DisplayOrder int    `json:"display_order"`
}
