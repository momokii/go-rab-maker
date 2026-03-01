package dashboard

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type DashboardRepo struct{}

func NewDashboardRepo() *DashboardRepo {
	return &DashboardRepo{}
}

// GetProjectsTotalCost gets total cost for all user's projects
func (r *DashboardRepo) GetProjectsTotalCost(tx *sql.Tx, userId int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(total_cost), 0) as total
		FROM (
			SELECT SUM(pic.total_cost) as total_cost
			FROM project_item_costs pic
			JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
			JOIN projects p ON pwi.project_id = p.project_id
			WHERE p.user_id = ?
		) as costs
	`

	var total float64
	if err := tx.QueryRow(query, userId).Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}

// GetProjectCount gets total number of projects for a user
func (r *DashboardRepo) GetProjectCount(tx *sql.Tx, userId int) (int, error) {
	query := "SELECT COUNT(*) FROM projects WHERE user_id = ?"

	var count int
	if err := tx.QueryRow(query, userId).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// GetRecentProjects gets recent projects for a user
func (r *DashboardRepo) GetRecentProjects(tx *sql.Tx, userId int, limit int) ([]models.Project, error) {
	query := `
		SELECT project_id, user_id, project_name, location, client_name, created_at, updated_at
		FROM projects
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := tx.Query(query, userId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		if err := rows.Scan(
			&project.ProjectId,
			&project.UserId,
			&project.ProjectName,
			&project.Location,
			&project.ClientName,
			&project.CreatedAt,
			&project.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// GetWorkItemsCount gets total count of work items for a user
func (r *DashboardRepo) GetWorkItemsCount(tx *sql.Tx, userId int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM project_work_items pwi
		JOIN projects p ON pwi.project_id = p.project_id
		WHERE p.user_id = ?
	`

	var count int
	if err := tx.QueryRow(query, userId).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// GetTypeCostBreakdown gets cost breakdown by item type (Material vs Labor)
func (r *DashboardRepo) GetTypeCostBreakdown(tx *sql.Tx, userId int) ([]models.TypeCostBreakdown, error) {
	query := `
		SELECT pic.item_type, COALESCE(SUM(pic.total_cost), 0) as total_cost
		FROM project_item_costs pic
		JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
		JOIN projects p ON pwi.project_id = p.project_id
		WHERE p.user_id = ?
		GROUP BY pic.item_type
	`

	rows, err := tx.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.TypeCostBreakdown
	for rows.Next() {
		var item models.TypeCostBreakdown
		if err := rows.Scan(&item.ItemType, &item.TotalCost); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}

// GetCategoryBreakdown gets statistics grouped by work category
func (r *DashboardRepo) GetCategoryBreakdown(tx *sql.Tx, userId int, limit int) ([]models.CategoryBreakdown, error) {
	query := `
		SELECT c.category_id, c.category_name,
		       COUNT(DISTINCT pwi.work_item_id) as item_count,
		       COALESCE(SUM(pwi.volume), 0) as total_volume,
		       COALESCE(SUM(pic.total_cost), 0) as total_cost
		FROM project_work_items pwi
		JOIN master_work_categories c ON pwi.category_id = c.category_id
		JOIN projects p ON pwi.project_id = p.project_id
		LEFT JOIN project_item_costs pic ON pwi.work_item_id = pic.work_item_id
		WHERE p.user_id = ?
		GROUP BY c.category_id, c.category_name
		ORDER BY total_cost DESC
	`

	if limit > 0 {
		query += " LIMIT ?"
		rows, err := tx.Query(query, userId, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var results []models.CategoryBreakdown
		for rows.Next() {
			var item models.CategoryBreakdown
			if err := rows.Scan(&item.CategoryID, &item.CategoryName, &item.ItemCount, &item.TotalVolume, &item.TotalCost); err != nil {
				return nil, err
			}
			results = append(results, item)
		}

		return results, nil
	}

	rows, err := tx.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.CategoryBreakdown
	for rows.Next() {
		var item models.CategoryBreakdown
		if err := rows.Scan(&item.CategoryID, &item.CategoryName, &item.ItemCount, &item.TotalVolume, &item.TotalCost); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}

// GetTopExpensiveItems gets the most expensive items across all projects
func (r *DashboardRepo) GetTopExpensiveItems(tx *sql.Tx, userId int, limit int) ([]models.TopExpensiveItem, error) {
	query := `
		SELECT pic.item_name,
		       pic.item_type,
		       COALESCE(SUM(pic.total_cost), 0) as total_cost,
		       COALESCE(SUM(pic.quantity_needed), 0) as total_quantity,
		       COALESCE(m.unit, l.unit, pic.unit) as unit
		FROM project_item_costs pic
		JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
		JOIN projects p ON pwi.project_id = p.project_id
		LEFT JOIN master_materials m ON pic.master_item_id = m.material_id AND pic.item_type = 'MATERIAL'
		LEFT JOIN master_labor_types l ON pic.master_item_id = l.labor_type_id AND pic.item_type = 'LABOR'
		WHERE p.user_id = ?
		GROUP BY pic.item_name, pic.item_type, COALESCE(m.unit, l.unit, pic.unit)
		ORDER BY total_cost DESC
		LIMIT ?
	`

	rows, err := tx.Query(query, userId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.TopExpensiveItem
	for rows.Next() {
		var item models.TopExpensiveItem
		if err := rows.Scan(&item.ItemName, &item.ItemType, &item.TotalCost, &item.TotalQty, &item.Unit); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}

// GetProjectBreakdown gets cost breakdown by project for Material Summary
func (r *DashboardRepo) GetProjectBreakdown(tx *sql.Tx, userId int) ([]models.ProjectBreakdown, error) {
	query := `
		SELECT p.project_id, p.project_name,
		       COUNT(DISTINCT pwi.work_item_id) as work_item_count,
		       COALESCE(SUM(CASE WHEN pic.item_type = 'MATERIAL' THEN pic.total_cost ELSE 0 END), 0) as material_cost,
		       COALESCE(SUM(CASE WHEN pic.item_type = 'LABOR' THEN pic.total_cost ELSE 0 END), 0) as labor_cost,
		       COALESCE(SUM(pic.total_cost), 0) as total_cost
		FROM projects p
		LEFT JOIN project_work_items pwi ON p.project_id = pwi.project_id
		LEFT JOIN project_item_costs pic ON pwi.work_item_id = pic.work_item_id
		WHERE p.user_id = ?
		GROUP BY p.project_id, p.project_name
		ORDER BY total_cost DESC
	`

	rows, err := tx.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.ProjectBreakdown
	for rows.Next() {
		var item models.ProjectBreakdown
		if err := rows.Scan(&item.ProjectID, &item.ProjectName, &item.WorkItemCount, &item.MaterialCost, &item.LaborCost, &item.TotalCost); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}

// GetEnhancedRecentProjects gets recent projects with additional statistics
func (r *DashboardRepo) GetEnhancedRecentProjects(tx *sql.Tx, userId int, limit int) ([]models.EnhancedProjectData, error) {
	query := `
		SELECT p.project_id, p.project_name, p.location, p.client_name, p.created_at, p.updated_at,
		       COUNT(DISTINCT pwi.work_item_id) as work_item_count,
		       COALESCE(SUM(pic.total_cost), 0) as total_cost
		FROM projects p
		LEFT JOIN project_work_items pwi ON p.project_id = pwi.project_id
		LEFT JOIN project_item_costs pic ON pwi.work_item_id = pic.work_item_id
		WHERE p.user_id = ?
		GROUP BY p.project_id, p.project_name, p.location, p.client_name, p.created_at, p.updated_at
		ORDER BY p.created_at DESC
		LIMIT ?
	`

	rows, err := tx.Query(query, userId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.EnhancedProjectData
	for rows.Next() {
		var item models.EnhancedProjectData
		if err := rows.Scan(
			&item.ProjectID,
			&item.ProjectName,
			&item.Location,
			&item.ClientName,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.WorkItemCount,
			&item.TotalCost,
		); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}

// GetMaterialSummaryStats gets overall statistics for Material Summary page
func (r *DashboardRepo) GetMaterialSummaryStats(tx *sql.Tx, userId int) (models.MaterialSummaryStats, error) {
	query := `
		SELECT COUNT(DISTINCT CONCAT(pic.master_item_id, '_', pic.item_type)) as total_items,
		       COALESCE(SUM(CASE WHEN pic.item_type = 'MATERIAL' THEN pic.total_cost ELSE 0 END), 0) as material_cost,
		       COALESCE(SUM(CASE WHEN pic.item_type = 'LABOR' THEN pic.total_cost ELSE 0 END), 0) as labor_cost,
		       COALESCE(SUM(pic.total_cost), 0) as total_cost,
		       COUNT(DISTINCT p.project_id) as unique_projects
		FROM project_item_costs pic
		JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
		JOIN projects p ON pwi.project_id = p.project_id
		WHERE p.user_id = ?
	`

	var stats models.MaterialSummaryStats
	if err := tx.QueryRow(query, userId).Scan(
		&stats.TotalItems,
		&stats.MaterialCost,
		&stats.LaborCost,
		&stats.TotalCost,
		&stats.UniqueProjects,
	); err != nil {
		return stats, err
	}

	return stats, nil
}
