# Google Docs MCP Server

A comprehensive Go-based MCP (Model Context Protocol) server for Google Docs that enables AI assistants like Claude to interact with Google Docs documents. This server provides complete document management, content manipulation, formatting, collaboration, and revision control capabilities.

## 🌟 Features

### 📄 Document Management
- **Create, read, update, and delete** Google Docs documents
- **List and search** documents with advanced filtering
- **Copy documents** with custom titles
- **Share documents** with users, groups, or make them public
- **Export documents** in multiple formats (PDF, DOCX, HTML, etc.)

### ✏️ Content Manipulation
- **Insert, replace, and delete** text at specific positions
- **Append text** to documents
- **Read text content** from documents or specific ranges
- **Find and replace** text with case-sensitive options
- **Bulk text operations** for efficient document editing

### 🎨 Text Formatting
- **Bold, italic, underline** text formatting
- **Font family and size** customization
- **Text and background colors** with hex color support
- **Paragraph styles** (headings, normal text, title, subtitle)
- **Line spacing** adjustment
- **Text alignment** (left, center, right, justify)

### 🏗️ Document Structure
- **Insert tables** with custom rows and columns
- **Update table cells** with new content
- **Create lists** (bulleted and numbered)
- **Insert page breaks** and horizontal rules
- **Add images** from URLs with size control
- **Generate table of contents** from document headings

### 🤝 Collaboration Features
- **Create and manage comments** on text ranges
- **Reply to comments** and resolve discussions
- **List all comments** with author and timestamp information
- **Share documents** with specific permissions (reader, writer, commenter)
- **Manage permissions** (update roles, remove access)
- **Create suggestions** for collaborative editing

### 📚 Revision Management
- **List document revisions** with modification history
- **Get detailed revision information** including author and changes
- **Compare revisions** to see what changed between versions
- **Restore previous revisions** (creates a copy)
- **Export specific revisions** in various formats

## 🚀 Quick Start

### Prerequisites

1. **Google Cloud Project** with Google Docs API enabled
2. **Authentication credentials** (Service Account or OAuth2)
3. **Go 1.23.2+** (for building from source)
4. **Docker** (for containerized deployment)

### Step 1: Set Up Google Cloud Authentication

#### Option A: Service Account (Recommended for Server Applications)

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the **Google Docs API** and **Google Drive API**
4. Go to **IAM & Admin > Service Accounts**
5. Create a new service account
6. Generate and download a JSON key file
7. Set the environment variable:
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json
   ```

#### Option B: OAuth2 Client (For User-Based Access)

1. Go to **APIs & Services > Credentials** in Google Cloud Console
2. Create **OAuth 2.0 Client ID** credentials
3. Download the client secrets JSON file
4. Set the environment variable:
   ```bash
   export GOOGLE_CLIENT_SECRETS=/path/to/client-secrets.json
   ```

### Step 2: Installation

#### 🐳 Docker (Recommended)

```bash
# Pull the latest image
docker pull ghcr.io/hdbrzgr/docs-mcp:latest

# Run with service account authentication
docker run -d \
  -v /path/to/credentials:/credentials \
  -e GOOGLE_APPLICATION_CREDENTIALS=/credentials/service-account-key.json \
  -p 8080:8080 \
  ghcr.io/hdbrzgr/docs-mcp:latest \
  --http_port 8080

# Or run with OAuth client secrets
docker run -d \
  -v /path/to/credentials:/credentials \
  -e GOOGLE_CLIENT_SECRETS=/credentials/client-secrets.json \
  -p 8080:8080 \
  ghcr.io/hdbrzgr/docs-mcp:latest \
  --http_port 8080
```

#### 📦 Binary Installation

```bash
# Download the latest binary from GitHub Releases
curl -L -o docs-mcp https://github.com/hdbrzgr/docs-mcp/releases/latest/download/docs-mcp-linux-amd64

# Make it executable
chmod +x docs-mcp

# Move to PATH
sudo mv docs-mcp /usr/local/bin/
```

#### 🛠️ Build from Source

```bash
# Clone the repository
git clone https://github.com/hdbrzgr/docs-mcp.git
cd docs-mcp

# Build the binary
go build -o docs-mcp .

# Run the server
./docs-mcp --http_port 8080
```

### Step 3: Configure Cursor

1. Open **Cursor IDE**
2. Go to **Settings > Features > Model Context Protocol**
3. Add a new MCP server configuration:

#### For Docker/HTTP Mode:
```json
{
  "mcpServers": {
    "docs": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

#### For Binary/Stdio Mode:
```json
{
  "mcpServers": {
    "docs": {
      "command": "/usr/local/bin/docs-mcp",
      "env": {
        "GOOGLE_APPLICATION_CREDENTIALS": "/path/to/service-account-key.json"
      }
    }
  }
}
```

### Step 4: Test Your Setup

1. **Restart Cursor** completely
2. Open a new chat with Claude
3. Try these test commands:

```
List my Google Docs documents
```

```
Create a new document called "Test Document"
```

```
Add some text to the document and format it as a heading
```

## 📖 Usage Examples

### Document Management

```
# Create a new document
Create a document titled "Project Proposal"

# List recent documents
Show me my last 5 Google Docs documents

# Share a document
Share document ID "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms" with john@example.com as a writer

# Copy a document
Make a copy of document "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms" with title "Project Proposal - Copy"
```

### Content Editing

```
# Insert text at the beginning
Insert "Executive Summary\n\n" at position 1 in document "doc-id"

# Replace text in a range
Replace text from position 100 to 150 with "Updated content" in document "doc-id"

# Find and replace
Find all instances of "old text" and replace with "new text" in document "doc-id"

# Append text to the end
Add "Conclusion\n\nThis concludes our analysis." to the end of document "doc-id"
```

### Formatting

```
# Make text bold
Make text from position 50 to 100 bold in document "doc-id"

# Change text color
Set text color to red (#FF0000) for positions 200-250 in document "doc-id"

# Apply heading style
Set paragraph style to HEADING_1 for text from position 1 to 20 in document "doc-id"

# Adjust line spacing
Set line spacing to 1.5 for text from position 300 to 500 in document "doc-id"
```

### Document Structure

```
# Insert a table
Insert a 3x4 table at position 100 in document "doc-id"

# Create a bulleted list
Insert a bulleted list with items ["Item 1", "Item 2", "Item 3"] at position 200 in document "doc-id"

# Add an image
Insert image from "https://example.com/image.jpg" at position 300 in document "doc-id"

# Create table of contents
Generate a table of contents at position 1 in document "doc-id"
```

### Collaboration

```
# Add a comment
Create a comment "Please review this section" on text from position 100 to 200 in document "doc-id"

# List all comments
Show me all comments in document "doc-id"

# Reply to a comment
Reply "I've made the changes" to comment "comment-id" in document "doc-id"

# Get sharing permissions
Show me who has access to document "doc-id"
```

### Revision Management

```
# List document versions
Show me the revision history for document "doc-id"

# Compare two versions
Compare revision "rev1" with "rev2" in document "doc-id"

# Restore a previous version
Restore document "doc-id" to revision "rev-id"

# Export a specific version
Export revision "rev-id" of document "doc-id" as PDF
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to service account JSON key | Yes (Option A) |
| `GOOGLE_CLIENT_SECRETS` | Path to OAuth2 client secrets JSON | Yes (Option B) |
| `GOOGLE_TOKEN_PATH` | Path to store OAuth2 tokens | No (default: token.json) |

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--env` | Path to environment file | None |
| `--http_port` | Port for HTTP server (stdio if not specified) | None |

### Using Environment Files

Create a `.env` file:

```bash
# For Service Account
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json

# Or for OAuth2
GOOGLE_CLIENT_SECRETS=/path/to/client-secrets.json
GOOGLE_TOKEN_PATH=/path/to/token.json
```

Run with environment file:

```bash
./docs-mcp --env .env --http_port 8080
```

## 🔒 Security & Permissions

### Google Cloud Permissions

Your service account or OAuth application needs these scopes:

- `https://www.googleapis.com/auth/documents` - Read and write Google Docs
- `https://www.googleapis.com/auth/drive` - Access Google Drive for file operations

### Document Access

- **Service Account**: Documents must be shared with the service account email
- **OAuth2**: User must have access to the documents
- **Public Documents**: No additional permissions needed for public documents

### Best Practices

1. **Use Service Accounts** for server applications
2. **Limit permissions** to only what's needed
3. **Rotate credentials** regularly
4. **Monitor API usage** in Google Cloud Console
5. **Use HTTPS** in production deployments

## 🛠️ Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/hdbrzgr/docs-mcp.git
cd docs-mcp

# Install dependencies
go mod download

# Run in development mode
go run main.go --env .env --http_port 8080

# Build for production
go build -o docs-mcp .
```

### Project Structure

```
docs-mcp/
├── main.go                 # Entry point and server setup
├── services/
│   └── google.go          # Google APIs client setup
├── tools/
│   ├── document.go        # Document management tools
│   ├── content.go         # Content manipulation tools
│   ├── formatting.go      # Text formatting tools
│   ├── structure.go       # Document structure tools
│   ├── collaboration.go   # Collaboration tools
│   └── revision.go        # Revision management tools
├── util/
│   ├── formatter.go       # Document formatting utilities
│   └── errors.go          # Error handling utilities
├── go.mod                 # Go module definition
├── Dockerfile            # Container build instructions
└── README.md             # This file
```

### Running Tests

```bash
# Run unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests (requires valid credentials)
go test -tags=integration ./...
```

## 🐛 Troubleshooting

### Common Issues

**❌ "Authentication failed"**
- Check that your credentials file path is correct
- Verify the service account has the required permissions
- Ensure the Google Docs and Drive APIs are enabled

**❌ "Document not found"**
- Verify the document ID is correct
- Check that the document is shared with your service account
- Ensure the document hasn't been deleted

**❌ "Permission denied"**
- Make sure your service account has edit access to the document
- Check that the required API scopes are included in your credentials
- Verify the document owner has granted appropriate permissions

**❌ "Rate limit exceeded"**
- Implement exponential backoff in your requests
- Check your API quotas in Google Cloud Console
- Consider upgrading your Google Cloud plan if needed

### Getting Help

1. **Check the logs**: Run with `--http_port` to see detailed error messages
2. **Verify credentials**: Test authentication with Google Cloud SDK
3. **Check API quotas**: Monitor usage in Google Cloud Console
4. **Review permissions**: Ensure proper document sharing settings

### Debug Mode

Enable detailed logging:

```bash
# Set log level to debug
export LOG_LEVEL=debug
./docs-mcp --env .env --http_port 8080
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Add tests for new functionality
- Update documentation for new features
- Use semantic commit messages
- Ensure all tests pass before submitting

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Google Docs API](https://developers.google.com/docs/api) for comprehensive document management
- [MCP Go SDK](https://github.com/mark3labs/mcp-go) for MCP server implementation
- [Google API Go Client](https://github.com/googleapis/google-api-go-client) for Google APIs integration

---

**Need help?** Open an issue on GitHub or check our [troubleshooting guide](#-troubleshooting) above.
