package qrcodes

import (
	"github.com/labstack/echo/v4"
	"github.com/troptropcontent/qr_code_maintenance/internal/routers"
	"github.com/troptropcontent/qr_code_maintenance/internal/services/pdf"
)

func NewRouter(parentGroup echo.Group, dependencies *routers.Dependencies, pdfConverter ...pdf.HtmlToPdfConverter) {
	var converter pdf.HtmlToPdfConverter
	if len(pdfConverter) > 0 {
		converter = pdfConverter[0]
	} else {
		converter = pdf.NewGotenbergService()
	}

	qr_code_routes := parentGroup.Group("/qr-codes")
	qr_code_routes.GET("", GetQRCodes(dependencies))
	qr_code_routes.GET("/batches/new", GetNewBatchForm(dependencies))
	qr_code_routes.POST("/batches", CreateBatch(dependencies))
	qr_code_routes.GET("/batches/:batch_id", GetBatch(dependencies))
	qr_code_routes.GET("/batches/:batch_id/pdf", DownloadBatchPDF(dependencies, converter))
	qr_code_routes.GET("/:uuid", ShowQRCode(dependencies))
	qr_code_routes.GET("/:uuid/pdf", DownloadQRCodePDF(dependencies, converter))
	qr_code_routes.POST("/:uuid/status", UpdateStatus(dependencies))
}
