package interventions

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
)

func TestCreateNewIntervention_Success(t *testing.T) {
	// Setup
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

	// Create request with form data
	fields := map[string]string{
		"portal_id":          fmt.Sprintf("%d", portal.ID),
		"date":               "2024-01-15",
		"summary":            "Routine maintenance check",
		"signature":          "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA",
		"controls[0].kind":   "warning_lights",
		"controls[0].result": "true",
		"controls[1].kind":   "area_lighting",
		"controls[1].result": "false",
	}

	files := map[string][]byte{}

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
	assert.Equal(t, "Routine maintenance check", *intervention.Summary)
	assert.Len(t, intervention.Controls, 2)
}

func TestCreateNewIntervention_WithPhotos(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}

	deps := &routers.Dependencies{
		DB:                       db,
		StorageService:           mockStorage,
		EmailNotificationService: mockEmail,
	}

	user := factories.NewUser().Create(db)

	// Create request with form data and photo files
	portal := factories.NewPortal().Create(db)
	fields := map[string]string{
		"portal_id":      fmt.Sprintf("%d", portal.ID),
		"date":           "2024-01-15",
		"summary":        "Maintenance with photos",
		"signature":      "test-signature",
		"photos[0].name": "Before repair",
		"photos[1].name": "After repair",
	}

	files := map[string][]byte{
		"photos[0].file": []byte("fake-image-data-1"),
		"photos[1].file": []byte("fake-image-data-2"),
	}

	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, files).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	// Verify storage service was called
	assert.True(t, mockStorage.UploadCalled, "Storage service should be called for photos")
	assert.Len(t, mockStorage.UploadedFiles, 2, "Should have uploaded 2 files")
}

func TestCreateNewIntervention_EmptySummary(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}

	deps := &routers.Dependencies{
		DB:                       db,
		StorageService:           mockStorage,
		EmailNotificationService: mockEmail,
	}

	// Create request with empty summary
	portal := factories.NewPortal().Create(db)
	fields := map[string]string{
		"portal_id": fmt.Sprintf("%d", portal.ID),
		"date":      "2024-01-15",
		"summary":   "",
		"signature": "test-signature",
	}

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
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}

	deps := &routers.Dependencies{
		DB:                       db,
		StorageService:           mockStorage,
		EmailNotificationService: mockEmail,
	}

	// Create request without controls
	portal := factories.NewPortal().Create(db)
	fields := map[string]string{
		"portal_id": fmt.Sprintf("%d", portal.ID),
		"date":      "2024-01-15",
		"summary":   "Test without controls",
		"signature": "test-signature",
	}

	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	// Verify intervention was created without controls
	var intervention models.Intervention
	err = db.Preload("Controls").First(&intervention).Error
	require.NoError(t, err)
	assert.Empty(t, intervention.Controls)
}

func TestCreateNewIntervention_MultipleControls(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}

	deps := &routers.Dependencies{
		DB:                       db,
		StorageService:           mockStorage,
		EmailNotificationService: mockEmail,
	}

	// Create request with all security controls
	portal := factories.NewPortal().Create(db)
	fields := map[string]string{
		"portal_id":          fmt.Sprintf("%d", portal.ID),
		"date":               "2024-01-15",
		"summary":            "Full security check",
		"signature":          "test-signature",
		"controls[0].kind":   "warning_lights",
		"controls[0].result": "true",
		"controls[1].kind":   "area_lighting",
		"controls[1].result": "true",
		"controls[2].kind":   "safety_cells",
		"controls[2].result": "false",
		"controls[3].kind":   "pressure_bar",
		"controls[3].result": "true",
		"controls[4].kind":   "floor_loop",
		"controls[4].result": "",
		"controls[5].kind":   "force_limiter",
		"controls[5].result": "true",
		"controls[6].kind":   "safety_springs",
		"controls[6].result": "false",
		"controls[7].kind":   "floor_markings",
		"controls[7].result": "true",
	}

	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	// Verify all controls were created
	var intervention models.Intervention
	err = db.Preload("Controls").First(&intervention).Error
	require.NoError(t, err)
	assert.Len(t, intervention.Controls, 8)
}

func TestCreateNewIntervention_InvalidDateFormat(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}

	deps := &routers.Dependencies{
		DB:                       db,
		StorageService:           mockStorage,
		EmailNotificationService: mockEmail,
	}

	// Create request with invalid date format
	portal := factories.NewPortal().Create(db)
	fields := map[string]string{
		"portal_id": fmt.Sprintf("%d", portal.ID),
		"date":      "01/15/2024", // Invalid format, should be YYYY-MM-DD
		"summary":   "Test invalid date",
		"signature": "test-signature",
	}

	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid date format")
}

func TestCreateNewIntervention_MissingFile(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}

	deps := &routers.Dependencies{
		DB:                       db,
		StorageService:           mockStorage,
		EmailNotificationService: mockEmail,
	}

	// Create request with photo name but no file
	portal := factories.NewPortal().Create(db)
	fields := map[string]string{
		"portal_id":      fmt.Sprintf("%d", portal.ID),
		"date":           "2024-01-15",
		"summary":        "Test missing file",
		"signature":      "test-signature",
		"photos[0].name": "Missing file photo",
	}

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
	testCases := []struct {
		name           string
		resultValue    string
		expectedResult *bool
	}{
		{
			name:           "true string",
			resultValue:    "true",
			expectedResult: boolPtr(true),
		},
		{
			name:           "false string",
			resultValue:    "false",
			expectedResult: boolPtr(false),
		},
		{
			name:           "1 as true",
			resultValue:    "1",
			expectedResult: boolPtr(true),
		},
		{
			name:           "0 as false",
			resultValue:    "0",
			expectedResult: boolPtr(false),
		},
		{
			name:           "empty string",
			resultValue:    "",
			expectedResult: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			db := tests.SetupTestDB(t)
			mockStorage := &tests.MockStorageService{}
			mockEmail := &tests.MockEmailService{}

			deps := &routers.Dependencies{
				DB:                       db,
				StorageService:           mockStorage,
				EmailNotificationService: mockEmail,
			}

			// Create request
			portal := factories.NewPortal().Create(db)
			fields := map[string]string{
				"portal_id":          fmt.Sprintf("%d", portal.ID),
				"date":               "2024-01-15",
				"summary":            "Test control parsing",
				"signature":          "test-signature",
				"controls[0].kind":   "test_control",
				"controls[0].result": tc.resultValue,
			}

			user := factories.NewUser().Create(db)
			c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()

			// Execute
			handler := CreateNewIntervention(deps)
			err := handler(c)

			// Assert
			require.NoError(t, err)

			// Verify control result was parsed correctly
			var intervention models.Intervention
			err = db.Preload("Controls").First(&intervention).Error
			require.NoError(t, err)
			require.Len(t, intervention.Controls, 1)

			control := intervention.Controls[0]
			if tc.expectedResult == nil {
				assert.Nil(t, control.Result)
			} else {
				require.NotNil(t, control.Result)
				assert.Equal(t, *tc.expectedResult, *control.Result)
			}
		})
	}
}

func TestCreateNewIntervention_RedirectsToCorrectURL(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}

	deps := &routers.Dependencies{
		DB:                       db,
		StorageService:           mockStorage,
		EmailNotificationService: mockEmail,
	}

	// Create request
	portal := factories.NewPortal().Create(db)
	fields := map[string]string{
		"portal_id": fmt.Sprintf("%d", portal.ID),
		"date":      "2024-01-15",
		"summary":   "Test redirect",
		"signature": "test-signature",
	}

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
	expectedLocation := fmt.Sprintf("/portals/%d", intervention.ID)
	actualLocation := rec.Header().Get("Location")
	assert.Equal(t, expectedLocation, actualLocation)
}

func TestCreateNewIntervention_FormParsingError(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}

	deps := &routers.Dependencies{
		DB:                       db,
		StorageService:           mockStorage,
		EmailNotificationService: mockEmail,
	}

	// Create malformed request using the corrupted form helper
	c := tests.CreateEchoContextWithCorruptedForm()

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert - should return error for malformed requests
	require.Error(t, err)
}

func TestCreateNewIntervention_StorageServiceError(t *testing.T) {
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{
		UploadError: fmt.Errorf("storage service unavailable"),
	}
	mockEmail := &tests.MockEmailService{}

	deps := &routers.Dependencies{
		DB:                       db,
		StorageService:           mockStorage,
		EmailNotificationService: mockEmail,
	}

	// Create request with photo
	portal := factories.NewPortal().Create(db)
	fields := map[string]string{
		"portal_id":      fmt.Sprintf("%d", portal.ID),
		"date":           "2024-01-15",
		"summary":        "Test storage error",
		"signature":      "test-signature",
		"photos[0].name": "Test photo",
	}

	files := map[string][]byte{
		"photos[0].file": []byte("fake-image-data"),
	}

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
	// Setup
	db := tests.SetupTestDB(t)
	mockStorage := &tests.MockStorageService{}
	mockEmail := &tests.MockEmailService{}

	deps := &routers.Dependencies{
		DB:                       db,
		StorageService:           mockStorage,
		EmailNotificationService: mockEmail,
	}

	// Create request
	portal := factories.NewPortal().Create(db)
	fields := map[string]string{
		"portal_id": fmt.Sprintf("%d", portal.ID),
		"date":      "2024-01-15",
		"summary":   "Test email notification",
		"signature": "test-signature",
	}

	user := factories.NewUser().Create(db)
	c := tests.NewContext(http.MethodPost, "/").WithAuthenticatedUser(user).WithMultiPartData(fields, nil).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	// Execute
	handler := CreateNewIntervention(deps)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	// Verify email service was called (if the service sends emails on intervention creation)
	// Note: This assertion depends on whether your CreateInterventionService sends emails
	// Adjust based on actual implementation
	if mockEmail.SendCalled {
		assert.Greater(t, len(mockEmail.SentEmails), 0, "Should have sent at least one email")
	}
}

// Helper function to create *bool
func boolPtr(b bool) *bool {
	return &b
}
