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

func TestNewSMTPServiceFromFile(t *testing.T) {
	t.Run("valid credentials file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_creds_*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		creds := Credentials{
			Username: "test@gmail.com",
			Password: "app_password",
		}
		data, _ := json.Marshal(creds)
		tmpFile.Write(data)
		tmpFile.Close()

		service, err := NewSMTPServiceFromFile("smtp.gmail.com", "587", tmpFile.Name())

		assert.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("missing credentials file", func(t *testing.T) {
		service, err := NewSMTPServiceFromFile("smtp.gmail.com", "587", "/nonexistent/file.json")

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "failed to load credentials from file")
	})

	t.Run("invalid json in credentials file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_creds_*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString("invalid json content")
		tmpFile.Close()

		service, err := NewSMTPServiceFromFile("smtp.gmail.com", "587", tmpFile.Name())

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "failed to load credentials from file")
	})

	t.Run("empty credentials in file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_creds_*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		creds := Credentials{
			Username: "",
			Password: "app_password",
		}
		data, _ := json.Marshal(creds)
		tmpFile.Write(data)
		tmpFile.Close()

		service, err := NewSMTPServiceFromFile("smtp.gmail.com", "587", tmpFile.Name())

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "credentials file must contain both username and password")
	})
}

func TestNewSMTPServiceFromEnv(t *testing.T) {
	t.Run("valid environment variables", func(t *testing.T) {
		os.Setenv("TEST_USERNAME", "test@gmail.com")
		os.Setenv("TEST_PASSWORD", "app_password")
		defer func() {
			os.Unsetenv("TEST_USERNAME")
			os.Unsetenv("TEST_PASSWORD")
		}()

		service, err := NewSMTPServiceFromEnv("smtp.gmail.com", "587", "TEST_USERNAME", "TEST_PASSWORD", "custom@example.com")

		assert.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("missing username environment variable", func(t *testing.T) {
		os.Setenv("TEST_PASSWORD", "app_password")
		defer os.Unsetenv("TEST_PASSWORD")

		service, err := NewSMTPServiceFromEnv("smtp.gmail.com", "587", "MISSING_USERNAME", "TEST_PASSWORD", "")

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "environment variable MISSING_USERNAME is not set or empty")
	})

	t.Run("missing password environment variable", func(t *testing.T) {
		os.Setenv("TEST_USERNAME", "test@gmail.com")
		defer os.Unsetenv("TEST_USERNAME")

		service, err := NewSMTPServiceFromEnv("smtp.gmail.com", "587", "TEST_USERNAME", "MISSING_PASSWORD", "")

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "environment variable MISSING_PASSWORD is not set or empty")
	})

	t.Run("empty environment variable values", func(t *testing.T) {
		os.Setenv("TEST_USERNAME", "")
		os.Setenv("TEST_PASSWORD", "app_password")
		defer func() {
			os.Unsetenv("TEST_USERNAME")
			os.Unsetenv("TEST_PASSWORD")
		}()

		service, err := NewSMTPServiceFromEnv("smtp.gmail.com", "587", "TEST_USERNAME", "TEST_PASSWORD", "")

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "environment variable TEST_USERNAME is not set or empty")
	})
}

func TestNewSMTPServiceGmail(t *testing.T) {
	// This test will fail unless the gmail_smtp_credentials.json file exists
	// We test that the function exists and handles missing file gracefully
	service, err := NewSMTPServiceGmail()

	// Either succeeds if file exists, or fails with expected error
	if err != nil {
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "failed to load credentials from file")
	} else {
		assert.NotNil(t, service)
	}
}

func TestNewGmailSMTPServiceFromEnv(t *testing.T) {
	t.Run("with gmail environment variables", func(t *testing.T) {
		os.Setenv("GMAIL_USERNAME", "test@gmail.com")
		os.Setenv("GMAIL_PASSWORD", "app_password")
		defer func() {
			os.Unsetenv("GMAIL_USERNAME")
			os.Unsetenv("GMAIL_PASSWORD")
		}()

		service, err := NewGmailSMTPServiceFromEnv()

		assert.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("without gmail environment variables", func(t *testing.T) {
		service, err := NewGmailSMTPServiceFromEnv()

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "environment variable GMAIL_USERNAME is not set or empty")
	})
}

func TestSMTPService_SendMessage(t *testing.T) {
	service := NewSMTPService(SMTPConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
	})

	t.Run("valid message without attachments", func(t *testing.T) {
		msg := &EmailMessage{
			To:      []string{"recipient@example.com"},
			Subject: "Test Subject",
			Body:    "Test body content",
		}

		err := service.SendMessage(msg)

		// This will fail due to no actual SMTP server, but we can verify validation passes
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to SMTP server")
	})

	t.Run("message validation - no recipients", func(t *testing.T) {
		msg := &EmailMessage{
			To:      []string{},
			Subject: "Test Subject",
			Body:    "Test body content",
		}

		err := service.SendMessage(msg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message validation failed")
		assert.Contains(t, err.Error(), "at least one recipient is required")
	})

	t.Run("message validation - invalid email", func(t *testing.T) {
		msg := &EmailMessage{
			To:      []string{"invalid-email-format"},
			Subject: "Test Subject",
			Body:    "Test body content",
		}

		err := service.SendMessage(msg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message validation failed")
		assert.Contains(t, err.Error(), "invalid recipient")
	})

	t.Run("message validation - empty subject", func(t *testing.T) {
		msg := &EmailMessage{
			To:      []string{"recipient@example.com"},
			Subject: "",
			Body:    "Test body content",
		}

		err := service.SendMessage(msg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message validation failed")
		assert.Contains(t, err.Error(), "subject cannot be empty")
	})

	t.Run("message validation - invalid attachment path", func(t *testing.T) {
		msg := &EmailMessage{
			To:      []string{"recipient@example.com"},
			Subject: "Test Subject",
			Body:    "Test body content",
			Attachments: []Attachment{
				{FilePath: "../../../etc/passwd"},
			},
		}

		err := service.SendMessage(msg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message validation failed")
		assert.Contains(t, err.Error(), "path traversal detected")
	})

	t.Run("message with valid attachment", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_attachment_*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString("test attachment content")
		tmpFile.Close()

		msg := &EmailMessage{
			To:      []string{"recipient@example.com"},
			Subject: "Test Subject",
			Body:    "Test body content",
			Attachments: []Attachment{
				{FilePath: tmpFile.Name()},
			},
		}

		err = service.SendMessage(msg)

		// This will fail due to no actual SMTP server, but validation should pass
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to SMTP server")
	})

	t.Run("message with custom attachment filename and content type", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_attachment_*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString("test attachment content")
		tmpFile.Close()

		msg := &EmailMessage{
			To:      []string{"recipient@example.com"},
			Subject: "Test Subject",
			Body:    "Test body content",
			Attachments: []Attachment{
				{
					FilePath:    tmpFile.Name(),
					FileName:    "custom_name.txt",
					ContentType: "text/plain",
				},
			},
		}

		err = service.SendMessage(msg)

		// This will fail due to no actual SMTP server, but validation should pass
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to SMTP server")
	})

	t.Run("message with multiple recipients", func(t *testing.T) {
		msg := &EmailMessage{
			To: []string{
				"recipient1@example.com",
				"recipient2@example.com",
				"recipient3@example.com",
			},
			Subject: "Test Subject",
			Body:    "Test body content",
		}

		err := service.SendMessage(msg)

		// This will fail due to no actual SMTP server, but validation should pass
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to SMTP server")
	})

	t.Run("message with nonexistent attachment file", func(t *testing.T) {
		msg := &EmailMessage{
			To:      []string{"recipient@example.com"},
			Subject: "Test Subject",
			Body:    "Test body content",
			Attachments: []Attachment{
				{FilePath: "/nonexistent/file.txt"},
			},
		}

		err := service.SendMessage(msg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message validation failed")
		assert.Contains(t, err.Error(), "attachment file not accessible")
	})

	t.Run("message with directory as attachment", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test_dir_*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		msg := &EmailMessage{
			To:      []string{"recipient@example.com"},
			Subject: "Test Subject",
			Body:    "Test body content",
			Attachments: []Attachment{
				{FilePath: tmpDir},
			},
		}

		err = service.SendMessage(msg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message validation failed")
		assert.Contains(t, err.Error(), "attachment path points to a directory")
	})
}

func TestSMTPService_Send(t *testing.T) {
	service := NewSMTPService(SMTPConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
	})

	t.Run("legacy send without attachment", func(t *testing.T) {
		err := service.Send("recipient@example.com", "Test Subject", "Test body", "")

		// This will fail due to no actual SMTP server, but validation should pass
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to SMTP server")
	})

	t.Run("legacy send with attachment", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_attachment_*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString("test attachment content")
		tmpFile.Close()

		err = service.Send("recipient@example.com", "Test Subject", "Test body", tmpFile.Name())

		// This will fail due to no actual SMTP server, but validation should pass
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to SMTP server")
	})

	t.Run("legacy send with invalid email", func(t *testing.T) {
		err := service.Send("invalid-email", "Test Subject", "Test body", "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message validation failed")
	})

	t.Run("legacy send with invalid attachment", func(t *testing.T) {
		err := service.Send("recipient@example.com", "Test Subject", "Test body", "/nonexistent/file.txt")

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

func BenchmarkSendMessage_Validation(b *testing.B) {
	service := NewSMTPService(SMTPConfig{
		Host:     "smtp.example.com",
		Port:     "587",
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
	})

	msg := &EmailMessage{
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Body:    "Test body content",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will fail at connection, but validation will run
		service.SendMessage(msg)
	}
}