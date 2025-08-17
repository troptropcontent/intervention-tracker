package models

import (
	"time"

	"gorm.io/gorm"
)

type QRCodeStatus string

const (
	QRCodeStatusAvailable  QRCodeStatus = "available"
	QRCodeStatusAssociated QRCodeStatus = "associated"
	QRCodeStatusDamaged    QRCodeStatus = "damaged"
	QRCodeStatusLost       QRCodeStatus = "lost"
)

type QRCode struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UUID         string         `json:"uuid" gorm:"type:uuid;unique;not null"`
	PortalID     *uint          `json:"portal_id" gorm:"index"`
	Status       QRCodeStatus   `json:"status" gorm:"type:varchar(20);default:available"`
	AssociatedAt *time.Time     `json:"associated_at"`
	GeneratedAt  time.Time      `json:"generated_at" gorm:"autoCreateTime"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Portal *Portal `json:"portal,omitempty" gorm:"foreignKey:PortalID"`
}

func (QRCode) TableName() string {
	return "qr_codes"
}
