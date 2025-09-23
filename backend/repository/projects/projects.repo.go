package projects

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type ProjectsRepo struct{}

func NewProjectsRepo() *ProjectsRepo {
	return &ProjectsRepo{}
}

// TODO: Implement FindById [Projects]
func (r *ProjectsRepo) FindById(tx *sql.Tx, projectId int) (models.Project, error) {

	return models.Project{}, nil
}

// TODO: implement Create [Projects
func (r *ProjectsRepo) Create(tx *sql.Tx, projectData models.ProjectCreate) error {

	return nil
}
