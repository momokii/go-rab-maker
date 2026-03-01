package handlers

import (
	"database/sql"

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
	var enhancedProjects []models.EnhancedProjectData
	var totalProjects int
	var totalCost float64
	var totalWorkItems int
	var typeCostBreakdown []models.TypeCostBreakdown
	var categoryBreakdown []models.CategoryBreakdown
	var topExpensiveItems []models.TopExpensiveItem

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Use the database service transaction
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get enhanced recent projects for the user
		enhancedProjectsData, err := h.dashboardRepo.GetEnhancedRecentProjects(tx, userData.ID, 10)
		if err != nil {
			enhancedProjects = []models.EnhancedProjectData{}
		} else {
			enhancedProjects = enhancedProjectsData
		}

		// Get total projects count
		totalProjectsData, err := h.dashboardRepo.GetProjectCount(tx, userData.ID)
		if err != nil {
			totalProjects = 0
		} else {
			totalProjects = totalProjectsData
		}

		// Get total work items count
		workItemsData, err := h.dashboardRepo.GetWorkItemsCount(tx, userData.ID)
		if err != nil {
			totalWorkItems = 0
		} else {
			totalWorkItems = workItemsData
		}

		// Get total cost of all projects
		totalCostData, err := h.dashboardRepo.GetProjectsTotalCost(tx, userData.ID)
		if err != nil {
			totalCost = 0
		} else {
			totalCost = totalCostData
		}

		// Get cost breakdown by type (Material vs Labor)
		typeCostData, err := h.dashboardRepo.GetTypeCostBreakdown(tx, userData.ID)
		if err != nil {
			typeCostBreakdown = []models.TypeCostBreakdown{}
		} else {
			typeCostBreakdown = typeCostData
		}

		// Get category breakdown (top 5)
		categoryData, err := h.dashboardRepo.GetCategoryBreakdown(tx, userData.ID, 5)
		if err != nil {
			categoryBreakdown = []models.CategoryBreakdown{}
		} else {
			categoryBreakdown = categoryData
		}

		// Get top 10 expensive items
		topItemsData, err := h.dashboardRepo.GetTopExpensiveItems(tx, userData.ID, 10)
		if err != nil {
			topExpensiveItems = []models.TopExpensiveItem{}
		} else {
			topExpensiveItems = topItemsData
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to load dashboard data")
	}

	// Render the dashboard page
	dashboardComponent := components.DashboardPage(
		enhancedProjects,
		totalProjects,
		totalWorkItems,
		totalCost,
		typeCostBreakdown,
		categoryBreakdown,
		topExpensiveItems,
	)
	return adaptor.HTTPHandler(templ.Handler(dashboardComponent))(c)
}
