package factories

import (
	"time"

	"github.com/google/uuid"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"gorm.io/gorm"
)

type PortalBuilder struct {
	portal *models.Portal
}

func NewPortal() *PortalBuilder {
	return &PortalBuilder{
		portal: &models.Portal{
			UUID:              uuid.New().String(),
			InternalId:        "TEST-001",
			Name:              "Test Portal",
			AddressStreet:     "123 Test Street",
			AddressZipcode:    "75001",
			AddressCity:       "Paris",
			ContractorCompany: "Test Contractor Inc.",
			ContactPhone:      "+33123456789",
			ContactEmail:      "contact@test.com",
			InstallationDate:  time.Now(),
		},
	}
}

func (b *PortalBuilder) WithUUID(uuid string) *PortalBuilder {
	b.portal.UUID = uuid
	return b
}

func (b *PortalBuilder) WithInternalId(internalId string) *PortalBuilder {
	b.portal.InternalId = internalId
	return b
}

func (b *PortalBuilder) WithName(name string) *PortalBuilder {
	b.portal.Name = name
	return b
}

func (b *PortalBuilder) WithAddress(street, zipcode, city string) *PortalBuilder {
	b.portal.AddressStreet = street
	b.portal.AddressZipcode = zipcode
	b.portal.AddressCity = city
	return b
}

func (b *PortalBuilder) WithContractorCompany(company string) *PortalBuilder {
	b.portal.ContractorCompany = company
	return b
}

func (b *PortalBuilder) WithContact(phone, email string) *PortalBuilder {
	b.portal.ContactPhone = phone
	b.portal.ContactEmail = email
	return b
}

func (b *PortalBuilder) WithInstallationDate(date time.Time) *PortalBuilder {
	b.portal.InstallationDate = date
	return b
}

func (b *PortalBuilder) Build() *models.Portal {
	return b.portal
}

func (b *PortalBuilder) Create(db *gorm.DB) *models.Portal {
	db.Create(b.portal)
	return b.portal
}
