package email

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GmailService implements EmailService using Gmail API
type GmailService struct {
	service *gmail.Service
}

// GmailConfig contains Gmail-specific configuration
type GmailConfig struct {
	CredentialsPath string
	TokenPath       string
}

// NewGmailService creates a new Gmail service instance
func NewGmailService(config GmailConfig) (*GmailService, error) {
	ctx := context.Background()

	// Read credentials
	b, err := os.ReadFile(config.CredentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}

	// Configure OAuth2
	oauthConfig, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %v", err)
	}

	// Get authenticated client
	client, err := getClient(oauthConfig, config.TokenPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get authenticated client: %v", err)
	}

	// Create Gmail service
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %v", err)
	}

	return &GmailService{service: srv}, nil
}

// Send implements the EmailService interface
func (g *GmailService) Send(config EmailConfig) error {
	var buf strings.Builder

	if config.AttachmentPath == "" {
		// Simple email without attachment
		buf.WriteString(fmt.Sprintf("To: %s\r\n", config.To))
		buf.WriteString(fmt.Sprintf("Subject: %s\r\n", config.Subject))
		buf.WriteString("MIME-version: 1.0\r\n")
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
		buf.WriteString(config.Body)
	} else {
		// Email with attachment using multipart
		writer := multipart.NewWriter(&buf)
		boundary := writer.Boundary()

		// Headers
		buf.WriteString(fmt.Sprintf("To: %s\r\n", config.To))
		buf.WriteString(fmt.Sprintf("Subject: %s\r\n", config.Subject))
		buf.WriteString("MIME-version: 1.0\r\n")
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))

		// Text body part
		textPart, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type": []string{"text/plain; charset=utf-8"},
		})
		if err != nil {
			return err
		}
		textPart.Write([]byte(config.Body))

		// Attachment part
		file, err := os.Open(config.AttachmentPath)
		if err != nil {
			return err
		}
		defer file.Close()

		filename := filepath.Base(config.AttachmentPath)
		attachPart, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type":              []string{"application/octet-stream"},
			"Content-Disposition":       []string{fmt.Sprintf("attachment; filename=\"%s\"", filename)},
			"Content-Transfer-Encoding": []string{"base64"},
		})
		if err != nil {
			return err
		}

		// Read file and encode to base64
		fileContent, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		attachPart.Write([]byte(base64.StdEncoding.EncodeToString(fileContent)))

		writer.Close()
	}

	// Create Gmail message
	gmailMessage := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(buf.String())),
	}

	// Send the email
	_, err := g.service.Users.Messages.Send("me", gmailMessage).Do()
	return err
}

// getClient retrieves an authenticated OAuth2 client
func getClient(config *oauth2.Config, tokenPath string) (*http.Client, error) {
	tok, err := tokenFromFile(tokenPath)
	if err != nil {
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		saveToken(tokenPath, tok)
	}
	return config.Client(context.Background(), tok), nil
}

// getTokenFromWeb requests a token from the web
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}
	return tok, nil
}

// tokenFromFile retrieves a token from a local file
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken saves a token to a file path
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}