package projects

import (
	"database/sql"
	"math"

	"github.com/momokii/go-rab-maker/backend/models"
)

type ProjectsRepo struct{}

func NewProjectsRepo() *ProjectsRepo {
	return &ProjectsRepo{}
}

// FindById retrieves a project by ID
func (r *ProjectsRepo) FindById(tx *sql.Tx, projectId int) (models.Project, error) {
	var project models.Project

	query := "SELECT project_id, user_id, project_name, location, client_name, created_at, updated_at FROM projects WHERE project_id = ?"

	if err := tx.QueryRow(
		query,
		projectId,
	).Scan(
		&project.ProjectId,
		&project.UserId,
		&project.ProjectName,
		&project.Location,
		&project.ClientName,
		&project.CreatedAt,
		&project.UpdatedAt,
	); err != nil && err != sql.ErrNoRows {
		return project, err
	}

	return project, nil
}

// Find retrieves projects with pagination and search
func (r *ProjectsRepo) Find(tx *sql.Tx, paginationInput models.TablePaginationDataInput) ([]models.Project, models.PaginationInfo, error) {
	var projects []models.Project
	var paginationData models.PaginationInfo
	var totalData int

	// Calculate offset for pagination
	offset := (paginationInput.Page - 1) * paginationInput.PerPage

	params := []interface{}{}
	base_query := "SELECT project_id, user_id, project_name, location, client_name, created_at, updated_at FROM projects WHERE 1=1"
	query_total := "SELECT COUNT(project_id) FROM projects WHERE 1=1"

	// if using search data
	if paginationInput.Search != "" {
		base_query += " AND (project_name LIKE ? OR location LIKE ? OR client_name LIKE ?)"
		query_total += " AND (project_name LIKE ? OR location LIKE ? OR client_name LIKE ?)"
		searchTerm := "%" + paginationInput.Search + "%"
		params = append(params, searchTerm, searchTerm, searchTerm)
	}

	// get total data
	if err := tx.QueryRow(
		query_total,
		params...,
	).Scan(&totalData); err != nil {
		return projects, paginationData, err
	}

	// set the offset for the main data
	base_query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	params = append(params, paginationInput.PerPage, offset)

	rows, err := tx.Query(base_query, params...)
	if err != nil {
		return projects, paginationData, err
	}

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
			return projects, paginationData, err
		} else {
			projects = append(projects, project)
		}
	}

	// pagination data
	paginationData = models.PaginationInfo{
		TotalItems:   totalData,
		ItemsPerPage: paginationInput.PerPage,
		CurrentPage:  paginationInput.Page,
		TotalPages:   int(math.Ceil(float64(totalData) / float64(paginationInput.PerPage))),
	}

	// if data nil, just return array
	if len(projects) == 0 {
		return []models.Project{}, paginationData, nil
	}

	return projects, paginationData, nil
}

// FindByUserId retrieves projects for a specific user with pagination and search
func (r *ProjectsRepo) FindByUserId(tx *sql.Tx, userId int, paginationInput models.TablePaginationDataInput) ([]models.Project, models.PaginationInfo, error) {
	var projects []models.Project
	var paginationData models.PaginationInfo
	var totalData int

	// Calculate offset for pagination
	offset := (paginationInput.Page - 1) * paginationInput.PerPage

	params := []interface{}{userId}
	base_query := "SELECT project_id, user_id, project_name, location, client_name, created_at, updated_at FROM projects WHERE user_id = ?"
	query_total := "SELECT COUNT(project_id) FROM projects WHERE user_id = ?"

	// if using search data
	if paginationInput.Search != "" {
		base_query += " AND (project_name LIKE ? OR location LIKE ? OR client_name LIKE ?)"
		query_total += " AND (project_name LIKE ? OR location LIKE ? OR client_name LIKE ?)"
		searchTerm := "%" + paginationInput.Search + "%"
		params = append(params, searchTerm, searchTerm, searchTerm)
	}

	// get total data
	if err := tx.QueryRow(
		query_total,
		params...,
	).Scan(&totalData); err != nil {
		return projects, paginationData, err
	}

	// set the offset for the main data
	base_query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	params = append(params, paginationInput.PerPage, offset)

	rows, err := tx.Query(base_query, params...)
	if err != nil {
		return projects, paginationData, err
	}

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
			return projects, paginationData, err
		} else {
			projects = append(projects, project)
		}
	}

	// pagination data
	paginationData = models.PaginationInfo{
		TotalItems:   totalData,
		ItemsPerPage: paginationInput.PerPage,
		CurrentPage:  paginationInput.Page,
		TotalPages:   int(math.Ceil(float64(totalData) / float64(paginationInput.PerPage))),
	}

	// if data nil, just return array
	if len(projects) == 0 {
		return []models.Project{}, paginationData, nil
	}

	return projects, paginationData, nil
}

// Create creates a new project
func (r *ProjectsRepo) Create(tx *sql.Tx, projectData models.ProjectCreate) error {
	query := "INSERT INTO projects (user_id, project_name, location, client_name) VALUES (?, ?, ?, ?)"
	if _, err := tx.Exec(
		query,
		projectData.UserId,
		projectData.ProjectName,
		projectData.Location,
		projectData.ClientName,
	); err != nil {
		return err
	}

	return nil
}

// Update updates an existing project
func (r *ProjectsRepo) Update(tx *sql.Tx, projectData models.Project) error {
	query := "UPDATE projects SET project_name = ?, location = ?, client_name = ? WHERE project_id = ? AND user_id = ?"
	if _, err := tx.Exec(
		query,
		projectData.ProjectName,
		projectData.Location,
		projectData.ClientName,
		projectData.ProjectId,
		projectData.UserId,
	); err != nil {
		return err
	}

	return nil
}

// Delete deletes a project
func (r *ProjectsRepo) Delete(tx *sql.Tx, projectData models.Project) error {
	// -------------------------------------------------------------------------
	// STEP 1: Delete Grandchild Data (project_item_costs)
	// -------------------------------------------------------------------------
	// We remove the lowest level data first using a subquery.
	// Logic: Delete costs where the work_item belongs to this project_id.
	query_delete_item_costs := `
        DELETE FROM project_item_costs 
        WHERE work_item_id IN (
            SELECT work_item_id 
            FROM project_work_items 
            WHERE project_id = ?
        )`

	if _, err := tx.Exec(query_delete_item_costs, projectData.ProjectId); err != nil {
		return err
	}

	// -------------------------------------------------------------------------
	// STEP 2: Delete Child Data (project_work_items)
	// -------------------------------------------------------------------------
	// Now that the costs are gone, we can safely remove the work items.
	query_delete_work_items := "DELETE FROM project_work_items WHERE project_id = ?"
	if _, err := tx.Exec(query_delete_work_items, projectData.ProjectId); err != nil {
		return err
	}

	// -------------------------------------------------------------------------
	// STEP 3: Delete Main Data (projects)
	// -------------------------------------------------------------------------
	// Finally, remove the project itself.
	query := "DELETE FROM projects WHERE project_id = ?"
	if _, err := tx.Exec(query, projectData.ProjectId); err != nil {
		return err
	}

	return nil
}
