package routers

import (
	"github.com/troptropcontent/qr_code_maintenance/internal/services/email"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/translation"
	"gorm.io/gorm"
)

type Dependencies struct {
	DB                       *gorm.DB
	EmailNotificationService email.EmailService
	StorageService           storage.StorageService
	TranslationService       translation.TranslatorService
}
