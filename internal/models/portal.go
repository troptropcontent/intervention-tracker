package models

import (
	"time"
)

type Portal struct {
	ID                int       `json:"id" db:"id"`
	UUID              string    `json:"uuid" db:"uuid"`
	Name              string    `json:"name" db:"name"`
	AddressStreet     string    `json:"address_street" db:"address_street"`
	AddressZipcode    string    `json:"address_zipcode" db:"address_zipcode"`
	AddressCity       string    `json:"address_city" db:"address_city"`
	ContractorCompany string    `json:"contractor_company" db:"contractor_company"`
	ContactPhone      string    `json:"contact_phone" db:"contact_phone"`
	ContactEmail      string    `json:"contact_email" db:"contact_email"`
	InstallationDate  time.Time `json:"installation_date" db:"installation_date"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
