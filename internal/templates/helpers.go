package templates

import (
	"fmt"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
)

// PaginationData holds pagination information
type PaginationData struct {
	CurrentPage int
	TotalPages  int
	TotalCount  int
	Limit       int
	HasPrev     bool
	HasNext     bool
	Search      string
}

func SplitControlKindsForDisplay() ([]models.ControlKind, []models.ControlKind) {
	mid := (len(models.ControlKinds) + 1) / 2 // Rounds up for odd numbers
	return models.ControlKinds[:mid], models.ControlKinds[mid:]
}

// FormatFloat formats a float64 value with the specified number of decimal places
func FormatFloat(value float64, decimals int) string {
	formatStr := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(formatStr, value)
}
