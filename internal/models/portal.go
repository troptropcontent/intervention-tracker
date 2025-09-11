package models

import (
	"time"

	"gorm.io/gorm"
)

type Portal struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	UUID              string         `json:"uuid" gorm:"type:uuid;unique;not null"`
	InternalId        string         `json:"internal_id" gorm:"type:string;unique"`
	Name              string         `json:"name" gorm:"not null"`
	AddressStreet     string         `json:"address_street" gorm:"not null"`
	AddressZipcode    string         `json:"address_zipcode" gorm:"size:10;not null"`
	AddressCity       string         `json:"address_city" gorm:"size:100;not null"`
	ContractorCompany string         `json:"contractor_company" gorm:"not null"`
	ContactPhone      string         `json:"contact_phone" gorm:"size:20;not null"`
	ContactEmail      string         `json:"contact_email"`
	InstallationDate  time.Time      `json:"installation_date" gorm:"not null"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	QRCodes []QRCode `json:"qr_codes,omitempty" gorm:"foreignKey:PortalID"`
}

func (Portal) TableName() string {
	return "portals"
}
