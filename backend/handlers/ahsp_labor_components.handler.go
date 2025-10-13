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
	ahsptemplates "github.com/momokii/go-rab-maker/backend/repository/ahsp_templates"
	"github.com/momokii/go-rab-maker/backend/repository/master_labor_types"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

type AhspLaborComponentHandler struct {
	dbService               databases.SQLiteServices
	ahspLaborComponentsRepo *ahsp_labor_components.AHSPLaborComponentsRepo
	ahspTemplatesRepo       *ahsptemplates.AhspTemplatesRepo
	laborTypesRepo          *master_labor_types.MasterLaborTypesRepo
}

func NewAhspLaborComponentHandler(
	dbService databases.SQLiteServices,
	ahspLaborComponentsRepo *ahsp_labor_components.AHSPLaborComponentsRepo,
	ahspTemplatesRepo *ahsptemplates.AhspTemplatesRepo,
	laborTypesRepo *master_labor_types.MasterLaborTypesRepo,
) *AhspLaborComponentHandler {
	return &AhspLaborComponentHandler{
		dbService:               dbService,
		ahspLaborComponentsRepo: ahspLaborComponentsRepo,
		ahspTemplatesRepo:       ahspTemplatesRepo,
		laborTypesRepo:          laborTypesRepo,
	}
}

// ==========================
// ========================== VIEWS
// ==========================

func (h *AhspLaborComponentHandler) AhspLaborComponentsPage(c *fiber.Ctx) error {
	templateIdStr := c.Params("templateId")
	templateId, err := strconv.Atoi(templateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid template ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var ahspTemplate models.AHSPTemplate
	var laborComponents []models.AHSPLaborComponentWithLabor
	var availableLaborTypes []models.MasterLaborType

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

		// Get labor components for this template
		laborComponents, err = h.ahspLaborComponentsRepo.FindByTemplateIdWithLaborInfo(tx, templateId)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		// Get all available labor types for this user
		paginationData := models.TablePaginationDataInput{
			Page:    1,
			PerPage: 1000, // Get all labor types
		}
		availableLaborTypes, _, err = h.laborTypesRepo.Find(tx, paginationData, userData.ID)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch data")
	}

	// Render the labor components tab content
	laborComponentsTab := components.AhspLaborComponentsTabContent(ahspTemplate, laborComponents, availableLaborTypes)
	return adaptor.HTTPHandler(templ.Handler(laborComponentsTab))(c)
}

func (h *AhspLaborComponentHandler) AhspLaborComponentCreateModalView(c *fiber.Ctx) error {
	templateIdStr := c.Params("templateId")
	templateId, err := strconv.Atoi(templateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid template ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var ahspTemplate models.AHSPTemplate
	var availableLaborTypes []models.MasterLaborType

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

		// Get all available labor types for this user
		paginationData := models.TablePaginationDataInput{
			Page:    1,
			PerPage: 1000, // Get all labor types
		}
		availableLaborTypes, _, err = h.laborTypesRepo.Find(tx, paginationData, userData.ID)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch data")
	}

	// Create an empty labor component for the form
	emptyComponent := models.AHSPLaborComponent{
		ComponentId: 0,
		Coefficient: 0.0,
	}

	// Render the labor component form modal
	modal := components.AhspLaborComponentFormModal(
		"Add Labor Component",
		"/ahsp_templates/"+templateIdStr+"/labor_components/new",
		"labor-component-form",
		"Add Labor Component",
		emptyComponent,
		ahspTemplate,
		availableLaborTypes,
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *AhspLaborComponentHandler) AhspLaborComponentEditModalView(c *fiber.Ctx) error {
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
	var laborComponent models.AHSPLaborComponent
	var availableLaborTypes []models.MasterLaborType

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

		// Get the labor component to edit
		laborComponent, err = h.ahspLaborComponentsRepo.FindById(tx, componentId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Labor component not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if component belongs to the template
		if laborComponent.TemplateId != templateId {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get all available labor types for this user
		paginationData := models.TablePaginationDataInput{
			Page:    1,
			PerPage: 1000, // Get all labor types
		}
		availableLaborTypes, _, err = h.laborTypesRepo.Find(tx, paginationData, userData.ID)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch data")
	}

	// Render the labor component form modal
	modal := components.AhspLaborComponentFormModal(
		"Edit Labor Component",
		"/ahsp_templates/"+templateIdStr+"/labor_components/"+componentIdStr+"/edit",
		"labor-component-form",
		"Update Labor Component",
		laborComponent,
		ahspTemplate,
		availableLaborTypes,
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *AhspLaborComponentHandler) AhspLaborComponentDeleteModalView(c *fiber.Ctx) error {
	templateIdStr := c.Params("templateId")
	componentIdStr := c.Params("componentId")

	templateId, err := strconv.Atoi(templateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid template ID")
	}

	_, err = strconv.Atoi(componentIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid component ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var ahspTemplate models.AHSPTemplate

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

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch data")
	}

	// For now, return a simple confirmation modal
	modal := components.ConfirmationDeleteModal(
		"Delete Labor Component",
		"Are you sure you want to delete this labor component?",
		"/ahsp_templates/"+templateIdStr+"/labor_components/"+componentIdStr+"/delete",
		"Delete Labor Component",
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

// ==========================
// ========================== FUNCTIONS
// ==========================

// CreateAhspLaborComponent handles the creation of a new AHSP labor component
func (h *AhspLaborComponentHandler) CreateAhspLaborComponent(c *fiber.Ctx) error {
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
	laborTypeIdStr := c.FormValue("labor_type_id")
	coefficientStr := c.FormValue("coefficient")

	// Validate input
	if laborTypeIdStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Labor type is required")
	}
	if coefficientStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Coefficient is required")
	}

	// Convert to proper types
	laborTypeId, err := strconv.Atoi(laborTypeIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid labor type ID")
	}

	coefficient, err := strconv.ParseFloat(coefficientStr, 64)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid coefficient format")
	}

	// Create AHSP labor component data
	componentData := models.AHSPLaborComponentCreate{
		TemplateId:  templateId,
		LaborTypeId: laborTypeId,
		Coefficient: coefficient,
	}

	// Create AHSP labor component in database
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

		if err := h.ahspLaborComponentsRepo.Create(tx, componentData); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to create labor component")
	}

	// Return success response with redirect to template detail page
	return utils.ResponseSuccessWithRedirect(c, "Success", "Labor component created successfully", "/ahsp_templates/"+templateIdStr)
}

// UpdateAhspLaborComponent handles the update of an existing AHSP labor component
func (h *AhspLaborComponentHandler) UpdateAhspLaborComponent(c *fiber.Ctx) error {
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

	// Update AHSP labor component in database
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

		// Update the labor component
		componentData := models.AHSPLaborComponentUpdate{
			Coefficient: coefficient,
		}

		if err := h.ahspLaborComponentsRepo.Update(tx, componentId, componentData); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to update labor component")
	}

	// Return success response with redirect to template detail page
	return utils.ResponseSuccessWithRedirect(c, "Success", "Labor component updated successfully", "/ahsp_templates/"+templateIdStr)
}

// DeleteAhspLaborComponent handles the deletion of an AHSP labor component
func (h *AhspLaborComponentHandler) DeleteAhspLaborComponent(c *fiber.Ctx) error {
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

	// Delete AHSP labor component from database
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

		// Delete the labor component
		if err := h.ahspLaborComponentsRepo.Delete(tx, componentId); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to delete labor component")
	}

	// Return success response with redirect to template detail page
	return utils.ResponseSuccessWithRedirect(c, "Success", "Labor component deleted successfully", "/ahsp_templates/"+templateIdStr)
}
