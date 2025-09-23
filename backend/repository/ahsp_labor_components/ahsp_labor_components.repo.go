package ahsp_labor_components

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type AHSPLaborComponentsRepo struct{}

func NewAHSPMaterialComponentsRepo() *AHSPLaborComponentsRepo {
	return &AHSPLaborComponentsRepo{}
}

// TODO:Implement FindById [AHSPLaborComponentsRepo]
func (r *AHSPLaborComponentsRepo) FindById(tx *sql.Tx, ahspLaborComponentId int) (models.AHSPLaborComponent, error) {
	return models.AHSPLaborComponent{}, nil
}

// TODO:Implement Create [AHSPLaborComponentsRepo]
func (r *AHSPLaborComponentsRepo) Create(tx *sql.Tx, userData models.AHSPLaborComponent) error {
	return nil
}
