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
