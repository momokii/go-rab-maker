package master_work_categories

import (
	"database/sql"
	"math"

	"github.com/momokii/go-rab-maker/backend/models"
)

type MasterWorkCategoriesRepo struct{}

func NewMasterWorkCategoriesRepo() *MasterWorkCategoriesRepo {
	return &MasterWorkCategoriesRepo{}
}

// TODO: Testing FindById [MasterWorkCategoriesRepo]
func (r *MasterWorkCategoriesRepo) FindById(tx *sql.Tx, masterWorkCategoryId int) (models.MasterWorkCategory, error) {

	var workCategory models.MasterWorkCategory

	query := "SELECT category_id, user_id, category_name, display_order, created_at, updated_at FROM master_work_categories WHERE category_id = ?"

	if err := tx.QueryRow(
		query,
		masterWorkCategoryId,
	).Scan(
		&workCategory.CategoryId,
		&workCategory.UserId,
		&workCategory.CategoryName,
		&workCategory.DisplayOrder,
		&workCategory.CreatedAt,
		&workCategory.UpdatedAt,
	); err != nil && err != sql.ErrNoRows {
		return workCategory, err
	}

	return workCategory, nil
}

// TODO: Testing Find [MasterWorkCategoriesRepo]
func (r *MasterWorkCategoriesRepo) Find(tx *sql.Tx, paginationInput models.TablePaginationDataInput) ([]models.MasterWorkCategory, models.PaginationInfo, error) {

	var masterWorkCategories []models.MasterWorkCategory
	var paginationData models.PaginationInfo
	var totalData int

	// Calculate offset for pagination
	offset := (paginationInput.Page - 1) * paginationInput.PerPage

	params := []interface{}{}
	base_query := "SELECT category_id, user_id, category_name, display_order, created_at, updated_at FROM master_work_categories WHERE 1=1"
	query_total := "SELECT COUNT(category_id) FROM master_work_categories WHERE 1=1"

	// if using search data
	if paginationInput.Search != "" {
		base_query += " AND category_name LIKE ?"
		query_total += " AND category_name LIKE ?"
		params = append(params, "%"+paginationInput.Search+"%")
	}

	// get total data
	if err := tx.QueryRow(
		query_total,
		params...,
	).Scan(&totalData); err != nil {
		return masterWorkCategories, paginationData, err
	}

	// set the offset for the main data
	base_query += " ORDER BY display_order, category_id LIMIT ? OFFSET ?"
	params = append(params, paginationInput.PerPage, offset)

	rows, err := tx.Query(base_query, params...)
	if err != nil {
		return masterWorkCategories, paginationData, err
	}

	for rows.Next() {
		var category models.MasterWorkCategory

		if err := rows.Scan(
			&category.CategoryId,
			&category.UserId,
			&category.CategoryName,
			&category.DisplayOrder,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return masterWorkCategories, paginationData, err
		} else {
			masterWorkCategories = append(masterWorkCategories, category)
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
	if len(masterWorkCategories) == 0 {
		return []models.MasterWorkCategory{}, paginationData, nil
	}

	return masterWorkCategories, paginationData, nil
}

// TODO: Testing Create [MasterWorkCategoriesRepo]
func (r *MasterWorkCategoriesRepo) Create(tx *sql.Tx, categoriesData models.MasterWorkCategoryCreate) error {

	query := "INSERT INTO master_work_categories (user_id, category_name, display_order) VALUES (?, ?, ?)"
	if _, err := tx.Exec(
		query,
		categoriesData.UserId,
		categoriesData.CategoryName,
		categoriesData.DisplayOrder,
	); err != nil {
		return err
	}

	return nil
}

// TODO: Testing Update [MasterWorkCategoriesRepo]
func (r *MasterWorkCategoriesRepo) Update(tx *sql.Tx, categoriesData models.MasterWorkCategory) error {

	query := "UPDATE master_work_categories SET category_name = ?, display_order = ? WHERE category_id = ? AND user_id = ?"
	if _, err := tx.Exec(
		query,
		categoriesData.CategoryName,
		categoriesData.DisplayOrder,
		categoriesData.CategoryId,
		categoriesData.UserId,
	); err != nil {
		return err
	}

	return nil
}

// Delete deletes a work category
func (r *MasterWorkCategoriesRepo) Delete(tx *sql.Tx, categoriesData models.MasterWorkCategory) error {
	query := "DELETE FROM master_work_categories WHERE category_id = ?"
	if _, err := tx.Exec(query, categoriesData.CategoryId); err != nil {
		return err
	}

	return nil
}
