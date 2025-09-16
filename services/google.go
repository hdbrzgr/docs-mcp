package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// AuthConfig holds the authentication configuration for Google APIs
type AuthConfig struct {
	CredentialsPath   string
	ClientSecretsPath string
	UseServiceAccount bool
}

// GoogleDocsClient provides a singleton Google Docs service client
var GoogleDocsClient = sync.OnceValue[*docs.Service](func() *docs.Service {
	config := loadGoogleCredentials()

	ctx := context.Background()
	var service *docs.Service

	if config.UseServiceAccount {
		// Use Service Account authentication
		log.Println("Using Service Account authentication for Google Docs API")

		credentialsData, err := ioutil.ReadFile(config.CredentialsPath)
		if err != nil {
			log.Fatalf("Failed to read service account credentials: %v", err)
		}

		creds, err := google.CredentialsFromJSON(ctx, credentialsData, docs.DocumentsScope, drive.DriveScope)
		if err != nil {
			log.Fatalf("Failed to create credentials from JSON: %v", err)
		}

		service, err = docs.NewService(ctx, option.WithCredentials(creds))
		if err != nil {
			log.Fatalf("Failed to create Google Docs service: %v", err)
		}
	} else {
		// Use OAuth 2.0 Client authentication
		log.Println("Using OAuth 2.0 Client authentication for Google Docs API")

		clientSecretsData, err := ioutil.ReadFile(config.ClientSecretsPath)
		if err != nil {
			log.Fatalf("Failed to read client secrets: %v", err)
		}

		oauthConfig, err := google.ConfigFromJSON(clientSecretsData, docs.DocumentsScope, drive.DriveScope)
		if err != nil {
			log.Fatalf("Failed to create OAuth config: %v", err)
		}

		// For server applications, you would typically implement a token storage mechanism
		// This is a simplified version - in production, implement proper token management
		client := getHTTPClient(ctx, oauthConfig)

		service, err = docs.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			log.Fatalf("Failed to create Google Docs service: %v", err)
		}
	}

	return service
})

// GoogleDriveClient provides a singleton Google Drive service client
var GoogleDriveClient = sync.OnceValue[*drive.Service](func() *drive.Service {
	config := loadGoogleCredentials()

	ctx := context.Background()
	var service *drive.Service

	if config.UseServiceAccount {
		// Use Service Account authentication
		log.Println("Using Service Account authentication for Google Drive API")

		credentialsData, err := ioutil.ReadFile(config.CredentialsPath)
		if err != nil {
			log.Fatalf("Failed to read service account credentials: %v", err)
		}

		creds, err := google.CredentialsFromJSON(ctx, credentialsData, docs.DocumentsScope, drive.DriveScope)
		if err != nil {
			log.Fatalf("Failed to create credentials from JSON: %v", err)
		}

		service, err = drive.NewService(ctx, option.WithCredentials(creds))
		if err != nil {
			log.Fatalf("Failed to create Google Drive service: %v", err)
		}
	} else {
		// Use OAuth 2.0 Client authentication
		log.Println("Using OAuth 2.0 Client authentication for Google Drive API")

		clientSecretsData, err := ioutil.ReadFile(config.ClientSecretsPath)
		if err != nil {
			log.Fatalf("Failed to read client secrets: %v", err)
		}

		oauthConfig, err := google.ConfigFromJSON(clientSecretsData, docs.DocumentsScope, drive.DriveScope)
		if err != nil {
			log.Fatalf("Failed to create OAuth config: %v", err)
		}

		client := getHTTPClient(ctx, oauthConfig)

		service, err = drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			log.Fatalf("Failed to create Google Drive service: %v", err)
		}
	}

	return service
})

// loadGoogleCredentials loads Google API credentials from environment variables
func loadGoogleCredentials() AuthConfig {
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	clientSecretsPath := os.Getenv("GOOGLE_CLIENT_SECRETS")

	// Check if we have service account credentials or OAuth client secrets
	hasServiceAccount := credentialsPath != ""
	hasClientSecrets := clientSecretsPath != ""

	if !hasServiceAccount && !hasClientSecrets {
		log.Fatal("Either GOOGLE_APPLICATION_CREDENTIALS or GOOGLE_CLIENT_SECRETS is required for authentication")
	}

	if hasServiceAccount && hasClientSecrets {
		log.Println("Both service account and client secrets provided, using service account authentication")
	}

	return AuthConfig{
		CredentialsPath:   credentialsPath,
		ClientSecretsPath: clientSecretsPath,
		UseServiceAccount: hasServiceAccount,
	}
}

// getHTTPClient gets an HTTP client for OAuth 2.0 authentication
// In a real application, you would implement proper token storage and refresh
func getHTTPClient(ctx context.Context, config *oauth2.Config) *http.Client {
	// This is a simplified implementation for demonstration
	// In production, you should:
	// 1. Implement proper token storage (file, database, etc.)
	// 2. Handle token refresh automatically
	// 3. Implement proper OAuth flow for initial authorization

	tokenPath := os.Getenv("GOOGLE_TOKEN_PATH")
	if tokenPath == "" {
		tokenPath = "token.json" // Default token file
	}

	token, err := tokenFromFile(tokenPath)
	if err != nil {
		token = getTokenFromWeb(config)
		saveToken(tokenPath, token)
	}

	return config.Client(ctx, token)
}

// getTokenFromWeb requests a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// tokenFromFile retrieves a token from a local file.
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

// saveToken saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
