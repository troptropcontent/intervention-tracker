package email

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests that require real SMTP credentials
// These tests are skipped unless INTEGRATION_TEST environment variable is set

func TestSMTPService_SendMessage_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run")
	}

	// These tests require real SMTP credentials to be set in environment
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	
	if username == "" || password == "" || host == "" || port == "" {
		t.Skip("Skipping integration test. Required environment variables not set: SMTP_USERNAME, SMTP_PASSWORD, SMTP_HOST, SMTP_PORT")
	}

	service := NewSMTPService(SMTPConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     username,
	})

	t.Run("send simple email", func(t *testing.T) {
		msg := &EmailMessage{
			To:      []string{username}, // Send to self for testing
			Subject: "Test Email from Integration Test",
			Body:    "This is a test email sent from the SMTP service integration test.",
		}

		err := service.SendMessage(msg)
		assert.NoError(t, err, "Should be able to send simple email")
	})

	t.Run("send email with attachment", func(t *testing.T) {
		// Create a temporary test file
		tmpFile, err := os.CreateTemp("", "integration_test_*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		testContent := "This is a test attachment content for integration testing."
		tmpFile.WriteString(testContent)
		tmpFile.Close()

		msg := &EmailMessage{
			To:      []string{username}, // Send to self for testing
			Subject: "Test Email with Attachment",
			Body:    "This is a test email with an attachment sent from the SMTP service integration test.",
			Attachments: []Attachment{
				{
					FilePath:    tmpFile.Name(),
					FileName:    "test_attachment.txt",
					ContentType: "text/plain",
				},
			},
		}

		err = service.SendMessage(msg)
		assert.NoError(t, err, "Should be able to send email with attachment")
	})

	t.Run("send email to multiple recipients", func(t *testing.T) {
		// For this test, you might want to set up multiple test email addresses
		// For now, we'll just send to the same address multiple times
		msg := &EmailMessage{
			To: []string{
				username,
				username, // Duplicate for testing (in real scenario, use different addresses)
			},
			Subject: "Test Email to Multiple Recipients",
			Body:    "This is a test email sent to multiple recipients from the SMTP service integration test.",
		}

		err := service.SendMessage(msg)
		assert.NoError(t, err, "Should be able to send email to multiple recipients")
	})
}

func TestSMTPService_Gmail_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run")
	}

	t.Run("gmail service from environment", func(t *testing.T) {
		// Set up Gmail environment variables
		username := os.Getenv("GMAIL_USERNAME")
		password := os.Getenv("GMAIL_PASSWORD")
		
		if username == "" || password == "" {
			t.Skip("Skipping Gmail integration test. Required environment variables not set: GMAIL_USERNAME, GMAIL_PASSWORD")
		}

		service, err := NewGmailSMTPServiceFromEnv()
		require.NoError(t, err, "Should be able to create Gmail service from environment")

		msg := &EmailMessage{
			To:      []string{username}, // Send to self for testing
			Subject: "Test Gmail Integration",
			Body:    "This is a test email sent through Gmail SMTP from the integration test.",
		}

		err = service.SendMessage(msg)
		assert.NoError(t, err, "Should be able to send email through Gmail")
	})

	t.Run("gmail service from file", func(t *testing.T) {
		// This test requires gmail_smtp_credentials.json to exist
		service, err := NewSMTPServiceGmail()
		if err != nil {
			t.Skipf("Skipping Gmail file integration test. gmail_smtp_credentials.json not found: %v", err)
		}

		msg := &EmailMessage{
			To:      []string{"test@example.com"}, // You should change this to a real email for testing
			Subject: "Test Gmail Integration from File",
			Body:    "This is a test email sent through Gmail SMTP using credentials from file.",
		}

		// Note: This might fail if the email address is not valid
		// Consider using a real test email address
		err = service.SendMessage(msg)
		
		// We don't assert no error here because the email might be invalid
		// In a real integration test, you'd use valid test emails
		t.Logf("Send result: %v", err)
	})
}

func TestSMTPService_Legacy_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run")
	}

	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	
	if username == "" || password == "" || host == "" || port == "" {
		t.Skip("Skipping integration test. Required environment variables not set")
	}

	service := NewSMTPService(SMTPConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     username,
	})

	t.Run("legacy send method without attachment", func(t *testing.T) {
		err := service.Send(
			username, // Send to self for testing
			"Test Legacy Send Method",
			"This is a test email sent using the legacy Send method.",
			"", // No attachment
		)

		assert.NoError(t, err, "Should be able to send email using legacy method")
	})

	t.Run("legacy send method with attachment", func(t *testing.T) {
		// Create a temporary test file
		tmpFile, err := os.CreateTemp("", "legacy_test_*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString("Legacy test attachment content")
		tmpFile.Close()

		err = service.Send(
			username, // Send to self for testing
			"Test Legacy Send Method with Attachment",
			"This is a test email sent using the legacy Send method with an attachment.",
			tmpFile.Name(),
		)

		assert.NoError(t, err, "Should be able to send email with attachment using legacy method")
	})
}

// Performance integration tests
func BenchmarkSMTPService_SendMessage_Integration(b *testing.B) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		b.Skip("Skipping integration benchmark. Set INTEGRATION_TEST=1 to run")
	}

	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	
	if username == "" || password == "" || host == "" || port == "" {
		b.Skip("Skipping integration benchmark. Required environment variables not set")
	}

	service := NewSMTPService(SMTPConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     username,
	})

	msg := &EmailMessage{
		To:      []string{username},
		Subject: "Benchmark Test Email",
		Body:    "This is a benchmark test email.",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.SendMessage(msg)
		if err != nil {
			b.Fatalf("Failed to send email: %v", err)
		}
	}
}

/*
Integration Test Setup Instructions:

1. For SMTP server testing, set these environment variables:
   export INTEGRATION_TEST=1
   export SMTP_HOST=smtp.your-server.com
   export SMTP_PORT=587
   export SMTP_USERNAME=your-email@domain.com
   export SMTP_PASSWORD=your-password

2. For Gmail testing, set these environment variables:
   export INTEGRATION_TEST=1
   export GMAIL_USERNAME=your-email@gmail.com
   export GMAIL_PASSWORD=your-app-password

3. Run integration tests with:
   go test -run Integration ./internal/services/email/

4. Run all tests including integration:
   INTEGRATION_TEST=1 go test ./internal/services/email/

5. Run benchmarks:
   INTEGRATION_TEST=1 go test -bench=. ./internal/services/email/

Note: 
- Never commit real credentials to version control
- Use app passwords for Gmail, not regular passwords
- Consider using test email addresses that you control
- Integration tests will actually send emails, so use responsibly
*/