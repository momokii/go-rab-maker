package models

type AHSPTemplate struct {
	TemplateId   int    `json:"template_id"`
	UserId       int    `json:"user_id"`
	TemplateName string `json:"template_name"`
	Unit         string `json:"unit"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type AHSPTemplateCreate struct {
	TemplateName string `json:"template_name" validate:"required,min=1,max=100"`
	UserId       int    `json:"user_id"`
	Unit         string `json:"unit" validate:"required,min=1,max=20"`
}
