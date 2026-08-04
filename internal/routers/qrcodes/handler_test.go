package qrcodes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/factories"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/tests/mocks"
)

func TestGetQRCodes_NoFilter_ReturnsAll(t *testing.T) {
	deps, db := setup(t)

	factories.NewQRCode().AsAvailable().Create(db)
	factories.NewQRCode().AsDamaged().Create(db)

	c := tests.NewContext(http.MethodGet, "/").Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := GetQRCodes(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "QR Codes")
}

func TestGetQRCodes_StatusFilter(t *testing.T) {
	deps, db := setup(t)

	available := factories.NewQRCode().AsAvailable().Create(db)
	damaged := factories.NewQRCode().AsDamaged().Create(db)

	c := tests.NewContext(http.MethodGet, "/").WithQueryParams(map[string]string{"status": "damaged"}).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := GetQRCodes(deps)
	err := handler(c)

	require.NoError(t, err)
	body := rec.Body.String()
	assert.Contains(t, body, damaged.UUID)
	assert.NotContains(t, body, available.UUID)
}

func TestGetQRCodes_Pagination(t *testing.T) {
	deps, db := setup(t)

	for i := 0; i < PAGE_LIMIT+5; i++ {
		factories.NewQRCode().AsAvailable().Create(db)
	}

	c := tests.NewContext(http.MethodGet, "/").WithQueryParams(map[string]string{"page": "2"}).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := GetQRCodes(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateBatch_Success(t *testing.T) {
	deps, db := setup(t)

	c := tests.NewContext(http.MethodPost, "/").WithFormData(map[string]string{"count": "5", "label": "Lot test"}).Build()
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := CreateBatch(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/admin/qr-codes/batches/")

	var count int64
	db.Model(&models.QRCode{}).Count(&count)
	assert.Equal(t, int64(5), count)
}

func TestCreateBatch_InvalidCount(t *testing.T) {
	deps, db := setup(t)

	c := tests.NewContext(http.MethodPost, "/").WithFormData(map[string]string{"count": "0"}).Build()

	handler := CreateBatch(deps)
	err := handler(c)

	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)

	var count int64
	db.Model(&models.QRCodeBatch{}).Count(&count)
	assert.Zero(t, count)
}

func TestGetBatch_Success(t *testing.T) {
	deps, db := setup(t)

	batch := factories.NewQRCodeBatch().WithCount(2).Create(db)
	factories.NewQRCode().AsAvailable().WithBatch(batch.ID).Create(db)
	factories.NewQRCode().AsAvailable().WithBatch(batch.ID).Create(db)

	c := tests.NewContext(http.MethodGet, "/").Build()
	c.SetParamNames("batch_id")
	c.SetParamValues(strconv.Itoa(int(batch.ID)))
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := GetBatch(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetBatch_NotFound(t *testing.T) {
	deps, _ := setup(t)

	c := tests.NewContext(http.MethodGet, "/").Build()
	c.SetParamNames("batch_id")
	c.SetParamValues("999")

	handler := GetBatch(deps)
	err := handler(c)

	require.Error(t, err)
}

func TestDownloadBatchPDF_Success(t *testing.T) {
	deps, db := setup(t)

	batch := factories.NewQRCodeBatch().WithCount(2).Create(db)
	factories.NewQRCode().AsAvailable().WithBatch(batch.ID).Create(db)
	factories.NewQRCode().AsAvailable().WithBatch(batch.ID).Create(db)

	mockGotemberg := &mocks.GotenbergService{PDFContent: []byte("%PDF-1.4 fake")}

	c := tests.NewContext(http.MethodGet, "/").Build()
	c.SetParamNames("batch_id")
	c.SetParamValues(strconv.Itoa(int(batch.ID)))
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := DownloadBatchPDF(deps, mockGotemberg)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Disposition"), fmt.Sprintf("qr-codes-batch-%d.pdf", batch.ID))
	assert.True(t, mockGotemberg.ConvertCalled)
}

func TestDownloadBatchPDF_GotenbergError(t *testing.T) {
	deps, db := setup(t)

	batch := factories.NewQRCodeBatch().WithCount(1).Create(db)
	factories.NewQRCode().AsAvailable().WithBatch(batch.ID).Create(db)

	mockGotemberg := &mocks.GotenbergService{ConvertError: fmt.Errorf("boom")}

	c := tests.NewContext(http.MethodGet, "/").Build()
	c.SetParamNames("batch_id")
	c.SetParamValues(strconv.Itoa(int(batch.ID)))

	handler := DownloadBatchPDF(deps, mockGotemberg)
	err := handler(c)

	require.Error(t, err)
}

func TestUpdateStatus_Success(t *testing.T) {
	deps, db := setup(t)

	code := factories.NewQRCode().AsAvailable().Create(db)

	c := tests.NewContext(http.MethodPost, "/").WithFormData(map[string]string{"status": "damaged"}).Build()
	c.SetParamNames("uuid")
	c.SetParamValues(code.UUID)
	rec := c.Response().Writer.(*httptest.ResponseRecorder)

	handler := UpdateStatus(deps)
	err := handler(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	var updated models.QRCode
	db.First(&updated, code.ID)
	assert.Equal(t, models.QRCodeStatusDamaged, updated.Status)
}

func TestUpdateStatus_RejectsAssociated(t *testing.T) {
	deps, db := setup(t)
	portal := factories.NewPortal().Create(db)
	code := factories.NewQRCode().WithPortalModel(portal).Create(db)

	c := tests.NewContext(http.MethodPost, "/").WithFormData(map[string]string{"status": "damaged"}).Build()
	c.SetParamNames("uuid")
	c.SetParamValues(code.UUID)

	handler := UpdateStatus(deps)
	err := handler(c)

	require.Error(t, err)
}

func TestUpdateStatus_InvalidStatus(t *testing.T) {
	deps, db := setup(t)
	code := factories.NewQRCode().AsAvailable().Create(db)

	c := tests.NewContext(http.MethodPost, "/").WithFormData(map[string]string{"status": "associated"}).Build()
	c.SetParamNames("uuid")
	c.SetParamValues(code.UUID)

	handler := UpdateStatus(deps)
	err := handler(c)

	require.Error(t, err)
}

func TestUpdateStatus_NotFound(t *testing.T) {
	deps, _ := setup(t)

	c := tests.NewContext(http.MethodPost, "/").WithFormData(map[string]string{"status": "damaged"}).Build()
	c.SetParamNames("uuid")
	c.SetParamValues("nonexistent-uuid")

	handler := UpdateStatus(deps)
	err := handler(c)

	require.Error(t, err)
}
