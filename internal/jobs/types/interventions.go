package types

// AttachReportPdfArgs contains the intervention ID to process
type AttachReportPdfArgs struct {
	InterventionID uint `json:"intervention_id"`
}

func (AttachReportPdfArgs) Kind() string { return "attach_report_pdf" }
