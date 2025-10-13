package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/momokii/go-rab-maker/backend/databases"
	"github.com/momokii/go-rab-maker/backend/middlewares"
	"github.com/momokii/go-rab-maker/backend/models"
	"github.com/momokii/go-rab-maker/backend/repository/ahsp_labor_components"
	ahsp_material_components "github.com/momokii/go-rab-maker/backend/repository/ahsp_material_components.repo.go"
	ahsptemplates "github.com/momokii/go-rab-maker/backend/repository/ahsp_templates"
	master_work_categories "github.com/momokii/go-rab-maker/backend/repository/masater_work_categories"
	"github.com/momokii/go-rab-maker/backend/repository/master_labor_types"
	"github.com/momokii/go-rab-maker/backend/repository/master_materials"
	"github.com/momokii/go-rab-maker/backend/repository/project_item_costs"
	"github.com/momokii/go-rab-maker/backend/repository/project_work_items"
	"github.com/momokii/go-rab-maker/backend/repository/projects"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

type ProjectWorkItemsHandler struct {
	dbService                  databases.SQLiteServices
	projectWorkItemsRepo       *project_work_items.ProjectWorkItemRepo
	projectItemCostsRepo       *project_item_costs.ProjectItemCostsRepo
	ahspTemplatesRepo          *ahsptemplates.AhspTemplatesRepo
	ahspMaterialComponentsRepo *ahsp_material_components.AHSPMaterialComponentsRepo
	masterMaterialsRepo        *master_materials.MasterMaterialsRepo
	masterLaborTypesRepo       *master_labor_types.MasterLaborTypesRepo
	ahspLaborComponentsRepo    *ahsp_labor_components.AHSPLaborComponentsRepo
}

func NewProjectWorkItemsHandler(
	dbService databases.SQLiteServices,
	projectWorkItemsRepo *project_work_items.ProjectWorkItemRepo,
	projectItemCostsRepo *project_item_costs.ProjectItemCostsRepo,
	ahspTemplatesRepo *ahsptemplates.AhspTemplatesRepo,
	ahspMaterialComponentsRepo *ahsp_material_components.AHSPMaterialComponentsRepo,
	masterMaterialsRepo *master_materials.MasterMaterialsRepo,
	masterLaborTypesRepo *master_labor_types.MasterLaborTypesRepo,
	ahspLaborComponentsRepo *ahsp_labor_components.AHSPLaborComponentsRepo,
) *ProjectWorkItemsHandler {
	return &ProjectWorkItemsHandler{
		dbService:                  dbService,
		projectWorkItemsRepo:       projectWorkItemsRepo,
		projectItemCostsRepo:       projectItemCostsRepo,
		ahspTemplatesRepo:          ahspTemplatesRepo,
		ahspMaterialComponentsRepo: ahspMaterialComponentsRepo,
		masterMaterialsRepo:        masterMaterialsRepo,
		masterLaborTypesRepo:       masterLaborTypesRepo,
		ahspLaborComponentsRepo:    ahspLaborComponentsRepo,
	}
}

// ==========================
// ========================== VIEWS
// ==========================

// ProjectDetailPage displays the project detail page with work items
func (h *ProjectWorkItemsHandler) ProjectDetailPage(c *fiber.Ctx) error {
	projectIdStr := c.Params("id")
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid project ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var project models.Project
	var workItems []models.ProjectWorkItemWithDetails
	var totalCost float64

	// Fetch project data
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get project details
		projectsRepo := projects.NewProjectsRepo()
		project, err = projectsRepo.FindById(tx, projectId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Project not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if project belongs to current user
		if project.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get work items with details
		workItems, err = h.projectWorkItemsRepo.FindByProjectIdWithDetails(tx, projectId)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		// Get total project cost
		totalCost, err = h.projectWorkItemsRepo.GetProjectTotalCost(tx, projectId)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch project details")
	}

	log.Println(fmt.Sprintf("Project %d has %d work items", projectId, len(workItems)))

	// Render the project detail page
	projectDetailComponent := components.ProjectDetailPage(project, workItems, totalCost)
	return adaptor.HTTPHandler(templ.Handler(projectDetailComponent))(c)
}

// ProjectWorkItemCreateModalView displays the modal to create a new work item
func (h *ProjectWorkItemsHandler) ProjectWorkItemCreateModalView(c *fiber.Ctx) error {
	projectIdStr := c.Params("id")
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid project ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var project models.Project
	var categories []models.MasterWorkCategory
	var templates []models.AHSPTemplate

	// Fetch required data
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get project details
		projectsRepo := projects.NewProjectsRepo()
		project, err = projectsRepo.FindById(tx, projectId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Project not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if project belongs to current user
		if project.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get work categories
		workCategoriesRepo := master_work_categories.NewMasterWorkCategoriesRepo()
		paginationData := models.TablePaginationDataInput{
			Page:    1,
			PerPage: 1000, // Get all categories
		}
		categoriesData, _, err := workCategoriesRepo.Find(tx, paginationData)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		// Get AHSP templates
		templatesData, _, err := h.ahspTemplatesRepo.Find(tx, paginationData, userData.ID)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		categories = categoriesData
		templates = templatesData

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch required data")
	}

	modal := components.ProjectWorkItemFormModal(
		projectId,
		nil, // No work item for create
		categories,
		templates,
		false, // Not edit mode
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

// ProjectWorkItemEditModalView displays the modal to edit a work item
func (h *ProjectWorkItemsHandler) ProjectWorkItemEditModalView(c *fiber.Ctx) error {
	projectIdStr := c.Params("id")
	workItemIdStr := c.Params("workItemId")

	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid project ID")
	}

	workItemId, err := strconv.Atoi(workItemIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid work item ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var project models.Project
	var workItem models.ProjectWorkItem
	var categories []models.MasterWorkCategory
	var templates []models.AHSPTemplate

	// Fetch required data
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get project details
		projectsRepo := projects.NewProjectsRepo()
		project, err = projectsRepo.FindById(tx, projectId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Project not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if project belongs to current user
		if project.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get work item details
		workItem, err = h.projectWorkItemsRepo.FindById(tx, workItemId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Work item not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if work item belongs to the project
		if workItem.ProjectId != projectId {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get work categories
		workCategoriesRepo := master_work_categories.NewMasterWorkCategoriesRepo()
		paginationData := models.TablePaginationDataInput{
			Page:    1,
			PerPage: 1000, // Get all categories
		}
		categories, _, err = workCategoriesRepo.Find(tx, paginationData)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		// Get AHSP templates
		templates, _, err = h.ahspTemplatesRepo.Find(tx, paginationData, userData.ID)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch required data")
	}

	// Convert work item to work item with details for the template
	workItemWithDetails := models.ProjectWorkItemWithDetails{
		WorkItemId:     workItem.WorkItemId,
		ProjectId:      workItem.ProjectId,
		CategoryId:     workItem.CategoryId,
		Description:    workItem.Description,
		Volume:         workItem.Volume,
		Unit:           workItem.Unit,
		AHSPTemplateId: workItem.AHSPTemplateId,
		CreatedAt:      workItem.CreatedAt,
		UpdatedAt:      workItem.UpdatedAt,
	}

	modal := components.ProjectWorkItemFormModal(
		projectId,
		&workItemWithDetails,
		categories,
		templates,
		true, // Edit mode
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

// ProjectWorkItemDeleteModalView displays the modal to delete a work item
func (h *ProjectWorkItemsHandler) ProjectWorkItemDeleteModalView(c *fiber.Ctx) error {
	projectIdStr := c.Params("id")
	workItemIdStr := c.Params("workItemId")

	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid project ID")
	}

	workItemId, err := strconv.Atoi(workItemIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid work item ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var workItem models.ProjectWorkItem

	// Fetch work item data
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get project details
		projectsRepo := projects.NewProjectsRepo()
		project, err := projectsRepo.FindById(tx, projectId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Project not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if project belongs to current user
		if project.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get work item details
		workItem, err = h.projectWorkItemsRepo.FindById(tx, workItemId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Work item not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if work item belongs to the project
		if workItem.ProjectId != projectId {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch work item")
	}

	modal := components.ConfirmationDeleteModal(
		"Delete Work Item",
		"Are you sure you want to delete this work item '"+workItem.Description+"'? This will also delete all associated cost calculations.",
		"/project/"+projectIdStr+"/work-items/"+workItemIdStr+"/delete",
		"Delete Work Item",
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

// ProjectWorkItemCostsView displays the cost breakdown for a work item
func (h *ProjectWorkItemsHandler) ProjectWorkItemCostsView(c *fiber.Ctx) error {
	workItemIdStr := c.Params("id")
	workItemId, err := strconv.Atoi(workItemIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid work item ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var costs []models.ProjectItemCostWithDetails

	// Fetch cost data
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get work item details to verify ownership
		workItem, err := h.projectWorkItemsRepo.FindById(tx, workItemId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Work item not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Get project details to verify ownership
		projectsRepo := projects.NewProjectsRepo()
		project, err := projectsRepo.FindById(tx, workItem.ProjectId)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		// Check if project belongs to current user
		if project.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get cost details for the work item
		costs, err = h.projectItemCostsRepo.FindByWorkItemId(tx, workItemId)
		if err != nil {
			return fiber.StatusInternalServerError, err
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch cost details")
	}

	// Render the cost breakdown
	costsComponent := components.ProjectItemCostsDisplay(costs)
	return adaptor.HTTPHandler(templ.Handler(costsComponent))(c)
}

// ==========================
// ========================== FUNCTIONS
// ==========================

// CreateProjectWorkItem handles the creation of a new work item
func (h *ProjectWorkItemsHandler) CreateProjectWorkItem(c *fiber.Ctx) error {
	// Add small delay for better UX
	time.Sleep(500 * time.Millisecond)

	projectIdStr := c.Params("id")
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid project ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	categoryIdStr := c.FormValue("category_id")
	description := c.FormValue("description")
	volumeStr := c.FormValue("volume")
	unit := c.FormValue("unit")
	ahspTemplateIdStr := c.FormValue("ahsp_template_id")

	// Validate input
	if categoryIdStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Category is required")
	}
	if description == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Description is required")
	}
	if volumeStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Volume is required")
	}
	if unit == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Unit is required")
	}

	// Convert values
	categoryId, err := strconv.Atoi(categoryIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid category ID")
	}

	volume, err := strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid volume value")
	}

	var ahspTemplateId *int
	if ahspTemplateIdStr != "" {
		templateId, err := strconv.Atoi(ahspTemplateIdStr)
		if err != nil {
			return utils.ResponseErrorModal(c, "Validation Error", "Invalid template ID")
		}
		ahspTemplateId = &templateId
	}

	// Create work item data
	workItemData := models.ProjectWorkItemCreate{
		ProjectId:      projectId,
		CategoryId:     categoryId,
		Description:    description,
		Volume:         volume,
		Unit:           unit,
		AHSPTemplateId: ahspTemplateId,
	}

	// Create work item and calculate costs in a transaction
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Verify project ownership
		projectsRepo := projects.NewProjectsRepo()
		project, err := projectsRepo.FindById(tx, projectId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Project not found")
			}
			return fiber.StatusInternalServerError, err
		}

		if project.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Create work item
		if err := h.projectWorkItemsRepo.Create(tx, workItemData); err != nil {
			log.Printf("Error creating work item: %v", err)
			return fiber.StatusInternalServerError, err
		}

		// Get the newly created work items to find the latest one
		workItems, err := h.projectWorkItemsRepo.FindByProjectId(tx, projectId)
		if err != nil {
			log.Printf("Error fetching work items after creation: %v", err)
			return fiber.StatusInternalServerError, err
		}

		// Find the most recent work item (the one we just created)
		var newWorkItemId int
		if len(workItems) > 0 {
			// Sort by work item ID descending to get the latest
			if workItems[0].WorkItemId > 0 {
				newWorkItemId = workItems[0].WorkItemId
			}
		}

		// If AHSP template is selected, calculate and create cost items
		if ahspTemplateId != nil && newWorkItemId > 0 {
			log.Printf("Calculating costs for template ID %d, work item ID %d", *ahspTemplateId, newWorkItemId)
			if err := h.calculateAndCreateCosts(tx, *ahspTemplateId, volume, newWorkItemId); err != nil {
				log.Printf("Error calculating costs: %v", err)
				// Don't fail the entire operation if cost calculation fails
				// Just log the error and continue
			}
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to create work item")
	}

	// Return success response with redirect
	return utils.ResponseSuccessWithRedirect(c, "Success", "Work item created successfully", "/project/"+projectIdStr)
}

// UpdateProjectWorkItem handles the update of an existing work item
func (h *ProjectWorkItemsHandler) UpdateProjectWorkItem(c *fiber.Ctx) error {
	// Add small delay for better UX
	time.Sleep(500 * time.Millisecond)

	projectIdStr := c.Params("id")
	workItemIdStr := c.Params("workItemId")

	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid project ID")
	}

	workItemId, err := strconv.Atoi(workItemIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid work item ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	categoryIdStr := c.FormValue("category_id")
	description := c.FormValue("description")
	volumeStr := c.FormValue("volume")
	unit := c.FormValue("unit")
	ahspTemplateIdStr := c.FormValue("ahsp_template_id")

	// Validate input
	if categoryIdStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Category is required")
	}
	if description == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Description is required")
	}
	if volumeStr == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Volume is required")
	}
	if unit == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Unit is required")
	}

	// Convert values
	categoryId, err := strconv.Atoi(categoryIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid category ID")
	}

	volume, err := strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return utils.ResponseErrorModal(c, "Validation Error", "Invalid volume value")
	}

	var ahspTemplateId *int
	if ahspTemplateIdStr != "" {
		templateId, err := strconv.Atoi(ahspTemplateIdStr)
		if err != nil {
			return utils.ResponseErrorModal(c, "Validation Error", "Invalid template ID")
		}
		ahspTemplateId = &templateId
	}

	// Update work item and recalculate costs in a transaction
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Verify project ownership
		projectsRepo := projects.NewProjectsRepo()
		project, err := projectsRepo.FindById(tx, projectId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Project not found")
			}
			return fiber.StatusInternalServerError, err
		}

		if project.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get existing work item
		existingWorkItem, err := h.projectWorkItemsRepo.FindById(tx, workItemId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Work item not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if work item belongs to the project
		if existingWorkItem.ProjectId != projectId {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Update work item
		updatedWorkItem := models.ProjectWorkItem{
			WorkItemId:     workItemId,
			ProjectId:      projectId,
			CategoryId:     categoryId,
			Description:    description,
			Volume:         volume,
			Unit:           unit,
			AHSPTemplateId: ahspTemplateId,
			CreatedAt:      existingWorkItem.CreatedAt,
			UpdatedAt:      time.Now().Format("2006-01-02 15:04:05"),
		}

		if err := h.projectWorkItemsRepo.Update(tx, updatedWorkItem); err != nil {
			return fiber.StatusInternalServerError, err
		}

		// Delete existing cost calculations
		if err := h.projectItemCostsRepo.DeleteByWorkItemId(tx, workItemId); err != nil {
			return fiber.StatusInternalServerError, err
		}

		// If AHSP template is selected, recalculate and create cost items
		if ahspTemplateId != nil {
			if err := h.calculateAndCreateCosts(tx, *ahspTemplateId, volume, workItemId); err != nil {
				log.Printf("Error recalculating costs: %v", err)
				// Don't fail the entire operation if cost calculation fails
				// Just log the error and continue
			}
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to update work item")
	}

	// Return success response with redirect
	return utils.ResponseSuccessWithRedirect(c, "Success", "Work item updated successfully", "/project/"+projectIdStr)
}

// DeleteProjectWorkItem handles the deletion of a work item
func (h *ProjectWorkItemsHandler) DeleteProjectWorkItem(c *fiber.Ctx) error {
	// Add small delay for better UX
	time.Sleep(500 * time.Millisecond)

	projectIdStr := c.Params("id")
	workItemIdStr := c.Params("workItemId")

	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid project ID")
	}

	workItemId, err := strconv.Atoi(workItemIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid work item ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Delete work item in a transaction
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Verify project ownership
		projectsRepo := projects.NewProjectsRepo()
		project, err := projectsRepo.FindById(tx, projectId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Project not found")
			}
			return fiber.StatusInternalServerError, err
		}

		if project.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get existing work item
		existingWorkItem, err := h.projectWorkItemsRepo.FindById(tx, workItemId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Work item not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if work item belongs to the project
		if existingWorkItem.ProjectId != projectId {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Delete work item (cascade will handle cost calculations)
		if err := h.projectWorkItemsRepo.Delete(tx, existingWorkItem); err != nil {
			return fiber.StatusInternalServerError, err
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to delete work item")
	}

	// Return success response with redirect
	return utils.ResponseSuccessWithRedirect(c, "Success", "Work item deleted successfully", "/project/"+projectIdStr)
}

// calculateAndCreateCosts calculates and creates cost items based on AHSP template
func (h *ProjectWorkItemsHandler) calculateAndCreateCosts(tx *sql.Tx, templateId int, volume float64, workItemId int) error {
	log.Printf("Starting cost calculation for template ID %d, volume %.2f, work item ID %d", templateId, volume, workItemId)

	// Get material components for the template
	materialComponents, err := h.ahspMaterialComponentsRepo.FindByTemplateId(tx, templateId)
	if err != nil {
		log.Printf("Error fetching material components for template %d: %v", templateId, err)
		return err
	}
	log.Printf("Found %d material components for template %d", len(materialComponents), templateId)

	// Get labor components for the template
	laborComponents, err := h.ahspLaborComponentsRepo.FindByTemplateId(tx, templateId)
	if err != nil {
		log.Printf("Error fetching labor components for template %d: %v", templateId, err)
		return err
	}
	log.Printf("Found %d labor components for template %d", len(laborComponents), templateId)

	var costItems []models.ProjectItemCostCreate

	// Process material components
	for i, component := range materialComponents {
		log.Printf("Processing material component %d: MaterialID %d, Coefficient %.4f", i+1, component.MaterialId, component.Coefficient)

		// Get material details
		material, err := h.masterMaterialsRepo.FindById(tx, component.MaterialId)
		if err != nil {
			log.Printf("Warning: Material with ID %d not found, skipping: %v", component.MaterialId, err)
			continue // Skip if material not found
		}

		quantityNeeded := component.Coefficient * volume
		totalCost := quantityNeeded * material.DefaultUnitPrice

		log.Printf("Material cost calculation - Quantity: %.4f, Unit Price: %.2f, Total: %.2f",
			quantityNeeded, material.DefaultUnitPrice, totalCost)

		costItems = append(costItems, models.ProjectItemCostCreate{
			WorkItemId:          workItemId,
			ItemType:            string(models.PROJECT_ITEM_TYPE_MATERIAL),
			ItemId:              component.MaterialId,
			QuantityNeeded:      quantityNeeded,
			UnitPriceAtCreation: material.DefaultUnitPrice,
			TotalCost:           totalCost,
		})
	}

	// Process labor components
	for i, component := range laborComponents {
		log.Printf("Processing labor component %d: LaborTypeID %d, Coefficient %.4f", i+1, component.LaborTypeId, component.Coefficient)

		// Get labor type details
		laborType, err := h.masterLaborTypesRepo.FindById(tx, component.LaborTypeId)
		if err != nil {
			log.Printf("Warning: Labor type with ID %d not found, skipping: %v", component.LaborTypeId, err)
			continue // Skip if labor type not found
		}

		quantityNeeded := component.Coefficient * volume
		totalCost := quantityNeeded * laborType.DefaultDailyWage

		log.Printf("Labor cost calculation - Quantity: %.4f, Daily Wage: %.2f, Total: %.2f",
			quantityNeeded, laborType.DefaultDailyWage, totalCost)

		costItems = append(costItems, models.ProjectItemCostCreate{
			WorkItemId:          workItemId,
			ItemType:            string(models.PROJECT_ITEM_TYPE_LABOR),
			ItemId:              component.LaborTypeId,
			QuantityNeeded:      quantityNeeded,
			UnitPriceAtCreation: laborType.DefaultDailyWage,
			TotalCost:           totalCost,
		})
	}

	// Create all cost items
	if len(costItems) > 0 {
		log.Printf("Creating %d cost items for work item %d", len(costItems), workItemId)
		if err := h.projectItemCostsRepo.CreateMultiple(tx, costItems); err != nil {
			log.Printf("Error creating cost items: %v", err)
			return err
		}
		log.Printf("Successfully created %d cost items", len(costItems))
	} else {
		log.Printf("No cost items to create for work item %d", workItemId)
	}

	return nil
}
