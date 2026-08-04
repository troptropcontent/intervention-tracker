package models

import "time"

type QRCodeBatch struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Label     *string   `json:"label" gorm:"type:varchar(255)"`
	Count     int       `json:"count" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	QRCodes []QRCode `json:"qr_codes,omitempty" gorm:"foreignKey:BatchID"`
}

func (QRCodeBatch) TableName() string {
	return "qr_code_batches"
}
