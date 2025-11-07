package templates

import "github.com/troptropcontent/qr_code_maintenance/internal/models"

func SplitControlKindsForDisplay() ([]models.ControlKind, []models.ControlKind) {
	mid := (len(models.ControlKinds) + 1) / 2 // Rounds up for odd numbers
	return models.ControlKinds[:mid], models.ControlKinds[mid:]
}
