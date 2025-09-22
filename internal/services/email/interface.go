package email

// EmailService defines the interface for sending emails with optional attachments.
// This interface abstracts email sending operations to allow for different implementations
// (Gmail, SMTP, etc.) while maintaining a consistent API.
type EmailService interface {
	// Send sends an email to the specified recipient with the given subject and body.
	// Parameters:
	//   - to: recipient email address
	//   - subject: email subject line
	//   - body: email content (can be plain text or HTML)
	//   - attachment: file path to attachment; if empty, sends email without attachment
	// Returns an error if the email fails to send.
	Send(to []string, subject string, body string, attachments []string) error
}
