package email

import (
	"fmt"
	"sync"
)

// MockEmailService is a mock implementation of EmailService for testing
type MockEmailService struct {
	mu sync.RWMutex
	
	// Configuration
	ShouldFailSend      bool
	SendError           error
	
	// Call tracking
	SendCalls           []SendCall
	SendMessageCalls    []SendMessageCall
	
	// Behavior simulation
	DelayDuration       int // milliseconds to simulate delay
}

// SendCall represents a call to the Send method
type SendCall struct {
	To         string
	Subject    string
	Body       string
	Attachment string
}

// SendMessageCall represents a call to the SendMessage method
type SendMessageCall struct {
	Message *EmailMessage
}

// NewMockEmailService creates a new mock email service
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SendCalls:        make([]SendCall, 0),
		SendMessageCalls: make([]SendMessageCall, 0),
	}
}

// Send implements the EmailService interface
func (m *MockEmailService) Send(to, subject, body, attachment string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Record the call
	m.SendCalls = append(m.SendCalls, SendCall{
		To:         to,
		Subject:    subject,
		Body:       body,
		Attachment: attachment,
	})
	
	// Simulate failure if configured
	if m.ShouldFailSend {
		if m.SendError != nil {
			return m.SendError
		}
		return fmt.Errorf("mock send failure")
	}
	
	return nil
}

// SendMessage provides additional functionality for testing (not part of interface)
func (m *MockEmailService) SendMessage(msg *EmailMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Record the call
	m.SendMessageCalls = append(m.SendMessageCalls, SendMessageCall{
		Message: msg,
	})
	
	// Simulate failure if configured
	if m.ShouldFailSend {
		if m.SendError != nil {
			return m.SendError
		}
		return fmt.Errorf("mock send message failure")
	}
	
	return nil
}

// Reset clears all recorded calls and resets configuration
func (m *MockEmailService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.SendCalls = make([]SendCall, 0)
	m.SendMessageCalls = make([]SendMessageCall, 0)
	m.ShouldFailSend = false
	m.SendError = nil
	m.DelayDuration = 0
}

// GetSendCallCount returns the number of times Send was called
func (m *MockEmailService) GetSendCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.SendCalls)
}

// GetSendMessageCallCount returns the number of times SendMessage was called
func (m *MockEmailService) GetSendMessageCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.SendMessageCalls)
}

// GetLastSendCall returns the last call to Send, or nil if no calls were made
func (m *MockEmailService) GetLastSendCall() *SendCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.SendCalls) == 0 {
		return nil
	}
	return &m.SendCalls[len(m.SendCalls)-1]
}

// GetLastSendMessageCall returns the last call to SendMessage, or nil if no calls were made
func (m *MockEmailService) GetLastSendMessageCall() *SendMessageCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if len(m.SendMessageCalls) == 0 {
		return nil
	}
	return &m.SendMessageCalls[len(m.SendMessageCalls)-1]
}

// GetAllSendCalls returns a copy of all Send calls
func (m *MockEmailService) GetAllSendCalls() []SendCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	calls := make([]SendCall, len(m.SendCalls))
	copy(calls, m.SendCalls)
	return calls
}

// GetAllSendMessageCalls returns a copy of all SendMessage calls
func (m *MockEmailService) GetAllSendMessageCalls() []SendMessageCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	calls := make([]SendMessageCall, len(m.SendMessageCalls))
	copy(calls, m.SendMessageCalls)
	return calls
}

// SetSendError configures the mock to return a specific error
func (m *MockEmailService) SetSendError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.ShouldFailSend = true
	m.SendError = err
}

// SetSendSuccess configures the mock to succeed
func (m *MockEmailService) SetSendSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.ShouldFailSend = false
	m.SendError = nil
}

// WasCalled returns true if any email sending method was called
func (m *MockEmailService) WasCalled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return len(m.SendCalls) > 0 || len(m.SendMessageCalls) > 0
}

// WasCalledWith checks if Send was called with specific parameters
func (m *MockEmailService) WasCalledWith(to, subject, body, attachment string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, call := range m.SendCalls {
		if call.To == to && call.Subject == subject && call.Body == body && call.Attachment == attachment {
			return true
		}
	}
	return false
}

// WasCalledWithMessage checks if SendMessage was called with a message matching the criteria
func (m *MockEmailService) WasCalledWithMessage(checkFunc func(*EmailMessage) bool) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, call := range m.SendMessageCalls {
		if checkFunc(call.Message) {
			return true
		}
	}
	return false
}