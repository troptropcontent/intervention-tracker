package config

import (
	"os"
	"path/filepath"

	"github.com/troptropcontent/qr_code_maintenance/internal/utils"
)

// EmailConfig holds email service configuration
type EmailConfig struct {
	Gmail GmailConfig `json:"gmail"`
}

// GmailConfig holds Gmail-specific configuration
type GmailConfig struct {
	CredentialsPath string `json:"credentials_path"`
	TokenPath       string `json:"token_path"`
}

// GetEmailConfig returns the email configuration based on environment variables
func GetEmailConfig() (*EmailConfig, error) {
	root, err := utils.GetProjectRootPath()
	if err != nil {
		return nil, err
	}

	return &EmailConfig{
		Gmail: GmailConfig{
			CredentialsPath: filepath.Join(root, "credentials.json"),
			TokenPath:       filepath.Join(root, "token.json"),
		},
	}, nil
}

// GetGmailCredentialsPath returns the path to Gmail credentials file
func GetGmailCredentialsPath() string {
	if path := os.Getenv("GMAIL_CREDENTIALS_PATH"); path != "" {
		return path
	}

	root, err := utils.GetProjectRootPath()
	if err != nil {
		return "credentials.json" // fallback to current directory
	}

	return filepath.Join(root, "credentials.json")
}

// GetGmailTokenPath returns the path to Gmail token file
func GetGmailTokenPath() string {
	if path := os.Getenv("GMAIL_TOKEN_PATH"); path != "" {
		return path
	}

	root, err := utils.GetProjectRootPath()
	if err != nil {
		return "token.json" // fallback to current directory
	}

	return filepath.Join(root, "token.json")
}
