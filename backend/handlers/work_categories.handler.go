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
	master_work_categories "github.com/momokii/go-rab-maker/backend/repository/masater_work_categories"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

type WorkCategoryHandler struct {
	dbService          databases.SQLiteServices
	workCategoriesRepo *master_work_categories.MasterWorkCategoriesRepo
}

func NewWorkCategoryHandler(
	dbService databases.SQLiteServices,
	workCategoriesRepo *master_work_categories.MasterWorkCategoriesRepo,
) *WorkCategoryHandler {
	return &WorkCategoryHandler{
		dbService:          dbService,
		workCategoriesRepo: workCategoriesRepo,
	}
}

// ==========================
// ========================== VIEWS
// ==========================

func (h *WorkCategoryHandler) WorkCategoriesMainPageTableView(c *fiber.Ctx) error {
	var workCategories []models.MasterWorkCategory
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

	// start transaction to get the data
	if _, err := h.dbService.Transaction(
		c.Context(),
		func(tx *sql.Tx) (int, error) {
			// get the work category data
			workCategoriesData, paginationData, err := h.workCategoriesRepo.Find(
				tx, paginationData,
			)
			if err != nil {
				return fiber.StatusInternalServerError, err
			}
			workCategories = workCategoriesData

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
		BaseURL:           "/work_categories",
		Title:             "Master Work Categories",
		SearchEnabled:     true,
		PaginationEnabled: true,
		PerPageEnabled:    true,
	}

	if c.Get("HX-Request") == "true" {
		tableComponents := components.WorkCategoriesTablePage(workCategories, paginationInfo, tableConfig)
		return adaptor.HTTPHandler(templ.Handler(tableComponents))(c)
	}

	workCategoriesComponent := components.WorkCategoriesPage(
		workCategories,
		paginationInfo,
		tableConfig,
	)
	return adaptor.HTTPHandler(templ.Handler(workCategoriesComponent))(c)
}

func (h *WorkCategoryHandler) WorkCategoryCreateModalView(c *fiber.Ctx) error {
	modal := components.WorkCategoriesFormModal(
		"Add New Work Category",
		"/work_categories/new",
		"new-work-category-form",
		"Add Work Category",
		models.MasterWorkCategory{},
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *WorkCategoryHandler) WorkCategoryEditModalView(c *fiber.Ctx) error {
	workCategoryIdStr := c.Params("id")
	workCategoryId, err := strconv.Atoi(workCategoryIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid work category ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var workCategory models.MasterWorkCategory

	// Fetch work category from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		workCategory, err = h.workCategoriesRepo.FindById(tx, workCategoryId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Work category not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if work category belongs to current user
		if workCategory.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch work category")
	}

	modal := components.WorkCategoriesFormModal(
		"Edit Work Category",
		"/work_categories/"+workCategoryIdStr+"/edit",
		"edit-work-category-form",
		"Update Work Category",
		workCategory,
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *WorkCategoryHandler) WorkCategoryDeleteModalView(c *fiber.Ctx) error {
	workCategoryIdStr := c.Params("id")
	workCategoryId, err := strconv.Atoi(workCategoryIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid work category ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var workCategory models.MasterWorkCategory

	// Fetch work category from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		workCategory, err = h.workCategoriesRepo.FindById(tx, workCategoryId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Work category not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if work category belongs to current user
		if workCategory.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch work category")
	}

	modal := components.ConfirmationDeleteModal(
		"Delete Work Category",
		"Are you sure you want to delete this work category "+workCategory.CategoryName+"?",
		"/work_categories/"+workCategoryIdStr+"/delete",
		"Delete Work Category",
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

// ==========================
// ========================== FUNCTIONS
// ==========================

// CreateWorkCategory handles the creation of a new work category
func (h *WorkCategoryHandler) CreateWorkCategory(c *fiber.Ctx) error {

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	categoryName := c.FormValue("category_name")
	displayOrderStr := c.FormValue("display_order")

	// Validate input
	if categoryName == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Category name is required")
	}
	if displayOrderStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Display order is required")
	}

	// Convert display order to int
	displayOrder, err := strconv.Atoi(displayOrderStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid display order format")
	}

	// Create work category data
	workCategoryData := models.MasterWorkCategoryCreate{
		CategoryName: categoryName,
		DisplayOrder: displayOrder,
		UserId:       userData.ID,
	}

	// Create work category in database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		if err := h.workCategoriesRepo.Create(tx, workCategoryData); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to create work category, make sure Category Name is unique")
	}

	// set for refresh table after success
	utils.SetRefreshTableTriggerHeader(c)

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "Work category created successfully", true)
}

// UpdateWorkCategory handles the update of an existing work category
func (h *WorkCategoryHandler) UpdateWorkCategory(c *fiber.Ctx) error {

	workCategoryIdStr := c.Params("id")
	workCategoryId, err := strconv.Atoi(workCategoryIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid work category ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	categoryName := c.FormValue("category_name")
	displayOrderStr := c.FormValue("display_order")

	// Validate input
	if categoryName == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Category name is required")
	}
	if displayOrderStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Display order is required")
	}

	// Convert display order to int
	displayOrder, err := strconv.Atoi(displayOrderStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid display order format")
	}

	// Update work category in database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// First, fetch the existing work category to ensure it belongs to the user
		existingWorkCategory, err := h.workCategoriesRepo.FindById(tx, workCategoryId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Work category not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if work category belongs to current user
		if existingWorkCategory.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Update the work category
		updatedWorkCategory := models.MasterWorkCategory{
			CategoryId:   workCategoryId,
			UserId:       userData.ID,
			CategoryName: categoryName,
			DisplayOrder: displayOrder,
			CreatedAt:    existingWorkCategory.CreatedAt,
			UpdatedAt:    time.Now().Format("2006-01-02 15:04:05"),
		}

		if err := h.workCategoriesRepo.Update(tx, updatedWorkCategory); err != nil {
			return fiber.StatusInternalServerError, fiber.NewError(fiber.StatusInternalServerError, "Make sure the category name is set to unique value")
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to update work category: "+err.Error())
	}

	// set for refresh table after success
	utils.SetRefreshTableTriggerHeader(c)

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "Work category updated successfully", true)
}

// DeleteWorkCategory handles the deletion of a work category
func (h *WorkCategoryHandler) DeleteWorkCategory(c *fiber.Ctx) error {

	workCategoryIdStr := c.Params("id")
	workCategoryId, err := strconv.Atoi(workCategoryIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid work category ID")
	}

	// Get user from session (using the same approach as in auth.handler.go)
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Delete work category from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// First, fetch the existing work category to ensure it belongs to the user
		existingWorkCategory, err := h.workCategoriesRepo.FindById(tx, workCategoryId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Work category not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if work category belongs to current user
		if existingWorkCategory.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Delete the work category
		if err := h.workCategoriesRepo.Delete(tx, existingWorkCategory); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to delete work category")
	}

	// set for refresh table after success
	utils.SetRefreshTableTriggerHeader(c)

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "Work category deleted successfully", true)
}
