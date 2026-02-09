package project_item_costs

import (
	"database/sql"
	"time"

	"github.com/momokii/go-rab-maker/backend/models"
)

type ProjectItemCostsRepo struct{}

func NewProjectItemCostsRepo() *ProjectItemCostsRepo {
	return &ProjectItemCostsRepo{}
}

// FindById retrieves a project item cost by its ID
func (r *ProjectItemCostsRepo) FindById(tx *sql.Tx, projectItemCostId int) (models.ProjectItemCost, error) {
	query := `
		SELECT
			cost_id, work_item_id, item_type, master_item_id, item_name,
			coefficient, quantity_needed, unit_price_at_creation, total_cost,
			created_at, updated_at
		FROM project_item_costs
		WHERE cost_id = ?
	`

	var cost models.ProjectItemCost
	err := tx.QueryRow(query, projectItemCostId).Scan(
		&cost.CostId,
		&cost.WorkItemId,
		&cost.ItemType,
		&cost.ItemId,
		&cost.ItemName,
		&cost.Coefficient,
		&cost.QuantityNeeded,
		&cost.UnitPriceAtCreation,
		&cost.TotalCost,
		&cost.CreatedAt,
		&cost.UpdatedAt,
	)

	if err != nil {
		return models.ProjectItemCost{}, err
	}

	return cost, nil
}

// FindByWorkItemId retrieves all item costs for a specific work item
func (r *ProjectItemCostsRepo) FindByWorkItemId(tx *sql.Tx, workItemId int) ([]models.ProjectItemCostWithDetails, error) {
	query := `
		SELECT
			pic.cost_id, pic.work_item_id, pic.item_type, pic.master_item_id, pic.item_name,
			pic.coefficient, pic.quantity_needed, pic.unit_price_at_creation, pic.total_cost,
			pic.created_at, pic.updated_at,
			CASE
				WHEN pic.item_type = 'MATERIAL' THEN mm.unit
				WHEN pic.item_type = 'LABOR' THEN mlt.unit
				ELSE ''
			END as unit
		FROM project_item_costs pic
		LEFT JOIN master_materials mm ON pic.item_type = 'MATERIAL' AND pic.master_item_id = mm.material_id
		LEFT JOIN master_labor_types mlt ON pic.item_type = 'LABOR' AND pic.master_item_id = mlt.labor_type_id
		WHERE pic.work_item_id = ?
		ORDER BY pic.item_type, pic.item_name
	`

	rows, err := tx.Query(query, workItemId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var costs []models.ProjectItemCostWithDetails
	for rows.Next() {
		var cost models.ProjectItemCostWithDetails
		err := rows.Scan(
			&cost.CostId,
			&cost.WorkItemId,
			&cost.ItemType,
			&cost.ItemId,
			&cost.ItemName,
			&cost.Coefficient,
			&cost.QuantityNeeded,
			&cost.UnitPriceAtCreation,
			&cost.TotalCost,
			&cost.CreatedAt,
			&cost.UpdatedAt,
			&cost.Unit,
		)
		if err != nil {
			return nil, err
		}
		costs = append(costs, cost)
	}

	return costs, nil
}

// FindByProjectId retrieves all item costs for a specific project
func (r *ProjectItemCostsRepo) FindByProjectId(tx *sql.Tx, projectId int) ([]models.ProjectItemCostWithDetails, error) {
	query := `
		SELECT
			pic.cost_id, pic.work_item_id, pic.item_type, pic.master_item_id, pic.item_name,
			pic.coefficient, pic.quantity_needed, pic.unit_price_at_creation, pic.total_cost,
			pic.created_at, pic.updated_at,
			pwi.description as work_item_description,
			CASE
				WHEN pic.item_type = 'MATERIAL' THEN mm.unit
				WHEN pic.item_type = 'LABOR' THEN mlt.unit
				ELSE ''
			END as unit
		FROM project_item_costs pic
		JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
		LEFT JOIN master_materials mm ON pic.item_type = 'MATERIAL' AND pic.master_item_id = mm.material_id
		LEFT JOIN master_labor_types mlt ON pic.item_type = 'LABOR' AND pic.master_item_id = mlt.labor_type_id
		WHERE pwi.project_id = ?
		ORDER BY pwi.description, pic.item_type, pic.item_name
	`

	rows, err := tx.Query(query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var costs []models.ProjectItemCostWithDetails
	for rows.Next() {
		var cost models.ProjectItemCostWithDetails
		err := rows.Scan(
			&cost.CostId,
			&cost.WorkItemId,
			&cost.ItemType,
			&cost.ItemId,
			&cost.ItemName,
			&cost.Coefficient,
			&cost.QuantityNeeded,
			&cost.UnitPriceAtCreation,
			&cost.TotalCost,
			&cost.CreatedAt,
			&cost.UpdatedAt,
			&cost.WorkItemDescription,
			&cost.Unit,
		)
		if err != nil {
			return nil, err
		}
		costs = append(costs, cost)
	}

	return costs, nil
}

// Create inserts a new project item cost
func (r *ProjectItemCostsRepo) Create(tx *sql.Tx, cost models.ProjectItemCostCreate) error {
	query := `
		INSERT INTO project_item_costs
		(work_item_id, item_type, master_item_id, item_name, coefficient,
		 quantity_needed, unit_price_at_creation, total_cost, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := tx.Exec(
		query,
		cost.WorkItemId,
		cost.ItemType,
		cost.MasterItemId,
		cost.ItemName,
		cost.Coefficient,
		cost.QuantityNeeded,
		cost.UnitPriceAtCreation,
		cost.TotalCost,
		now,
		now,
	)

	return err
}

// CreateMultiple inserts multiple project item costs in a single transaction
func (r *ProjectItemCostsRepo) CreateMultiple(tx *sql.Tx, costs []models.ProjectItemCostCreate) error {
	if len(costs) == 0 {
		return nil
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	for _, cost := range costs {
		query := `
			INSERT INTO project_item_costs
			(work_item_id, item_type, master_item_id, item_name, coefficient,
			 quantity_needed, unit_price_at_creation, total_cost, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		_, err := tx.Exec(
			query,
			cost.WorkItemId,
			cost.ItemType,
			cost.MasterItemId,
			cost.ItemName,
			cost.Coefficient,
			cost.QuantityNeeded,
			cost.UnitPriceAtCreation,
			cost.TotalCost,
			now,
			now,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteByWorkItemId deletes all item costs for a specific work item
func (r *ProjectItemCostsRepo) DeleteByWorkItemId(tx *sql.Tx, workItemId int) error {
	query := `DELETE FROM project_item_costs WHERE work_item_id = ?`
	_, err := tx.Exec(query, workItemId)
	return err
}

// GetMaterialSummaryByProjectId retrieves a summary of all materials needed for a project
func (r *ProjectItemCostsRepo) GetMaterialSummaryByProjectId(tx *sql.Tx, projectId int) ([]models.MaterialSummary, error) {
	query := `
		SELECT
			pic.master_item_id, pic.item_name,
			SUM(pic.quantity_needed) as total_quantity,
			CASE
				WHEN pic.item_type = 'MATERIAL' THEN mm.unit
				WHEN pic.item_type = 'LABOR' THEN mlt.unit
				ELSE ''
			END as unit,
			pic.item_type,
			SUM(pic.total_cost) as total_cost
		FROM project_item_costs pic
		JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
		LEFT JOIN master_materials mm ON pic.item_type = 'MATERIAL' AND pic.master_item_id = mm.material_id
		LEFT JOIN master_labor_types mlt ON pic.item_type = 'LABOR' AND pic.master_item_id = mlt.labor_type_id
		WHERE pwi.project_id = ?
		GROUP BY pic.master_item_id, pic.item_name,
		         CASE
		             WHEN pic.item_type = 'MATERIAL' THEN mm.unit
		             WHEN pic.item_type = 'LABOR' THEN mlt.unit
		             ELSE ''
		         END,
		         pic.item_type
		ORDER BY pic.item_type, pic.item_name
	`

	rows, err := tx.Query(query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summary []models.MaterialSummary
	for rows.Next() {
		var item models.MaterialSummary
		err := rows.Scan(
			&item.ItemId,
			&item.ItemName,
			&item.TotalQuantity,
			&item.Unit,
			&item.ItemType,
			&item.TotalCost,
		)
		if err != nil {
			return nil, err
		}
		summary = append(summary, item)
	}

	return summary, nil
}

// GetDetailedMaterialSummaryByProjectId retrieves a detailed summary with work item breakdown
func (r *ProjectItemCostsRepo) GetDetailedMaterialSummaryByProjectId(tx *sql.Tx, projectId int) ([]models.DetailedMaterialSummary, error) {
	// First get all unique items for the project
	query := `
		SELECT DISTINCT
			pic.master_item_id, pic.item_name,
			CASE
				WHEN pic.item_type = 'MATERIAL' THEN mm.unit
				WHEN pic.item_type = 'LABOR' THEN mlt.unit
				ELSE ''
			END as unit,
			pic.item_type,
			SUM(pic.total_cost) as total_cost,
			SUM(pic.quantity_needed) as total_quantity
		FROM project_item_costs pic
		JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
		LEFT JOIN master_materials mm ON pic.item_type = 'MATERIAL' AND pic.master_item_id = mm.material_id
		LEFT JOIN master_labor_types mlt ON pic.item_type = 'LABOR' AND pic.master_item_id = mlt.labor_type_id
		WHERE pwi.project_id = ?
		GROUP BY pic.master_item_id, pic.item_name,
		         CASE
		             WHEN pic.item_type = 'MATERIAL' THEN mm.unit
		             WHEN pic.item_type = 'LABOR' THEN mlt.unit
		             ELSE ''
		         END,
		         pic.item_type
		ORDER BY pic.item_type, pic.item_name
	`

	rows, err := tx.Query(query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var detailedSummary []models.DetailedMaterialSummary
	for rows.Next() {
		var item models.DetailedMaterialSummary
		err := rows.Scan(
			&item.ItemId,
			&item.ItemName,
			&item.Unit,
			&item.ItemType,
			&item.TotalCost,
			&item.TotalQuantity,
		)
		if err != nil {
			return nil, err
		}

		// Get work item breakdown for this item
		workItemBreakdown, err := r.getWorkItemBreakdownForItem(tx, projectId, item.ItemId, item.ItemType)
		if err != nil {
			return nil, err
		}
		item.WorkItemBreakdown = workItemBreakdown

		detailedSummary = append(detailedSummary, item)
	}

	return detailedSummary, nil
}

// getWorkItemBreakdownForItem gets the breakdown of a specific item across work items
func (r *ProjectItemCostsRepo) getWorkItemBreakdownForItem(tx *sql.Tx, projectId, itemId int, itemType string) ([]models.WorkItemBreakdown, error) {
	query := `
		SELECT
			pic.work_item_id,
			pwi.description,
			pic.quantity_needed,
			pic.total_cost,
			pwi.volume,
			pic.coefficient
		FROM project_item_costs pic
		JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
		WHERE pwi.project_id = ? AND pic.master_item_id = ? AND pic.item_type = ?
		ORDER BY pwi.created_at
	`

	rows, err := tx.Query(query, projectId, itemId, itemType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var breakdown []models.WorkItemBreakdown
	for rows.Next() {
		var item models.WorkItemBreakdown
		err := rows.Scan(
			&item.WorkItemId,
			&item.WorkItemDesc,
			&item.Quantity,
			&item.Cost,
			&item.Volume,
			&item.Coefficient,
		)
		if err != nil {
			return nil, err
		}
		breakdown = append(breakdown, item)
	}

	return breakdown, nil
}
