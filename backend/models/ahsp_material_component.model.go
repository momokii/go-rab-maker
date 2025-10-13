package models

type AHSPMaterialComponent struct {
	ComponentId int     `json:"component_id"`
	TemplateId  int     `json:"template_id"`
	MaterialId  int     `json:"material_id"`
	Coefficient float64 `json:"coefficient"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type AHSPMaterialComponentCreate struct {
	TemplateId  int     `json:"template_id" validate:"required"`
	MaterialId  int     `json:"material_id" validate:"required"`
	Coefficient float64 `json:"coefficient" validate:"required,gt=0"`
}

type AHSPMaterialComponentUpdate struct {
	Coefficient float64 `json:"coefficient" validate:"required,gt=0"`
}

type AHSPMaterialComponentWithMaterial struct {
	ComponentId   int     `json:"component_id"`
	TemplateId    int     `json:"template_id"`
	MaterialId    int     `json:"material_id"`
	Coefficient   float64 `json:"coefficient"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	MaterialName  string  `json:"material_name"`
	MaterialUnit  string  `json:"material_unit"`
	MaterialPrice float64 `json:"material_price"`
}
