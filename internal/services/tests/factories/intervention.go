package factories

import (
	"time"

	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"gorm.io/gorm"
)

type InterventionBuilder struct {
	intervention *models.Intervention
	controls     []models.Control
}

func NewIntervention() *InterventionBuilder {
	summary := "Test intervention summary"
	return &InterventionBuilder{
		intervention: &models.Intervention{
			Date:      time.Now(),
			Summary:   &summary,
			UserID:    1,
			UserName:  "John Doe",
			PortalID:  1,
			Signature: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA",
		},
		controls: []models.Control{},
	}
}

func (b *InterventionBuilder) WithDate(date time.Time) *InterventionBuilder {
	b.intervention.Date = date
	return b
}

func (b *InterventionBuilder) WithSummary(summary string) *InterventionBuilder {
	b.intervention.Summary = &summary
	return b
}

func (b *InterventionBuilder) WithUser(userID uint, userName string) *InterventionBuilder {
	b.intervention.UserID = userID
	b.intervention.UserName = userName
	return b
}

func (b *InterventionBuilder) WithPortal(portalID uint) *InterventionBuilder {
	b.intervention.PortalID = portalID
	return b
}

func (b *InterventionBuilder) WithPortalModel(portal *models.Portal) *InterventionBuilder {
	b.intervention.PortalID = portal.ID
	b.intervention.Portal = *portal
	return b
}

func (b *InterventionBuilder) WithUserModel(user *models.User) *InterventionBuilder {
	b.intervention.UserID = user.ID
	b.intervention.UserName = user.FullName()
	b.intervention.User = *user
	return b
}

func (b *InterventionBuilder) WithSignature(signature string) *InterventionBuilder {
	b.intervention.Signature = signature
	return b
}

func (b *InterventionBuilder) WithControls(controls ...models.Control) *InterventionBuilder {
	b.controls = append(b.controls, controls...)
	return b
}

func (b *InterventionBuilder) Build() *models.Intervention {
	if len(b.controls) > 0 {
		b.intervention.Controls = b.controls
	}
	return b.intervention
}

func (b *InterventionBuilder) Create(db *gorm.DB) *models.Intervention {
	if len(b.controls) > 0 {
		b.intervention.Controls = b.controls
	}
	db.Create(b.intervention)
	return b.intervention
}
