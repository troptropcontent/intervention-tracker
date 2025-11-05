package interventions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
	"gorm.io/gorm"
)

// testFixture creates portal and user for tests
type testFixture struct {
	portal *models.Portal
	user   *models.User
}

func setupTestFixture(db *gorm.DB) *testFixture {
	portal := factories.NewPortal().Create(db)
	user := factories.NewUser().Create(db)
	return &testFixture{
		portal: portal,
		user:   user,
	}
}

func TestCreateInterventionService_Create_Success(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := &CreateInterventionService{DB: db}
	fixture := setupTestFixture(db)

	args := &CreateArgs{
		Type:      models.InterventionTypeMaintenance,
		Date:      "2024-01-15",
		Summary:   "Routine maintenance check",
		Signature: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA",
		PortalID:  fixture.portal.ID,
		UserID:    fixture.user.ID,
		UserName:  fixture.user.FullName(),
		Controls: []struct {
			Kind   string
			Result string
		}{
			{Kind: "warning_lights", Result: "true"},
			{Kind: "area_lighting", Result: "false"},
			{Kind: "safety_cells", Result: ""},
		},
	}

	intervention, err := service.Create(args)

	require.NoError(t, err, "Create should succeed")
	assert.NotNil(t, intervention, "Intervention should not be nil")
	assert.NotZero(t, intervention.ID, "Intervention ID should be set")

	// Verify date parsing
	expectedDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedDate, intervention.Date, "Date should be parsed correctly")

	// Verify intervention type
	assert.Equal(t, models.InterventionTypeMaintenance, intervention.Type, "Type should be set to 'maintenance'")

	// Verify summary
	assert.NotNil(t, intervention.Summary, "Summary should not be nil")
	assert.Equal(t, "Routine maintenance check", *intervention.Summary, "Summary should match")

	// Verify signature
	assert.Equal(t, "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA", intervention.Signature, "Signature should match")

	// Verify controls were created
	assert.Len(t, intervention.Controls, 3, "Should have 3 controls")

	// Verify first control (true)
	assert.Equal(t, "warning_lights", intervention.Controls[0].Kind)
	assert.NotNil(t, intervention.Controls[0].Result, "Result should not be nil")
	assert.True(t, *intervention.Controls[0].Result, "Result should be true")
	assert.NotZero(t, intervention.Controls[0].ID, "Control ID should be set")

	// Verify second control (false)
	assert.Equal(t, "area_lighting", intervention.Controls[1].Kind)
	assert.NotNil(t, intervention.Controls[1].Result, "Result should not be nil")
	assert.False(t, *intervention.Controls[1].Result, "Result should be false")

	// Verify third control (nil/unchecked)
	assert.Equal(t, "safety_cells", intervention.Controls[2].Kind)
	assert.Nil(t, intervention.Controls[2].Result, "Result should be nil for unchecked")
}

func TestCreateInterventionService_CreateRepair_Success(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := &CreateInterventionService{DB: db}
	fixture := setupTestFixture(db)

	args := &CreateArgs{
		Type:      models.InterventionTypeRepair,
		Date:      "2024-01-15",
		Summary:   "Routine maintenance check",
		Signature: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA",
		PortalID:  fixture.portal.ID,
		UserID:    fixture.user.ID,
		UserName:  fixture.user.FullName(),
	}

	intervention, err := service.Create(args)

	require.NoError(t, err, "Create should succeed")
	assert.NotNil(t, intervention, "Intervention should not be nil")
	assert.NotZero(t, intervention.ID, "Intervention ID should be set")

	// Verify date parsing
	expectedDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedDate, intervention.Date, "Date should be parsed correctly")

	// Verify intervention type
	assert.Equal(t, models.InterventionTypeRepair, intervention.Type, "Type should be set to 'maintenance'")

	// Verify summary
	assert.NotNil(t, intervention.Summary, "Summary should not be nil")
	assert.Equal(t, "Routine maintenance check", *intervention.Summary, "Summary should match")

	// Verify signature
	assert.Equal(t, "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA", intervention.Signature, "Signature should match")

	// Verify no controls were created
	assert.Len(t, intervention.Controls, 0, "Should have 0 controls")
}

func TestCreateInterventionService_Create_InvalidDateFormat(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := &CreateInterventionService{DB: db}

	testCases := []struct {
		name        string
		dateString  string
		expectError string
	}{
		{
			name:        "invalid format",
			dateString:  "01/15/2024",
			expectError: "invalid date format",
		},
		{
			name:        "incomplete date",
			dateString:  "2024-01",
			expectError: "invalid date format",
		},
		{
			name:        "invalid date",
			dateString:  "2024-13-45",
			expectError: "invalid date format",
		},
		{
			name:        "empty date",
			dateString:  "",
			expectError: "invalid date format",
		},
		{
			name:        "text instead of date",
			dateString:  "not a date",
			expectError: "invalid date format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := &CreateArgs{
				Type:      models.InterventionTypeMaintenance,
				Date:      tc.dateString,
				Summary:   "Test summary",
				Signature: "test-signature",
			}

			intervention, err := service.Create(args)

			assert.Error(t, err, "Should return error for invalid date")
			assert.Nil(t, intervention, "Intervention should be nil on error")
			assert.Contains(t, err.Error(), tc.expectError, "Error message should mention invalid date")
		})
	}
}

func TestCreateInterventionService_Create_ControlResultParsing(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := &CreateInterventionService{DB: db}
	fixture := setupTestFixture(db)

	testCases := []struct {
		name           string
		resultString   string
		expectedResult *bool
	}{
		{
			name:           "true string",
			resultString:   "true",
			expectedResult: boolPtr(true),
		},
		{
			name:           "false string",
			resultString:   "false",
			expectedResult: boolPtr(false),
		},
		{
			name:           "1 as true",
			resultString:   "1",
			expectedResult: boolPtr(true),
		},
		{
			name:           "0 as false",
			resultString:   "0",
			expectedResult: boolPtr(false),
		},
		{
			name:           "empty string",
			resultString:   "",
			expectedResult: nil,
		},
		{
			name:           "invalid string",
			resultString:   "invalid",
			expectedResult: nil,
		},
		{
			name:           "T as true",
			resultString:   "T",
			expectedResult: boolPtr(true),
		},
		{
			name:           "F as false",
			resultString:   "F",
			expectedResult: boolPtr(false),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := &CreateArgs{
				Type:      models.InterventionTypeMaintenance,
				Date:      "2024-01-15",
				Summary:   "Test",
				Signature: "sig",
				PortalID:  fixture.portal.ID,
				UserID:    fixture.user.ID,
				UserName:  fixture.user.FullName(),
				Controls: []struct {
					Kind   string
					Result string
				}{
					{Kind: "test_control", Result: tc.resultString},
				},
			}

			intervention, err := service.Create(args)

			require.NoError(t, err, "Create should succeed")
			require.NotNil(t, intervention, "Intervention should not be nil")
			require.Len(t, intervention.Controls, 1, "Should have 1 control")

			control := intervention.Controls[0]
			if tc.expectedResult == nil {
				assert.Nil(t, control.Result, "Result should be nil")
			} else {
				require.NotNil(t, control.Result, "Result should not be nil")
				assert.Equal(t, *tc.expectedResult, *control.Result, "Result should match expected value")
			}
		})
	}
}

func TestCreateInterventionService_Create_EmptySummary(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := &CreateInterventionService{DB: db}
	fixture := setupTestFixture(db)

	args := &CreateArgs{
		Type:      models.InterventionTypeMaintenance,
		Date:      "2024-01-15",
		Summary:   "",
		Signature: "test-signature",
		PortalID:  fixture.portal.ID,
		UserID:    fixture.user.ID,
		UserName:  fixture.user.FullName(),
	}

	intervention, err := service.Create(args)

	require.NoError(t, err, "Create should succeed with empty summary")
	assert.NotNil(t, intervention, "Intervention should not be nil")
	assert.NotNil(t, intervention.Summary, "Summary pointer should not be nil")
	assert.Equal(t, "", *intervention.Summary, "Summary should be empty string")
}

func TestCreateInterventionService_Create_NoControls(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := &CreateInterventionService{DB: db}
	fixture := setupTestFixture(db)

	args := &CreateArgs{
		Type:      models.InterventionTypeMaintenance,
		Date:      "2024-01-15",
		Summary:   "Test without controls",
		Signature: "test-signature",
		PortalID:  fixture.portal.ID,
		UserID:    fixture.user.ID,
		UserName:  fixture.user.FullName(),
		Controls: []struct {
			Kind   string
			Result string
		}{},
	}

	intervention, err := service.Create(args)

	require.NoError(t, err, "Create should succeed without controls")
	assert.NotNil(t, intervention, "Intervention should not be nil")
	assert.Empty(t, intervention.Controls, "Controls should be empty")
}

func TestCreateInterventionService_Create_MultipleControls(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := &CreateInterventionService{DB: db}
	fixture := setupTestFixture(db)

	// Create all security controls
	securityControls := []struct {
		Kind   string
		Result string
	}{
		{Kind: "warning_lights", Result: "true"},
		{Kind: "area_lighting", Result: "true"},
		{Kind: "safety_cells", Result: "false"},
		{Kind: "pressure_bar", Result: "true"},
		{Kind: "floor_loop", Result: ""},
		{Kind: "force_limiter", Result: "true"},
		{Kind: "safety_springs", Result: "false"},
		{Kind: "floor_markings", Result: "true"},
	}

	args := &CreateArgs{
		Type:      models.InterventionTypeMaintenance,
		Date:      "2024-01-15",
		Summary:   "Full security check",
		Signature: "test-signature",
		PortalID:  fixture.portal.ID,
		UserID:    fixture.user.ID,
		UserName:  fixture.user.FullName(),
		Controls:  securityControls,
	}

	intervention, err := service.Create(args)

	require.NoError(t, err, "Create should succeed with multiple controls")
	assert.NotNil(t, intervention, "Intervention should not be nil")
	assert.Len(t, intervention.Controls, 8, "Should have 8 controls")

	// Verify all controls were saved
	for i, expected := range securityControls {
		assert.Equal(t, expected.Kind, intervention.Controls[i].Kind, "Control kind should match")
	}
}

func TestCreateInterventionService_Create_DatabasePersistence(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := &CreateInterventionService{DB: db}
	fixture := setupTestFixture(db)

	args := &CreateArgs{
		Type:      models.InterventionTypeMaintenance,
		Date:      "2024-01-15",
		Summary:   "Persistence test",
		Signature: "test-signature",
		PortalID:  fixture.portal.ID,
		UserID:    fixture.user.ID,
		UserName:  fixture.user.FullName(),
		Controls: []struct {
			Kind   string
			Result string
		}{
			{Kind: "warning_lights", Result: "true"},
		},
	}

	// Create intervention
	intervention, err := service.Create(args)
	require.NoError(t, err)
	require.NotNil(t, intervention)

	// Query database to verify it was saved
	var savedIntervention models.Intervention
	err = db.Preload("Controls").First(&savedIntervention, intervention.ID).Error
	require.NoError(t, err, "Should find saved intervention")

	// Verify the data matches
	assert.Equal(t, intervention.ID, savedIntervention.ID)
	assert.Equal(t, intervention.Date.Unix(), savedIntervention.Date.Unix())
	assert.Equal(t, *intervention.Summary, *savedIntervention.Summary)
	assert.Equal(t, intervention.Signature, savedIntervention.Signature)
	assert.Len(t, savedIntervention.Controls, 1)
	assert.Equal(t, "warning_lights", savedIntervention.Controls[0].Kind)
}

func TestCreateInterventionService_Create_TimestampsSet(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := &CreateInterventionService{DB: db}
	fixture := setupTestFixture(db)

	beforeCreate := time.Now()

	args := &CreateArgs{
		Type:      models.InterventionTypeMaintenance,
		Date:      "2024-01-15",
		Summary:   "Timestamp test",
		Signature: "test-signature",
		PortalID:  fixture.portal.ID,
		UserID:    fixture.user.ID,
		UserName:  fixture.user.FullName(),
	}

	intervention, err := service.Create(args)
	require.NoError(t, err)

	afterCreate := time.Now()

	// Verify timestamps were set by GORM
	assert.False(t, intervention.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, intervention.UpdatedAt.IsZero(), "UpdatedAt should be set")
	assert.True(t, intervention.CreatedAt.After(beforeCreate) || intervention.CreatedAt.Equal(beforeCreate))
	assert.True(t, intervention.CreatedAt.Before(afterCreate) || intervention.CreatedAt.Equal(afterCreate))
}

// Helper function to create *bool
func boolPtr(b bool) *bool {
	return &b
}
