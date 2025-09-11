package email

// EmailConfig contains all email configuration parameters
type EmailConfig struct {
	To             string
	Subject        string
	Body           string
	AttachmentPath string // Optional - empty means no attachment
}

// EmailService defines the interface for email operations
type EmailService interface {
	// Send sends an email based on the provided configuration
	// If AttachmentPath is empty, sends a simple text email
	// If AttachmentPath is provided, sends email with attachment
	Send(config EmailConfig) error
}