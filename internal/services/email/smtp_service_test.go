package email

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSMTPService(t *testing.T) {
	config := SMTPConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		Username: "test@example.com",
		Password: "password123",
		From:     "sender@example.com",
	}

	service := NewSMTPService(config)

	assert.NotNil(t, service)
	// Note: We can't test private fields directly, but we can test behavior
	// The service should be ready to use with the provided config
}

func TestSMTPService_Send(t *testing.T) {
	service := NewSMTPService(SMTPConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
	})

	t.Run("send without attachement", func(t *testing.T) {
		to := []string{"test@example.com"}
		attachements := []string{}
		subject := "Test Subject"
		body := "Test body"
		err := service.Send(to, subject, body, attachements)

		// This will fail due to no actual SMTP server, but validation should pass
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to SMTP server")
	})

	t.Run("send with attachment", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_attachment_*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString("test attachment content")
		tmpFile.Close()

		to := []string{"test@example.com"}
		attachements := []string{tmpFile.Name()}
		subject := "Test Subject"
		body := "Test body"
		err = service.Send(to, subject, body, attachements)

		// This will fail due to no actual SMTP server, but validation should pass
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to SMTP server")
	})

	t.Run("send with invalid email", func(t *testing.T) {
		err := service.Send([]string{"invalid-email"}, "Test Subject", "Test body", []string{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message validation failed")
	})

	t.Run("send with invalid attachment", func(t *testing.T) {
		err := service.Send([]string{"recipient@example.com"}, "Test Subject", "Test body", []string{"/nonexistent/file.txt"})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message validation failed")
	})
}

func TestEmailMessage_Struct(t *testing.T) {
	t.Run("create valid email message", func(t *testing.T) {
		msg := &EmailMessage{
			To:      []string{"test@example.com"},
			Subject: "Test Subject",
			Body:    "Test body",
		}

		assert.Equal(t, []string{"test@example.com"}, msg.To)
		assert.Equal(t, "Test Subject", msg.Subject)
		assert.Equal(t, "Test body", msg.Body)
		assert.Empty(t, msg.Attachments)
	})

	t.Run("create email message with attachments", func(t *testing.T) {
		attachment := Attachment{
			FilePath:    "/path/to/file.pdf",
			FileName:    "document.pdf",
			ContentType: "application/pdf",
		}

		msg := &EmailMessage{
			To:          []string{"test@example.com"},
			Subject:     "Test Subject",
			Body:        "Test body",
			Attachments: []Attachment{attachment},
		}

		assert.Len(t, msg.Attachments, 1)
		assert.Equal(t, "/path/to/file.pdf", msg.Attachments[0].FilePath)
		assert.Equal(t, "document.pdf", msg.Attachments[0].FileName)
		assert.Equal(t, "application/pdf", msg.Attachments[0].ContentType)
	})
}

func TestAttachment_Struct(t *testing.T) {
	t.Run("create basic attachment", func(t *testing.T) {
		attachment := Attachment{
			FilePath: "/path/to/file.txt",
		}

		assert.Equal(t, "/path/to/file.txt", attachment.FilePath)
		assert.Empty(t, attachment.FileName)
		assert.Empty(t, attachment.ContentType)
	})

	t.Run("create attachment with custom properties", func(t *testing.T) {
		attachment := Attachment{
			FilePath:    "/path/to/file.txt",
			FileName:    "custom_name.txt",
			ContentType: "text/plain",
		}

		assert.Equal(t, "/path/to/file.txt", attachment.FilePath)
		assert.Equal(t, "custom_name.txt", attachment.FileName)
		assert.Equal(t, "text/plain", attachment.ContentType)
	})
}

func TestSMTPConfig_Struct(t *testing.T) {
	config := SMTPConfig{
		Host:     "smtp.gmail.com",
		Port:     "587",
		Username: "user@gmail.com",
		Password: "app_password",
		From:     "sender@gmail.com",
	}

	assert.Equal(t, "smtp.gmail.com", config.Host)
	assert.Equal(t, "587", config.Port)
	assert.Equal(t, "user@gmail.com", config.Username)
	assert.Equal(t, "app_password", config.Password)
	assert.Equal(t, "sender@gmail.com", config.From)
}

func TestCredentials_Struct(t *testing.T) {
	creds := Credentials{
		Username: "user@gmail.com",
		Password: "app_password",
	}

	assert.Equal(t, "user@gmail.com", creds.Username)
	assert.Equal(t, "app_password", creds.Password)

	// Test JSON marshaling/unmarshaling
	data, err := json.Marshal(creds)
	assert.NoError(t, err)

	var unmarshaled Credentials
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, creds.Username, unmarshaled.Username)
	assert.Equal(t, creds.Password, unmarshaled.Password)
}

// Integration test helpers for when testing with real SMTP servers
func createTestCredentialsFile(t *testing.T, username, password string) (string, func()) {
	tmpFile, err := os.CreateTemp("", "test_creds_*.json")
	require.NoError(t, err)

	creds := Credentials{
		Username: username,
		Password: password,
	}
	data, err := json.Marshal(creds)
	require.NoError(t, err)

	_, err = tmpFile.Write(data)
	require.NoError(t, err)
	tmpFile.Close()

	return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
}

func createTestAttachment(t *testing.T, content string) (string, func()) {
	tmpFile, err := os.CreateTemp("", "test_attachment_*.txt")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
}

// Benchmark tests for performance
func BenchmarkNewSMTPService(b *testing.B) {
	config := SMTPConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewSMTPService(config)
	}
}

func BenchmarkSend_Validation(b *testing.B) {
	service := NewSMTPService(SMTPConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will fail at connection, but validation will run
		service.Send([]string{"recipient@example.com"}, "Test Subject", "Test body content", []string{})
	}
}
