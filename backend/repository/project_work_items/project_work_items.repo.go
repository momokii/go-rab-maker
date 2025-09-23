package project_work_items

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type ProjectWorkItemRepo struct{}

func NewProjectWorkItemRepo() *ProjectWorkItemRepo {
	return &ProjectWorkItemRepo{}
}

// TODO: Implement FindById [ProjectWorkItemRepo]
func (r *ProjectWorkItemRepo) FindById(tx *sql.Tx, projectWorkItemId int) (models.ProjectWorkItem, error) {

	return models.ProjectWorkItem{}, nil
}

// TODO: implement Create [ProjectWorkItemRepo]
func (r *ProjectWorkItemRepo) Create(tx *sql.Tx, userData models.ProjectWorkItem) error {

	return nil
}
