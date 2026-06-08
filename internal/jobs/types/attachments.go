package types

// UploadArgs contains the attachment ID to process
// The attachment record contains a temporary file path in StorageKey until the upload completes
type UploadArgs struct {
	AttachmentID uint `json:"attachment_id"`
}

func (UploadArgs) Kind() string { return "attachment_upload" }
