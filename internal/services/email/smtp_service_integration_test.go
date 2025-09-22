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

		err := service.Send(msg.To, msg.Subject, msg.Body, []string{})
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

		err = service.Send(msg.To, msg.Subject, msg.Body, []string{msg.Attachments[0].FilePath})
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

		err := service.Send(msg.To, msg.Subject, msg.Body, []string{})
		assert.NoError(t, err, "Should be able to send email to multiple recipients")
	})
}

func TestSMTPService_Gmail_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run")
	}

	t.Run("gmail service from environment", func(t *testing.T) {
		// Set up SMTP environment variables
		username := os.Getenv("SMTP_USERNAME")
		password := os.Getenv("SMTP_PASSWORD")

		if username == "" || password == "" {
			t.Skip("Skipping SMTP integration test. Required environment variables not set: SMTP_USERNAME, SMTP_PASSWORD")
		}

		service, err := NewSMTPServiceFromEnv(&NewSMTPServiceFromEnvOptions{})
		require.NoError(t, err, "Should be able to create SMTP service from environment")

		err = service.Send([]string{username}, "Test SMTP Integration", "This is a test email sent through SMTP from the integration test.", []string{})
		assert.NoError(t, err, "Should be able to send email through SMTP")
	})

}

func TestSMTPService_Send_Integration(t *testing.T) {
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

	t.Run("send method without attachment", func(t *testing.T) {
		err := service.Send(
			[]string{username}, // Send to self for testing
			"Test Send Method",
			"This is a test email sent using the Send method.",
			[]string{}, // No attachment
		)

		assert.NoError(t, err, "Should be able to send email using send method")
	})

	t.Run("send method with attachment", func(t *testing.T) {
		// Create a temporary test file
		tmpFile, err := os.CreateTemp("", "test_*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString("Test attachment content")
		tmpFile.Close()

		err = service.Send(
			[]string{username}, // Send to self for testing
			"Test Send Method with Attachment",
			"This is a test email sent using the Send method with an attachment.",
			[]string{tmpFile.Name()},
		)

		assert.NoError(t, err, "Should be able to send email with attachment using send method")
	})
}

// Performance integration tests
func BenchmarkSMTPService_Send_Integration(b *testing.B) {
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
		err := service.Send(msg.To, msg.Subject, msg.Body, []string{})
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