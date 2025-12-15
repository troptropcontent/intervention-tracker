package templates

import (
	"fmt"
	"net/url"

	"github.com/troptropcontent/qr_code_maintenance/internal/models"
)

func AddQueryParamsToURL(baseURL string, queryParams map[string]string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	values := u.Query()
	for key, value := range queryParams {
		values.Set(key, value)
	}

	u.RawQuery = values.Encode()
	return u.String(), nil
}
func MustAddQueryParamsToURL(baseURL string, queryParams map[string]string) string {
	u, err := AddQueryParamsToURL(baseURL, queryParams)
	if err != nil {
		return baseURL
	}
	return u
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
