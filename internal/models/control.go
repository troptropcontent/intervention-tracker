package models

import (
	"slices"
	"time"

	"gorm.io/gorm"
)

type ControlKind string

const (
	ControlKindWarningLights   ControlKind = "warning_lights"
	ControlKindAreaLighting    ControlKind = "area_lighting"
	ControlKindSafetyCells     ControlKind = "safety_cells"
	ControlKindPressureBar     ControlKind = "pressure_bar"
	ControlKindFloorLoop       ControlKind = "floor_loop"
	ControlKindForceLimiter    ControlKind = "force_limiter"
	ControlKindSafetySprings   ControlKind = "safety_springs"
	ControlKindFloorMarkings   ControlKind = "floor_markings"
	ControlKindApronCondition  ControlKind = "apron_condition"
	ControlKindHorizontalRails ControlKind = "horizontal_rails"
	ControlKindVerticalRails   ControlKind = "vertical_rails"
	ControlKindRollerCondition ControlKind = "roller_condition"
	ControlKindDriveSystem     ControlKind = "drive_system"
	ControlKindLimitSwitches   ControlKind = "limit_switches"
	ControlKindControlDevices  ControlKind = "control_devices"
	ControlKindControlPanel    ControlKind = "control_panel"
	ControlKindManualOverride  ControlKind = "manual_override"
)

var ControlKinds = []ControlKind{
	ControlKindWarningLights,
	ControlKindAreaLighting,
	ControlKindSafetyCells,
	ControlKindPressureBar,
	ControlKindFloorLoop,
	ControlKindForceLimiter,
	ControlKindSafetySprings,
	ControlKindFloorMarkings,
	ControlKindApronCondition,
	ControlKindHorizontalRails,
	ControlKindVerticalRails,
	ControlKindRollerCondition,
	ControlKindDriveSystem,
	ControlKindLimitSwitches,
	ControlKindControlDevices,
	ControlKindControlPanel,
	ControlKindManualOverride,
}

func (c *ControlKind) IsValid() bool {
	return slices.Contains(ControlKinds, *c)
}

type ControlResult string

const (
	ControlResultCompliant      ControlResult = "compliant"
	ControlResultNonCompliant   ControlResult = "non_compliant"
	ControlResultNeedsAttention ControlResult = "needs_attention"
	ControlResultSkipped        ControlResult = "skipped"
)

func (c ControlResult) IsValid() bool {
	switch c {
	case ControlResultCompliant, ControlResultNonCompliant, ControlResultNeedsAttention, ControlResultSkipped:
		return true
	}
	return false
}

type Control struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Kind           ControlKind    `json:"kind" gorm:"type:varchar(20);not null"`
	Result         ControlResult  `json:"result"`
	InterventionID uint           `json:"intervention_id" gorm:"not null"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Intervention Intervention `json:"intervention,omitempty"`
}

func (Control) TableName() string {
	return "controls"
}
