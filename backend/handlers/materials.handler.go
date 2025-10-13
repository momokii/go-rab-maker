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
	"github.com/momokii/go-rab-maker/backend/repository/master_materials"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

type MaterialHandler struct {
	dbService     databases.SQLiteServices
	materialsRepo master_materials.MasterMaterialsRepo
}

func NewMaterialsHandler(
	dbService databases.SQLiteServices,
	materialsRepo master_materials.MasterMaterialsRepo,
) *MaterialHandler {
	return &MaterialHandler{
		dbService:     dbService,
		materialsRepo: materialsRepo,
	}
}

// ==========================
// ========================== VIEWS
// ==========================
func (h *MaterialHandler) MaterialsMainPageTableView(c *fiber.Ctx) error {

	var materials []models.MasterMaterial
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

			// get the material data
			materialsData, paginationData, err := h.materialsRepo.Find(
				tx, paginationData, userData.ID,
			)
			if err != nil {
				return fiber.StatusInternalServerError, err
			}

			paginationInfo = paginationData
			materials = materialsData

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
		BaseURL:           "/materials",
		Title:             "Master Materials",
		SearchEnabled:     true,
		PaginationEnabled: true,
		PerPageEnabled:    true,
	}

	if c.Get("HX-Request") == "true" {
		tableComponents := components.MaterialsTablePage(materials, paginationInfo, tableConfig)
		return adaptor.HTTPHandler(templ.Handler(tableComponents))(c)
	}

	materialsComponent := components.MaterialsPage(
		materials,
		paginationInfo,
		tableConfig,
	)
	return adaptor.HTTPHandler(templ.Handler(materialsComponent))(c)
}

func (h *MaterialHandler) MaterialCreateModalView(c *fiber.Ctx) error {
	modal := components.MaterialsFormModal(
		"Add New Material",
		"/materials/new",
		"new-materials-form",
		"Add Materials",
		models.MasterMaterial{},
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *MaterialHandler) MaterialEditModalView(c *fiber.Ctx) error {
	materialIdStr := c.Params("id")
	materialId, err := strconv.Atoi(materialIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid material ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var material models.MasterMaterial

	// Fetch material from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		material, err = h.materialsRepo.FindById(tx, materialId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Material not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if material belongs to current user
		if material.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch material: "+err.Error())
	}

	modal := components.MaterialsFormModal(
		"Edit Material",
		"/materials/"+materialIdStr+"/edit",
		"edit-material-form",
		"Update Material",
		material,
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *MaterialHandler) MaterialDeleteModalView(c *fiber.Ctx) error {
	materialIdStr := c.Params("id")
	materialId, err := strconv.Atoi(materialIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid material ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var material models.MasterMaterial

	// Fetch material from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		material, err = h.materialsRepo.FindById(tx, materialId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Material not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if material belongs to current user
		if material.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch material: "+err.Error())
	}

	modal := components.ConfirmationDeleteModal(
		"Delete Material",
		"Are you sure you want to delete this material "+material.MaterialName+"?",
		"/materials/"+materialIdStr+"/delete",
		"Delete Material",
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

// ==========================
// ========================== FUNCTIONS
// ==========================

// CreateMaterial handles the creation of a new material
func (h *MaterialHandler) CreateMaterial(c *fiber.Ctx) error {

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	materialName := c.FormValue("material_name")
	unit := c.FormValue("material_unit")
	defaultPriceStr := c.FormValue("material_defaultUnitPrice")

	// Validate input
	if materialName == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Material name is required")
	}
	if unit == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Unit is required")
	}
	if defaultPriceStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Default unit price is required")
	}

	// Convert price to float64
	defaultPrice, err := strconv.ParseFloat(defaultPriceStr, 64)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid default price format")
	}

	// Create material data
	materialData := models.MasterMaterialCreate{
		MaterialName:     materialName,
		Unit:             unit,
		DefaultUnitPrice: defaultPrice,
		UserId:           userData.ID,
	}

	// Create material in database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		if err := h.materialsRepo.Create(tx, materialData); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to create material, make sure Material Name is Unique")
	}

	// refresh table
	utils.SetRefreshTableTriggerHeader(c)

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "Material created successfully", true)
}

// UpdateMaterial handles the update of an existing material
func (h *MaterialHandler) UpdateMaterial(c *fiber.Ctx) error {

	materialIdStr := c.Params("id")
	materialId, err := strconv.Atoi(materialIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid material ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	materialName := c.FormValue("material_name")
	unit := c.FormValue("material_unit")
	defaultPriceStr := c.FormValue("material_defaultUnitPrice")

	// Validate input
	if materialName == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Material name is required")
	}
	if unit == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Unit is required")
	}
	if defaultPriceStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Default unit price is required")
	}

	// Convert price to float64
	defaultPrice, err := strconv.ParseFloat(defaultPriceStr, 64)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid default price format")
	}

	// Update material in database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// First, fetch the existing material to ensure it belongs to the user
		existingMaterial, err := h.materialsRepo.FindById(tx, materialId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Material not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if material belongs to current user
		if existingMaterial.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Update the material
		updatedMaterial := models.MasterMaterial{
			MaterialId:       materialId,
			UserId:           userData.ID,
			MaterialName:     materialName,
			Unit:             unit,
			DefaultUnitPrice: defaultPrice,
			CreatedAt:        existingMaterial.CreatedAt,
			UpdatedAt:        time.Now().Format("2006-01-02 15:04:05"),
		}

		if err := h.materialsRepo.Update(tx, updatedMaterial); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to update material")
	}

	// refresh table data
	utils.SetRefreshTableTriggerHeader(c)

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "Material updated successfully", true)
}

// DeleteMaterial handles the deletion of a material
func (h *MaterialHandler) DeleteMaterial(c *fiber.Ctx) error {

	materialIdStr := c.Params("id")
	materialId, err := strconv.Atoi(materialIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid material ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Delete material from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// First, fetch the existing material to ensure it belongs to the user
		existingMaterial, err := h.materialsRepo.FindById(tx, materialId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Material not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if material belongs to current user
		if existingMaterial.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Delete the material
		if err := h.materialsRepo.Delete(tx, existingMaterial); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to delete material")
	}

	// refresh table data
	utils.SetRefreshTableTriggerHeader(c)

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "Material deleted successfully", true)
}
