package models

import (
	"time"

	"gorm.io/gorm"
)

type ControlKind string

const (
	ControlKindSecurity ControlKind = "security"
	ControlKindOther    ControlKind = "other"
)

var ControlTypesByKind = map[ControlKind][]string{
	ControlKindSecurity: {
		"warning_lights",
		"area_lighting",
		"safety_cells",
		"pressure_bar",
		"floor_loop",
		"force_limiter",
		"safety_springs",
		"floor_markings",
	},
	ControlKindOther: {
		"apron_condition",
		"horizontal_rails",
		"vertical_rails",
		"roller_condition",
		"drive_system",
		"limit_switches",
		"control_devices",
		"control_panel",
		"manual_override",
	},
}

type ControlResult *bool

type Intervention struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Date      time.Time      `json:"date" gorm:"not null"`
	Summary   *string        `json:"summary"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	UserName  string         `json:"user_name" gorm:"not null"`
	PortalID  uint           `json:"portal_id" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Portal   Portal    `json:"portal,omitempty" gorm:"foreignKey:PortalID"`
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Controls []Control `json:"controls,omitempty" gorm:"foreignKey:intervention_id"`
}

type Control struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Kind           string         `json:"kind" gorm:"type:varchar(20);not null"`
	Result         ControlResult  `json:"result"`
	InterventionID uint           `json:"intervention_id" gorm:"not null"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Intervention Intervention `json:"intervention,omitempty"`
}

func (Intervention) TableName() string {
	return "interventions"
}

func (Control) TableName() string {
	return "controls"
}
