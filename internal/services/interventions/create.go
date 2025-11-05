package interventions

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"time"

	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/attachments"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/email"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/storage"
	"gorm.io/gorm"
)

type CreateInterventionService struct {
	DB                       *gorm.DB
	StorageService           storage.StorageService
	EmailNotificationService email.EmailService
}

type PhotoData struct {
	Name string
	File *multipart.FileHeader
}

type CreateArgs struct {
	Date      string
	Summary   string
	Signature string
	PortalID  uint
	UserID    uint
	UserName  string
	Type      models.InterventionType
	Photos    []PhotoData
	Controls  []struct {
		Kind   string
		Result string
	}
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

		if err := s.attachPhotos(tx, intervention.ID, args.Photos); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	go AttachReportPdf(s.DB, s.StorageService, s.EmailNotificationService, intervention.ID)

	return intervention, nil
}

// buildIntervention constructs an intervention model from the provided arguments
func (s *CreateInterventionService) buildIntervention(args *CreateArgs) (*models.Intervention, error) {
	date, err := time.Parse("2006-01-02", args.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	intervention := &models.Intervention{
		Date:      date,
		Type:      args.Type,
		Summary:   &args.Summary,
		Signature: args.Signature,
		PortalID:  args.PortalID,
		UserID:    args.UserID,
		UserName:  args.UserName,
	}

	// Only build controls for maintenance interventions
	if args.Type == models.InterventionTypeMaintenance {
		intervention.Controls = s.buildControls(args.Controls)
	}

	return intervention, nil
}

// buildControls converts control arguments to control models
func (s *CreateInterventionService) buildControls(ctrlArgs []struct {
	Kind   string
	Result string
}) []models.Control {
	controls := make([]models.Control, 0, len(ctrlArgs))
	for _, ctrl := range ctrlArgs {
		var result models.ControlResult
		if ctrl.Result != "" {
			if boolVal, err := strconv.ParseBool(ctrl.Result); err == nil {
				result = &boolVal
			}
		}
		// If Result is empty or invalid, result stays nil (not checked)

		controls = append(controls, models.Control{
			Kind:   ctrl.Kind,
			Result: result,
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
