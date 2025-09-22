package email

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockEmailService(t *testing.T) {
	t.Run("implements EmailService interface", func(t *testing.T) {
		var service EmailService = NewMockEmailService()
		assert.NotNil(t, service)
	})

	t.Run("successful send", func(t *testing.T) {
		mock := NewMockEmailService()
		
		err := mock.Send("test@example.com", "Test Subject", "Test Body", "")
		
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.GetSendCallCount())
		assert.True(t, mock.WasCalled())
		
		lastCall := mock.GetLastSendCall()
		assert.NotNil(t, lastCall)
		assert.Equal(t, "test@example.com", lastCall.To)
		assert.Equal(t, "Test Subject", lastCall.Subject)
		assert.Equal(t, "Test Body", lastCall.Body)
		assert.Equal(t, "", lastCall.Attachment)
	})

	t.Run("send with attachment", func(t *testing.T) {
		mock := NewMockEmailService()
		
		err := mock.Send("test@example.com", "Test Subject", "Test Body", "/path/to/file.pdf")
		
		assert.NoError(t, err)
		assert.True(t, mock.WasCalledWith("test@example.com", "Test Subject", "Test Body", "/path/to/file.pdf"))
	})

	t.Run("send failure", func(t *testing.T) {
		mock := NewMockEmailService()
		mock.SetSendError(errors.New("network error"))
		
		err := mock.Send("test@example.com", "Test Subject", "Test Body", "")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
		assert.Equal(t, 1, mock.GetSendCallCount())
	})

	t.Run("send message", func(t *testing.T) {
		mock := NewMockEmailService()
		
		msg := &EmailMessage{
			To:      []string{"test@example.com"},
			Subject: "Test Subject",
			Body:    "Test Body",
		}
		
		err := mock.SendMessage(msg)
		
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.GetSendMessageCallCount())
		
		lastCall := mock.GetLastSendMessageCall()
		assert.NotNil(t, lastCall)
		assert.Equal(t, msg, lastCall.Message)
	})

	t.Run("send message with multiple recipients", func(t *testing.T) {
		mock := NewMockEmailService()
		
		msg := &EmailMessage{
			To:      []string{"test1@example.com", "test2@example.com"},
			Subject: "Test Subject",
			Body:    "Test Body",
		}
		
		err := mock.SendMessage(msg)
		
		assert.NoError(t, err)
		
		found := mock.WasCalledWithMessage(func(m *EmailMessage) bool {
			return len(m.To) == 2 && 
				   m.To[0] == "test1@example.com" && 
				   m.To[1] == "test2@example.com"
		})
		assert.True(t, found)
	})

	t.Run("multiple calls tracking", func(t *testing.T) {
		mock := NewMockEmailService()
		
		// Make multiple calls
		mock.Send("test1@example.com", "Subject 1", "Body 1", "")
		mock.Send("test2@example.com", "Subject 2", "Body 2", "/file.pdf")
		mock.SendMessage(&EmailMessage{
			To:      []string{"test3@example.com"},
			Subject: "Subject 3",
			Body:    "Body 3",
		})
		
		assert.Equal(t, 2, mock.GetSendCallCount())
		assert.Equal(t, 1, mock.GetSendMessageCallCount())
		
		allSendCalls := mock.GetAllSendCalls()
		assert.Len(t, allSendCalls, 2)
		assert.Equal(t, "test1@example.com", allSendCalls[0].To)
		assert.Equal(t, "test2@example.com", allSendCalls[1].To)
		
		allMessageCalls := mock.GetAllSendMessageCalls()
		assert.Len(t, allMessageCalls, 1)
		assert.Equal(t, "test3@example.com", allMessageCalls[0].Message.To[0])
	})

	t.Run("reset functionality", func(t *testing.T) {
		mock := NewMockEmailService()
		
		// Make some calls and set failure
		mock.Send("test@example.com", "Subject", "Body", "")
		mock.SetSendError(errors.New("test error"))
		
		assert.Equal(t, 1, mock.GetSendCallCount())
		assert.True(t, mock.WasCalled())
		
		// Reset
		mock.Reset()
		
		assert.Equal(t, 0, mock.GetSendCallCount())
		assert.Equal(t, 0, mock.GetSendMessageCallCount())
		assert.False(t, mock.WasCalled())
		assert.Nil(t, mock.GetLastSendCall())
		assert.Nil(t, mock.GetLastSendMessageCall())
		
		// Should succeed after reset
		err := mock.Send("test@example.com", "Subject", "Body", "")
		assert.NoError(t, err)
	})

	t.Run("thread safety", func(t *testing.T) {
		mock := NewMockEmailService()
		
		// Simulate concurrent access
		done := make(chan bool, 2)
		
		go func() {
			for i := 0; i < 10; i++ {
				mock.Send("test@example.com", "Subject", "Body", "")
			}
			done <- true
		}()
		
		go func() {
			for i := 0; i < 10; i++ {
				mock.GetSendCallCount()
				mock.WasCalled()
			}
			done <- true
		}()
		
		// Wait for both goroutines
		<-done
		<-done
		
		assert.Equal(t, 10, mock.GetSendCallCount())
	})
}

// Example of how to use the mock in application tests
func TestEmailNotificationService(t *testing.T) {
	// This is an example of how you might test a service that uses EmailService
	mock := NewMockEmailService()
	
	// Example service that sends notifications
	notificationService := &ExampleNotificationService{
		emailService: mock,
	}
	
	t.Run("send welcome email", func(t *testing.T) {
		err := notificationService.SendWelcomeEmail("user@example.com", "John Doe")
		
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.GetSendCallCount())
		
		lastCall := mock.GetLastSendCall()
		assert.Equal(t, "user@example.com", lastCall.To)
		assert.Contains(t, lastCall.Subject, "Welcome")
		assert.Contains(t, lastCall.Body, "John Doe")
	})
	
	t.Run("handle email failure", func(t *testing.T) {
		mock.Reset()
		mock.SetSendError(errors.New("SMTP server unavailable"))
		
		err := notificationService.SendWelcomeEmail("user@example.com", "John Doe")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send welcome email")
	})
}

// ExampleNotificationService demonstrates how a service might use EmailService
type ExampleNotificationService struct {
	emailService EmailService
}

func (s *ExampleNotificationService) SendWelcomeEmail(email, name string) error {
	subject := "Welcome to our service!"
	body := "Hello " + name + ",\n\nWelcome to our service. We're excited to have you on board!"
	
	err := s.emailService.Send(email, subject, body, "")
	if err != nil {
		return errors.New("failed to send welcome email: " + err.Error())
	}
	
	return nil
}

// Example usage patterns for testing with the mock
func ExampleMockEmailService_usage() {
	// Create a mock service
	mock := NewMockEmailService()
	
	// Use it in your application code
	var emailService EmailService = mock
	
	// Send an email
	emailService.Send("user@example.com", "Test", "Hello", "")
	
	// Verify the call was made
	if mock.WasCalled() {
		lastCall := mock.GetLastSendCall()
		println("Email sent to:", lastCall.To)
	}
	
	// Configure failure for testing error handling
	mock.SetSendError(errors.New("network error"))
	
	// Reset for next test
	mock.Reset()
}