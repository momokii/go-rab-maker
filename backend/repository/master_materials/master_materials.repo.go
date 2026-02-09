package master_materials

import (
	"database/sql"
	"log"
	"math"

	"github.com/momokii/go-rab-maker/backend/models"
)

type MasterMaterialsRepo struct{}

func NewMasterMaterialsRepo() *MasterMaterialsRepo {
	return &MasterMaterialsRepo{}
}

func (r *MasterMaterialsRepo) FindById(tx *sql.Tx, masterMaterialId int) (models.MasterMaterial, error) {

	var material models.MasterMaterial
	var userId sql.NullInt64

	query := "SELECT material_id, user_id, material_name, unit, default_unit_price, created_at, updated_at FROM master_materials WHERE material_id = ?"
	if err := tx.QueryRow(
		query,
		masterMaterialId,
	).Scan(
		&material.MaterialId,
		&userId,
		&material.MaterialName,
		&material.Unit,
		&material.DefaultUnitPrice,
		&material.CreatedAt,
		&material.UpdatedAt,
	); err != nil && err != sql.ErrNoRows {
		// makesure also error is not err because not found

		return material, err
	}

	// Convert sql.NullInt64 to int (0 if NULL)
	if userId.Valid {
		material.UserId = int(userId.Int64)
	} else {
		material.UserId = 0
	}

	return material, nil
}

func (r *MasterMaterialsRepo) Find(tx *sql.Tx, paginationInput models.TablePaginationDataInput, userId int) ([]models.MasterMaterial, models.PaginationInfo, error) {
	var materials []models.MasterMaterial
	var paginationData models.PaginationInfo
	var totalData int

	// Calculate offset for pagination
	offset := (paginationInput.Page - 1) * paginationInput.PerPage

	params := []interface{}{}
	base_query := "SELECT material_id, user_id, material_name, unit, default_unit_price, created_at, updated_at FROM master_materials WHERE 1=1"
	query_total := "SELECT COUNT(material_id) FROM master_materials WHERE 1=1"

	// if using search data
	if paginationInput.Search != "" {
		base_query += " AND material_name LIKE ?"
		query_total += " AND material_name LIKE ?"
		params = append(params, "%"+paginationInput.Search+"%")
	}

	// if using user_id
	if userId != 0 {
		// Include both user-specific items and system-wide defaults (user_id IS NULL)
		base_query += " AND (user_id = ? OR user_id IS NULL)"
		query_total += " AND (user_id = ? OR user_id IS NULL)"
		params = append(params, userId)
	}

	// get total data
	if err := tx.QueryRow(
		query_total,
		params...,
	).Scan(&totalData); err != nil {
		return materials, paginationData, err
	}

	// set the offset for the main data
	base_query += " ORDER BY material_id LIMIT ? OFFSET ?"
	params = append(params, paginationInput.PerPage, offset)

	rows, err := tx.Query(base_query, params...)
	if err != nil {
		return materials, paginationData, err
	}

	for rows.Next() {
		var mateial models.MasterMaterial
		var userId sql.NullInt64

		if err := rows.Scan(
			&mateial.MaterialId,
			&userId,
			&mateial.MaterialName,
			&mateial.Unit,
			&mateial.DefaultUnitPrice,
			&mateial.CreatedAt,
			&mateial.UpdatedAt,
		); err != nil {
			return materials, paginationData, err
		}

		// Convert sql.NullInt64 to int (0 if NULL)
		if userId.Valid {
			mateial.UserId = int(userId.Int64)
		} else {
			mateial.UserId = 0
		}

		materials = append(materials, mateial)
	}

	// pagination data
	paginationData = models.PaginationInfo{
		TotalItems:   totalData,
		ItemsPerPage: paginationInput.PerPage,
		CurrentPage:  paginationInput.Page,
		TotalPages:   int(math.Ceil(float64(totalData) / float64(paginationInput.PerPage))),
	}

	// if data nil, just return array
	if len(materials) == 0 {
		return []models.MasterMaterial{}, paginationData, nil
	}

	return materials, paginationData, nil
}

func (r *MasterMaterialsRepo) Create(tx *sql.Tx, materialData models.MasterMaterialCreate) error {

	query := "INSERT INTO master_materials (material_name, unit, default_unit_price, user_id) VALUES (?, ?, ?, ?)"
	if _, err := tx.Exec(
		query,
		materialData.MaterialName,
		materialData.Unit,
		materialData.DefaultUnitPrice,
		materialData.UserId,
	); err != nil {
		return err
	}

	return nil
}

// TODO: [update] testing
func (r *MasterMaterialsRepo) Update(tx *sql.Tx, materialData models.MasterMaterial) error {

	// update main data
	query := "UPDATE master_materials SET material_name = ?, unit = ?, default_unit_price = ? WHERE material_id = ? AND user_id = ?"
	if _, err := tx.Exec(
		query,
		materialData.MaterialName,
		materialData.Unit,
		materialData.DefaultUnitPrice,
		materialData.MaterialId,
		materialData.UserId,
	); err != nil {
		return err
	}

	// add below if this data will related to another table

	// make sure also update form the project_item_costs table
	query_update_project_item_costs := `
		UPDATE 
			project_item_costs 
		SET
			item_name = ?,
			unit_price_at_creation = ?,
			total_cost = quantity_needed * ?
		WHERE master_item_id = ?`
	if _, err := tx.Exec(
		query_update_project_item_costs,
		materialData.MaterialName,
		materialData.DefaultUnitPrice,
		materialData.DefaultUnitPrice,
		materialData.MaterialId,
	); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (r *MasterMaterialsRepo) Delete(tx *sql.Tx, materialData models.MasterMaterial) error {

	// delete main data
	query_delete_material := "DELETE FROM master_materials WHERE material_id = ?"
	if _, err := tx.Exec(
		query_delete_material,
		materialData.MaterialId,
	); err != nil {
		return err
	}

	// add below if this data will related to another table

	// make sure to delete the related id at ahsp_material_component table
	query_delete_ahsp_material := "DELETE FROM ahsp_material_components WHERE material_id = ?"
	if _, err := tx.Exec(
		query_delete_ahsp_material,
		materialData.MaterialId,
	); err != nil {
		return err
	}

	// make sure also delete from the project_item_costs table
	query_delete_project_item_costs := "DELETE FROM project_item_costs WHERE master_item_id = ?"
	if _, err := tx.Exec(
		query_delete_project_item_costs,
		materialData.MaterialId,
	); err != nil {
		return err
	}

	return nil
}
