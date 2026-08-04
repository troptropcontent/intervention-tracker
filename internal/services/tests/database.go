package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB creates an in-memory SQLite database for testing
// It automatically runs migrations for all common models
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to create test database")

	// Enable foreign key constraints to match PostgreSQL behavior
	db.Exec("PRAGMA foreign_keys = ON")

	// Run migrations for all models
	err = db.AutoMigrate(
		&models.Intervention{},
		&models.Control{},
		&models.Attachment{},
		&models.Portal{},
		&models.QRCodeBatch{},
		&models.QRCode{},
	)
	require.NoError(t, err, "Failed to run migrations")

	return db
}
