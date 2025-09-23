package ahsptemplates

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type AhspTemplatesRepo struct{}

func NewAhspTemplatesRepo() *AhspTemplatesRepo {
	return &AhspTemplatesRepo{}
}

// TODO: Implement FindById [AhspTemplatesRepo]
func (r *AhspTemplatesRepo) FindById(tx *sql.Tx, ahspTemplateId int) (models.AHSPTemplate, error) {
	return models.AHSPTemplate{}, nil
}

// TODO: Implement Create [AhspTemplatesRepo]
func (r *AhspTemplatesRepo) Create(tx *sql.Tx, userData models.AHSPTemplate) error {
	return nil
}
