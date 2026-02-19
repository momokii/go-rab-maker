package utils

import (
	"fmt"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

// ExcelExporter handles Excel file generation
type ExcelExporter struct {
	file *excelize.File
}

// NewExcelExporter creates a new Excel exporter
func NewExcelExporter() *ExcelExporter {
	return &ExcelExporter{
		file: excelize.NewFile(),
	}
}

// AddSheet adds a new sheet with data
func (e *ExcelExporter) AddSheet(sheetName string, headers []string, rows [][]interface{}) error {
	// Create a new sheet
	index, err := e.file.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}

	// Write headers
	for i, header := range headers {
		cell := string(rune('A'+i)) + "1"
		if err := e.file.SetCellValue(sheetName, cell, header); err != nil {
			return fmt.Errorf("failed to set header: %w", err)
		}
	}

	// Write data rows
	for rowIdx, row := range rows {
		for colIdx, value := range row {
			cell := string(rune('A'+colIdx)) + fmt.Sprintf("%d", rowIdx+2)
			if err := e.file.SetCellValue(sheetName, cell, value); err != nil {
				return fmt.Errorf("failed to set cell value: %w", err)
			}
		}
	}

	// Set active sheet
	e.file.SetActiveSheet(index)

	// Auto-fit columns
	for i := range headers {
		col := string(rune('A' + i))
		if err := e.file.SetColWidth(sheetName, col, col, 15); err != nil {
			return fmt.Errorf("failed to set column width: %w", err)
		}
	}

	return nil
}

// Write outputs the Excel file as bytes
func (e *ExcelExporter) Write() ([]byte, error) {
	buffer, err := e.file.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel buffer: %w", err)
	}
	return buffer.Bytes(), nil
}

// PDFExporter handles PDF file generation
type PDFExporter struct {
	pdf *gofpdf.Fpdf
}

// NewPDFExporter creates a new PDF exporter
func NewPDFExporter(orientation, unit, size string) *PDFExporter {
	pdf := gofpdf.New(orientation, unit, size, "")
	pdf.AddPage()
	return &PDFExporter{pdf: pdf}
}

// AddTitle adds a title to the PDF
func (p *PDFExporter) AddTitle(title string) {
	p.pdf.SetFont("Arial", "B", 16)
	p.pdf.Cell(40, 10, title)
	p.pdf.Ln(12)
}

// AddTable adds a table to the PDF
func (p *PDFExporter) AddTable(headers []string, rows [][]string) {
	// Set font for table
	p.pdf.SetFont("Arial", "", 10)

	// Calculate column widths
	colWidth := 190.0 / float64(len(headers))

	// Write headers
	p.pdf.SetFont("Arial", "B", 10)
	for _, header := range headers {
		p.pdf.CellFormat(colWidth, 7, header, "1", 0, "C", false, 0, "")
	}
	p.pdf.Ln(-1)
	p.pdf.SetFont("Arial", "", 10)

	// Write data rows
	for _, row := range rows {
		for _, cell := range row {
			p.pdf.CellFormat(colWidth, 6, cell, "1", 0, "L", false, 0, "")
		}
		p.pdf.Ln(-1)
	}
}

// Write outputs the PDF file as bytes
func (p *PDFExporter) Write() ([]byte, error) {
	var buf []byte
	writer := &bytesWriter{buf: &buf}
	if err := p.pdf.Output(writer); err != nil {
		return nil, err
	}
	return buf, nil
}

// bytesWriter is a writer that captures bytes
type bytesWriter struct {
	buf *[]byte
}

func (w *bytesWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

func (w *bytesWriter) Close() error {
	return nil
}
