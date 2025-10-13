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
	"github.com/momokii/go-rab-maker/backend/repository/projects"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

type ProjectsHandler struct {
	dbService    databases.SQLiteServices
	projectsRepo *projects.ProjectsRepo
}

func NewProjectsHandler(
	dbService databases.SQLiteServices,
	projectsRepo *projects.ProjectsRepo,
) *ProjectsHandler {
	return &ProjectsHandler{
		dbService:    dbService,
		projectsRepo: projectsRepo,
	}
}

// ==========================
// ========================== VIEWS
// ==========================

func (h *ProjectsHandler) ProjectsMainPageTableView(c *fiber.Ctx) error {
	var projects []models.Project
	var paginationInfo models.PaginationInfo

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

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
			// get the projects data for current user
			projectsData, paginationData, err := h.projectsRepo.FindByUserId(
				tx, userData.ID, paginationData,
			)
			if err != nil {
				return fiber.StatusInternalServerError, err
			}

			projects = projectsData

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
		BaseURL:           "/projects",
		Title:             "My Projects",
		SearchEnabled:     true,
		PaginationEnabled: true,
		PerPageEnabled:    true,
	}

	if c.Get("HX-Request") == "true" {
		tableComponents := components.ProjectsPage(projects, paginationInfo, tableConfig)
		return adaptor.HTTPHandler(templ.Handler(tableComponents))(c)
	}

	projectsComponent := components.ProjectsPage(
		projects,
		paginationInfo,
		tableConfig,
	)
	return adaptor.HTTPHandler(templ.Handler(projectsComponent))(c)
}

func (h *ProjectsHandler) ProjectCreateModalView(c *fiber.Ctx) error {
	modal := components.ProjectsFormModal(
		"Create New Project",
		"/projects/new",
		"new-project-form",
		"Create Project",
		models.Project{},
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *ProjectsHandler) ProjectEditModalView(c *fiber.Ctx) error {
	projectIdStr := c.Params("id")
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid project ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var project models.Project

	// Fetch project from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		project, err = h.projectsRepo.FindById(tx, projectId)
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

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch project")
	}

	modal := components.ProjectsFormModal(
		"Edit Project",
		"/projects/"+projectIdStr+"/edit",
		"edit-project-form",
		"Update Project",
		project,
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

func (h *ProjectsHandler) ProjectDeleteModalView(c *fiber.Ctx) error {
	projectIdStr := c.Params("id")
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid project ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	var project models.Project

	// Fetch project from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		project, err = h.projectsRepo.FindById(tx, projectId)
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

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch project")
	}

	modal := components.ConfirmationDeleteModal(
		"Delete Project",
		"Are you sure you want to delete this project "+project.ProjectName+"? This will also delete all associated work items.",
		"/projects/"+projectIdStr+"/delete",
		"Delete Project",
	)

	return adaptor.HTTPHandler(templ.Handler(modal))(c)
}

// ==========================
// ========================== FUNCTIONS
// ==========================

// CreateProject handles the creation of a new project
func (h *ProjectsHandler) CreateProject(c *fiber.Ctx) error {
	// Add small delay for better UX
	time.Sleep(500 * time.Millisecond)

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Extract form data
	projectName := c.FormValue("project_name")
	location := c.FormValue("location")
	clientName := c.FormValue("client_name")

	// Validate input
	if projectName == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Project name is required")
	}
	if location == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Location is required")
	}
	if clientName == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Client name is required")
	}

	// Create project data
	projectData := models.ProjectCreate{
		ProjectName: projectName,
		Location:    location,
		ClientName:  clientName,
		UserId:      userData.ID,
	}

	// Create project in database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		if err := h.projectsRepo.Create(tx, projectData); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to create project")
	}

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "Project created successfully", true)
}

// UpdateProject handles the update of an existing project
func (h *ProjectsHandler) UpdateProject(c *fiber.Ctx) error {
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
	projectName := c.FormValue("project_name")
	location := c.FormValue("location")
	clientName := c.FormValue("client_name")

	// Validate input
	if projectName == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Project name is required")
	}
	if location == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Location is required")
	}
	if clientName == "" {
		return utils.ResponseErrorModal(c, "Validation Error", "Client name is required")
	}

	// Update project in database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// First, fetch the existing project to ensure it belongs to the user
		existingProject, err := h.projectsRepo.FindById(tx, projectId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Project not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if project belongs to current user
		if existingProject.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Update the project
		updatedProject := models.Project{
			ProjectId:   projectId,
			UserId:      userData.ID,
			ProjectName: projectName,
			Location:    location,
			ClientName:  clientName,
			CreatedAt:   existingProject.CreatedAt,
			UpdatedAt:   time.Now().Format("2006-01-02 15:04:05"),
		}

		if err := h.projectsRepo.Update(tx, updatedProject); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to update project")
	}

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "Project updated successfully", true)
}

// DeleteProject handles the deletion of a project
func (h *ProjectsHandler) DeleteProject(c *fiber.Ctx) error {
	// Add small delay for better UX
	time.Sleep(500 * time.Millisecond)

	projectIdStr := c.Params("id")
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid project ID")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Delete project from database
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// First, fetch the existing project to ensure it belongs to the user
		existingProject, err := h.projectsRepo.FindById(tx, projectId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Project not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if project belongs to current user
		if existingProject.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Delete the project (cascade will handle related records)
		if err := h.projectsRepo.Delete(tx, existingProject); err != nil {
			return fiber.StatusInternalServerError, err
		}
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to delete project")
	}

	// Return success response with refresh
	return utils.ResponseSuccessModal(c, "Success", "Project deleted successfully", true)
}
