package components

import (
	"fmt"
	"math"
	"strings"
)

// formatCurrency formats a float64 value as Indonesian Rupiah with thousand separators
// Example: 1000000 -> "Rp 1.000.000"
func formatCurrency(value float64) string {
	// Round to nearest integer
	rounded := int(math.Round(value))

	// Handle negative values
	if rounded < 0 {
		rounded = -rounded
	}

	// Format with thousand separators
	str := fmt.Sprintf("%d", rounded)

	// Insert dots from right to left every 3 digits
	var result strings.Builder
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(".")
		}
		result.WriteRune(c)
	}

	return "Rp " + result.String()
}
