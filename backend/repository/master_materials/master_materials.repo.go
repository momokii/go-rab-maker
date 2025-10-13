package master_materials

import (
	"database/sql"
	"math"

	"github.com/momokii/go-rab-maker/backend/models"
)

type MasterMaterialsRepo struct{}

func NewMasterMaterialsRepo() *MasterMaterialsRepo {
	return &MasterMaterialsRepo{}
}

// TODO: testing FindById [MasterMaterialsRepo]
func (r *MasterMaterialsRepo) FindById(tx *sql.Tx, masterMaterialId int) (models.MasterMaterial, error) {

	var material models.MasterMaterial

	query := "SELECT material_id, user_id, material_name, unit, default_unit_price, created_at, updated_at FROM master_materials WHERE material_id = ?"
	if err := tx.QueryRow(
		query,
		masterMaterialId,
	).Scan(
		&material.MaterialId,
		&material.UserId,
		&material.MaterialName,
		&material.Unit,
		&material.DefaultUnitPrice,
		&material.CreatedAt,
		&material.UpdatedAt,
	); err != nil && err != sql.ErrNoRows {
		// makesure also error is not err because not found

		return material, err
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
		base_query += " AND material_name = ?"
		query_total += " AND material_name = ?"
		params = append(params, paginationInput.Search)
	}

	// if using user_id
	if userId != 0 {
		base_query += " AND user_id = ?"
		query_total += " AND user_id = ?"
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

		if err := rows.Scan(
			&mateial.MaterialId,
			&mateial.UserId,
			&mateial.MaterialName,
			&mateial.Unit,
			&mateial.DefaultUnitPrice,
			&mateial.CreatedAt,
			&mateial.UpdatedAt,
		); err != nil {
			return materials, paginationData, err
		} else {
			materials = append(materials, mateial)
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
	if len(materials) == 0 {
		return []models.MasterMaterial{}, paginationData, nil
	}

	return materials, paginationData, nil
}

// TODO: testing Create [MasterMaterialsRepo]
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

// TODO: testing update materials
func (r *MasterMaterialsRepo) Update(tx *sql.Tx, materialData models.MasterMaterial) error {

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

	return nil
}

// TODO: testing delete materials
func (r *MasterMaterialsRepo) Delete(tx *sql.Tx, materialData models.MasterMaterial) error {

	query := "DELETE FROM master_materials WHERE material_id = ?"
	if _, err := tx.Exec(query, materialData.MaterialId); err != nil {
		return nil
	}

	// add below if this data will related to another table

	return nil
}
