package project_work_items

import (
	"database/sql"
	"time"

	"github.com/momokii/go-rab-maker/backend/models"
)

type ProjectWorkItemRepo struct{}

func NewProjectWorkItemRepo() *ProjectWorkItemRepo {
	return &ProjectWorkItemRepo{}
}

// FindById retrieves a project work item by its ID
func (r *ProjectWorkItemRepo) FindById(tx *sql.Tx, projectWorkItemId int) (models.ProjectWorkItem, error) {
	query := `
		SELECT
			work_item_id, project_id, category_id, description,
			volume, unit, ahsp_template_id, created_at, updated_at
		FROM project_work_items
		WHERE work_item_id = ?
	`

	var workItem models.ProjectWorkItem
	err := tx.QueryRow(query, projectWorkItemId).Scan(
		&workItem.WorkItemId,
		&workItem.ProjectId,
		&workItem.CategoryId,
		&workItem.Description,
		&workItem.Volume,
		&workItem.Unit,
		&workItem.AHSPTemplateId,
		&workItem.CreatedAt,
		&workItem.UpdatedAt,
	)

	if err != nil {
		return models.ProjectWorkItem{}, err
	}

	return workItem, nil
}

// FindByProjectId retrieves all work items for a specific project
func (r *ProjectWorkItemRepo) FindByProjectId(tx *sql.Tx, projectId int) ([]models.ProjectWorkItem, error) {
	query := `
		SELECT
			work_item_id, project_id, category_id, description,
			volume, unit, ahsp_template_id, created_at, updated_at
		FROM project_work_items
		WHERE project_id = ?
		ORDER BY created_at DESC
	`

	rows, err := tx.Query(query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workItems []models.ProjectWorkItem
	for rows.Next() {
		var workItem models.ProjectWorkItem
		err := rows.Scan(
			&workItem.WorkItemId,
			&workItem.ProjectId,
			&workItem.CategoryId,
			&workItem.Description,
			&workItem.Volume,
			&workItem.Unit,
			&workItem.AHSPTemplateId,
			&workItem.CreatedAt,
			&workItem.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		workItems = append(workItems, workItem)
	}

	return workItems, nil
}

// FindByProjectIdWithDetails retrieves work items with category and template details
func (r *ProjectWorkItemRepo) FindByProjectIdWithDetails(tx *sql.Tx, projectId int) ([]models.ProjectWorkItemWithDetails, error) {
	query := `
		SELECT
			pwi.work_item_id, pwi.project_id, pwi.category_id, pwi.description,
			pwi.volume, pwi.unit, pwi.ahsp_template_id, pwi.created_at, pwi.updated_at,
			mwc.category_name,
			at.template_name
		FROM project_work_items pwi
		LEFT JOIN master_work_categories mwc ON pwi.category_id = mwc.category_id
		LEFT JOIN ahsp_templates at ON pwi.ahsp_template_id = at.template_id
		WHERE pwi.project_id = ?
		ORDER BY mwc.display_order ASC, pwi.created_at DESC
	`

	rows, err := tx.Query(query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workItems []models.ProjectWorkItemWithDetails
	for rows.Next() {
		var workItem models.ProjectWorkItemWithDetails
		var categoryName, templateName sql.NullString
		err := rows.Scan(
			&workItem.WorkItemId,
			&workItem.ProjectId,
			&workItem.CategoryId,
			&workItem.Description,
			&workItem.Volume,
			&workItem.Unit,
			&workItem.AHSPTemplateId,
			&workItem.CreatedAt,
			&workItem.UpdatedAt,
			&categoryName,
			&templateName,
		)
		if err != nil {
			return nil, err
		}

		if categoryName.Valid {
			workItem.CategoryName = categoryName.String
		}
		if templateName.Valid {
			workItem.TemplateName = templateName.String
		}

		workItems = append(workItems, workItem)
	}

	return workItems, nil
}

// Create inserts a new project work item and returns the ID
func (r *ProjectWorkItemRepo) Create(tx *sql.Tx, workItem models.ProjectWorkItemCreate) (int, error) {
	query := `
		INSERT INTO project_work_items
		(project_id, category_id, description, volume, unit, ahsp_template_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := tx.Exec(
		query,
		workItem.ProjectId,
		workItem.CategoryId,
		workItem.Description,
		workItem.Volume,
		workItem.Unit,
		workItem.AHSPTemplateId,
		now,
		now,
	)

	if err != nil {
		return 0, err
	}

	// Get the ID of the newly inserted work item
	workItemId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(workItemId), nil
}

// Update updates an existing project work item
func (r *ProjectWorkItemRepo) Update(tx *sql.Tx, workItem models.ProjectWorkItem) error {
	query := `
		UPDATE project_work_items
		SET category_id = ?, description = ?, volume = ?, unit = ?,
		    ahsp_template_id = ?, updated_at = ?
		WHERE work_item_id = ?
	`

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := tx.Exec(
		query,
		workItem.CategoryId,
		workItem.Description,
		workItem.Volume,
		workItem.Unit,
		workItem.AHSPTemplateId,
		now,
		workItem.WorkItemId,
	)

	return err
}

// Delete deletes a project work item
func (r *ProjectWorkItemRepo) Delete(tx *sql.Tx, workItem models.ProjectWorkItem) error {
	query := `DELETE FROM project_work_items WHERE work_item_id = ?`
	_, err := tx.Exec(query, workItem.WorkItemId)
	return err
}

// DeleteByProjectId deletes all work items for a specific project
func (r *ProjectWorkItemRepo) DeleteByProjectId(tx *sql.Tx, projectId int) error {
	query := `DELETE FROM project_work_items WHERE project_id = ?`
	_, err := tx.Exec(query, projectId)
	return err
}

// GetProjectTotalCost calculates the total cost of all work items in a project
func (r *ProjectWorkItemRepo) GetProjectTotalCost(tx *sql.Tx, projectId int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(total_cost), 0)
		FROM project_item_costs pic
		JOIN project_work_items pwi ON pic.work_item_id = pwi.work_item_id
		WHERE pwi.project_id = ?
	`

	var totalCost float64
	err := tx.QueryRow(query, projectId).Scan(&totalCost)
	if err != nil {
		return 0, err
	}

	return totalCost, nil
}
