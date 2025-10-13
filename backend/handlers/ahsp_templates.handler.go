package handlers

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/momokii/go-rab-maker/backend/databases"
	"github.com/momokii/go-rab-maker/backend/middlewares"
	"github.com/momokii/go-rab-maker/backend/models"
	ahsptemplates "github.com/momokii/go-rab-maker/backend/repository/ahsp_templates"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

type AhspTemplateHandler struct {
	dbService         databases.SQLiteServices
	ahspTemplatesRepo *ahsptemplates.AhspTemplatesRepo
}

func NewAhspTemplateHandler(
	dbService databases.SQLiteServices,
	ahspTemplatesRepo *ahsptemplates.AhspTemplatesRepo,
) *AhspTemplateHandler {
	return &AhspTemplateHandler{
		dbService:         dbService,
		ahspTemplatesRepo: ahspTemplatesRepo,
	}
}

// ==========================
// ========================== VIEWS
// ==========================

func (h *AhspTemplateHandler) AhspTemplatesMainPageTableView(c *fiber.Ctx) error {
	var ahspTemplates []models.AHSPTemplate
	var paginationInfo models.PaginationInfo

	// get pagination data
	paginationData, err := utils.GetPaginationData(c)
	if err != nil {
		return utils.ResponseErrorModal(
			c,
			"Error",
			"Failed process to get pagination data",
		)
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)
	log.Println(userData)

	// start transaction to get the data
	if _, err := h.dbService.Transaction(
		c.Context(),
		func(tx *sql.Tx) (int, error) {
			// get the AHSP template data
			ahspTemplatesData, paginationData, err := h.ahspTemplatesRepo.Find(
				tx, paginationData, userData.ID,
			)
			if err != nil {
				return fiber.StatusInternalServerError, err
			}

			ahspTemplates = ahspTemplatesData

			paginationInfo = paginationData

			return fiber.StatusOK, nil
		},
	); err != nil {
		return utils.ResponseErrorModal(
			c,
			"Error",
			err.Error(),
		)
	}

	// base data for table
	tableConfig := models.TableConfig{
		BaseURL:           "/ahsp_templates",
		Title:             "AHSP Templates",
		SearchEnabled:     true,
		PaginationEnabled: true,
		PerPageEnabled:    true,
	}

	if c.Get("HX-Request") == "true" {
		tableComponents := components.AhspTemplatesTablePage(ahspTemplates, paginationInfo, tableConfig)
		return adaptor.HTTPHandler(templ.Handler(tableComponents))(c)
	}

	ahspTemplatesComponent := components.AhspTemplatesPage(
		ahspTemplates,
		paginationInfo,
		tableConfig,
	)
	return adaptor.HTTPHandler(templ.Handler(ahspTemplatesComponent))(c)
}

func (h *AhspTemplateHandler) AhspTemplateCreateModalView(c *fiber.Ctx) error {
	modal := components.AhspTemplatesFormModal(
		"Add New AHSP Template",
		"/ahsp_templates/new",
		"new-ahsp-template-form",
		"Add AHSP Template",
		models.AHSPTemplate{},
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *AhspTemplateHandler) AhspTemplateEditModalView(c *fiber.Ctx) error {
	ahspTemplateIdStr := c.Params("id")
	ahspTemplateId, err := strconv.Atoi(ahspTemplateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid AHSP template ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var ahspTemplate models.AHSPTemplate

	// Fetch AHSP template from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		ahspTemplate, err = h.ahspTemplatesRepo.FindById(tx, ahspTemplateId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "AHSP template not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if AHSP template belongs to current user
		if ahspTemplate.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch AHSP template")
	}

	modal := components.AhspTemplatesFormModal(
		"Edit AHSP Template",
		"/ahsp_templates/"+ahspTemplateIdStr+"/edit",
		"edit-ahsp-template-form",
		"Update AHSP Template",
		ahspTemplate,
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *AhspTemplateHandler) AhspTemplateDeleteModalView(c *fiber.Ctx) error {
	ahspTemplateIdStr := c.Params("id")
	ahspTemplateId, err := strconv.Atoi(ahspTemplateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid AHSP template ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var ahspTemplate models.AHSPTemplate

	// Fetch AHSP template from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		ahspTemplate, err = h.ahspTemplatesRepo.FindById(tx, ahspTemplateId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "AHSP template not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if AHSP template belongs to current user
		if ahspTemplate.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch AHSP template")
	}

	modal := components.ConfirmationDeleteModal(
		"Delete AHSP Template",
		"Are you sure you want to delete this AHSP template "+ahspTemplate.TemplateName+"?",
		"/ahsp_templates/"+ahspTemplateIdStr+"/delete",
		"Delete AHSP Template",
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

// ==========================
// ========================== FUNCTIONS
// ==========================

// CreateAhspTemplate handles the creation of a new AHSP template
func (h *AhspTemplateHandler) CreateAhspTemplate(c *fiber.Ctx) error {
	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	templateName := c.FormValue("template_name")
	unit := c.FormValue("unit")

	// Validate input
	if templateName == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Template name is required")
	}
	if unit == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Unit is required")
	}

	// Create AHSP template data
	ahspTemplateData := models.AHSPTemplateCreate{
		TemplateName: templateName,
		Unit:         unit,
		UserId:       userData.ID,
	}

	// Create AHSP template in database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		if err := h.ahspTemplatesRepo.Create(tx, ahspTemplateData); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to create AHSP template, make sure template name is unique")
	}

	// refresh table data
	utils.SetRefreshTableTriggerHeader(c)

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "AHSP template created successfully", true)
}

// UpdateAhspTemplate handles the update of an existing AHSP template
func (h *AhspTemplateHandler) UpdateAhspTemplate(c *fiber.Ctx) error {
	ahspTemplateIdStr := c.Params("id")
	ahspTemplateId, err := strconv.Atoi(ahspTemplateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid AHSP template ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	templateName := c.FormValue("template_name")
	unit := c.FormValue("unit")

	// Validate input
	if templateName == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Template name is required")
	}
	if unit == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Unit is required")
	}

	// Update AHSP template in database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// First, fetch the existing AHSP template to ensure it belongs to the user
		existingAhspTemplate, err := h.ahspTemplatesRepo.FindById(tx, ahspTemplateId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "AHSP template not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if AHSP template belongs to current user
		if existingAhspTemplate.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Update the AHSP template
		updatedAhspTemplate := models.AHSPTemplate{
			TemplateId:   ahspTemplateId,
			UserId:       userData.ID,
			TemplateName: templateName,
			Unit:         unit,
			CreatedAt:    existingAhspTemplate.CreatedAt,
			UpdatedAt:    time.Now().Format("2006-01-02 15:04:05"),
		}

		if err := h.ahspTemplatesRepo.Update(tx, updatedAhspTemplate); err != nil {
			return fiber.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "Make sure template name is unique")
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to update AHSP template: "+err.Error())
	}

	// refresh table data
	utils.SetRefreshTableTriggerHeader(c)

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "AHSP template updated successfully", true)
}

// DeleteAhspTemplate handles the deletion of an AHSP template
func (h *AhspTemplateHandler) DeleteAhspTemplate(c *fiber.Ctx) error {
	ahspTemplateIdStr := c.Params("id")
	ahspTemplateId, err := strconv.Atoi(ahspTemplateIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid AHSP template ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Delete AHSP template from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// First, fetch the existing AHSP template to ensure it belongs to the user
		existingAhspTemplate, err := h.ahspTemplatesRepo.FindById(tx, ahspTemplateId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "AHSP template not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if AHSP template belongs to current user
		if existingAhspTemplate.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Delete the AHSP template
		if err := h.ahspTemplatesRepo.Delete(tx, existingAhspTemplate); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to delete AHSP template")
	}

	// refresh table data
	utils.SetRefreshTableTriggerHeader(c)

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "AHSP template deleted successfully", true)
}
