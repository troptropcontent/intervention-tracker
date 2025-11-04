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
			Kind:           "warning_lights",
			Result:         nil,
			InterventionID: 1,
		},
	}
}

func (b *ControlBuilder) WithKind(kind string) *ControlBuilder {
	b.control.Kind = kind
	return b
}

func (b *ControlBuilder) WithResult(result bool) *ControlBuilder {
	b.control.Result = &result
	return b
}

func (b *ControlBuilder) WithNilResult() *ControlBuilder {
	b.control.Result = nil
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

func NewSecurityControl(kind string, result *bool) *ControlBuilder {
	builder := NewControl().WithKind(kind)
	if result != nil {
		builder = builder.WithResult(*result)
	} else {
		builder = builder.WithNilResult()
	}
	return builder
}

func NewWarningLightsControl(result bool) *ControlBuilder {
	return NewControl().WithKind("warning_lights").WithResult(result)
}

func NewAreaLightingControl(result bool) *ControlBuilder {
	return NewControl().WithKind("area_lighting").WithResult(result)
}

func NewSafetyCellsControl(result bool) *ControlBuilder {
	return NewControl().WithKind("safety_cells").WithResult(result)
}

func NewPressureBarControl(result bool) *ControlBuilder {
	return NewControl().WithKind("pressure_bar").WithResult(result)
}

func NewFloorLoopControl(result bool) *ControlBuilder {
	return NewControl().WithKind("floor_loop").WithResult(result)
}

func NewForceLimiterControl(result bool) *ControlBuilder {
	return NewControl().WithKind("force_limiter").WithResult(result)
}

func NewSafetySpringsControl(result bool) *ControlBuilder {
	return NewControl().WithKind("safety_springs").WithResult(result)
}

func NewFloorMarkingsControl(result bool) *ControlBuilder {
	return NewControl().WithKind("floor_markings").WithResult(result)
}
