package interventions

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/attachments"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/email"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/jobs"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
	"gorm.io/gorm"
)

type CreateInterventionService struct {
	DB                       *gorm.DB
	StorageService           storage.StorageService
	EmailNotificationService email.EmailService
	JobRunner                jobs.BackgroundJobRunner
}

// NewInterventionService creates a new intervention service with validation
func NewCreateInterventionService(
	db *gorm.DB,
	storageService storage.StorageService,
	emailNotificationService email.EmailService,
	jobRunner jobs.BackgroundJobRunner,
) (*CreateInterventionService, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	if storageService == nil {
		return nil, fmt.Errorf("storageService cannot be nil")
	}
	if emailNotificationService == nil {
		return nil, fmt.Errorf("emailNotificationService cannot be nil")
	}
	if jobRunner == nil {
		return nil, fmt.Errorf("jobRunner cannot be nil")
	}

	return &CreateInterventionService{
		DB:                       db,
		StorageService:           storageService,
		EmailNotificationService: emailNotificationService,
		JobRunner:                jobRunner,
	}, nil
}

type PhotoData struct {
	Name string
	File *multipart.FileHeader
}

type ControlData struct {
	Kind   models.ControlKind
	Result models.ControlResult
}

type CreateArgs struct {
	Date      time.Time
	Summary   string
	TimeSpent *int
	Signature string
	PortalID  uint
	UserID    uint
	UserName  string
	Type      models.InterventionType
	Photos    []PhotoData
	Controls  []ControlData
}

type BaseInterventionCreateArgs struct {
	Date      string
	Summary   string
	Signature string
	PortalID  uint
	UserID    uint
	UserName  string
	Photos    []PhotoData
}

type MaintenanceInterventionCreateArgs struct {
	BaseInterventionCreateArgs
	Controls []struct {
		Kind   string
		Result string
	}
}

type RepairInterventionCreateArgs struct {
	BaseInterventionCreateArgs
	TimeSpent *float64
}

// Create creates a new intervention (maintenance or repair) and handles photo uploads
// For maintenance interventions, controls are included; for repair interventions, they are omitted
// All operations (intervention, controls, photo uploads) are done synchronously in a transaction
// This ensures atomicity: either everything succeeds or everything rolls back
func (s *CreateInterventionService) Create(args *CreateArgs) (*models.Intervention, error) {
	intervention, err := s.buildIntervention(args)
	if err != nil {
		return nil, err
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(intervention).Error; err != nil {
			return fmt.Errorf("failed to create intervention: %w", err)
		}

		utils.PP(args.Photos)

		if err := s.attachPhotos(tx, intervention.ID, args.Photos); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.JobRunner.Perform(func() { AttachReportPdf(s.DB, s.StorageService, s.EmailNotificationService, intervention.ID) })

	return intervention, nil
}

// buildIntervention constructs an intervention model from the provided arguments
func (s *CreateInterventionService) buildIntervention(args *CreateArgs) (*models.Intervention, error) {
	intervention := &models.Intervention{
		Date:      args.Date,
		Type:      args.Type,
		Summary:   &args.Summary,
		TimeSpent: args.TimeSpent,
		Signature: args.Signature,
		PortalID:  args.PortalID,
		UserID:    args.UserID,
		UserName:  args.UserName,
	}

	// Only build controls for maintenance interventions
	if args.Type == models.InterventionTypeMaintenance {
		intervention.Controls = s.buildControls(args.Controls)
	}

	validationError := intervention.Validate()
	if validationError != nil {
		return nil, validationError
	}

	return intervention, nil
}

// buildControls converts control arguments to control models
func (s *CreateInterventionService) buildControls(ctrlArgs []ControlData) []models.Control {
	controls := make([]models.Control, 0, len(ctrlArgs))
	for _, ctrl := range ctrlArgs {
		controls = append(controls, models.Control{
			Kind:   ctrl.Kind,
			Result: ctrl.Result,
		})
	}
	return controls
}

// attachPhotos uploads photos and creates attachment records within a transaction
func (s *CreateInterventionService) attachPhotos(tx *gorm.DB, interventionID uint, photos []PhotoData) error {
	if len(photos) == 0 {
		return nil
	}

	ctx := context.Background() // TODO: Pass context from handler when available
	attachmentService := attachments.NewAttachmentService(tx, s.StorageService)

	for _, photo := range photos {
		if err := s.attachSinglePhoto(ctx, attachmentService, photo, interventionID); err != nil {
			return err
		}
	}

	return nil
}

// attachSinglePhoto handles the upload and attachment of a single photo
func (s *CreateInterventionService) attachSinglePhoto(ctx context.Context, attachmentService attachments.Attacher, photo PhotoData, interventionID uint) error {
	src, err := photo.File.Open()
	if err != nil {
		return fmt.Errorf("failed to open photo file %s: %w", photo.Name, err)
	}
	defer src.Close()

	ext := filepath.Ext(photo.File.Filename)
	fileName := fmt.Sprintf("%s%s", photo.Name, ext)

	if err := attachmentService.Attach(ctx, src, fileName, interventionID, "interventions", "photo"); err != nil {
		// This will rollback the entire transaction AND cleanup the S3 upload
		return fmt.Errorf("failed to attach photo %s: %w", fileName, err)
	}

	return nil
}
