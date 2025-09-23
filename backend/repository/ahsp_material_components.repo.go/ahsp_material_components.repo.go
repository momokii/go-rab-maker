package ahsp_material_components

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type AHSPMaterialComponentsRepo struct{}

func NewAHSPMaterialComponentsRepo() *AHSPMaterialComponentsRepo {
	return &AHSPMaterialComponentsRepo{}
}

// TODO: Implement FindById [AHSPMaterialComponentsRepo]
func (r *AHSPMaterialComponentsRepo) FindById(tx *sql.Tx, ahspMaterialComponentId int) (models.AHSPMaterialComponent, error) {
	return models.AHSPMaterialComponent{}, nil
}

// TODO: Implement Create [AHSPMaterialComponentsRepo]
func (r *AHSPMaterialComponentsRepo) Create(tx *sql.Tx, userData models.AHSPMaterialComponent) error {
	return nil
}
