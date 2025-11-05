# Test Factories

This package provides factory builders for creating test data in a Rails factory_bot style.

## Usage Examples

### Basic Usage

```go
import (
    "github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
)

func TestExample(t *testing.T) {
    db := tests.SetupTestDB(t)

    // Build without saving to database
    portal := factories.NewPortal().
        WithName("Main Gate").
        Build()

    // Create and save to database
    user := factories.NewUser().
        WithEmail("john@example.com").
        WithName("John", "Doe").
        Create(db)
}
```

### Portal Factory

```go
// Default portal
portal := factories.NewPortal().Create(db)

// Custom portal
portal := factories.NewPortal().
    WithName("North Entrance").
    WithAddress("456 North St", "75002", "Paris").
    WithContact("+33987654321", "north@example.com").
    Create(db)
```

### User Factory

```go
// Default user with password "password123"
user := factories.NewUser().Create(db)

// Custom user
user := factories.NewUser().
    WithEmail("admin@example.com").
    WithName("Admin", "User").
    WithPassword("secure-password").
    WithIsActive(true).
    Create(db)
```

### Intervention Factory

```go
// With related models
portal := factories.NewPortal().Create(db)
user := factories.NewUser().Create(db)

intervention := factories.NewIntervention().
    WithPortalModel(portal).
    WithUserModel(user).
    WithSummary("Routine inspection").
    Create(db)

// With controls
control1 := factories.NewWarningLightsControl(true).Build()
control2 := factories.NewAreaLightingControl(false).Build()

intervention := factories.NewIntervention().
    WithPortalModel(portal).
    WithUserModel(user).
    WithControls(control1, control2).
    Create(db)
```

### Control Factory

```go
// Using builder pattern
control := factories.NewControl().
    WithKind("warning_lights").
    WithResult(true).
    Create(db)

// Using helper functions
control := factories.NewWarningLightsControl(true).Create(db)
control := factories.NewSafetyCellsControl(false).Create(db)

// Nil result (unchecked)
control := factories.NewControl().
    WithKind("pressure_bar").
    WithNilResult().
    Create(db)
```

### QRCode Factory

```go
// Available QR code
qrCode := factories.NewQRCode().Create(db)

// Associated with portal
portal := factories.NewPortal().Create(db)
qrCode := factories.NewQRCode().
    WithPortalModel(portal).
    Create(db)

// Different statuses
qrCode := factories.NewQRCode().AsDamaged().Create(db)
qrCode := factories.NewQRCode().AsLost().Create(db)
```

### Complex Example

```go
func TestInterventionWithFullData(t *testing.T) {
    db := tests.SetupTestDB(t)

    // Create related records
    portal := factories.NewPortal().
        WithName("Test Portal").
        Create(db)

    user := factories.NewUser().
        WithEmail("tech@example.com").
        WithName("Tech", "Support").
        Create(db)

    // Create controls
    controls := []models.Control{
        *factories.NewWarningLightsControl(true).Build(),
        *factories.NewAreaLightingControl(true).Build(),
        *factories.NewSafetyCellsControl(false).Build(),
    }

    // Create intervention with all data
    intervention := factories.NewIntervention().
        WithPortalModel(portal).
        WithUserModel(user).
        WithSummary("Full security inspection").
        WithDate(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)).
        WithControls(controls...).
        Create(db)

    // Assert
    assert.NotZero(t, intervention.ID)
    assert.Equal(t, portal.ID, intervention.PortalID)
    assert.Equal(t, user.ID, intervention.UserID)
    assert.Len(t, intervention.Controls, 3)
}
```

## Benefits

- **Readable**: Clear, chainable API
- **Flexible**: Override any field as needed
- **Type-safe**: Compile-time checking
- **Consistent**: Default values for all fields
- **No external dependencies**: Pure Go implementation
