package factories

import (
	"time"

	"github.com/google/uuid"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"gorm.io/gorm"
)

type QRCodeBuilder struct {
	qrCode *models.QRCode
}

func NewQRCode() *QRCodeBuilder {
	return &QRCodeBuilder{
		qrCode: &models.QRCode{
			UUID:        uuid.New().String(),
			Status:      models.QRCodeStatusAvailable,
			GeneratedAt: time.Now(),
		},
	}
}

func (b *QRCodeBuilder) WithUUID(uuid string) *QRCodeBuilder {
	b.qrCode.UUID = uuid
	return b
}

func (b *QRCodeBuilder) WithStatus(status models.QRCodeStatus) *QRCodeBuilder {
	b.qrCode.Status = status
	return b
}

func (b *QRCodeBuilder) WithPortal(portalID uint) *QRCodeBuilder {
	b.qrCode.PortalID = &portalID
	b.qrCode.Status = models.QRCodeStatusAssociated
	now := time.Now()
	b.qrCode.AssociatedAt = &now
	return b
}

func (b *QRCodeBuilder) WithPortalModel(portal *models.Portal) *QRCodeBuilder {
	b.qrCode.PortalID = &portal.ID
	b.qrCode.Portal = portal
	b.qrCode.Status = models.QRCodeStatusAssociated
	now := time.Now()
	b.qrCode.AssociatedAt = &now
	return b
}

func (b *QRCodeBuilder) AsAvailable() *QRCodeBuilder {
	b.qrCode.Status = models.QRCodeStatusAvailable
	b.qrCode.PortalID = nil
	b.qrCode.AssociatedAt = nil
	return b
}

func (b *QRCodeBuilder) AsDamaged() *QRCodeBuilder {
	b.qrCode.Status = models.QRCodeStatusDamaged
	return b
}

func (b *QRCodeBuilder) AsLost() *QRCodeBuilder {
	b.qrCode.Status = models.QRCodeStatusLost
	return b
}

func (b *QRCodeBuilder) WithGeneratedAt(generatedAt time.Time) *QRCodeBuilder {
	b.qrCode.GeneratedAt = generatedAt
	return b
}

func (b *QRCodeBuilder) Build() *models.QRCode {
	return b.qrCode
}

func (b *QRCodeBuilder) Create(db *gorm.DB) *models.QRCode {
	db.Create(b.qrCode)
	return b.qrCode
}
