package factories

import (
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"gorm.io/gorm"
)

type ControlBuilder struct {
	control *models.Control
}

func NewControl() *ControlBuilder {
	return &ControlBuilder{
		control: &models.Control{
			Kind:           models.ControlKindWarningLights,
			Result:         models.ControlResultSkipped,
			InterventionID: 1,
		},
	}
}

func (b *ControlBuilder) WithKind(kind models.ControlKind) *ControlBuilder {
	b.control.Kind = kind
	return b
}

func (b *ControlBuilder) WithResult(result models.ControlResult) *ControlBuilder {
	b.control.Result = result
	return b
}

func (b *ControlBuilder) WithCompliantResult() *ControlBuilder {
	b.control.Result = models.ControlResultCompliant
	return b
}

func (b *ControlBuilder) WithNonCompliantResult() *ControlBuilder {
	b.control.Result = models.ControlResultNonCompliant
	return b
}

func (b *ControlBuilder) WithNeedsAttentionResult() *ControlBuilder {
	b.control.Result = models.ControlResultNeedsAttention
	return b
}

func (b *ControlBuilder) WithSkippedResult() *ControlBuilder {
	b.control.Result = models.ControlResultSkipped
	return b
}

func (b *ControlBuilder) WithIntervention(interventionID uint) *ControlBuilder {
	b.control.InterventionID = interventionID
	return b
}

func (b *ControlBuilder) Build() *models.Control {
	return b.control
}

func (b *ControlBuilder) Create(db *gorm.DB) *models.Control {
	db.Create(b.control)
	return b.control
}

// Helper functions for common control types

func NewWarningLightsControl(result models.ControlResult) *ControlBuilder {
	return NewControl().WithKind(models.ControlKindWarningLights).WithResult(result)
}

func NewAreaLightingControl(result models.ControlResult) *ControlBuilder {
	return NewControl().WithKind(models.ControlKindAreaLighting).WithResult(result)
}

func NewSafetyCellsControl(result models.ControlResult) *ControlBuilder {
	return NewControl().WithKind(models.ControlKindSafetyCells).WithResult(result)
}

func NewPressureBarControl(result models.ControlResult) *ControlBuilder {
	return NewControl().WithKind(models.ControlKindPressureBar).WithResult(result)
}

func NewFloorLoopControl(result models.ControlResult) *ControlBuilder {
	return NewControl().WithKind(models.ControlKindFloorLoop).WithResult(result)
}

func NewForceLimiterControl(result models.ControlResult) *ControlBuilder {
	return NewControl().WithKind(models.ControlKindForceLimiter).WithResult(result)
}

func NewSafetySpringsControl(result models.ControlResult) *ControlBuilder {
	return NewControl().WithKind(models.ControlKindSafetySprings).WithResult(result)
}

func NewFloorMarkingsControl(result models.ControlResult) *ControlBuilder {
	return NewControl().WithKind(models.ControlKindFloorMarkings).WithResult(result)
}
