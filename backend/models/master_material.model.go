package models

type MasterMaterial struct {
	MaterialId       int     `json:"material_id"`
	UserId           int     `json:"user_id"`
	MaterialName     string  `json:"material_name"`
	Unit             string  `json:"unit"`
	DefaultUnitPrice float64 `json:"default_unit_price"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type MasterMaterialCreate struct {
	MaterialName     string  `json:"material_name" validate:"required,min=1,max=100"`
	UserId           int     `json:"user_id"`
	Unit             string  `json:"unit" validate:"required,min=1,max=20"`
	DefaultUnitPrice float64 `json:"default_unit_price" validate:"required,gte=0"`
}
