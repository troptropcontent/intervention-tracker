package tests

import (
	"testing"
	"time"
)

func MustParseDate(t *testing.T, dateStr string) time.Time {
	t.Helper()
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t.Fatalf("failed to parse date %q: %v", dateStr, err)
	}
	return date
}
