package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
)

// Credentials represents the Gmail SMTP credentials structure
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Attachment represents an email attachment
type Attachment struct {
	FilePath    string
	FileName    string // Override default filename
	ContentType string // Override auto-detected content type
}

// EmailMessage represents a structured email message
type EmailMessage struct {
	To          []string
	Subject     string
	Body        string
	Attachments []Attachment
}

// SMTPService implements EmailService using SMTP
type SMTPService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// SMTPConfig contains SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// NewSMTPService creates a new SMTP service instance
func NewSMTPService(config SMTPConfig) *SMTPService {
	return &SMTPService{
		host:     config.Host,
		port:     config.Port,
		username: config.Username,
		password: config.Password,
		from:     config.From,
	}
}

// loadCredentialsFromFile loads SMTP credentials from a JSON file
func loadCredentialsFromFile(path string) (*Credentials, error) {
	var fullPath string

	// If path is absolute, use as-is; otherwise, resolve from project root
	if filepath.IsAbs(path) {
		fullPath = path
	} else {
		fullPath = utils.MustGetPathFromRoot(path)
	}

	jsonFile, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SMTP credentials file at %s: %v", fullPath, err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read SMTP credentials file: %v", err)
	}

	var creds Credentials
	if err := json.Unmarshal(byteValue, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse SMTP credentials: %v", err)
	}

	// Validate credentials
	if creds.Username == "" || creds.Password == "" {
		return nil, fmt.Errorf("credentials file must contain both username and password")
	}

	return &creds, nil
}

// loadCredentialsFromEnv loads SMTP credentials from environment variables
func loadCredentialsFromEnv(usernameVar, passwordVar string) (*Credentials, error) {
	username := os.Getenv(usernameVar)
	password := os.Getenv(passwordVar)

	if username == "" {
		return nil, fmt.Errorf("environment variable %s is not set or empty", usernameVar)
	}
	if password == "" {
		return nil, fmt.Errorf("environment variable %s is not set or empty", passwordVar)
	}

	return &Credentials{
		Username: username,
		Password: password,
	}, nil
}

// NewSMTPServiceFromFile creates a new SMTP service by loading credentials from a file
func NewSMTPServiceFromFile(host, port string, credentialsPath string) (*SMTPService, error) {
	creds, err := loadCredentialsFromFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials from file: %v", err)
	}

	return NewSMTPService(SMTPConfig{
		Host:     host,
		Port:     port,
		Username: creds.Username,
		Password: creds.Password,
		From:     creds.Username, // Use username as from address by default
	}), nil
}

// NewSMTPServiceFromEnv creates a new SMTP service by loading credentials from environment variables
func NewSMTPServiceFromEnv(host, port, usernameVar, passwordVar, fromAddress string) (*SMTPService, error) {
	creds, err := loadCredentialsFromEnv(usernameVar, passwordVar)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials from environment: %v", err)
	}

	from := fromAddress
	if from == "" {
		from = creds.Username // Use username as from address if not specified
	}

	return NewSMTPService(SMTPConfig{
		Host:     host,
		Port:     port,
		Username: creds.Username,
		Password: creds.Password,
		From:     from,
	}), nil
}

// NewGmailSMTPService creates a new Gmail SMTP service instance by loading credentials from file
func NewSMTPServiceGmail() (*SMTPService, error) {
	return NewSMTPServiceFromFile("smtp.gmail.com", "587", "gmail_smtp_credentials.json")
}

// NewGmailSMTPServiceFromEnv creates a new Gmail SMTP service from environment variables
func NewGmailSMTPServiceFromEnv() (*SMTPService, error) {
	return NewSMTPServiceFromEnv("smtp.gmail.com", "587", "GMAIL_USERNAME", "GMAIL_PASSWORD", "")
}

// validateEmail checks if an email address is valid
func (s *SMTPService) validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email address cannot be empty")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email address: %v", err)
	}
	return nil
}

// validateAttachment checks if an attachment is accessible and secure
func (s *SMTPService) validateAttachment(attachment Attachment) error {
	if attachment.FilePath == "" {
		return nil // Empty attachment is valid
	}

	// Security: prevent path traversal
	cleanPath := filepath.Clean(attachment.FilePath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid attachment path: path traversal detected")
	}

	// Check if file exists and is readable
	info, err := os.Stat(cleanPath)
	if err != nil {
		return fmt.Errorf("attachment file not accessible: %v", err)
	}

	if info.IsDir() {
		return fmt.Errorf("attachment path points to a directory, not a file")
	}

	return nil
}

// validateMessage validates the email message structure
func (s *SMTPService) validateMessage(msg *EmailMessage) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	for _, recipient := range msg.To {
		if err := s.validateEmail(recipient); err != nil {
			return fmt.Errorf("invalid recipient %s: %v", recipient, err)
		}
	}

	if msg.Subject == "" {
		return fmt.Errorf("subject cannot be empty")
	}

	for i, attachment := range msg.Attachments {
		if err := s.validateAttachment(attachment); err != nil {
			return fmt.Errorf("invalid attachment %d: %v", i, err)
		}
	}

	return nil
}

// createConnection establishes and authenticates SMTP connection
func (s *SMTPService) createConnection() (*smtp.Client, error) {
	// Create TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.host,
	}

	// Connect to server with plain connection
	conn, err := smtp.Dial(s.host + ":" + s.port)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SMTP server: %v", err)
	}

	// Start TLS
	if err = conn.StartTLS(tlsConfig); err != nil {
		conn.Quit()
		return nil, fmt.Errorf("failed to start TLS: %v", err)
	}

	// Authenticate
	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	if err = conn.Auth(auth); err != nil {
		conn.Quit()
		return nil, fmt.Errorf("SMTP authentication failed: %v", err)
	}

	return conn, nil
}

// sendMessage sends the email message through the SMTP connection
func (s *SMTPService) sendMessage(conn *smtp.Client, msg *EmailMessage) error {
	// Set sender
	if err := conn.Mail(s.from); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	// Set recipients
	for _, recipient := range msg.To {
		if err := conn.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %v", recipient, err)
		}
	}

	// Get data writer
	writer, err := conn.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %v", err)
	}
	defer writer.Close()

	// Build and send message
	message, err := s.buildMessageFromStruct(msg)
	if err != nil {
		return fmt.Errorf("failed to build message: %v", err)
	}

	_, err = writer.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	return nil
}

// SendMessage sends an email using the structured EmailMessage type
func (s *SMTPService) SendMessage(msg *EmailMessage) error {
	// Validate message
	if err := s.validateMessage(msg); err != nil {
		return fmt.Errorf("message validation failed: %v", err)
	}

	// Create connection
	conn, err := s.createConnection()
	if err != nil {
		return err
	}
	defer conn.Quit()

	// Send message
	return s.sendMessage(conn, msg)
}

// Send implements the EmailService interface (backward compatibility)
func (s *SMTPService) Send(to, subject, body, attachments string) error {
	// Convert legacy parameters to new structure
	msg := &EmailMessage{
		To:      []string{to},
		Subject: subject,
		Body:    body,
	}

	// Add attachment if provided
	if attachments != "" {
		msg.Attachments = []Attachment{
			{FilePath: attachments},
		}
	}

	// Use new SendMessage method
	return s.SendMessage(msg)
}

// buildMessageFromStruct creates an email message from EmailMessage struct
func (s *SMTPService) buildMessageFromStruct(msg *EmailMessage) (string, error) {
	if len(msg.Attachments) == 0 {
		return s.buildSimpleMessage(msg), nil
	}
	return s.buildMultipartMessage(msg)
}

// buildSimpleMessage creates a simple email without attachments
func (s *SMTPService) buildSimpleMessage(msg *EmailMessage) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	buf.WriteString(msg.Body)
	return buf.String()
}

// buildMultipartMessage creates an email with attachments
func (s *SMTPService) buildMultipartMessage(msg *EmailMessage) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Write email headers
	headers := fmt.Sprintf("From: %s\r\n", s.from)
	headers += fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", "))
	headers += fmt.Sprintf("Subject: %s\r\n", msg.Subject)
	headers += "MIME-Version: 1.0\r\n"
	headers += fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", writer.Boundary())

	buf.WriteString(headers)

	// Text body part
	textPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type": []string{"text/plain; charset=utf-8"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create text part: %v", err)
	}
	if _, err := textPart.Write([]byte(msg.Body)); err != nil {
		return "", fmt.Errorf("failed to write text body: %v", err)
	}

	// Process attachments
	for i, attachment := range msg.Attachments {
		if err := s.addAttachment(writer, attachment); err != nil {
			return "", fmt.Errorf("failed to add attachment %d: %v", i, err)
		}
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %v", err)
	}

	return buf.String(), nil
}

// addAttachment adds a single attachment to the multipart writer
func (s *SMTPService) addAttachment(writer *multipart.Writer, attachment Attachment) error {
	file, err := os.Open(attachment.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open attachment file: %v", err)
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read attachment file: %v", err)
	}

	// Determine filename
	filename := attachment.FileName
	if filename == "" {
		filename = filepath.Base(attachment.FilePath)
	}

	// Determine content type
	contentType := attachment.ContentType
	if contentType == "" {
		contentType = getContentType(filename)
	}

	// Create attachment part
	attachPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              []string{contentType},
		"Content-Disposition":       []string{fmt.Sprintf("attachment; filename=\"%s\"", filename)},
		"Content-Transfer-Encoding": []string{"base64"},
	})
	if err != nil {
		return fmt.Errorf("failed to create attachment part: %v", err)
	}

	// Encode and write file content
	encoded := base64.StdEncoding.EncodeToString(fileContent)

	// Split base64 into 76-character lines as per RFC 2045
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		if _, err := attachPart.Write([]byte(encoded[i:end] + "\r\n")); err != nil {
			return fmt.Errorf("failed to write attachment content: %v", err)
		}
	}

	return nil
}

// getContentType returns the MIME content type based on file extension
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".html":
		return "text/html"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".zip":
		return "application/zip"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}
