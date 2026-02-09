package ahsp_material_components

import (
	"database/sql"

	"github.com/momokii/go-rab-maker/backend/models"
)

type AHSPMaterialComponentsRepo struct{}

func NewAHSPMaterialComponentsRepo() *AHSPMaterialComponentsRepo {
	return &AHSPMaterialComponentsRepo{}
}

// FindById retrieves an AHSP material component by ID
func (r *AHSPMaterialComponentsRepo) FindById(tx *sql.Tx, ahspMaterialComponentId int) (models.AHSPMaterialComponent, error) {
	var component models.AHSPMaterialComponent

	query := "SELECT component_id, template_id, material_id, coefficient, created_at, updated_at FROM ahsp_material_components WHERE component_id = ?"

	if err := tx.QueryRow(
		query,
		ahspMaterialComponentId,
	).Scan(
		&component.ComponentId,
		&component.TemplateId,
		&component.MaterialId,
		&component.Coefficient,
		&component.CreatedAt,
		&component.UpdatedAt,
	); err != nil && err != sql.ErrNoRows {
		return component, err
	}

	return component, nil
}

// FindByTemplateId retrieves all AHSP material components for a template
func (r *AHSPMaterialComponentsRepo) FindByTemplateId(tx *sql.Tx, templateId int) ([]models.AHSPMaterialComponent, error) {
	var components []models.AHSPMaterialComponent

	query := "SELECT component_id, template_id, material_id, coefficient, created_at, updated_at FROM ahsp_material_components WHERE template_id = ? ORDER BY component_id"

	rows, err := tx.Query(query, templateId)
	if err != nil {
		return components, err
	}
	defer rows.Close()

	for rows.Next() {
		var component models.AHSPMaterialComponent

		if err := rows.Scan(
			&component.ComponentId,
			&component.TemplateId,
			&component.MaterialId,
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
		return []models.AHSPMaterialComponent{}, nil
	}

	return components, nil
}

// FindByTemplateIdWithMaterialInfo retrieves all AHSP material components for a template with material information
func (r *AHSPMaterialComponentsRepo) FindByTemplateIdWithMaterialInfo(tx *sql.Tx, templateId int) ([]models.AHSPMaterialComponentWithMaterial, error) {
	var components []models.AHSPMaterialComponentWithMaterial

	query := `SELECT
				amc.component_id,
				amc.template_id,
				amc.material_id,
				amc.coefficient,
				amc.created_at,
				amc.updated_at,
				mm.material_name,
				mm.unit,
				mm.default_unit_price
			  FROM ahsp_material_components amc
			  LEFT JOIN master_materials mm ON amc.material_id = mm.material_id
			  WHERE amc.template_id = ?
			  ORDER BY amc.component_id`

	rows, err := tx.Query(query, templateId)
	if err != nil {
		return components, err
	}
	defer rows.Close()

	for rows.Next() {
		var component models.AHSPMaterialComponentWithMaterial

		if err := rows.Scan(
			&component.ComponentId,
			&component.TemplateId,
			&component.MaterialId,
			&component.Coefficient,
			&component.CreatedAt,
			&component.UpdatedAt,
			&component.MaterialName,
			&component.MaterialUnit,
			&component.MaterialPrice,
		); err != nil {
			return components, err
		} else {
			components = append(components, component)
		}
	}

	// if data nil, just return array
	if len(components) == 0 {
		return []models.AHSPMaterialComponentWithMaterial{}, nil
	}

	return components, nil
}

// Create creates a new AHSP material component
func (r *AHSPMaterialComponentsRepo) Create(tx *sql.Tx, componentData models.AHSPMaterialComponentCreate) error {
	query := "INSERT INTO ahsp_material_components (template_id, material_id, coefficient) VALUES (?, ?, ?)"
	if _, err := tx.Exec(
		query,
		componentData.TemplateId,
		componentData.MaterialId,
		componentData.Coefficient,
	); err != nil {
		return err
	}

	return nil
}

// Update updates an existing AHSP material component
func (r *AHSPMaterialComponentsRepo) Update(tx *sql.Tx, componentId int, componentData models.AHSPMaterialComponentUpdate) error {
	query := "UPDATE ahsp_material_components SET coefficient = ? WHERE component_id = ?"
	if _, err := tx.Exec(
		query,
		componentData.Coefficient,
		componentId,
	); err != nil {
		return err
	}

	return nil
}

// Delete deletes an AHSP material component
func (r *AHSPMaterialComponentsRepo) Delete(tx *sql.Tx, componentId int) error {
	query := "DELETE FROM ahsp_material_components WHERE component_id = ?"
	if _, err := tx.Exec(query, componentId); err != nil {
		return err
	}

	return nil
}

// DeleteByTemplateId deletes all AHSP material components for a template
func (r *AHSPMaterialComponentsRepo) DeleteByTemplateId(tx *sql.Tx, templateId int) error {
	query := "DELETE FROM ahsp_material_components WHERE template_id = ?"
	if _, err := tx.Exec(query, templateId); err != nil {
		return err
	}

	return nil
}
