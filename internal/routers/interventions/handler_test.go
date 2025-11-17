package interventions

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/translation"
	"gorm.io/gorm"
)

func setupDependencies(db *gorm.DB) *routers.Dependencies {
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}
	jobRunner := &tests.MockBackgroundJobRunner{}
	translator := translation.NewTranslator()
	deps, _ := routers.NewRouterDependencies(db, mockEmail, mockStorage, translator, jobRunner)
	return deps
}

type SetupArgs struct {
	DB             *gorm.DB
	StorageService *tests.MockStorageService
	EmailService   *tests.MockEmailService
}

func setup(t *testing.T, args ...SetupArgs) (deps *routers.Dependencies, db *gorm.DB, storageService *tests.MockStorageService, emailService *tests.MockEmailService) {
	// Set defaults
	db = tests.SetupTestDB(t)
	storageService = &tests.MockStorageService{}
	emailService = &tests.MockEmailService{}
	jobRunner := &tests.MockBackgroundJobRunner{}
	translator := translation.NewTranslator()

	// Override with provided args if present
	if len(args) > 0 {
		if args[0].DB != nil {
			db = args[0].DB
		}
		if args[0].StorageService != nil {
			// Use provided storage service instead of mock
			storageService = args[0].StorageService
		}
		if args[0].EmailService != nil {
			emailService = args[0].EmailService
		}
	}

	deps = &routers.Dependencies{
		DB:                       db,
		StorageService:           storageService,
		EmailNotificationService: emailService,
		TranslationService:       translator,
		BackGroundJobRunner:      jobRunner,
	}

	return deps, db, storageService, emailService
}

func TestCreateNewIntervention_Success(t *testing.T) {
	// Setup
	deps, db, _, _ := setup(t)

	user := factories.NewUser().Create(db)
	portal := factories.NewPortal().Create(db)
	fields, files := NewCreateInterventionFormBuilder(fmt.Sprintf("%d", portal.ID)).Build()

	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, files).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	// Verify redirect URL format
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "/portals/")
	assert.NotEqual(t, "/portals/0", location)

	// Verify intervention was created in database
	var intervention models.Intervention
	err = db.Preload("Controls").First(&intervention).Error
	require.NoError(t, err)
	assert.NotNil(t, intervention.Summary)
	assert.Equal(t, "Test intervention", *intervention.Summary)
	assert.Len(t, intervention.Controls, 17)
}

func TestCreateNewIntervention_RepairType_Success(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)

	deps := setupDependencies(db)

	user := factories.NewUser().Create(db)
	portal := factories.NewPortal().Create(db)

	// Create request with type: "repair"
	fields := map[string]string{
		"portal_id":        fmt.Sprintf("%d", portal.ID),
		"type":             "repair",
		"date":             "2024-01-15",
		"summary":          "Door spring replacement",
		"signature":        "test-signature",
		"time_spent_hours": "2.5",
	}

	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	// Verify repair intervention was created with correct type
	var intervention models.Intervention
	err = db.First(&intervention).Error
	require.NoError(t, err)
	assert.Equal(t, "repair", string(intervention.Type))
	assert.NotNil(t, intervention.Summary)
	assert.Equal(t, "Door spring replacement", *intervention.Summary)
}

func TestCreateNewIntervention_WithPhotos(t *testing.T) {
	// Setup
	deps, db, mockedStorageService, _ := setup(t)

	user := factories.NewUser().Create(db)

	// Create request with form data and photo files
	portal := factories.NewPortal().Create(db)
	portalIDString := fmt.Sprintf("%d", portal.ID)
	fields, files := NewCreateInterventionFormBuilder(portalIDString).WithPhoto(0, "Before repair", []byte("fake-image-data-1")).WithPhoto(1, "After repair", []byte("fake-image-data-2")).Build()

	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, files).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	// Verify storage service was called
	assert.True(t, mockedStorageService.UploadCalled, "Storage service should be called for photos")
	assert.Len(t, mockedStorageService.UploadedFiles, 3, "Should have uploaded 3 files 2 photos and 1 report")
}

func TestCreateNewIntervention_EmptySummary(t *testing.T) {
	// Setup
	deps, db, _, _ := setup(t)

	// Create request with empty summary
	portal := factories.NewPortal().Create(db)
	portalIDString := fmt.Sprintf("%d", portal.ID)
	fields, _ := NewCreateInterventionFormBuilder(portalIDString).WithSummary("").Build()

	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	// Verify intervention was created with empty summary
	var intervention models.Intervention
	err = db.First(&intervention).Error
	require.NoError(t, err)
	assert.NotNil(t, intervention.Summary)
	assert.Equal(t, "", *intervention.Summary)
}

func TestCreateNewIntervention_NoControls(t *testing.T) {
	// Setup
	deps, db, _, _ := setup(t)

	// Create request without controls
	portal := factories.NewPortal().Create(db)
	portalIDString := fmt.Sprintf("%d", portal.ID)
	fields, _ := NewCreateInterventionFormBuilder(portalIDString).WithOutControls().Build()

	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)

	// Verify no intervention was created
	var intervention models.Intervention
	db.Find(&intervention)
	require.Equal(t, intervention.ID, uint(0))
}

func TestCreateNewIntervention_InvalidDateFormat(t *testing.T) {
	// Setup
	deps, db, _, _ := setup(t)

	// Create request without controls
	portal := factories.NewPortal().Create(db)
	fields, _ := NewCreateInterventionFormBuilder(fmt.Sprintf("%d", portal.ID)).WithDate("invalid").Build()
	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Contains(t, httpErr.Message, "ERROR:parsing time \"invalid\" as \"2006-01-02\"")
}

func TestCreateNewIntervention_MissingFile(t *testing.T) {
	// Setup
	deps, db, _, _ := setup(t)

	// Create request with photo name but no file
	portal := factories.NewPortal().Create(db)
	fields, _ := NewCreateInterventionFormBuilder(fmt.Sprintf("%d", portal.ID)).WithKeyValue("photos[0].name", "photo 1").Build()

	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert - should succeed, just skip the photo without file
	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestCreateNewIntervention_ControlResultParsing(t *testing.T) {
	// Setup
	deps, db, _, _ := setup(t)

	// Create request with an invalid controls result
	portal := factories.NewPortal().Create(db)
	fields, _ := NewCreateInterventionFormBuilder(fmt.Sprintf("%d", portal.ID)).WithControl(16, string(models.ControlKinds[16]), "invalid").Build()
	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Contains(t, httpErr.Message, "invalid control result: 'invalid'")
}

func TestCreateNewIntervention_RedirectsToCorrectURL(t *testing.T) {
	// Setup
	deps, db, _, _ := setup(t)

	// Create request with an invalid controls result
	portal := factories.NewPortal().Create(db)
	fields, _ := NewCreateInterventionFormBuilder(fmt.Sprintf("%d", portal.ID)).Build()

	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	// Get the created intervention ID from database
	var intervention models.Intervention
	err = db.First(&intervention).Error
	require.NoError(t, err)

	// Verify redirect URL
	expectedLocation := fmt.Sprintf("/admin/portals/%d", intervention.ID)
	actualLocation := rec.Header().Get("Location")
	assert.Equal(t, expectedLocation, actualLocation)
}

func TestCreateNewIntervention_FormParsingError(t *testing.T) {
	// Setup
	deps, db, _, _ := setup(t)

	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithInvalidFormData().Build()

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Contains(t, httpErr.Message, "failed to parse form data")
}

func TestCreateNewIntervention_StorageServiceError(t *testing.T) {
	// Setup
	deps, db, _, _ := setup(t, SetupArgs{StorageService: &tests.MockStorageService{
		UploadError: fmt.Errorf("storage service unavailable"),
	}})

	// Create request with an invalid controls result
	portal := factories.NewPortal().Create(db)
	fields, files := NewCreateInterventionFormBuilder(fmt.Sprintf("%d", portal.ID)).WithPhotos("photo 1", "photo 2").Build()

	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, files).Build()

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert - should return error from storage service
	require.Error(t, err)
	assert.Contains(t, err.Error(), "storage service unavailable")
}

func TestCreateNewIntervention_EmailServiceCalled(t *testing.T) {
	deps, db, _, emailService := setup(t)

	// Create request
	portal := factories.NewPortal().Create(db)
	fields, _ := NewCreateInterventionFormBuilder(fmt.Sprintf("%d", portal.ID)).Build()

	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, len(emailService.SentEmails), 1, "Should have sent one email")
}

func TestGetNewInterventionForm(t *testing.T) {
	// Shared setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}

	deps := &routers.Dependencies{
		DB:                       db,
		StorageService:           mockStorage,
		EmailNotificationService: mockEmail,
	}
	user := factories.NewUser().Create(db)
	portal := factories.NewPortal().Create(db)

	t.Run("SuccessWithValidParams", func(t *testing.T) {
		handler := GetNewInterventionForm(deps)
		c := tests.NewContext().
			WithAuthenticatedUser(user).
			WithQueryParams(map[string]string{
				"portal_id":         fmt.Sprintf("%d", portal.ID),
				"intervention_type": "repair",
			}).
			Build()
		err := handler(c)
		require.NoError(t, err)
	})

	t.Run("ErrorWhenPortalIdIsNotProvided", func(t *testing.T) {
		handler := GetNewInterventionForm(deps)
		c := tests.NewContext().WithAuthenticatedUser(user).Build()
		err := handler(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "query param portal_id is invalid or missing")
	})

	t.Run("ErrorWhenInterventionTypeIsNotProvided", func(t *testing.T) {
		handler := GetNewInterventionForm(deps)
		c := tests.NewContext().
			WithAuthenticatedUser(user).
			WithQueryParams(map[string]string{"portal_id": fmt.Sprintf("%d", portal.ID)}).
			Build()
		err := handler(c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "query param intervention_type is invalid or missing")
	})
}
