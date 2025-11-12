package interventions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/jobs"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
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

func setupService(db *gorm.DB) *CreateInterventionService {
	service, _ := NewCreateInterventionService(db, &tests.MockStorageService{}, &tests.MockEmailService{}, jobs.NewSyncJobRunner())
	return service
}

func TestCreateInterventionService_Create_Success(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := setupService(db)
	fixture := setupTestFixture(db)

	args := &CreateArgs{
		Type:      models.InterventionTypeMaintenance,
		Date:      tests.MustParseDate(t, "2024-01-15"),
		Summary:   "Routine maintenance check",
		Signature: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA",
		PortalID:  fixture.portal.ID,
		UserID:    fixture.user.ID,
		UserName:  fixture.user.FullName(),
		Controls: []ControlData{
			{Kind: models.ControlKindApronCondition, Result: models.ControlResultCompliant},
			{Kind: models.ControlKindControlDevices, Result: models.ControlResultNeedsAttention},
			{Kind: models.ControlKindAreaLighting, Result: models.ControlResultNonCompliant},
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

	// Verify first control (compliant)
	assert.Equal(t, models.ControlKindApronCondition, intervention.Controls[0].Kind)
	assert.Equal(t, models.ControlResultCompliant, intervention.Controls[0].Result, "Result should be compliant")
	assert.NotZero(t, intervention.Controls[0].ID, "Control ID should be set")

	// Verify second control (needs attention)
	assert.Equal(t, models.ControlKindControlDevices, intervention.Controls[1].Kind)
	assert.Equal(t, models.ControlResultNeedsAttention, intervention.Controls[1].Result, "Result should be needs_attention")

	// Verify third control (non-compliant)
	assert.Equal(t, models.ControlKindAreaLighting, intervention.Controls[2].Kind)
	assert.Equal(t, models.ControlResultNonCompliant, intervention.Controls[2].Result, "Result should be non_compliant")
}

func TestCreateInterventionService_CreateRepair_Success(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := setupService(db)
	fixture := setupTestFixture(db)

	args := &CreateArgs{
		Type:      models.InterventionTypeRepair,
		Date:      tests.MustParseDate(t, "2024-01-15"),
		TimeSpent: utils.Ptr(120),
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

// Note: Date validation and control result parsing now happen at the handler/parsing level
// before CreateArgs is constructed, so we don't test those conversions at the service level

func TestCreateInterventionService_Create_ControlResultTypes(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := setupService(db)
	fixture := setupTestFixture(db)

	testCases := []struct {
		name           string
		result         models.ControlResult
		expectedResult models.ControlResult
	}{
		{
			name:           "compliant",
			result:         models.ControlResultCompliant,
			expectedResult: models.ControlResultCompliant,
		},
		{
			name:           "non_compliant",
			result:         models.ControlResultNonCompliant,
			expectedResult: models.ControlResultNonCompliant,
		},
		{
			name:           "needs_attention",
			result:         models.ControlResultNeedsAttention,
			expectedResult: models.ControlResultNeedsAttention,
		},
		{
			name:           "skipped",
			result:         models.ControlResultSkipped,
			expectedResult: models.ControlResultSkipped,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args := &CreateArgs{
				Type:      models.InterventionTypeMaintenance,
				Date:      tests.MustParseDate(t, "2024-01-15"),
				Summary:   "Test",
				Signature: "sig",
				PortalID:  fixture.portal.ID,
				UserID:    fixture.user.ID,
				UserName:  fixture.user.FullName(),
				Controls: []ControlData{
					{Kind: models.ControlKindWarningLights, Result: tc.result},
				},
			}

			intervention, err := service.Create(args)

			require.NoError(t, err, "Create should succeed")
			require.NotNil(t, intervention, "Intervention should not be nil")
			require.Len(t, intervention.Controls, 1, "Should have 1 control")

			control := intervention.Controls[0]
			assert.Equal(t, tc.expectedResult, control.Result, "Result should match expected value")
		})
	}
}

func TestCreateInterventionService_Create_EmptySummary(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := setupService(db)
	fixture := setupTestFixture(db)

	args := &CreateArgs{
		Type:      models.InterventionTypeMaintenance,
		Date:      tests.MustParseDate(t, "2024-01-15"),
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
	service := setupService(db)
	fixture := setupTestFixture(db)

	args := &CreateArgs{
		Type:      models.InterventionTypeMaintenance,
		Date:      tests.MustParseDate(t, "2024-01-15"),
		Summary:   "Test without controls",
		Signature: "test-signature",
		PortalID:  fixture.portal.ID,
		UserID:    fixture.user.ID,
		UserName:  fixture.user.FullName(),
		Controls:  []ControlData{},
	}

	intervention, err := service.Create(args)

	require.NoError(t, err, "Create should succeed without controls")
	assert.NotNil(t, intervention, "Intervention should not be nil")
	assert.Empty(t, intervention.Controls, "Controls should be empty")
}

func TestCreateInterventionService_Create_MultipleControls(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := setupService(db)
	fixture := setupTestFixture(db)

	// Create all security controls
	securityControls := []ControlData{
		{Kind: models.ControlKindWarningLights, Result: models.ControlResultCompliant},
		{Kind: models.ControlKindAreaLighting, Result: models.ControlResultCompliant},
		{Kind: models.ControlKindSafetyCells, Result: models.ControlResultNonCompliant},
		{Kind: models.ControlKindPressureBar, Result: models.ControlResultCompliant},
		{Kind: models.ControlKindFloorLoop, Result: models.ControlResultSkipped},
		{Kind: models.ControlKindForceLimiter, Result: models.ControlResultCompliant},
		{Kind: models.ControlKindSafetySprings, Result: models.ControlResultNonCompliant},
		{Kind: models.ControlKindFloorMarkings, Result: models.ControlResultCompliant},
	}

	args := &CreateArgs{
		Type:      models.InterventionTypeMaintenance,
		Date:      tests.MustParseDate(t, "2024-01-15"),
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
		assert.Equal(t, expected.Result, intervention.Controls[i].Result, "Control result should match")
	}
}

func TestCreateInterventionService_Create_DatabasePersistence(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := setupService(db)
	fixture := setupTestFixture(db)

	args := &CreateArgs{
		Type:      models.InterventionTypeMaintenance,
		Date:      tests.MustParseDate(t, "2024-01-15"),
		Summary:   "Persistence test",
		Signature: "test-signature",
		PortalID:  fixture.portal.ID,
		UserID:    fixture.user.ID,
		UserName:  fixture.user.FullName(),
		Controls: []ControlData{
			{Kind: models.ControlKindWarningLights, Result: models.ControlResultCompliant},
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
	assert.Equal(t, models.ControlKindWarningLights, savedIntervention.Controls[0].Kind)
	assert.Equal(t, models.ControlResultCompliant, savedIntervention.Controls[0].Result)
}

func TestCreateInterventionService_Create_TimestampsSet(t *testing.T) {
	db := tests.SetupTestDB(t)
	service := setupService(db)
	fixture := setupTestFixture(db)

	beforeCreate := time.Now()

	args := &CreateArgs{
		Type:      models.InterventionTypeMaintenance,
		Date:      tests.MustParseDate(t, "2024-01-15"),
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
