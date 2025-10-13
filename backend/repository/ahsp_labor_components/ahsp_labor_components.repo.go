package ahsp_labor_components

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type AHSPLaborComponentsRepo struct{}

func NewAHSPLaborComponentsRepo() *AHSPLaborComponentsRepo {
	return &AHSPLaborComponentsRepo{}
}

// FindById retrieves an AHSP labor component by ID
func (r *AHSPLaborComponentsRepo) FindById(tx *sql.Tx, ahspLaborComponentId int) (models.AHSPLaborComponent, error) {
	var component models.AHSPLaborComponent

	query := "SELECT component_id, template_id, labor_type_id, coefficient, created_at, updated_at FROM ahsp_labor_components WHERE component_id = ?"

	if err := tx.QueryRow(
		query,
		ahspLaborComponentId,
	).Scan(
		&component.ComponentId,
		&component.TemplateId,
		&component.LaborTypeId,
		&component.Coefficient,
		&component.CreatedAt,
		&component.UpdatedAt,
	); err != nil && err != sql.ErrNoRows {
		return component, err
	}

	return component, nil
}

// FindByTemplateId retrieves all AHSP labor components for a template
func (r *AHSPLaborComponentsRepo) FindByTemplateId(tx *sql.Tx, templateId int) ([]models.AHSPLaborComponent, error) {
	var components []models.AHSPLaborComponent

	query := "SELECT component_id, template_id, labor_type_id, coefficient, created_at, updated_at FROM ahsp_labor_components WHERE template_id = ? ORDER BY component_id"

	rows, err := tx.Query(query, templateId)
	if err != nil {
		return components, err
	}
	defer rows.Close()

	for rows.Next() {
		var component models.AHSPLaborComponent

		if err := rows.Scan(
			&component.ComponentId,
			&component.TemplateId,
			&component.LaborTypeId,
			&component.Coefficient,
			&component.CreatedAt,
			&component.UpdatedAt,
		); err != nil {
			return components, err
		} else {
			components = append(components, component)
		}
	}

	// if data nil, just return array
	if len(components) == 0 {
		return []models.AHSPLaborComponent{}, nil
	}

	return components, nil
}

// FindByTemplateIdWithLaborInfo retrieves all AHSP labor components for a template with labor type information
func (r *AHSPLaborComponentsRepo) FindByTemplateIdWithLaborInfo(tx *sql.Tx, templateId int) ([]models.AHSPLaborComponentWithLabor, error) {
	var components []models.AHSPLaborComponentWithLabor

	query := `SELECT
				alc.component_id,
				alc.template_id,
				alc.labor_type_id,
				alc.coefficient,
				alc.created_at,
				alc.updated_at,
				mlt.role_name,
				mlt.unit,
				mlt.default_daily_wage
			  FROM ahsp_labor_components alc
			  LEFT JOIN master_labor_types mlt ON alc.labor_type_id = mlt.labor_type_id
			  WHERE alc.template_id = ?
			  ORDER BY alc.component_id`

	rows, err := tx.Query(query, templateId)
	if err != nil {
		return components, err
	}
	defer rows.Close()

	for rows.Next() {
		var component models.AHSPLaborComponentWithLabor

		if err := rows.Scan(
			&component.ComponentId,
			&component.TemplateId,
			&component.LaborTypeId,
			&component.Coefficient,
			&component.CreatedAt,
			&component.UpdatedAt,
			&component.LaborTypeName,
			&component.LaborUnit,
			&component.LaborWage,
		); err != nil {
			return components, err
		} else {
			components = append(components, component)
		}
	}

	// if data nil, just return array
	if len(components) == 0 {
		return []models.AHSPLaborComponentWithLabor{}, nil
	}

	return components, nil
}

// Create creates a new AHSP labor component
func (r *AHSPLaborComponentsRepo) Create(tx *sql.Tx, componentData models.AHSPLaborComponentCreate) error {
	query := "INSERT INTO ahsp_labor_components (template_id, labor_type_id, coefficient) VALUES (?, ?, ?)"
	if _, err := tx.Exec(
		query,
		componentData.TemplateId,
		componentData.LaborTypeId,
		componentData.Coefficient,
	); err != nil {
		return err
	}

	return nil
}

// Update updates an existing AHSP labor component
func (r *AHSPLaborComponentsRepo) Update(tx *sql.Tx, componentId int, componentData models.AHSPLaborComponentUpdate) error {
	query := "UPDATE ahsp_labor_components SET coefficient = ? WHERE component_id = ?"
	if _, err := tx.Exec(
		query,
		componentData.Coefficient,
		componentId,
	); err != nil {
		return err
	}

	return nil
}

// Delete deletes an AHSP labor component
func (r *AHSPLaborComponentsRepo) Delete(tx *sql.Tx, componentId int) error {
	query := "DELETE FROM ahsp_labor_components WHERE component_id = ?"
	if _, err := tx.Exec(query, componentId); err != nil {
		return err
	}

	return nil
}

// DeleteByTemplateId deletes all AHSP labor components for a template
func (r *AHSPLaborComponentsRepo) DeleteByTemplateId(tx *sql.Tx, templateId int) error {
	query := "DELETE FROM ahsp_labor_components WHERE template_id = ?"
	if _, err := tx.Exec(query, templateId); err != nil {
		return err
	}

	return nil
}
