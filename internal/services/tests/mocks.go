package tests

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/riverqueue/river"
)

// MockStorageService is a mock implementation of storage.StorageService for testing
type MockStorageService struct {
	UploadCalled  bool
	UploadError   error
	UploadedFiles []struct {
		Key         string
		ContentType string
	}

	DeleteCalled bool
	DeleteError  error
	DeletedFiles []string
	DeleteErrors map[string]error // Map storage keys to specific errors for individual file failures
}

func (m *MockStorageService) UploadFile(ctx context.Context, key string, file io.Reader, contentType string) (string, error) {
	m.UploadCalled = true
	m.UploadedFiles = append(m.UploadedFiles, struct {
		Key         string
		ContentType string
	}{Key: key, ContentType: contentType})

	if m.UploadError != nil {
		return "", m.UploadError
	}
	return fmt.Sprintf("https://storage.example.com/%s", key), nil
}

func (m *MockStorageService) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockStorageService) DeleteFile(ctx context.Context, key string) error {
	m.DeleteCalled = true
	m.DeletedFiles = append(m.DeletedFiles, key)

	// Check for specific error for this key
	if m.DeleteErrors != nil {
		if err, exists := m.DeleteErrors[key]; exists {
			return err
		}
	}

	return m.DeleteError
}

func (m *MockStorageService) GetFileURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	return fmt.Sprintf("https://storage.example.com/%s", key), nil
}

func (m *MockStorageService) FileExists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

// MockEmailService is a mock implementation of email.EmailService for testing
type MockEmailService struct {
	SendCalled bool
	SendError  error
	SentEmails []struct {
		To          []string
		Subject     string
		Body        string
		Attachments []string
	}
}

func (m *MockEmailService) Send(to []string, subject string, body string, attachments []string) error {
	m.SendCalled = true
	m.SentEmails = append(m.SentEmails, struct {
		To          []string
		Subject     string
		Body        string
		Attachments []string
	}{To: to, Subject: subject, Body: body, Attachments: attachments})

	return m.SendError
}

// MockTranslationService is a mock implementation of translation.TranslatorService for testing
type MockTranslationService struct{}

func (m *MockTranslationService) Translate(key string, args ...interface{}) string {
	return key
}

// MockBackgroundJobRunner is a mock implementation of jobs.BackgroundJobRunner for testing
type MockBackgroundJobRunner struct {
	EnqueueCalled bool
	EnqueueError  error
	EnqueuedJobs  []river.JobArgs
}

func (m *MockBackgroundJobRunner) Enqueue(ctx context.Context, args river.JobArgs) error {
	m.EnqueueCalled = true
	m.EnqueuedJobs = append(m.EnqueuedJobs, args)

	return m.EnqueueError
}
