package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

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
	tokenPath := os.Getenv("GOOGLE_TOKEN_PATH")
	if tokenPath == "" {
		tokenPath = "token.json" // Default token file
	}

	token, err := tokenFromFile(tokenPath)
	if err != nil {
		// Check if we should use callback server or manual flow
		useCallback := os.Getenv("OAUTH_USE_CALLBACK")
		if useCallback == "true" || useCallback == "1" {
			token = getTokenFromWebWithCallback(config)
		} else {
			token = getTokenFromWeb(config)
		}
		saveToken(tokenPath, token)
	}

	return config.Client(ctx, token)
}

// getTokenFromWeb requests a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	fmt.Println("🔐 Attempting to authorize...")
	fmt.Println()

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("📋 Authorize this app by visiting this url:")
	fmt.Println()
	fmt.Printf("🔗 %s\n", authURL)
	fmt.Println()
	fmt.Println("📖 Instructions:")
	fmt.Println("1. Copy the entire URL above")
	fmt.Println("2. Paste it into your web browser and press Enter")
	fmt.Println("3. Log in with your Google account")
	fmt.Println("4. Review and click 'Allow' or 'Grant' to authorize the app")
	fmt.Println("5. After clicking Allow, your browser will redirect to http://localhost")
	fmt.Println("6. Look at the URL in your browser's address bar")
	fmt.Println("7. Copy the authorization code (the long string after 'code=' and before '&scope')")
	fmt.Println()
	fmt.Print("🔑 Enter the code from that page here: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("❌ Unable to read authorization code: %v", err)
	}

	// Validate the authorization code
	authCode = strings.TrimSpace(authCode)
	if authCode == "" {
		log.Fatalf("❌ Authorization code cannot be empty")
	}

	// Basic validation - Google auth codes typically start with "4/"
	if !strings.HasPrefix(authCode, "4/") {
		fmt.Println("⚠️  Warning: Authorization code doesn't look like a typical Google auth code.")
		fmt.Println("   Make sure you copied the code correctly from the URL.")
		fmt.Println("   The code should start with '4/' and be quite long.")
		fmt.Print("   Continue anyway? (y/N): ")

		var confirm string
		if _, err := fmt.Scan(&confirm); err != nil || strings.ToLower(confirm) != "y" {
			log.Fatalf("❌ Authorization cancelled by user")
		}
	}

	fmt.Println()
	fmt.Println("🔄 Exchanging authorization code for access token...")

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		fmt.Println("❌ Failed to exchange authorization code for access token.")
		fmt.Println("   This usually means:")
		fmt.Println("   - The authorization code is incorrect or expired")
		fmt.Println("   - The code was already used (they can only be used once)")
		fmt.Println("   - There was a network issue")
		fmt.Println()
		fmt.Println("   Please try the authorization process again.")
		log.Fatalf("❌ Unable to retrieve token from web: %v", err)
	}

	fmt.Println("✅ Authentication successful!")
	return tok
}

// getTokenFromWebWithCallback requests a token using OAuth callback server
func getTokenFromWebWithCallback(config *oauth2.Config) *oauth2.Token {
	fmt.Println("🔐 Starting OAuth authorization with callback server...")
	fmt.Println()

	// Set up OAuth config with callback URL
	callbackPort := os.Getenv("OAUTH_CALLBACK_PORT")
	if callbackPort == "" {
		callbackPort = "8080"
	}

	callbackURL := fmt.Sprintf("http://localhost:%s/oauth/callback", callbackPort)
	config.RedirectURL = callbackURL

	// Generate state for security
	state := fmt.Sprintf("state-%d", time.Now().Unix())

	// Start callback server
	server := &http.Server{
		Addr: ":" + callbackPort,
	}

	// Channel to receive the authorization code
	codeChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	// Set up callback handler
	http.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		// Check state parameter for security
		if r.URL.Query().Get("state") != state {
			errorChan <- fmt.Errorf("invalid state parameter")
			return
		}

		// Get authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			errorChan <- fmt.Errorf("authorization code not found")
			return
		}

		// Send success response to browser
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Authorization Successful</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .success { color: #4CAF50; font-size: 24px; }
        .message { color: #666; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="success">✅ Authorization Successful!</div>
    <div class="message">You can now close this window and return to your terminal.</div>
    <div class="message">The server will continue starting up...</div>
</body>
</html>`)

		// Send code to main goroutine
		codeChan <- code
	})

	// Start server in goroutine
	go func() {
		fmt.Printf("🌐 Starting callback server on port %s\n", callbackPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errorChan <- fmt.Errorf("callback server error: %v", err)
		}
	}()

	// Give server a moment to start
	time.Sleep(1 * time.Second)

	// Generate authorization URL
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Println("📋 Please visit this URL to authorize the application:")
	fmt.Println()
	fmt.Printf("🔗 %s\n", authURL)
	fmt.Println()
	fmt.Println("📖 What will happen:")
	fmt.Println("1. Click the link above or copy it to your browser")
	fmt.Println("2. Log in with your Google account")
	fmt.Println("3. Review and click 'Allow' or 'Grant'")
	fmt.Println("4. You'll be redirected back to this application")
	fmt.Println("5. The authorization will complete automatically")
	fmt.Println()
	fmt.Println("⏳ Waiting for authorization...")

	// Wait for either the code or an error
	select {
	case code := <-codeChan:
		// Shutdown the callback server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)

		fmt.Println("🔄 Received authorization code, exchanging for token...")

		// Exchange code for token
		tok, err := config.Exchange(context.TODO(), code)
		if err != nil {
			fmt.Println("❌ Failed to exchange authorization code for access token.")
			fmt.Println("   This usually means:")
			fmt.Println("   - The authorization code is incorrect or expired")
			fmt.Println("   - The code was already used (they can only be used once)")
			fmt.Println("   - There was a network issue")
			fmt.Println()
			fmt.Println("   Please try the authorization process again.")
			log.Fatalf("❌ Unable to retrieve token from web: %v", err)
		}

		fmt.Println("✅ Authentication successful!")
		return tok

	case err := <-errorChan:
		// Shutdown the callback server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)

		log.Fatalf("❌ OAuth callback error: %v", err)

	case <-time.After(5 * time.Minute):
		// Timeout after 5 minutes
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)

		log.Fatalf("❌ OAuth authorization timed out after 5 minutes")
	}

	return nil // This should never be reached
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
	fmt.Printf("💾 Token stored to %s\n", path)

	// Create the file with secure permissions (read/write for owner only)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("❌ Unable to cache oauth token: %v", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		log.Fatalf("❌ Unable to encode token to file: %v", err)
	}

	fmt.Println()
	fmt.Println("⚠️  SECURITY WARNING:")
	fmt.Println("This token.json file contains the key that allows the server to access your Google account without asking again.")
	fmt.Println("Protect it like a password. Do not commit it to GitHub.")
	fmt.Println("The included .gitignore file should prevent this automatically.")
	fmt.Println()
}
