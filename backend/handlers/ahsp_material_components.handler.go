package handlers

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/momokii/go-rab-maker/backend/databases"
	"github.com/momokii/go-rab-maker/backend/middlewares"
	"github.com/momokii/go-rab-maker/backend/models"
	ahsp_labor_components "github.com/momokii/go-rab-maker/backend/repository/ahsp_labor_components"
	ahsp_material_components "github.com/momokii/go-rab-maker/backend/repository/ahsp_material_components"
	ahsptemplates "github.com/momokii/go-rab-maker/backend/repository/ahsp_templates"
	"github.com/momokii/go-rab-maker/backend/repository/master_materials"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

type AhspMaterialComponentHandler struct {
	dbService                  databases.SQLiteServices
	ahspMaterialComponentsRepo *ahsp_material_components.AHSPMaterialComponentsRepo
	ahspTemplatesRepo          *ahsptemplates.AhspTemplatesRepo
	materialsRepo              *master_materials.MasterMaterialsRepo
	ahspLaborComponentsRepo    *ahsp_labor_components.AHSPLaborComponentsRepo
}

func NewAhspMaterialComponentHandler(
	dbService databases.SQLiteServices,
	ahspMaterialComponentsRepo *ahsp_material_components.AHSPMaterialComponentsRepo,
	ahspTemplatesRepo *ahsptemplates.AhspTemplatesRepo,
	materialsRepo *master_materials.MasterMaterialsRepo,
	ahspLaborComponentsRepo *ahsp_labor_components.AHSPLaborComponentsRepo,

) *AhspMaterialComponentHandler {
	return &AhspMaterialComponentHandler{
		dbService:                  dbService,
		ahspMaterialComponentsRepo: ahspMaterialComponentsRepo,
		ahspTemplatesRepo:          ahspTemplatesRepo,
		materialsRepo:              materialsRepo,
		ahspLaborComponentsRepo:    ahspLaborComponentsRepo,
	}
}

// ==========================
// ========================== VIEWS
// ==========================

func (h *AhspMaterialComponentHandler) AhspMaterialComponentsPage(c *fiber.Ctx) error {
	templateIdStr := c.Params("templateId")
	templateId, err := strconv.Atoi(templateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid template ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var ahspTemplate models.AHSPTemplate
	var materialComponents []models.AHSPMaterialComponentWithMaterial
	var laborComponents []models.AHSPLaborComponentWithLabor
	var availableMaterials []models.MasterMaterial

	// Fetch data from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get AHSP template
		ahspTemplate, err = h.ahspTemplatesRepo.FindById(tx, templateId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "AHSP template not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if template belongs to current user
		if ahspTemplate.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get material components for this template
		materialComponents, err = h.ahspMaterialComponentsRepo.FindByTemplateIdWithMaterialInfo(tx, templateId)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		// get labor componetns
		laborComponents, err = h.ahspLaborComponentsRepo.FindByTemplateIdWithLaborInfo(tx, templateId)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		// Get available materials
		paginationData := models.TablePaginationDataInput{
			Page:    1,
			PerPage: 1000, // Get all materials
			Search:  "",
		}
		availableMaterials, _, err = h.materialsRepo.Find(tx, paginationData, userData.ID)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch data")
	}

	if c.Get("HX-Request") == "true" {
		components := components.AhspMaterialComponentsTabContent(ahspTemplate, materialComponents)
		return adaptor.HTTPHandler(templ.Handler(components))(c)
	}

	component := components.AhspMaterialComponentsPage(
		ahspTemplate,
		materialComponents,
		availableMaterials,
		laborComponents,
	)
	return adaptor.HTTPHandler(templ.Handler(component))(c)
}

func (h *AhspMaterialComponentHandler) AhspMaterialComponentCreateModalView(c *fiber.Ctx) error {
	templateIdStr := c.Params("templateId")
	templateId, err := strconv.Atoi(templateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid template ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var ahspTemplate models.AHSPTemplate
	var availableMaterials []models.MasterMaterial

	// Fetch data from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get AHSP template
		ahspTemplate, err = h.ahspTemplatesRepo.FindById(tx, templateId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "AHSP template not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if template belongs to current user
		if ahspTemplate.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get available materials
		paginationData := models.TablePaginationDataInput{
			Page:    1,
			PerPage: 1000, // Get all materials
			Search:  "",
		}
		availableMaterials, _, err = h.materialsRepo.Find(tx, paginationData, userData.ID)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch data")
	}

	modal := components.AhspMaterialComponentFormModal(
		"Add Material Component",
		"/ahsp_templates/"+templateIdStr+"/material_components/new",
		"new-material-component-form",
		"Add Material Component",
		models.AHSPMaterialComponent{},
		ahspTemplate,
		availableMaterials,
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *AhspMaterialComponentHandler) AhspMaterialComponentEditModalView(c *fiber.Ctx) error {
	templateIdStr := c.Params("templateId")
	componentIdStr := c.Params("componentId")

	templateId, err := strconv.Atoi(templateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid template ID")
	}

	componentId, err := strconv.Atoi(componentIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid component ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var ahspTemplate models.AHSPTemplate
	var materialComponent models.AHSPMaterialComponent
	var availableMaterials []models.MasterMaterial

	// Fetch data from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get AHSP template
		ahspTemplate, err = h.ahspTemplatesRepo.FindById(tx, templateId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "AHSP template not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if template belongs to current user
		if ahspTemplate.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get material component
		materialComponent, err = h.ahspMaterialComponentsRepo.FindById(tx, componentId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Material component not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Get available materials
		paginationData := models.TablePaginationDataInput{
			Page:    1,
			PerPage: 1000, // Get all materials
			Search:  "",
		}
		availableMaterials, _, err = h.materialsRepo.Find(tx, paginationData, userData.ID)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch data")
	}

	modal := components.AhspMaterialComponentFormModal(
		"Edit Material Component",
		"/ahsp_templates/"+templateIdStr+"/material_components/"+componentIdStr+"/edit",
		"edit-material-component-form",
		"Update Material Component",
		materialComponent,
		ahspTemplate,
		availableMaterials,
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *AhspMaterialComponentHandler) AhspMaterialComponentDeleteModalView(c *fiber.Ctx) error {
	templateIdStr := c.Params("templateId")
	componentIdStr := c.Params("componentId")

	templateId, err := strconv.Atoi(templateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid template ID")
	}

	componentId, err := strconv.Atoi(componentIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid component ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var ahspTemplate models.AHSPTemplate
	var materialComponent models.AHSPMaterialComponentWithMaterial

	// Fetch data from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get AHSP template
		ahspTemplate, err = h.ahspTemplatesRepo.FindById(tx, templateId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "AHSP template not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if template belongs to current user
		if ahspTemplate.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get material component with material info
		components, err := h.ahspMaterialComponentsRepo.FindByTemplateIdWithMaterialInfo(tx, templateId)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		// Find the specific component
		for _, comp := range components {
			if comp.ComponentId == componentId {
				materialComponent = comp
				break
			}
		}

		if materialComponent.ComponentId == 0 {
			return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Material component not found")
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch data")
	}

	modal := components.ConfirmationDeleteModal(
		"Delete Material Component",
		"Are you sure you want to delete this material component "+materialComponent.MaterialName+"?",
		"/ahsp_templates/"+templateIdStr+"/material_components/"+componentIdStr+"/delete",
		"Delete Material Component",
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

// ==========================
// ========================== FUNCTIONS
// ==========================

// CreateAhspMaterialComponent handles the creation of a new AHSP material component
func (h *AhspMaterialComponentHandler) CreateAhspMaterialComponent(c *fiber.Ctx) error {
	// Add small delay for better UX
	time.Sleep(500 * time.Millisecond)

	templateIdStr := c.Params("templateId")
	templateId, err := strconv.Atoi(templateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid template ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	materialIdStr := c.FormValue("material_id")
	coefficientStr := c.FormValue("coefficient")

	// Validate input
	if materialIdStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Material is required")
	}
	if coefficientStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Coefficient is required")
	}

	// Convert to proper types
	materialId, err := strconv.Atoi(materialIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid material ID")
	}

	coefficient, err := strconv.ParseFloat(coefficientStr, 64)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid coefficient format")
	}

	// Create AHSP material component data
	componentData := models.AHSPMaterialComponentCreate{
		TemplateId:  templateId,
		MaterialId:  materialId,
		Coefficient: coefficient,
	}

	// Create AHSP material component in database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Check if template belongs to current user
		template, err := h.ahspTemplatesRepo.FindById(tx, templateId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "AHSP template not found")
			}
			return fiber.StatusInternalServerError, err
		}

		if template.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		if err := h.ahspMaterialComponentsRepo.Create(tx, componentData); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to create material component")
	}

	// Return success response with redirect to template detail page
	return utils.ResponseSuccessWithRedirect(c, "Success", "Material component created successfully", "/ahsp_templates/"+templateIdStr)
}

// UpdateAhspMaterialComponent handles the update of an existing AHSP material component
func (h *AhspMaterialComponentHandler) UpdateAhspMaterialComponent(c *fiber.Ctx) error {
	// Add small delay for better UX
	time.Sleep(500 * time.Millisecond)

	templateIdStr := c.Params("templateId")
	componentIdStr := c.Params("componentId")

	templateId, err := strconv.Atoi(templateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid template ID")
	}

	componentId, err := strconv.Atoi(componentIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid component ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	coefficientStr := c.FormValue("coefficient")

	// Validate input
	if coefficientStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Coefficient is required")
	}

	// Convert to proper types
	coefficient, err := strconv.ParseFloat(coefficientStr, 64)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid coefficient format")
	}

	// Update AHSP material component in database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Check if template belongs to current user
		template, err := h.ahspTemplatesRepo.FindById(tx, templateId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "AHSP template not found")
			}
			return fiber.StatusInternalServerError, err
		}

		if template.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Update the material component
		componentData := models.AHSPMaterialComponentUpdate{
			Coefficient: coefficient,
		}

		if err := h.ahspMaterialComponentsRepo.Update(tx, componentId, componentData); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to update material component")
	}

	// Return success response with redirect to template detail page
	return utils.ResponseSuccessWithRedirect(c, "Success", "Material component updated successfully", "/ahsp_templates/"+templateIdStr)
}

// DeleteAhspMaterialComponent handles the deletion of an AHSP material component
func (h *AhspMaterialComponentHandler) DeleteAhspMaterialComponent(c *fiber.Ctx) error {
	// Add small delay for better UX
	time.Sleep(500 * time.Millisecond)

	templateIdStr := c.Params("templateId")
	componentIdStr := c.Params("componentId")

	templateId, err := strconv.Atoi(templateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid template ID")
	}

	componentId, err := strconv.Atoi(componentIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid component ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Delete AHSP material component from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Check if template belongs to current user
		template, err := h.ahspTemplatesRepo.FindById(tx, templateId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "AHSP template not found")
			}
			return fiber.StatusInternalServerError, err
		}

		if template.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Delete the material component
		if err := h.ahspMaterialComponentsRepo.Delete(tx, componentId); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to delete material component")
	}

	// Return success response with redirect to template detail page
	return utils.ResponseSuccessWithRedirect(c, "Success", "Material component deleted successfully", "/ahsp_templates/"+templateIdStr)
}
