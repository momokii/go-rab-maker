package master_work_categories

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type MasterWorkCategoriesRepo struct{}

func NewMasterWorkCategoriesRepo() *MasterWorkCategoriesRepo {
	return &MasterWorkCategoriesRepo{}
}

// TODO: Implement FindById [MasterWorkCategoriesRepo]
func (r *MasterWorkCategoriesRepo) FindById(tx *sql.Tx, masterWorkCategoryId int) (models.MasterWorkCategory, error) {
	return models.MasterWorkCategory{}, nil
}

// TODO: Implement Create [MasterWorkCategoriesRepo]
func (r *MasterWorkCategoriesRepo) Create(tx *sql.Tx, userData models.MasterWorkCategory) error {

	return nil
}
