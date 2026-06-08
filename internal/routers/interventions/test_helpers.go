package interventions

import (
	"fmt"
	"regexp"

	"github.com/troptropcontent/qr_code_maintenance/internal/models"
)

// InterventionFormBuilder helps build test form data for intervention creation
type InterventionFormBuilder struct {
	fields map[string]string
	files  map[string][]byte
}

// NewInterventionFormBuilder creates a new builder with sensible defaults
func NewCreateInterventionFormBuilder(portalID string) *InterventionFormBuilder {
	builder := InterventionFormBuilder{
		fields: map[string]string{},
		files:  map[string][]byte{},
	}
	return builder.WithPortalId(portalID).WithType("maintenance").WithDate("2024-01-15").WithSummary("Test intervention").WithSignature("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA").WithAllCompliantControls()
}

// WithType sets the intervention type
func (b *InterventionFormBuilder) WithType(interventionType string) *InterventionFormBuilder {
	b.fields["type"] = interventionType
	return b
}

// WithDate sets the intervention date
func (b *InterventionFormBuilder) WithPortalId(portalID string) *InterventionFormBuilder {
	b.fields["portal_id"] = portalID
	return b
}

// WithDate sets the intervention date
func (b *InterventionFormBuilder) WithDate(date string) *InterventionFormBuilder {
	b.fields["date"] = date
	return b
}

// WithSummary sets the summary
func (b *InterventionFormBuilder) WithSummary(summary string) *InterventionFormBuilder {
	b.fields["summary"] = summary
	return b
}

// WithSignature sets the signature
func (b *InterventionFormBuilder) WithSignature(signature string) *InterventionFormBuilder {
	b.fields["signature"] = signature
	return b
}

// WithTimeSpent sets the time spent in hours
func (b *InterventionFormBuilder) WithTimeSpent(hours string) *InterventionFormBuilder {
	b.fields["time_spent_hours"] = hours
	return b
}

// WithControl adds a single control to the form
func (b *InterventionFormBuilder) WithControl(index int, kind string, result string) *InterventionFormBuilder {
	b.fields[fmt.Sprintf("controls[%d].kind", index)] = kind
	b.fields[fmt.Sprintf("controls[%d].result", index)] = result
	return b
}

// WithAllCompliantControls adds all 17 controls with compliant result
func (b *InterventionFormBuilder) WithAllCompliantControls() *InterventionFormBuilder {
	for i, kind := range models.ControlKinds {
		b.WithControl(i, string(kind), string(models.ControlResultCompliant))
	}
	return b
}

// WithAllControls adds all 17 controls with the specified result
func (b *InterventionFormBuilder) WithAllControls(result models.ControlResult) *InterventionFormBuilder {
	for i, kind := range models.ControlKinds {
		b.WithControl(i, string(kind), string(result))
	}
	return b
}

// WithControlResults adds all 17 controls with custom results
// If fewer than 17 results are provided, remaining controls are set to compliant
func (b *InterventionFormBuilder) WithControlResults(results []models.ControlResult) *InterventionFormBuilder {
	for i, kind := range models.ControlKinds {
		result := models.ControlResultCompliant
		if i < len(results) {
			result = results[i]
		}
		b.WithControl(i, string(kind), string(result))
	}
	return b
}

// WithOutControls remove all controls
func (b *InterventionFormBuilder) WithOutControls() *InterventionFormBuilder {
	regex := regexp.MustCompile(`^controls\[\d+\].(kind|result)$`)
	for key := range b.fields {
		if regex.MatchString(key) {
			delete(b.fields, key)
		}
	}
	return b
}

// WithPhoto adds a photo to the form
func (b *InterventionFormBuilder) WithPhoto(index int, name string, data []byte) *InterventionFormBuilder {
	b.fields[fmt.Sprintf("photos[%d].name", index)] = name
	if data != nil {
		b.files[fmt.Sprintf("photos[%d].file", index)] = data
	}
	return b
}

// WithPhotos adds multiple photos with default data
func (b *InterventionFormBuilder) WithPhotos(names ...string) *InterventionFormBuilder {
	for i, name := range names {
		b.WithPhoto(i, name, []byte(fmt.Sprintf("fake-image-data-%d", i+1)))
	}
	return b
}

// WithKeyValue adds an arbitrary key/value to the fields
func (b *InterventionFormBuilder) WithKeyValue(key string, value string) *InterventionFormBuilder {
	b.fields[key] = value

	return b
}

// Build returns the fields and files maps
func (b *InterventionFormBuilder) Build() (map[string]string, map[string][]byte) {
	return b.fields, b.files
}

// BuildFields returns only the fields map (for tests without files)
func (b *InterventionFormBuilder) BuildFields() map[string]string {
	return b.fields
}
