package models

type ItemType string

const (
	PROJECT_ITEM_TYPE_MATERIAL ItemType = "material"
	PROJECT_ITEM_TYPE_LABOR    ItemType = "labor"
)

type ProjectItemCost struct {
	CostId              int     `json:"cost_id"`
	WorkItemId          int     `json:"work_item_id"`
	ItemType            string  `json:"item_type"` // "material" or "labor"
	ItemId              int     `json:"item_id"`
	QuantityNeeded      float64 `json:"quantity_needed"`
	UnitPriceAtCreation float64 `json:"unit_price_at_creation"`
	TotalCost           float64 `json:"total_cost"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

type ProjectItemCostCreate struct {
	WorkItemId          int     `json:"work_item_id" validate:"required"`
	ItemType            string  `json:"item_type" validate:"required,oneof=material labor"`
	ItemId              int     `json:"item_id" validate:"required"`
	QuantityNeeded      float64 `json:"quantity_needed" validate:"required,gt=0"`
	UnitPriceAtCreation float64 `json:"unit_price_at_creation" validate:"required,gt=0"`
	TotalCost           float64 `json:"total_cost" validate:"required,gt=0"`
}
