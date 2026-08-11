package settings

import (
	"testing"

	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/translation"
	"gorm.io/gorm"
)

func setup(t *testing.T) (deps *routers.Dependencies, db *gorm.DB) {
	db = tests.SetupTestDB(t)
	deps = &routers.Dependencies{
		DB:                       db,
		StorageService:           &tests.MockStorageService{},
		EmailNotificationService: &tests.MockEmailService{},
		TranslationService:       translation.NewTranslator(),
		BackGroundJobRunner:      &tests.MockBackgroundJobRunner{},
	}
	return deps, db
}
