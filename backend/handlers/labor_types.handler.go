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
	"github.com/momokii/go-rab-maker/backend/repository/master_labor_types"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

type LaborTypeHandler struct {
	dbService      databases.SQLiteServices
	laborTypesRepo master_labor_types.MasterLaborTypesRepo
}

func NewLaborTypeHandler(
	dbService databases.SQLiteServices,
	laborTypesRepo master_labor_types.MasterLaborTypesRepo,
) *LaborTypeHandler {
	return &LaborTypeHandler{
		dbService:      dbService,
		laborTypesRepo: laborTypesRepo,
	}
}

// ==========================
// ========================== VIEWS
// ==========================

func (h *LaborTypeHandler) LaborTypesMainPageTableView(c *fiber.Ctx) error {
	var laborTypes []models.MasterLaborType
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

	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// start transaction to get the data
	if _, err := h.dbService.Transaction(
		c.Context(),
		func(tx *sql.Tx) (int, error) {
			// get the labor type data
			laborTypesData, paginationData, err := h.laborTypesRepo.Find(
				tx, paginationData, userData.ID,
			)
			if err != nil {
				return fiber.StatusInternalServerError, err
			}

			laborTypes = laborTypesData

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
		BaseURL:           "/labor_types",
		Title:             "Master Labor Types",
		SearchEnabled:     true,
		PaginationEnabled: true,
		PerPageEnabled:    true,
	}

	if c.Get("HX-Request") == "true" {
		tableComponents := components.LaborTypesTablePage(laborTypes, paginationInfo, tableConfig)
		return adaptor.HTTPHandler(templ.Handler(tableComponents))(c)
	}

	laborTypesComponent := components.LaborTypesPage(
		laborTypes,
		paginationInfo,
		tableConfig,
	)

	return adaptor.HTTPHandler(templ.Handler(laborTypesComponent))(c)
}

func (h *LaborTypeHandler) LaborTypeCreateModalView(c *fiber.Ctx) error {
	modal := components.LaborTypesFormModal(
		"Add New Labor Type",
		"/labor_types/new",
		"new-labor-type-form",
		"Add Labor Type",
		models.MasterLaborType{},
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *LaborTypeHandler) LaborTypeEditModalView(c *fiber.Ctx) error {
	laborTypeIdStr := c.Params("id")
	laborTypeId, err := strconv.Atoi(laborTypeIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid labor type ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var laborType models.MasterLaborType

	// Fetch labor type from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		laborType, err = h.laborTypesRepo.FindById(tx, laborTypeId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Labor type not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if labor type belongs to current user
		if laborType.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch labor type")
	}

	modal := components.LaborTypesFormModal(
		"Edit Labor Type",
		"/labor_types/"+laborTypeIdStr+"/edit",
		"edit-labor-type-form",
		"Update Labor Type",
		laborType,
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *LaborTypeHandler) LaborTypeDeleteModalView(c *fiber.Ctx) error {
	laborTypeIdStr := c.Params("id")
	laborTypeId, err := strconv.Atoi(laborTypeIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid labor type ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var laborType models.MasterLaborType

	// Fetch labor type from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		laborType, err = h.laborTypesRepo.FindById(tx, laborTypeId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Labor type not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if labor type belongs to current user
		if laborType.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch labor type")
	}

	modal := components.ConfirmationDeleteModal(
		"Delete Labor Type",
		"Are you sure you want to delete this labor type "+laborType.RoleName+"?",
		"/labor_types/"+laborTypeIdStr+"/delete",
		"Delete Labor Type",
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

// ==========================
// ========================== FUNCTIONS
// ==========================

// CreateLaborType handles the creation of a new labor type
func (h *LaborTypeHandler) CreateLaborType(c *fiber.Ctx) error {

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	roleName := c.FormValue("role_name")
	unit := c.FormValue("unit")
	defaultWageStr := c.FormValue("default_daily_wage")

	// Validate input
	if roleName == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Role name is required")
	}
	if unit == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Unit is required")
	}
	if defaultWageStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Default daily wage is required")
	}

	// Convert wage to float64
	defaultWage, err := strconv.ParseFloat(defaultWageStr, 64)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid default wage format")
	}

	// Create labor type data
	laborTypeData := models.MasterLaborTypeCreate{
		RoleName:         roleName,
		Unit:             unit,
		DefaultDailyWage: defaultWage,
		UserId:           userData.ID,
	}

	// Create labor type in database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		if err := h.laborTypesRepo.Create(tx, laborTypeData); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to create labor type, make sure the role name is unique")
	}

	// refresh table
	utils.SetRefreshTableTriggerHeader(c)

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "Labor type created successfully", true)
}

// UpdateLaborType handles the update of an existing labor type
func (h *LaborTypeHandler) UpdateLaborType(c *fiber.Ctx) error {

	laborTypeIdStr := c.Params("id")
	laborTypeId, err := strconv.Atoi(laborTypeIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid labor type ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	roleName := c.FormValue("role_name")
	unit := c.FormValue("unit")
	defaultWageStr := c.FormValue("default_daily_wage")

	// Validate input
	if roleName == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Role name is required")
	}
	if unit == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Unit is required")
	}
	if defaultWageStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Default daily wage is required")
	}

	// Convert wage to float64
	defaultWage, err := strconv.ParseFloat(defaultWageStr, 64)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid default wage format")
	}

	// Update labor type in database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// First, fetch the existing labor type to ensure it belongs to the user
		existingLaborType, err := h.laborTypesRepo.FindById(tx, laborTypeId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Labor type not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if labor type belongs to current user
		if existingLaborType.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Update the labor type
		updatedLaborType := models.MasterLaborType{
			LaborTypeId:      laborTypeId,
			UserId:           userData.ID,
			RoleName:         roleName,
			Unit:             unit,
			DefaultDailyWage: defaultWage,
			CreatedAt:        existingLaborType.CreatedAt,
			UpdatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		}

		if err := h.laborTypesRepo.Update(tx, updatedLaborType); err != nil {
			return fiber.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "Make sure Role Name is Unique")
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to update labor type: "+err.Error())
	}

	// refresh table
	utils.SetRefreshTableTriggerHeader(c)

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "Labor type updated successfully", true)
}

// DeleteLaborType handles the deletion of a labor type
func (h *LaborTypeHandler) DeleteLaborType(c *fiber.Ctx) error {

	laborTypeIdStr := c.Params("id")
	laborTypeId, err := strconv.Atoi(laborTypeIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid labor type ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Delete labor type from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// First, fetch the existing labor type to ensure it belongs to the user
		existingLaborType, err := h.laborTypesRepo.FindById(tx, laborTypeId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Labor type not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if labor type belongs to current user
		if existingLaborType.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Delete the labor type
		if err := h.laborTypesRepo.Delete(tx, existingLaborType); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to delete labor type: "+err.Error())
	}

	// refresh table
	utils.SetRefreshTableTriggerHeader(c)

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "Labor type deleted successfully", true)
}
