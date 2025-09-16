# Google Docs MCP Server Development Commands

# Default recipe - show available commands
default:
    @just --list

# Build the project
build:
    go build -o docs-mcp .

# Run in development mode with environment file
dev:
    go run main.go -env .env -http_port 3000

# Run in stdio mode for MCP testing
stdio:
    go run main.go -env .env

# First-time OAuth setup (one-time authorization)
oauth-setup:
    @echo "🔐 Starting OAuth setup for Google Docs MCP Server..."
    @echo "This will guide you through the one-time authorization process."
    @echo ""
    @echo "📋 Prerequisites:"
    @echo "1. Make sure you have GOOGLE_CLIENT_SECRETS environment variable set"
    @echo "2. Or create a .env file with GOOGLE_CLIENT_SECRETS=/path/to/client-secrets.json"
    @echo ""
    @echo "🚀 Starting server for OAuth authorization..."
    go run main.go -env .env

# OAuth setup with callback server (automatic redirect)
oauth-callback:
    @echo "🔐 Starting OAuth setup with callback server..."
    @echo "This will use automatic browser redirect for easier authorization."
    @echo ""
    @echo "📋 Prerequisites:"
    @echo "1. Make sure you have GOOGLE_CLIENT_SECRETS environment variable set"
    @echo "2. Or create a .env file with GOOGLE_CLIENT_SECRETS=/path/to/client-secrets.json"
    @echo ""
    @echo "🚀 Starting server with OAuth callback..."
    OAUTH_USE_CALLBACK=true go run main.go -env .env

# Run tests
test:
    go test ./...

# Run tests with coverage
test-coverage:
    go test -cover ./...

# Format code
fmt:
    go fmt ./...

# Lint code
lint:
    golangci-lint run

# Clean build artifacts
clean:
    rm -f docs-mcp
    go clean

# Install dependencies
deps:
    go mod download
    go mod tidy

# Build Docker image
docker-build:
    docker build -t docs-mcp:latest .

# Run Docker container with service account
docker-run-sa credentials_path:
    docker run --rm -it \
        -v {{credentials_path}}:/credentials \
        -e GOOGLE_APPLICATION_CREDENTIALS=/credentials/service-account-key.json \
        -p 8080:8080 \
        docs-mcp:latest \
        --http_port 8080

# Run Docker container with OAuth client secrets
docker-run-oauth credentials_path:
    docker run --rm -it \
        -v {{credentials_path}}:/credentials \
        -e GOOGLE_CLIENT_SECRETS=/credentials/client-secrets.json \
        -p 8080:8080 \
        docs-mcp:latest \
        --http_port 8080

# Test MCP connection (requires npx and @modelcontextprotocol/inspector)
test-mcp:
    npx @modelcontextprotocol/inspector http://localhost:3000/mcp

# Show project structure
tree:
    tree -I 'node_modules|.git|vendor'

# Create a new release build
release:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o docs-mcp-linux-amd64 .
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -installsuffix cgo -o docs-mcp-darwin-amd64 .
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -installsuffix cgo -o docs-mcp-windows-amd64.exe .
