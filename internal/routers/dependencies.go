package routers

import (
	"fmt"

	"github.com/troptropcontent/qr_code_maintenance/internal/jobs"
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
	BackGroundJobRunner      jobs.BackgroundJobRunner
}

func NewRouterDependencies(db *gorm.DB, emailService email.EmailService, storageService storage.StorageService, translationService translation.TranslatorService, backGroundJobRunner jobs.BackgroundJobRunner) (d *Dependencies, err error) {
	if db == nil {
		return nil, fmt.Errorf("db can not be nil")
	}
	if emailService == nil {
		return nil, fmt.Errorf("emailService can not be nil")
	}
	if storageService == nil {
		return nil, fmt.Errorf("storageService can not be nil")
	}
	if translationService == nil {
		return nil, fmt.Errorf("translationService can not be nil")
	}
	if backGroundJobRunner == nil {
		return nil, fmt.Errorf("backGroundJobRunner can not be nil")
	}

	return &Dependencies{
		DB:                       db,
		EmailNotificationService: emailService,
		StorageService:           storageService,
		TranslationService:       translationService,
		BackGroundJobRunner:      backGroundJobRunner,
	}, nil
}
