package factories_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
)

// TestFactoryExample demonstrates how to use the factory pattern
func TestFactoryExample(t *testing.T) {
	db := tests.SetupTestDB(t)

	// Create a portal using the factory
	portal := factories.NewPortal().
		WithName("Main Entrance").
		WithAddress("123 Main St", "75001", "Paris").
		WithContact("+33123456789", "contact@example.com").
		Create(db)

	assert.NotZero(t, portal.ID)
	assert.Equal(t, "Main Entrance", portal.Name)
	assert.Equal(t, "75001", portal.AddressZipcode)
}

// TestFactoryWithRelationships demonstrates creating related records
func TestFactoryWithRelationships(t *testing.T) {
	db := tests.SetupTestDB(t)

	// Create portal and user first
	portal := factories.NewPortal().
		WithName("Test Portal").
		Create(db)

	user := factories.NewUser().
		WithEmail("tech@example.com").
		WithName("Tech", "Support").
		Create(db)

	// Create intervention with relationships
	intervention := factories.NewIntervention().
		WithPortalModel(portal).
		WithUserModel(user).
		WithSummary("Routine inspection").
		WithDate(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)).
		Create(db)

	require.NotZero(t, intervention.ID)
	assert.Equal(t, portal.ID, intervention.PortalID)
	assert.Equal(t, user.ID, intervention.UserID)
	assert.Equal(t, "Tech Support", intervention.UserName)
}

// TestFactoryWithControls demonstrates creating interventions with controls
func TestFactoryWithControls(t *testing.T) {
	db := tests.SetupTestDB(t)

	portal := factories.NewPortal().Create(db)
	user := factories.NewUser().Create(db)

	// Create controls using helper functions
	control1 := *factories.NewWarningLightsControl(true).Build()
	control2 := *factories.NewAreaLightingControl(false).Build()
	control3 := *factories.NewControl().WithKind("safety_cells").WithNilResult().Build()

	// Create intervention with controls
	intervention := factories.NewIntervention().
		WithPortalModel(portal).
		WithUserModel(user).
		WithControls(control1, control2, control3).
		Create(db)

	// Verify controls were created
	assert.Len(t, intervention.Controls, 3)
	assert.Equal(t, "warning_lights", intervention.Controls[0].Kind)
	assert.NotNil(t, intervention.Controls[0].Result)
	assert.True(t, *intervention.Controls[0].Result)

	assert.Equal(t, "area_lighting", intervention.Controls[1].Kind)
	assert.NotNil(t, intervention.Controls[1].Result)
	assert.False(t, *intervention.Controls[1].Result)

	assert.Equal(t, "safety_cells", intervention.Controls[2].Kind)
	assert.Nil(t, intervention.Controls[2].Result)
}

// TestFactoryQRCode demonstrates QR code factory usage
func TestFactoryQRCode(t *testing.T) {
	db := tests.SetupTestDB(t)

	// Create an available QR code
	qrCode1 := factories.NewQRCode().
		AsAvailable().
		Create(db)

	assert.Equal(t, "available", string(qrCode1.Status))
	assert.Nil(t, qrCode1.PortalID)

	// Create an associated QR code
	portal := factories.NewPortal().Create(db)
	qrCode2 := factories.NewQRCode().
		WithPortalModel(portal).
		Create(db)

	assert.Equal(t, "associated", string(qrCode2.Status))
	require.NotNil(t, qrCode2.PortalID)
	assert.Equal(t, portal.ID, *qrCode2.PortalID)
	assert.NotNil(t, qrCode2.AssociatedAt)
}

// TestFactoryBuildVsCreate demonstrates the difference between Build and Create
func TestFactoryBuildVsCreate(t *testing.T) {
	db := tests.SetupTestDB(t)

	// Build creates the object but doesn't save to DB
	portal1 := factories.NewPortal().
		WithName("Not Saved").
		Build()
	assert.Zero(t, portal1.ID) // ID not set yet

	// Create saves to DB
	portal2 := factories.NewPortal().
		WithName("Saved").
		Create(db)
	assert.NotZero(t, portal2.ID) // ID is set by database
}
