package models

type ItemType string

const (
	PROJECT_ITEM_TYPE_MATERIAL ItemType = "MATERIAL"
	PROJECT_ITEM_TYPE_LABOR    ItemType = "LABOR"
)

type ProjectItemCost struct {
	CostId              int     `json:"cost_id"`
	WorkItemId          int     `json:"work_item_id"`
	ItemType            string  `json:"item_type"` // "material" or "labor"
	ItemId              int     `json:"item_id"`
	ItemName            string  `json:"item_name"`
	Coefficient         float64 `json:"coefficient"`
	QuantityNeeded      float64 `json:"quantity_needed"`
	UnitPriceAtCreation float64 `json:"unit_price_at_creation"`
	TotalCost           float64 `json:"total_cost"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

type ProjectItemCostCreate struct {
	WorkItemId          int     `json:"work_item_id" validate:"required"`
	ItemType            string  `json:"item_type" validate:"required,oneof=MATERIAL LABOR"`
	MasterItemId        int     `json:"master_item_id" validate:"required"`
	ItemName            string  `json:"item_name"`
	Coefficient         float64 `json:"coefficient"`
	QuantityNeeded      float64 `json:"quantity_needed" validate:"required,gt=0"`
	UnitPriceAtCreation float64 `json:"unit_price_at_creation" validate:"required,gt=0"`
	TotalCost           float64 `json:"total_cost" validate:"required,gt=0"`
}

type ProjectItemCostWithDetails struct {
	CostId              int     `json:"cost_id"`
	WorkItemId          int     `json:"work_item_id"`
	ItemType            string  `json:"item_type"` // "material" or "labor"
	ItemId              int     `json:"item_id"`
	ItemName            string  `json:"item_name"`
	Coefficient         float64 `json:"coefficient"`
	QuantityNeeded      float64 `json:"quantity_needed"`
	UnitPriceAtCreation float64 `json:"unit_price_at_creation"`
	TotalCost           float64 `json:"total_cost"`
	Unit                string  `json:"unit"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
	WorkItemDescription string  `json:"work_item_description"`
}

type MaterialSummary struct {
	ItemId        int     `json:"item_id"`
	ItemName      string  `json:"item_name"`
	TotalQuantity float64 `json:"total_quantity"`
	Unit          string  `json:"unit"`
	ItemType      string  `json:"item_type"`
	TotalCost     float64 `json:"total_cost"`
}

type DetailedMaterialSummary struct {
	ItemId            int                 `json:"item_id"`
	ItemName          string              `json:"item_name"`
	TotalQuantity     float64             `json:"total_quantity"`
	Unit              string              `json:"unit"`
	ItemType          string              `json:"item_type"`
	TotalCost         float64             `json:"total_cost"`
	WorkItemBreakdown []WorkItemBreakdown `json:"work_item_breakdown"`
}

type WorkItemBreakdown struct {
	WorkItemId   int     `json:"work_item_id"`
	WorkItemDesc string  `json:"work_item_desc"`
	Quantity     float64 `json:"quantity"`
	Cost         float64 `json:"cost"`
	Volume       float64 `json:"volume"`
	Coefficient  float64 `json:"coefficient"`
}
