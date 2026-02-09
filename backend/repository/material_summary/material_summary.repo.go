package material_summary

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type MaterialSummaryRepo struct{}

func NewMaterialSummaryRepo() *MaterialSummaryRepo {
	return &MaterialSummaryRepo{}
}

// GetAllMaterialsSummary gets material summary across all projects for a user
func (r *MaterialSummaryRepo) GetAllMaterialsSummary(tx *sql.Tx, userId int) ([]models.MaterialSummary, error) {
	// Get materials
	materialsQuery := `
		SELECT
			m.material_id as item_id,
			m.material_name as item_name,
			SUM(pic.quantity_needed) as total_quantity,
			m.unit,
			'MATERIAL' as item_type,
			SUM(pic.total_cost) as total_cost
		FROM project_item_costs pic
		JOIN master_materials m ON pic.master_item_id = m.material_id
		JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
		JOIN projects p ON pwi.project_id = p.project_id
		WHERE p.user_id = ? AND pic.item_type = 'MATERIAL'
		GROUP BY m.material_id, m.material_name, m.unit
	`

	// Get labor
	laborQuery := `
		SELECT
			lt.labor_type_id as item_id,
			lt.role_name as item_name,
			SUM(pic.quantity_needed) as total_quantity,
			lt.unit,
			'LABOR' as item_type,
			SUM(pic.total_cost) as total_cost
		FROM project_item_costs pic
		JOIN master_labor_types lt ON pic.master_item_id = lt.labor_type_id
		JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
		JOIN projects p ON pwi.project_id = p.project_id
		WHERE p.user_id = ? AND pic.item_type = 'LABOR'
		GROUP BY lt.labor_type_id, lt.role_name, lt.unit
	`

	// Execute materials query
	materialsRows, err := tx.Query(materialsQuery, userId)
	if err != nil {
		return nil, err
	}
	defer materialsRows.Close()

	var summaries []models.MaterialSummary
	for materialsRows.Next() {
		var summary models.MaterialSummary
		if err := materialsRows.Scan(
			&summary.ItemId,
			&summary.ItemName,
			&summary.TotalQuantity,
			&summary.Unit,
			&summary.ItemType,
			&summary.TotalCost,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	// Execute labor query
	laborRows, err := tx.Query(laborQuery, userId)
	if err != nil {
		return nil, err
	}
	defer laborRows.Close()

	for laborRows.Next() {
		var summary models.MaterialSummary
		if err := laborRows.Scan(
			&summary.ItemId,
			&summary.ItemName,
			&summary.TotalQuantity,
			&summary.Unit,
			&summary.ItemType,
			&summary.TotalCost,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// GetProjectMaterialSummary gets material summary for a specific project
func (r *MaterialSummaryRepo) GetProjectMaterialSummary(tx *sql.Tx, projectId int) ([]models.MaterialSummary, error) {
	// Get materials
	materialsQuery := `
		SELECT
			m.material_id as item_id,
			m.material_name as item_name,
			SUM(pic.quantity_needed) as total_quantity,
			m.unit,
			'MATERIAL' as item_type,
			SUM(pic.total_cost) as total_cost
		FROM project_item_costs pic
		JOIN master_materials m ON pic.master_item_id = m.material_id
		JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
		WHERE pwi.project_id = ? AND pic.item_type = 'MATERIAL'
		GROUP BY m.material_id, m.material_name, m.unit
	`

	// Get labor
	laborQuery := `
		SELECT
			lt.labor_type_id as item_id,
			lt.role_name as item_name,
			SUM(pic.quantity_needed) as total_quantity,
			lt.unit,
			'LABOR' as item_type,
			SUM(pic.total_cost) as total_cost
		FROM project_item_costs pic
		JOIN master_labor_types lt ON pic.master_item_id = lt.labor_type_id
		JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
		WHERE pwi.project_id = ? AND pic.item_type = 'LABOR'
		GROUP BY lt.labor_type_id, lt.role_name, lt.unit
	`

	// Execute materials query
	materialsRows, err := tx.Query(materialsQuery, projectId)
	if err != nil {
		return nil, err
	}
	defer materialsRows.Close()

	var summaries []models.MaterialSummary
	for materialsRows.Next() {
		var summary models.MaterialSummary
		if err := materialsRows.Scan(
			&summary.ItemId,
			&summary.ItemName,
			&summary.TotalQuantity,
			&summary.Unit,
			&summary.ItemType,
			&summary.TotalCost,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	// Execute labor query
	laborRows, err := tx.Query(laborQuery, projectId)
	if err != nil {
		return nil, err
	}
	defer laborRows.Close()

	for laborRows.Next() {
		var summary models.MaterialSummary
		if err := laborRows.Scan(
			&summary.ItemId,
			&summary.ItemName,
			&summary.TotalQuantity,
			&summary.Unit,
			&summary.ItemType,
			&summary.TotalCost,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}
