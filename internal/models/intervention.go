package models

import (
	"time"

	"gorm.io/gorm"
)

type InterventionType string

const (
	InterventionTypeMaintenance InterventionType = "maintenance"
	InterventionTypeRepair      InterventionType = "repair"
)

func (t InterventionType) IsValid() bool {
	switch t {
	case InterventionTypeMaintenance, InterventionTypeRepair:
		return true
	}
	return false
}

type Intervention struct {
	ID        uint             `json:"id" gorm:"primaryKey"`
	Date      time.Time        `json:"date" gorm:"not null"`
	Type      InterventionType `json:"type" gorm:"not null"`
	Summary   *string          `json:"summary"`
	UserID    uint             `json:"user_id" gorm:"not null"`
	UserName  string           `json:"user_name" gorm:"not null"`
	PortalID  uint             `json:"portal_id" gorm:"not null"`
	Signature string           `json:"signature" gorm:"type:text"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	DeletedAt gorm.DeletedAt   `json:"-" gorm:"index"`

	// Relationships
	Portal      Portal       `json:"portal,omitempty" gorm:"foreignKey:PortalID"`
	User        User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Controls    []Control    `json:"controls,omitempty" gorm:"foreignKey:intervention_id"`
	Attachments []Attachment `gorm:"polymorphic:Holder;"`
}

func (Intervention) TableName() string {
	return "interventions"
}
