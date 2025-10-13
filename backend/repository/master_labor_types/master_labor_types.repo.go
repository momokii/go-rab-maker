package master_labor_types

import (
	"database/sql"
	"math"

	"github.com/momokii/go-rab-maker/backend/models"
)

type MasterLaborTypesRepo struct{}

func NewMasterLaborTypesRepo() *MasterLaborTypesRepo {
	return &MasterLaborTypesRepo{}
}

// TODO: Testing FindById [MasterLaborTypesRepo]
func (r *MasterLaborTypesRepo) FindById(tx *sql.Tx, masterLaborTypeId int) (models.MasterLaborType, error) {

	var laborData models.MasterLaborType

	query := "SELECT labor_type_id, user_id, role_name, unit, default_daily_wage, created_at, updated_at FROM master_labor_types WHERE labor_type_id = ?"
	if err := tx.QueryRow(
		query,
		masterLaborTypeId,
	).Scan(
		&laborData.LaborTypeId,
		&laborData.UserId,
		&laborData.RoleName,
		&laborData.Unit,
		&laborData.DefaultDailyWage,
		&laborData.CreatedAt,
		&laborData.UpdatedAt,
	); err != nil && err != sql.ErrNoRows {
		return laborData, err
	}

	return laborData, nil
}

// Find finds labor types with pagination
func (r *MasterLaborTypesRepo) Find(tx *sql.Tx, paginationInput models.TablePaginationDataInput, user_id int) ([]models.MasterLaborType, models.PaginationInfo, error) {

	var laborTypes []models.MasterLaborType
	var paginationData models.PaginationInfo
	var totalData int

	// Calculate offset for pagination
	offset := (paginationInput.Page - 1) * paginationInput.PerPage

	params := []interface{}{}
	base_query := "SELECT labor_type_id, user_id, role_name, unit, default_daily_wage, created_at, updated_at FROM master_labor_types WHERE 1=1"
	query_total := "SELECT COUNT(labor_type_id) FROM master_labor_types WHERE 1=1"

	// if using search data
	if paginationInput.Search != "" {
		base_query += " AND role_name LIKE ?"
		query_total += " AND role_name LIKE ?"
		params = append(params, "%"+paginationInput.Search+"%")
	}

	// if using for per user
	if user_id != 0 {
		base_query += " AND user_id = ?"
		query_total += " AND user_id = ?"
		params = append(params, user_id)
	}

	// get total data
	if err := tx.QueryRow(
		query_total,
		params...,
	).Scan(&totalData); err != nil {
		return laborTypes, paginationData, err
	}

	// set the offset for the main data
	base_query += " ORDER BY labor_type_id LIMIT ? OFFSET ?"
	params = append(params, paginationInput.PerPage, offset)

	rows, err := tx.Query(base_query, params...)
	if err != nil {
		return laborTypes, paginationData, err
	}

	for rows.Next() {
		var laborData models.MasterLaborType

		if err := rows.Scan(
			&laborData.LaborTypeId,
			&laborData.UserId,
			&laborData.RoleName,
			&laborData.Unit,
			&laborData.DefaultDailyWage,
			&laborData.CreatedAt,
			&laborData.UpdatedAt,
		); err != nil {
			return laborTypes, paginationData, err
		} else {
			laborTypes = append(laborTypes, laborData)
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
	if len(laborTypes) == 0 {
		return []models.MasterLaborType{}, paginationData, nil
	}

	return laborTypes, paginationData, nil
}

// Create creates a new labor type
func (r *MasterLaborTypesRepo) Create(tx *sql.Tx, laborData models.MasterLaborTypeCreate) error {

	query := "INSERT INTO master_labor_types (role_name, unit, default_daily_wage, user_id) VALUES (?, ?, ?, ?)"
	if _, err := tx.Exec(
		query,
		laborData.RoleName,
		laborData.Unit,
		laborData.DefaultDailyWage,
		laborData.UserId,
	); err != nil {
		return err
	}

	return nil
}

// Update updates an existing labor type
func (r *MasterLaborTypesRepo) Update(tx *sql.Tx, laborData models.MasterLaborType) error {

	query := "UPDATE master_labor_types SET role_name = ?, unit = ?, default_daily_wage = ? WHERE labor_type_id = ? AND user_id = ?"
	if _, err := tx.Exec(
		query,
		laborData.RoleName,
		laborData.Unit,
		laborData.DefaultDailyWage,
		laborData.LaborTypeId,
		laborData.UserId,
	); err != nil {
		return err
	}

	return nil
}

// Delete deletes a labor type
func (r *MasterLaborTypesRepo) Delete(tx *sql.Tx, laborData models.MasterLaborType) error {
	query := "DELETE FROM master_labor_types WHERE labor_type_id = ?"
	if _, err := tx.Exec(query, laborData.LaborTypeId); err != nil {
		return err
	}

	return nil
}
