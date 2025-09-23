package master_labor_types

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type MasterLaborTypesRepo struct{}

func NewMasterLaborTypesRepo() *MasterLaborTypesRepo {
	return &MasterLaborTypesRepo{}
}

// TODO: Implement FindById [MasterLaborTypesRepo]
func (r *MasterLaborTypesRepo) FindById(tx *sql.Tx, masterLaborTypeId int) (models.MasterLaborType, error) {

	return models.MasterLaborType{}, nil
}

// TODO: implement Create [MasterLaborTypesRepo]
func (r *MasterLaborTypesRepo) Create(tx *sql.Tx, userData models.MasterLaborType) error {

	return nil
}
