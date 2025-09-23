package project_item_costs

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type ProjectItemCostsRepo struct{}

func NewProjectItemCostsRepo() *ProjectItemCostsRepo {
	return &ProjectItemCostsRepo{}
}

// TODO: Implement FindById [ProjectItemCostsRepo]
func (r *ProjectItemCostsRepo) FindById(tx *sql.Tx, projectItemCostId int) (models.ProjectItemCost, error) {

	return models.ProjectItemCost{}, nil
}

// TODO: implement Create [ProjectItemCostsRepo]
func (r *ProjectItemCostsRepo) Create(tx *sql.Tx, userData models.ProjectItemCost) error {

	return nil
}
