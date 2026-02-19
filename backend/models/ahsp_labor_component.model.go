package models

type AHSPLaborComponent struct {
	ComponentId int     `json:"component_id"`
	TemplateId  int     `json:"template_id"`
	LaborTypeId int     `json:"labor_type_id"`
	Coefficient float64 `json:"coefficient"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type AHSPLaborComponentCreate struct {
	TemplateId  int     `json:"template_id" validate:"required"`
	LaborTypeId int     `json:"labor_type_id" validate:"required"`
	Coefficient float64 `json:"coefficient" validate:"required,gt=0"`
}

type AHSPLaborComponentUpdate struct {
	LaborTypeId int     `json:"labor_type_id" validate:"required"`
	Coefficient float64 `json:"coefficient" validate:"required,gt=0"`
}

type AHSPLaborComponentWithLabor struct {
	ComponentId   int     `json:"component_id"`
	TemplateId    int     `json:"template_id"`
	LaborTypeId   int     `json:"labor_type_id"`
	Coefficient   float64 `json:"coefficient"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	LaborTypeName string  `json:"labor_type_name"`
	LaborUnit     string  `json:"labor_unit"`
	LaborWage     float64 `json:"labor_wage"`
}
