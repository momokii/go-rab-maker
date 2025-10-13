package handlers

import (
	"database/sql"
	"log"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/momokii/go-rab-maker/backend/databases"
	"github.com/momokii/go-rab-maker/backend/middlewares"
	"github.com/momokii/go-rab-maker/backend/models"
	"github.com/momokii/go-rab-maker/backend/repository/dashboard"
	"github.com/momokii/go-rab-maker/backend/repository/projects"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

type DashboardHandler struct {
	dbService     databases.SQLiteServices
	dashboardRepo dashboard.DashboardRepo
	projectsRepo  projects.ProjectsRepo
}

func NewDashboardHandler(
	dbService databases.SQLiteServices,
	dashboardRepo dashboard.DashboardRepo,
	projectsRepo projects.ProjectsRepo,
) *DashboardHandler {
	return &DashboardHandler{
		dbService:     dbService,
		dashboardRepo: dashboardRepo,
		projectsRepo:  projectsRepo,
	}
}

// Dashboard displays the main dashboard with project overview
func (h *DashboardHandler) Dashboard(c *fiber.Ctx) error {
	var projects []models.Project
	var totalProjects int
	var totalCost float64

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Use the database service transaction
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get recent projects for the user
		projectsData, err := h.dashboardRepo.GetRecentProjects(tx, userData.ID, 10)
		if err != nil {
			log.Printf("Error fetching projects: %v", err)
			projects = []models.Project{}
		} else {
			projects = projectsData
		}

		// Get total projects count
		totalProjectsData, err := h.dashboardRepo.GetProjectCount(tx, userData.ID)
		if err != nil {
			log.Printf("Error getting project count: %v", err)
			totalProjects = 0
		} else {
			totalProjects = totalProjectsData
		}

		// Get total cost of all projects
		totalCostData, err := h.dashboardRepo.GetProjectsTotalCost(tx, userData.ID)
		if err != nil {
			log.Printf("Error getting total cost: %v", err)
			totalCost = 0
		} else {
			totalCost = totalCostData
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to load dashboard data")
	}

	// Render the dashboard page
	dashboardComponent := components.DashboardPage(projects, totalProjects, totalCost)
	return adaptor.HTTPHandler(templ.Handler(dashboardComponent))(c)
}
