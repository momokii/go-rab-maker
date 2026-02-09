package ahsptemplates

import (
	"database/sql"
	"math"

	"github.com/momokii/go-rab-maker/backend/models"
)

type AhspTemplatesRepo struct{}

func NewAhspTemplatesRepo() *AhspTemplatesRepo {
	return &AhspTemplatesRepo{}
}

// FindById retrieves an AHSP template by ID
func (r *AhspTemplatesRepo) FindById(tx *sql.Tx, ahspTemplateId int) (models.AHSPTemplate, error) {
	var template models.AHSPTemplate
	var userId sql.NullInt64

	query := "SELECT template_id, user_id, template_name, unit, created_at, updated_at FROM ahsp_templates WHERE template_id = ?"

	if err := tx.QueryRow(
		query,
		ahspTemplateId,
	).Scan(
		&template.TemplateId,
		&userId,
		&template.TemplateName,
		&template.Unit,
		&template.CreatedAt,
		&template.UpdatedAt,
	); err != nil && err != sql.ErrNoRows {
		return template, err
	}

	// Convert sql.NullInt64 to int (0 if NULL)
	if userId.Valid {
		template.UserId = int(userId.Int64)
	} else {
		template.UserId = 0
	}

	return template, nil
}

// Find retrieves AHSP templates with pagination and search
func (r *AhspTemplatesRepo) Find(tx *sql.Tx, paginationInput models.TablePaginationDataInput, user_id int) ([]models.AHSPTemplate, models.PaginationInfo, error) {
	var templates []models.AHSPTemplate
	var paginationData models.PaginationInfo
	var totalData int

	// Calculate offset for pagination
	offset := (paginationInput.Page - 1) * paginationInput.PerPage

	params := []interface{}{}
	base_query := "SELECT template_id, user_id, template_name, unit, created_at, updated_at FROM ahsp_templates WHERE 1=1"
	query_total := "SELECT COUNT(template_id) FROM ahsp_templates WHERE 1=1"

	// if using search data
	if paginationInput.Search != "" {
		base_query += " AND template_name LIKE ?"
		query_total += " AND template_name LIKE ?"
		params = append(params, "%"+paginationInput.Search+"%")
	}

	if user_id != 0 {
		// Include both user-specific items and system-wide defaults (user_id IS NULL)
		base_query += " AND (user_id = ? OR user_id IS NULL)"
		query_total += " AND (user_id = ? OR user_id IS NULL)"
		params = append(params, user_id)
	}

	// get total data
	if err := tx.QueryRow(
		query_total,
		params...,
	).Scan(&totalData); err != nil {
		return templates, paginationData, err
	}

	// set the offset for the main data
	base_query += " ORDER BY template_id LIMIT ? OFFSET ?"
	params = append(params, paginationInput.PerPage, offset)

	rows, err := tx.Query(base_query, params...)
	if err != nil {
		return templates, paginationData, err
	}

	for rows.Next() {
		var template models.AHSPTemplate
		var userId sql.NullInt64

		if err := rows.Scan(
			&template.TemplateId,
			&userId,
			&template.TemplateName,
			&template.Unit,
			&template.CreatedAt,
			&template.UpdatedAt,
		); err != nil {
			return templates, paginationData, err
		}

		// Convert sql.NullInt64 to int (0 if NULL)
		if userId.Valid {
			template.UserId = int(userId.Int64)
		} else {
			template.UserId = 0
		}

		templates = append(templates, template)
	}

	// pagination data
	paginationData = models.PaginationInfo{
		TotalItems:   totalData,
		ItemsPerPage: paginationInput.PerPage,
		CurrentPage:  paginationInput.Page,
		TotalPages:   int(math.Ceil(float64(totalData) / float64(paginationInput.PerPage))),
	}

	// if data nil, just return array
	if len(templates) == 0 {
		return []models.AHSPTemplate{}, paginationData, nil
	}

	return templates, paginationData, nil
}

// Create creates a new AHSP template
func (r *AhspTemplatesRepo) Create(tx *sql.Tx, templateData models.AHSPTemplateCreate) error {
	query := "INSERT INTO ahsp_templates (user_id, template_name, unit) VALUES (?, ?, ?)"
	if _, err := tx.Exec(
		query,
		templateData.UserId,
		templateData.TemplateName,
		templateData.Unit,
	); err != nil {
		return err
	}

	return nil
}

// Update updates an existing AHSP template
func (r *AhspTemplatesRepo) Update(tx *sql.Tx, templateData models.AHSPTemplate) error {
	query := "UPDATE ahsp_templates SET template_name = ?, unit = ? WHERE template_id = ? AND user_id = ?"
	if _, err := tx.Exec(
		query,
		templateData.TemplateName,
		templateData.Unit,
		templateData.TemplateId,
		templateData.UserId,
	); err != nil {
		return err
	}

	return nil
}

func (r *AhspTemplatesRepo) Delete(tx *sql.Tx, templateData models.AHSPTemplate) error {
	// -------------------------------------------------------------------------
	// OPTIMIZATION: REMOVE LOOP (N+1 Query)
	// -------------------------------------------------------------------------

	// 1. Delete data from the "grandchild" table (project_item_costs) first.
	// We use a subquery to identify which costs to delete based on the template ID.
	// CRITICAL: This must be executed BEFORE deleting 'project_work_items',
	// otherwise the subquery will return empty results.
	query_delete_project_item_costs := `
        DELETE FROM project_item_costs 
        WHERE work_item_id IN (
            SELECT work_item_id 
            FROM project_work_items 
            WHERE ahsp_template_id = ?
        )`

	if _, err := tx.Exec(query_delete_project_item_costs, templateData.TemplateId); err != nil {
		return err
	}

	// 2. Delete from 'project_work_items' (the child table)
	query_delete_project_work_items := "DELETE FROM project_work_items WHERE ahsp_template_id = ?"
	if _, err := tx.Exec(query_delete_project_work_items, templateData.TemplateId); err != nil {
		return err
	}

	// -------------------------------------------------------------------------
	// STANDARD DELETIONS
	// -------------------------------------------------------------------------

	// 3. Delete related material components
	query_delete_material := "DELETE FROM ahsp_material_components WHERE template_id = ?"
	if _, err := tx.Exec(query_delete_material, templateData.TemplateId); err != nil {
		return err
	}

	// 4. Delete related labor components
	query_delete_labor := "DELETE FROM ahsp_labor_components WHERE template_id = ?"
	if _, err := tx.Exec(query_delete_labor, templateData.TemplateId); err != nil {
		return err
	}

	// 5. Finally, delete the main template data
	// It is best practice to delete the parent row last to maintain referential integrity
	// logic during the transaction.
	query := "DELETE FROM ahsp_templates WHERE template_id = ?"
	if _, err := tx.Exec(query, templateData.TemplateId); err != nil {
		return err
	}

	return nil
}
