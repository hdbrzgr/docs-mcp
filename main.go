package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hdbrzgr/docs-mcp/services"
	"github.com/hdbrzgr/docs-mcp/tools"
	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	envFile := flag.String("env", "", "Path to environment file (optional when environment variables are set directly)")
	httpPort := flag.String("http_port", "", "Port for HTTP server. If not provided, will use stdio")

	// Add usage information for environment variables
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] [KEY=VALUE ...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  You can also pass environment variables as arguments in KEY=VALUE format.\n")
		fmt.Fprintf(os.Stderr, "  Example: %s GOOGLE_APPLICATION_CREDENTIALS=/path/to/creds.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Example: %s -http_port=8080 GOOGLE_CLIENT_SECRETS=/path/to/secrets.json\n", os.Args[0])
	}

	// Parse environment variables from command line arguments FIRST
	parseEnvArgs()

	// Then parse flags
	flag.Parse()

	// Load environment file if specified
	if *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			fmt.Printf("⚠️  Warning: Error loading env file %s: %v\n", *envFile, err)
		} else {
			fmt.Printf("✅ Loaded environment variables from %s\n", *envFile)
		}
	}

	// Check required environment variables
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	clientSecretsPath := os.Getenv("GOOGLE_CLIENT_SECRETS")

	missingEnvs := []string{}

	// Check authentication: either service account credentials or OAuth client secrets
	hasServiceAccount := credentialsPath != ""
	hasClientSecrets := clientSecretsPath != ""

	if !hasServiceAccount && !hasClientSecrets {
		if credentialsPath == "" {
			missingEnvs = append(missingEnvs, "GOOGLE_APPLICATION_CREDENTIALS (for service account auth)")
		}
		if clientSecretsPath == "" {
			missingEnvs = append(missingEnvs, "GOOGLE_CLIENT_SECRETS (for OAuth client auth)")
		}
	}

	if len(missingEnvs) > 0 {
		fmt.Println("❌ Configuration Error: Missing required environment variables")
		fmt.Println()
		fmt.Println("Missing variables:")
		for _, env := range missingEnvs {
			fmt.Printf("  - %s\n", env)
		}
		fmt.Println()
		fmt.Println("📋 Setup Instructions:")
		fmt.Println("Choose one of the following authentication methods:")
		fmt.Println()
		fmt.Println("🔑 Method 1: Service Account (Recommended for server applications)")
		fmt.Println("1. Go to Google Cloud Console > IAM & Admin > Service Accounts")
		fmt.Println("2. Create a new service account or use an existing one")
		fmt.Println("3. Enable Google Docs API for your project")
		fmt.Println("4. Generate and download a JSON key file")
		fmt.Println("5. Set the environment variable:")
		fmt.Println("   GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json")
		fmt.Println()
		fmt.Println("🔑 Method 2: OAuth Client Secrets (For user-based access)")
		fmt.Println("1. Go to Google Cloud Console > APIs & Services > Credentials")
		fmt.Println("2. Create OAuth 2.0 Client ID credentials")
		fmt.Println("3. Download the client secrets JSON file")
		fmt.Println("4. Set the environment variable:")
		fmt.Println("   GOOGLE_CLIENT_SECRETS=/path/to/client-secrets.json")
		fmt.Println()
		fmt.Println("📁 Configuration Options:")
		fmt.Println("   Option A - Using .env file:")
		fmt.Println("   Create a .env file with one of the authentication methods above")
		fmt.Println()
		fmt.Println("   Option B - Using environment variables:")
		fmt.Println("   # For service account:")
		fmt.Println("   export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json")
		fmt.Println("   # OR for OAuth client:")
		fmt.Println("   export GOOGLE_CLIENT_SECRETS=/path/to/client-secrets.json")
		fmt.Println()
		fmt.Println("   Option C - Using command line arguments:")
		fmt.Println("   # For service account:")
		fmt.Println("   ./docs-mcp GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json")
		fmt.Println("   # OR for OAuth client:")
		fmt.Println("   ./docs-mcp GOOGLE_CLIENT_SECRETS=/path/to/client-secrets.json")
		fmt.Println("   # You can also combine with other flags:")
		fmt.Println("   ./docs-mcp -http_port=8080 GOOGLE_APPLICATION_CREDENTIALS=/path/to/creds.json")
		fmt.Println()
		fmt.Println("   Option D - Using Docker:")
		fmt.Println("   # For service account:")
		fmt.Printf("   docker run -v /path/to/credentials:/credentials \\\n")
		fmt.Printf("              -e GOOGLE_APPLICATION_CREDENTIALS=/credentials/service-account-key.json \\\n")
		fmt.Printf("              ghcr.io/hdbrzgr/docs-mcp:latest\n")
		fmt.Println()
		os.Exit(1)
	}

	fmt.Println("✅ All required environment variables are set")

	// Show which authentication method is being used
	if hasServiceAccount {
		fmt.Println("🔑 Using Service Account authentication")
		fmt.Printf("📄 Credentials file: %s\n", credentialsPath)
	} else {
		fmt.Println("🔑 Using OAuth Client Secrets authentication")
		fmt.Printf("📄 Client secrets file: %s\n", clientSecretsPath)

		// Check if this is a first run (no token.json file exists)
		tokenPath := os.Getenv("GOOGLE_TOKEN_PATH")
		if tokenPath == "" {
			tokenPath = "token.json"
		}

		if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
			fmt.Println()
			fmt.Println("🚀 First Run Detected!")
			fmt.Println("This appears to be your first time running the server with OAuth authentication.")
			fmt.Println("You will need to authorize the app to access your Google account.")
			fmt.Println()
			fmt.Println("📋 What will happen next:")
			fmt.Println("1. The server will start a temporary callback server")
			fmt.Println("2. You'll be prompted to visit a Google authorization URL")
			fmt.Println("3. You'll log in and grant permissions in your browser")
			fmt.Println("4. Google will redirect back to the callback server automatically")
			fmt.Println("5. A token.json file will be created for future use")
			fmt.Println()
			fmt.Println("⚠️  Important: This is a one-time setup process.")
			fmt.Println("   After this, the server will use the saved token automatically.")
			fmt.Println()

			// Enable callback mode for OAuth
			os.Setenv("OAUTH_USE_CALLBACK", "true")
		}
	}

	mcpServer := server.NewMCPServer(
		"Google Docs MCP",
		"1.0.0",
		server.WithLogging(),
		server.WithPromptCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithRecovery(),
	)

	// Register available Google Docs tools
	tools.RegisterDocumentTools(mcpServer)
	tools.RegisterContentTools(mcpServer)
	tools.RegisterFormattingTools(mcpServer)
	tools.RegisterStructureTools(mcpServer)
	tools.RegisterCollaborationTools(mcpServer)
	tools.RegisterRevisionTools(mcpServer)

	if *httpPort != "" {
		fmt.Println()
		fmt.Println("🚀 Starting Google Docs MCP Server in HTTP mode...")
		fmt.Printf("📡 Server will be available at: http://localhost:%s/mcp\n", *httpPort)
		fmt.Println()
		fmt.Println("📋 Cursor Configuration:")
		fmt.Println("Add the following to your Cursor MCP settings (.cursor/mcp.json):")
		fmt.Println()
		fmt.Println("```json")
		fmt.Println("{")
		fmt.Println("  \"mcpServers\": {")
		fmt.Println("    \"docs\": {")
		fmt.Printf("      \"url\": \"http://localhost:%s/mcp\"\n", *httpPort)
		fmt.Println("    }")
		fmt.Println("  }")
		fmt.Println("}")
		fmt.Println("```")
		fmt.Println()
		fmt.Println("💡 Tips:")
		fmt.Println("- Restart Cursor after adding the configuration")
		fmt.Println("- Test the connection by asking Claude: 'List my Google Docs'")
		fmt.Println("- Use '@docs' in Cursor to reference Google Docs-related context")
		fmt.Println()
		fmt.Println("🔄 Server starting...")

		httpServer := server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath("/mcp"))
		if err := httpServer.Start(fmt.Sprintf(":%s", *httpPort)); err != nil && !isContextCanceled(err) {
			log.Fatalf("❌ Server error: %v", err)
		}
	} else {
		fmt.Println()
		fmt.Println("🚀 Starting Google Docs MCP Server in stdio mode...")
		fmt.Println("📡 Server is ready and awaiting MCP client connection via stdio")
		fmt.Println()
		fmt.Println("💡 Tips:")
		fmt.Println("- The server is now running and ready to handle MCP requests")
		fmt.Println("- You can stop the server with Ctrl+C")
		fmt.Println("- If this was your first run, check that token.json was created successfully")
		fmt.Println()
		services.GoogleDriveClient()
		if err := server.ServeStdio(mcpServer); err != nil && !isContextCanceled(err) {
			log.Fatalf("❌ Server error: %v", err)
		}
	}
}

// parseEnvArgs parses environment variables from command line arguments
// Arguments should be in the format: KEY=VALUE
func parseEnvArgs() {
	args := os.Args[1:] // Skip the program name

	for _, arg := range args {
		// Skip flags that start with - or --
		if strings.HasPrefix(arg, "-") {
			continue
		}

		// Check if argument contains an equals sign (KEY=VALUE format)
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Only set if not already set in environment
				if os.Getenv(key) == "" {
					os.Setenv(key, value)
					fmt.Printf("✅ Set environment variable from argument: %s\n", key)
				}
			}
		}
	}
}

// isContextCanceled checks if the error is related to context cancellation
func isContextCanceled(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's directly context.Canceled
	if errors.Is(err, context.Canceled) {
		return true
	}

	// Check if the error message contains context canceled
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "context canceled") ||
		strings.Contains(errMsg, "operation was canceled") ||
		strings.Contains(errMsg, "context deadline exceeded")
}
