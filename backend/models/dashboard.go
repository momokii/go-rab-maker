package models

// DashboardStats represents overall statistics for the main dashboard
type DashboardStats struct {
	TotalProjects     int     `db:"total_projects"`
	TotalWorkItems    int     `db:"total_work_items"`
	TotalCost         float64 `db:"total_cost"`
	ActiveUsersCount  int     `db:"active_users_count"`
}

// TypeCostBreakdown represents cost breakdown by type (Material vs Labor)
type TypeCostBreakdown struct {
	ItemType  string  `db:"item_type"` // "MATERIAL" or "LABOR"
	TotalCost float64 `db:"total_cost"`
}

// CategoryBreakdown represents statistics grouped by work category
type CategoryBreakdown struct {
	CategoryID   int     `db:"category_id"`
	CategoryName string  `db:"category_name"`
	ItemCount    int     `db:"item_count"`
	TotalVolume  float64 `db:"total_volume"`
	TotalCost    float64 `db:"total_cost"`
}

// TopExpensiveItem represents the most expensive items across all projects
type TopExpensiveItem struct {
	ProjectName string  `db:"project_name"`
	ItemName    string  `db:"item_name"`
	ItemType    string  `db:"item_type"`
	TotalCost   float64 `db:"total_cost"`
	TotalQty    float64 `db:"total_quantity"`
	Unit        string  `db:"unit"`
}

// ProjectBreakdown represents cost breakdown by project
type ProjectBreakdown struct {
	ProjectID      int     `db:"project_id"`
	ProjectName    string  `db:"project_name"`
	WorkItemCount  int     `db:"work_item_count"`
	MaterialCost   float64 `db:"material_cost"`
	LaborCost      float64 `db:"labor_cost"`
	TotalCost      float64 `db:"total_cost"`
}

// EnhancedProjectData represents project data with additional statistics
type EnhancedProjectData struct {
	ProjectID        int     `db:"project_id"`
	ProjectName      string  `db:"project_name"`
	Location         string  `db:"location"`
	ClientName       string  `db:"client_name"`
	WorkItemCount    int     `db:"work_item_count"`
	TotalCost        float64 `db:"total_cost"`
	CreatedAt        string  `db:"created_at"`
	UpdatedAt        string  `db:"updated_at"`
}

// MaterialSummaryStats represents statistics for the Material Summary page
type MaterialSummaryStats struct {
	TotalItems      int     `db:"total_items"`
	MaterialCost    float64 `db:"material_cost"`
	LaborCost       float64 `db:"labor_cost"`
	TotalCost       float64 `db:"total_cost"`
	UniqueProjects  int     `db:"unique_projects"`
}

// ProjectMaterialBreakdown represents material/labor breakdown for a specific project
type ProjectMaterialBreakdown struct {
	ProjectID       int              `db:"project_id"`
	ProjectName     string           `db:"project_name"`
	MaterialItems   []MaterialSummary `db:"-"` // Items for this project
	LaborItems      []MaterialSummary `db:"-"` // Items for this project
	MaterialCost    float64          `db:"material_cost"`
	LaborCost       float64          `db:"labor_cost"`
	TotalCost       float64          `db:"total_cost"`
	ItemCount       int              `db:"item_count"`
}
