package models

type ProjectWorkItem struct {
	WorkItemId     int     `json:"work_item_id"`
	ProjectId      int     `json:"project_id"`
	CategoryId     int     `json:"category_id"`
	Description    string  `json:"description"`
	Volume         float64 `json:"volume"`
	Unit           string  `json:"unit"`
	AHSPTemplateId *int    `json:"ahsp_template_id,omitempty"` // Pointer to allow null value
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type ProjectWorkItemCreate struct {
	ProjectId      int     `json:"project_id" validate:"required"`
	CategoryId     int     `json:"category_id" validate:"required"`
	Description    string  `json:"description" validate:"required,min=1,max=255"`
	Volume         float64 `json:"volume" validate:"required,gt=0"`
	Unit           string  `json:"unit" validate:"required,min=1,max=50"`
	AHSPTemplateId *int    `json:"ahsp_template_id,omitempty"` // Pointer to allow null value
}
