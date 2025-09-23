package master_materials

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type MasterMaterialsRepo struct{}

func NewMasterMaterialsRepo() *MasterMaterialsRepo {
	return &MasterMaterialsRepo{}
}

// TODO: Implement FindById [MasterMaterialsRepo]
func (r *MasterMaterialsRepo) FindById(tx *sql.Tx, masterMaterialId int) (models.MasterMaterial, error) {

	return models.MasterMaterial{}, nil
}

// TODO: implement Create [MasterMaterialsRepo]
func (r *MasterMaterialsRepo) Create(tx *sql.Tx, userData models.MasterMaterial) error {

	return nil
}
