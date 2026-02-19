package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/momokii/go-rab-maker/backend/databases"
	"github.com/momokii/go-rab-maker/backend/middlewares"
	"github.com/momokii/go-rab-maker/backend/models"
	"github.com/momokii/go-rab-maker/backend/repository/material_summary"
	"github.com/momokii/go-rab-maker/backend/repository/project_item_costs"
	"github.com/momokii/go-rab-maker/backend/repository/projects"
	"github.com/momokii/go-rab-maker/backend/utils"
	"github.com/momokii/go-rab-maker/frontend/components"
)

type MaterialSummaryHandler struct {
	dbService            databases.SQLiteServices
	materialSummaryRepo  material_summary.MaterialSummaryRepo
	projectsRepo         projects.ProjectsRepo
	projectItemCostsRepo *project_item_costs.ProjectItemCostsRepo
}

func NewMaterialSummaryHandler(
	dbService databases.SQLiteServices,
	materialSummaryRepo material_summary.MaterialSummaryRepo,
	projectsRepo projects.ProjectsRepo,
	projectItemCostsRepo *project_item_costs.ProjectItemCostsRepo,
) *MaterialSummaryHandler {
	return &MaterialSummaryHandler{
		dbService:            dbService,
		materialSummaryRepo:  materialSummaryRepo,
		projectsRepo:         projectsRepo,
		projectItemCostsRepo: projectItemCostsRepo,
	}
}

// MaterialSummary displays a summary of all materials needed across projects
func (h *MaterialSummaryHandler) MaterialSummary(c *fiber.Ctx) error {
	var materialSummaries []models.MaterialSummary

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Use the database service transaction
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get material summary data
		materialSummariesData, err := h.materialSummaryRepo.GetAllMaterialsSummary(tx, userData.ID)
		if err != nil {
			log.Printf("Error fetching material summary: %v", err)
			materialSummaries = []models.MaterialSummary{}
		} else {
			materialSummaries = materialSummariesData
		}

		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch material summary")
	}

	// Render the material summary page
	materialSummaryComponent := components.MaterialSummaryPage(materialSummaries)
	return adaptor.HTTPHandler(templ.Handler(materialSummaryComponent))(c)
}

// ExportMaterialSummary exports the material summary to a file (PDF or Excel)
func (h *MaterialSummaryHandler) ExportMaterialSummary(c *fiber.Ctx) error {
	// Get export format from query parameter
	format := c.Query("format", "pdf")

	if format != "pdf" && format != "excel" {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid format. Use 'pdf' or 'excel'")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Use the database service transaction
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Get material summary data
		materialSummaries, err := h.materialSummaryRepo.GetAllMaterialsSummary(tx, userData.ID)
		if err != nil {
			log.Printf("Error fetching material summary: %v", err)
			return fiber.StatusInternalServerError, err
		}

		// Export based on format
		if format == "pdf" {
			return h.exportToPDF(c, materialSummaries)
		} else {
			return h.exportToExcel(c, materialSummaries)
		}
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Export failed")
	}

	return nil
}

// exportToPDF exports the material summary to PDF format
func (h *MaterialSummaryHandler) exportToPDF(c *fiber.Ctx, summaries []models.MaterialSummary) (int, error) {
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=material-summary.pdf")

	pdf := utils.NewPDFExporter("L", "mm", "A4")
	pdf.AddTitle("Material Summary")

	// Prepare headers and data
	headers := []string{"Item Name", "Type", "Total Quantity", "Unit", "Total Cost"}
	var rows [][]string

	for _, summary := range summaries {
		rows = append(rows, []string{
			summary.ItemName,
			string(summary.ItemType),
			fmt.Sprintf("%.2f", summary.TotalQuantity),
			summary.Unit,
			fmt.Sprintf("%.2f", summary.TotalCost),
		})
	}

	pdf.AddTable(headers, rows)

	// Write PDF
	pdfData, err := pdf.Write()
	if err != nil {
		return fiber.StatusInternalServerError, err
	}

	return fiber.StatusOK, c.Send(pdfData)
}

// exportToExcel exports the material summary to Excel format
func (h *MaterialSummaryHandler) exportToExcel(c *fiber.Ctx, summaries []models.MaterialSummary) (int, error) {
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=material-summary.xlsx")

	excel := utils.NewExcelExporter()

	// Prepare headers and data
	headers := []string{"Item Name", "Type", "Total Quantity", "Unit", "Total Cost"}
	var rows [][]interface{}

	for _, summary := range summaries {
		rows = append(rows, []interface{}{
			summary.ItemName,
			string(summary.ItemType),
			summary.TotalQuantity,
			summary.Unit,
			summary.TotalCost,
		})
	}

	if err := excel.AddSheet("Material Summary", headers, rows); err != nil {
		return fiber.StatusInternalServerError, err
	}

	// Write Excel
	excelData, err := excel.Write()
	if err != nil {
		return fiber.StatusInternalServerError, err
	}

	return fiber.StatusOK, c.Send(excelData)
}

// ProjectMaterialSummary displays a summary of materials needed for a specific project
func (h *MaterialSummaryHandler) ProjectMaterialSummary(c *fiber.Ctx) error {
	projectIdStr := c.Params("id")
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid project ID")
	}

	var materialSummaries []models.MaterialSummary
	var project models.Project

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Use the database service transaction
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Verify project ownership
		projectData, err := h.projectsRepo.FindById(tx, projectId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "Project not found")
			}
			return fiber.StatusInternalServerError, err
		}

		// Check if project belongs to current user
		if projectData.UserId != userData.ID {
			return fiber.StatusForbidden, fiber.NewError(fiber.StatusForbidden, "Access denied")
		}

		// Get detailed material summary data for the specific project
		detailedSummaries, err := h.projectItemCostsRepo.GetDetailedMaterialSummaryByProjectId(tx, projectId)
		if err != nil {
			log.Printf("Error fetching detailed project material summary: %v", err)
			// Fallback to regular summary if detailed fails
			materialSummariesData, err := h.materialSummaryRepo.GetProjectMaterialSummary(tx, projectId)
			if err != nil {
				log.Printf("Error fetching project material summary: %v", err)
				materialSummaries = []models.MaterialSummary{}
			} else {
				materialSummaries = materialSummariesData
			}
		} else {
			// Convert detailed summaries to regular summaries for compatibility
			for _, detailed := range detailedSummaries {
				materialSummaries = append(materialSummaries, models.MaterialSummary{
					ItemId:        detailed.ItemId,
					ItemName:      detailed.ItemName,
					TotalQuantity: detailed.TotalQuantity,
					Unit:          detailed.Unit,
					ItemType:      detailed.ItemType,
					TotalCost:     detailed.TotalCost,
				})
			}
		}

		project = projectData
		return fiber.StatusOK, nil
	}); err != nil {
		return utils.ResponseErrorModal(c, "Error", "Failed to fetch project material summary")
	}

	// Render the project material summary component
	projectMaterialSummaryComponent := components.ProjectMaterialSummary(materialSummaries, project)
	return adaptor.HTTPHandler(templ.Handler(projectMaterialSummaryComponent))(c)
}

// ExportProjectMaterialSummary exports the material summary for a specific project to a file
func (h *MaterialSummaryHandler) ExportProjectMaterialSummary(c *fiber.Ctx) error {
	projectIdStr := c.Params("id")
	projectId, err := strconv.Atoi(projectIdStr)
	if err != nil {
		return utils.ResponseErrorModal(c, "Error", "Invalid project ID")
	}

	// Get export format from query parameter
	format := c.Query("format", "pdf")

	if format != "pdf" && format != "excel" {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid format. Use 'pdf' or 'excel'")
	}

	// Get user from session
	userData := c.Locals(middlewares.SESSION_USER_NAME).(models.SessionUser)

	// Use the database service transaction
	if _, err := h.dbService.Transaction(c.Context(), func(tx *sql.Tx) (int, error) {
		// Verify project ownership
		project, err := h.projectsRepo.FindById(tx, projectId)
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

		// Get material summary data for the specific project
		materialSummaries, err := h.materialSummaryRepo.GetProjectMaterialSummary(tx, projectId)
		if err != nil {
			log.Printf("Error fetching project material summary: %v", err)
			return fiber.StatusInternalServerError, err
		}

		// Export based on format
		if format == "pdf" {
			return h.exportProjectToPDF(c, materialSummaries, project)
		} else {
			return h.exportProjectToExcel(c, materialSummaries, project)
		}
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Export failed")
	}

	return nil
}

// exportProjectToPDF exports the project material summary to PDF format
func (h *MaterialSummaryHandler) exportProjectToPDF(c *fiber.Ctx, summaries []models.MaterialSummary, project models.Project) (int, error) {
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=material-summary-"+project.ProjectName+".pdf")

	pdf := utils.NewPDFExporter("L", "mm", "A4")
	pdf.AddTitle(fmt.Sprintf("Material Summary - %s", project.ProjectName))

	// Prepare headers and data
	headers := []string{"Item Name", "Type", "Total Quantity", "Unit", "Total Cost"}
	var rows [][]string

	for _, summary := range summaries {
		rows = append(rows, []string{
			summary.ItemName,
			string(summary.ItemType),
			fmt.Sprintf("%.2f", summary.TotalQuantity),
			summary.Unit,
			fmt.Sprintf("%.2f", summary.TotalCost),
		})
	}

	pdf.AddTable(headers, rows)

	// Write PDF
	pdfData, err := pdf.Write()
	if err != nil {
		return fiber.StatusInternalServerError, err
	}

	return fiber.StatusOK, c.Send(pdfData)
}

// exportProjectToExcel exports the project material summary to Excel format
func (h *MaterialSummaryHandler) exportProjectToExcel(c *fiber.Ctx, summaries []models.MaterialSummary, project models.Project) (int, error) {
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=material-summary-"+project.ProjectName+".xlsx")

	excel := utils.NewExcelExporter()

	// Prepare headers and data
	headers := []string{"Item Name", "Type", "Total Quantity", "Unit", "Total Cost"}
	var rows [][]interface{}

	for _, summary := range summaries {
		rows = append(rows, []interface{}{
			summary.ItemName,
			string(summary.ItemType),
			summary.TotalQuantity,
			summary.Unit,
			summary.TotalCost,
		})
	}

	if err := excel.AddSheet(fmt.Sprintf("Materials - %s", project.ProjectName), headers, rows); err != nil {
		return fiber.StatusInternalServerError, err
	}

	// Write Excel
	excelData, err := excel.Write()
	if err != nil {
		return fiber.StatusInternalServerError, err
	}

	return fiber.StatusOK, c.Send(excelData)
}
