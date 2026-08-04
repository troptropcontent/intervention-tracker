package qrcodes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
)

func TestGenerateService_GenerateBatch_Success(t *testing.T) {
	db := tests.SetupTestDB(t)
	service, err := NewGenerateService(db)
	require.NoError(t, err)

	label := "Lot test"
	batch, err := service.GenerateBatch(10, &label)
	require.NoError(t, err)
	assert.NotZero(t, batch.ID)
	assert.Equal(t, 10, batch.Count)
	assert.Equal(t, "Lot test", *batch.Label)

	var codes []models.QRCode
	require.NoError(t, db.Where("batch_id = ?", batch.ID).Find(&codes).Error)
	require.Len(t, codes, 10)

	seenUUIDs := make(map[string]bool)
	for _, code := range codes {
		assert.Equal(t, models.QRCodeStatusAvailable, code.Status)
		require.NotEmpty(t, code.UUID)
		assert.False(t, seenUUIDs[code.UUID], "expected unique UUIDs")
		seenUUIDs[code.UUID] = true
	}
}

func TestGenerateService_GenerateBatch_WithoutLabel(t *testing.T) {
	db := tests.SetupTestDB(t)
	service, err := NewGenerateService(db)
	require.NoError(t, err)

	batch, err := service.GenerateBatch(3, nil)
	require.NoError(t, err)
	assert.Nil(t, batch.Label)
}

func TestGenerateService_GenerateBatch_CountTooLow(t *testing.T) {
	db := tests.SetupTestDB(t)
	service, err := NewGenerateService(db)
	require.NoError(t, err)

	_, err = service.GenerateBatch(0, nil)
	assert.Error(t, err)

	var batchCount int64
	db.Model(&models.QRCodeBatch{}).Count(&batchCount)
	assert.Zero(t, batchCount)
}

func TestGenerateService_GenerateBatch_CountTooHigh(t *testing.T) {
	db := tests.SetupTestDB(t)
	service, err := NewGenerateService(db)
	require.NoError(t, err)

	_, err = service.GenerateBatch(MaxBatchCount+1, nil)
	assert.Error(t, err)

	var batchCount int64
	db.Model(&models.QRCodeBatch{}).Count(&batchCount)
	assert.Zero(t, batchCount)
}

func TestNewGenerateService_NilDB(t *testing.T) {
	_, err := NewGenerateService(nil)
	assert.Error(t, err)
}
