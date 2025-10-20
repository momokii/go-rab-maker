package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/momokii/go-rab-maker/backend/databases"
	"github.com/momokii/go-rab-maker/backend/handlers"
	"github.com/momokii/go-rab-maker/backend/middlewares"
	"github.com/momokii/go-rab-maker/backend/repository/ahsp_labor_components"
	ahsp_material_components "github.com/momokii/go-rab-maker/backend/repository/ahsp_material_components.repo.go"
	ahsptemplates "github.com/momokii/go-rab-maker/backend/repository/ahsp_templates"
	"github.com/momokii/go-rab-maker/backend/repository/dashboard"
	master_work_categories "github.com/momokii/go-rab-maker/backend/repository/masater_work_categories"
	"github.com/momokii/go-rab-maker/backend/repository/master_labor_types"
	"github.com/momokii/go-rab-maker/backend/repository/master_materials"
	"github.com/momokii/go-rab-maker/backend/repository/material_summary"
	"github.com/momokii/go-rab-maker/backend/repository/project_item_costs"
	"github.com/momokii/go-rab-maker/backend/repository/project_work_items"
	"github.com/momokii/go-rab-maker/backend/repository/projects"
	"github.com/momokii/go-rab-maker/backend/utils"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	log.Println(utils.GetBaseDir())

	// databases setup using sqlite
	databases.InitDatabaseSQLite()
	dbServices, err := databases.NewSQLiteDatabases(databases.DATABASE_SQLITE_PATH)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// repositories
	materialsRepo := master_materials.NewMasterMaterialsRepo()
	laborTypesRepo := master_labor_types.NewMasterLaborTypesRepo()
	workCategoriesRepo := master_work_categories.NewMasterWorkCategoriesRepo()
	ahspTemplatesRepo := ahsptemplates.NewAhspTemplatesRepo()
	ahspMaterialComponentsRepo := ahsp_material_components.NewAHSPMaterialComponentsRepo()
	ahspLaborComponentsRepo := ahsp_labor_components.NewAHSPLaborComponentsRepo()
	projectWorkItemsRepo := project_work_items.NewProjectWorkItemRepo()
	projectItemCostsRepo := project_item_costs.NewProjectItemCostsRepo()
	projectsRepo := projects.NewProjectsRepo()
	dashboardRepo := dashboard.NewDashboardRepo()
	materialSummaryRepo := material_summary.NewMaterialSummaryRepo()

	// handlers
	authHandler := handlers.NewAuthHandler(dbServices)
	masterMaterialsHandler := handlers.NewMaterialsHandler(
		dbServices,
		*materialsRepo,
	)
	masterLaborTypesHandler := handlers.NewLaborTypeHandler(
		dbServices,
		*laborTypesRepo,
	)
	masterWorkCategoriesHandler := handlers.NewWorkCategoryHandler(
		dbServices,
		workCategoriesRepo,
	)
	ahspTemplatesHandler := handlers.NewAhspTemplateHandler(
		dbServices,
		ahspTemplatesRepo,
	)
	ahspMaterialComponentHandler := handlers.NewAhspMaterialComponentHandler(
		dbServices,
		ahspMaterialComponentsRepo,
		ahspTemplatesRepo,
		materialsRepo,
		ahspLaborComponentsRepo,
	)
	ahspLaborComponentHandler := handlers.NewAhspLaborComponentHandler(
		dbServices,
		ahspLaborComponentsRepo,
		ahspTemplatesRepo,
		laborTypesRepo,
	)
	projectsHandler := handlers.NewProjectsHandler(
		dbServices,
		projectsRepo,
	)
	projectWorkItemsHandler := handlers.NewProjectWorkItemsHandler(
		dbServices,
		projectWorkItemsRepo,
		projectItemCostsRepo,
		ahspTemplatesRepo,
		ahspMaterialComponentsRepo,
		materialsRepo,
		laborTypesRepo,
		ahspLaborComponentsRepo,
	)
	dashboardHandler := handlers.NewDashboardHandler(
		dbServices,
		*dashboardRepo,
		*projectsRepo,
	)
	materialSummaryHandler := handlers.NewMaterialSummaryHandler(
		dbServices,
		*materialSummaryRepo,
		*projectsRepo,
		projectItemCostsRepo,
	)

	// session for auth
	session := middlewares.NewSessionMiddleware()

	app := fiber.New(fiber.Config{
		ServerHeader:    "RAB Maker",
		AppName:         "RAB Maker v1.1.0",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				// Allow redirects to pass through - don't treat them as errors
				if code >= 300 && code < 400 {
					return err
				}

			}
			// handle actual error
			return utils.ResponseErrorModal(c, "Error "+string(rune(code)), err.Error())
		},
		DisableStartupMessage: true,
	})

	middlewares.SetupAppMiddlewares(app)

	// Authentication routes
	app.Get("/login", session.IsNotAuth, authHandler.LoginView)
	app.Post("/login", session.IsNotAuth, authHandler.Login)
	app.Get("/register", session.IsNotAuth, authHandler.RegisterView)
	app.Post("/register", session.IsNotAuth, authHandler.Register)
	app.Post("/logout", session.IsAuth, authHandler.Logout)

	app.Get("/", session.IsAuth, dashboardHandler.Dashboard)
	app.Get("/countries/new", session.IsAuth, handlers.CreateCountriesModal)
	app.Get("/countries/:id/edit", session.IsAuth, handlers.EditCountriesModal)
	app.Post("/countries/:id/edit", session.IsAuth, handlers.EditCountries)
	app.Get("/countries/:id/delete", session.IsAuth, handlers.DeleteCountriesModal)
	app.Delete("/countries/:id/delete", session.IsAuth, handlers.DeleteCountries)
	app.Post("/countries/new", session.IsAuth, handlers.AddNewCountries)

	// materials
	app.Get("/materials", session.IsAuth, masterMaterialsHandler.MaterialsMainPageTableView)
	app.Get("/materials/new", session.IsAuth, masterMaterialsHandler.MaterialCreateModalView)
	app.Post("/materials/new", session.IsAuth, masterMaterialsHandler.CreateMaterial)
	app.Get("/materials/:id/edit", session.IsAuth, masterMaterialsHandler.MaterialEditModalView)
	app.Post("/materials/:id/edit", session.IsAuth, masterMaterialsHandler.UpdateMaterial)
	app.Get("/materials/:id/delete", session.IsAuth, masterMaterialsHandler.MaterialDeleteModalView)
	app.Delete("/materials/:id/delete", session.IsAuth, masterMaterialsHandler.DeleteMaterial)

	// labor types
	app.Get("/labor_types", session.IsAuth, masterLaborTypesHandler.LaborTypesMainPageTableView)
	app.Get("/labor_types/new", session.IsAuth, masterLaborTypesHandler.LaborTypeCreateModalView)
	app.Post("/labor_types/new", session.IsAuth, masterLaborTypesHandler.CreateLaborType)
	app.Get("/labor_types/:id/edit", session.IsAuth, masterLaborTypesHandler.LaborTypeEditModalView)
	app.Post("/labor_types/:id/edit", session.IsAuth, masterLaborTypesHandler.UpdateLaborType)
	app.Get("/labor_types/:id/delete", session.IsAuth, masterLaborTypesHandler.LaborTypeDeleteModalView)
	app.Delete("/labor_types/:id/delete", session.IsAuth, masterLaborTypesHandler.DeleteLaborType)

	// work categories
	app.Get("/work_categories", session.IsAuth, masterWorkCategoriesHandler.WorkCategoriesMainPageTableView)
	app.Get("/work_categories/new", session.IsAuth, masterWorkCategoriesHandler.WorkCategoryCreateModalView)
	app.Post("/work_categories/new", session.IsAuth, masterWorkCategoriesHandler.CreateWorkCategory)
	app.Get("/work_categories/:id/edit", session.IsAuth, masterWorkCategoriesHandler.WorkCategoryEditModalView)
	app.Post("/work_categories/:id/edit", session.IsAuth, masterWorkCategoriesHandler.UpdateWorkCategory)
	app.Get("/work_categories/:id/delete", session.IsAuth, masterWorkCategoriesHandler.WorkCategoryDeleteModalView)
	app.Delete("/work_categories/:id/delete", session.IsAuth, masterWorkCategoriesHandler.DeleteWorkCategory)

	// AHSP templates
	app.Get("/ahsp_templates", session.IsAuth, ahspTemplatesHandler.AhspTemplatesMainPageTableView)
	app.Get("/ahsp_templates/new", session.IsAuth, ahspTemplatesHandler.AhspTemplateCreateModalView)
	app.Post("/ahsp_templates/new", session.IsAuth, ahspTemplatesHandler.CreateAhspTemplate)
	app.Get("/ahsp_templates/:id/edit", session.IsAuth, ahspTemplatesHandler.AhspTemplateEditModalView)
	app.Post("/ahsp_templates/:id/edit", session.IsAuth, ahspTemplatesHandler.UpdateAhspTemplate)
	app.Get("/ahsp_templates/:id/delete", session.IsAuth, ahspTemplatesHandler.AhspTemplateDeleteModalView)
	app.Delete("/ahsp_templates/:id/delete", session.IsAuth, ahspTemplatesHandler.DeleteAhspTemplate)

	// AHSP template material components
	app.Get("/ahsp_templates/:templateId", session.IsAuth, ahspMaterialComponentHandler.AhspMaterialComponentsPage)
	app.Get("/ahsp_templates/:templateId/material_components/new", session.IsAuth, ahspMaterialComponentHandler.AhspMaterialComponentCreateModalView)
	app.Post("/ahsp_templates/:templateId/material_components/new", session.IsAuth, ahspMaterialComponentHandler.CreateAhspMaterialComponent)
	app.Get("/ahsp_templates/:templateId/material_components/:componentId/edit", session.IsAuth, ahspMaterialComponentHandler.AhspMaterialComponentEditModalView)
	app.Post("/ahsp_templates/:templateId/material_components/:componentId/edit", session.IsAuth, ahspMaterialComponentHandler.UpdateAhspMaterialComponent)
	app.Get("/ahsp_templates/:templateId/material_components/:componentId/delete", session.IsAuth, ahspMaterialComponentHandler.AhspMaterialComponentDeleteModalView)
	app.Delete("/ahsp_templates/:templateId/material_components/:componentId/delete", session.IsAuth, ahspMaterialComponentHandler.DeleteAhspMaterialComponent)

	// AHSP template labor components
	app.Get("/ahsp_templates/:templateId/labor_components", session.IsAuth, ahspLaborComponentHandler.AhspLaborComponentsPage)
	app.Get("/ahsp_templates/:templateId/labor_components/new", session.IsAuth, ahspLaborComponentHandler.AhspLaborComponentCreateModalView)
	app.Post("/ahsp_templates/:templateId/labor_components/new", session.IsAuth, ahspLaborComponentHandler.CreateAhspLaborComponent)
	app.Get("/ahsp_templates/:templateId/labor_components/:componentId/edit", session.IsAuth, ahspLaborComponentHandler.AhspLaborComponentEditModalView)
	app.Post("/ahsp_templates/:templateId/labor_components/:componentId/edit", session.IsAuth, ahspLaborComponentHandler.UpdateAhspLaborComponent)
	app.Get("/ahsp_templates/:templateId/labor_components/:componentId/delete", session.IsAuth, ahspLaborComponentHandler.AhspLaborComponentDeleteModalView)
	app.Delete("/ahsp_templates/:templateId/labor_components/:componentId/delete", session.IsAuth, ahspLaborComponentHandler.DeleteAhspLaborComponent)

	// projects
	app.Get("/projects", session.IsAuth, projectsHandler.ProjectsMainPageTableView)
	app.Get("/projects/new", session.IsAuth, projectsHandler.ProjectCreateModalView)
	app.Post("/projects/new", session.IsAuth, projectsHandler.CreateProject)
	app.Get("/projects/:id/edit", session.IsAuth, projectsHandler.ProjectEditModalView)
	app.Post("/projects/:id/edit", session.IsAuth, projectsHandler.UpdateProject)
	app.Get("/projects/:id/delete", session.IsAuth, projectsHandler.ProjectDeleteModalView)
	app.Delete("/projects/:id/delete", session.IsAuth, projectsHandler.DeleteProject)

	// project detail page
	app.Get("/project/:id", session.IsAuth, projectWorkItemsHandler.ProjectDetailPage)

	// project work items
	app.Get("/project/:id/work-items/new", session.IsAuth, projectWorkItemsHandler.ProjectWorkItemCreateModalView)
	app.Post("/project/:id/work-items", session.IsAuth, projectWorkItemsHandler.CreateProjectWorkItem)
	app.Get("/project/:id/work-items/:workItemId/edit", session.IsAuth, projectWorkItemsHandler.ProjectWorkItemEditModalView)
	app.Post("/project/:id/work-items/:workItemId/edit", session.IsAuth, projectWorkItemsHandler.UpdateProjectWorkItem)
	app.Get("/project/:id/work-items/:workItemId/delete", session.IsAuth, projectWorkItemsHandler.ProjectWorkItemDeleteModalView)
	app.Delete("/project/:id/work-items/:workItemId/delete", session.IsAuth, projectWorkItemsHandler.DeleteProjectWorkItem)

	// project work item costs
	app.Get("/work-items/:id/costs", session.IsAuth, projectWorkItemsHandler.ProjectWorkItemCostsView)

	app.Get("/countries", session.IsAuth, handlers.CountriesView)

	// Dashboard route (replacing countries as the main landing page)
	app.Get("/dashboard", session.IsAuth, dashboardHandler.Dashboard)

	// Material Summary routes
	app.Get("/material-summary", session.IsAuth, materialSummaryHandler.MaterialSummary)
	app.Get("/material-summary/export", session.IsAuth, materialSummaryHandler.ExportMaterialSummary)

	// Project Material Summary routes
	app.Get("/projects/:id/material-summary", session.IsAuth, materialSummaryHandler.ProjectMaterialSummary)
	app.Get("/projects/:id/material-summary/export", session.IsAuth, materialSummaryHandler.ExportProjectMaterialSummary)

	app.Get("/countries/:id", handlers.GetCountryDetails)
	app.Get("/search", handlers.SearchCountries)

	app.Get("/demo/loading-test", handlers.DemoLoadingTest)
	app.Get("/demo/success-test", handlers.SucceessModalTest)
	app.Get("/demo/error-test", handlers.ErrorModalTest)

	startServerWithGracefulShutdown(app)
}

func startServerWithGracefulShutdown(app *fiber.App) {
	// channel for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002" // default port
	}
	go func() {
		log.Println("Starting server on port " + port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Server startup error: %v", err)
		}
	}()

	log.Println("Server started successfully.")

	// Wait for shutdown signal
	<-quit
	log.Println("Shutdown signal received...")

	// Wait 5 seconds before starting graceful shutdown
	log.Println("Waiting 5 seconds before graceful shutdown...")
	time.Sleep(5 * time.Second)

	// Graceful shutdown
	log.Println("Gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Server forced to shutdown due to error: %v", err)
	} else {
		log.Println("Server shutdown complete")
	}

	log.Println("Application terminated")
}
