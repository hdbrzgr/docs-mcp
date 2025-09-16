package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/hdbrzgr/docs-mcp/services"
	"github.com/hdbrzgr/docs-mcp/util"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

// Input types for document tools
type CreateDocumentInput struct {
	Title string `json:"title" validate:"required"`
}

type GetDocumentInput struct {
	DocumentID string `json:"document_id" validate:"required"`
}

type ListDocumentsInput struct {
	Query    string `json:"query,omitempty"`
	MaxCount int64  `json:"max_count,omitempty"`
}

type DeleteDocumentInput struct {
	DocumentID string `json:"document_id" validate:"required"`
}

type CopyDocumentInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	NewTitle   string `json:"new_title" validate:"required"`
}

type ShareDocumentInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	Email      string `json:"email" validate:"required"`
	Role       string `json:"role,omitempty"` // reader, writer, commenter
	Type       string `json:"type,omitempty"` // user, group, domain, anyone
}

func RegisterDocumentTools(s *server.MCPServer) {
	// Create document tool
	createDocTool := mcp.NewTool("create_document",
		mcp.WithDescription("Create a new Google Docs document with the specified title"),
		mcp.WithString("title", mcp.Required(), mcp.Description("The title of the new document")),
	)
	s.AddTool(createDocTool, mcp.NewTypedToolHandler(createDocumentHandler))

	// Get document tool
	getDocTool := mcp.NewTool("get_document",
		mcp.WithDescription("Retrieve detailed information about a specific Google Docs document including its content, structure, and metadata"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the Google Docs document")),
	)
	s.AddTool(getDocTool, mcp.NewTypedToolHandler(getDocumentHandler))

	// List documents tool
	listDocsTool := mcp.NewTool("list_documents",
		mcp.WithDescription("List Google Docs documents accessible to the authenticated user. Can filter by query and limit results"),
		mcp.WithString("query", mcp.Description("Search query to filter documents (e.g., 'name contains \"report\"', 'modifiedTime > \"2023-01-01\"')")),
		mcp.WithNumber("max_count", mcp.Description("Maximum number of documents to return (default: 10, max: 100)")),
	)
	s.AddTool(listDocsTool, mcp.NewTypedToolHandler(listDocumentsHandler))

	// Delete document tool
	deleteDocTool := mcp.NewTool("delete_document",
		mcp.WithDescription("Move a Google Docs document to trash. The document can be restored from trash if needed"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document to delete")),
	)
	s.AddTool(deleteDocTool, mcp.NewTypedToolHandler(deleteDocumentHandler))

	// Copy document tool
	copyDocTool := mcp.NewTool("copy_document",
		mcp.WithDescription("Create a copy of an existing Google Docs document with a new title"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document to copy")),
		mcp.WithString("new_title", mcp.Required(), mcp.Description("The title for the copied document")),
	)
	s.AddTool(copyDocTool, mcp.NewTypedToolHandler(copyDocumentHandler))

	// Share document tool
	shareDocTool := mcp.NewTool("share_document",
		mcp.WithDescription("Share a Google Docs document with a user or group by email address"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document to share")),
		mcp.WithString("email", mcp.Required(), mcp.Description("Email address of the user or group to share with")),
		mcp.WithString("role", mcp.Description("Permission role: 'reader', 'writer', or 'commenter' (default: 'reader')")),
		mcp.WithString("type", mcp.Description("Type of permission: 'user', 'group', 'domain', or 'anyone' (default: 'user')")),
	)
	s.AddTool(shareDocTool, mcp.NewTypedToolHandler(shareDocumentHandler))
}

func createDocumentHandler(ctx context.Context, request mcp.CallToolRequest, input CreateDocumentInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	// Create a new document
	doc := &docs.Document{
		Title: input.Title,
	}

	createdDoc, err := docsService.Documents.Create(doc).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("create document", err), nil
	}

	result := fmt.Sprintf("Document created successfully!\n\nTitle: %s\nDocument ID: %s\nURL: https://docs.google.com/document/d/%s/edit",
		createdDoc.Title, createdDoc.DocumentId, createdDoc.DocumentId)

	return mcp.NewToolResultText(result), nil
}

func getDocumentHandler(ctx context.Context, request mcp.CallToolRequest, input GetDocumentInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	doc, err := docsService.Documents.Get(input.DocumentID).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("get document", err), nil
	}

	// Format the document using the utility function
	formattedDoc := util.FormatGoogleDoc(doc)

	return mcp.NewToolResultText(formattedDoc), nil
}

func listDocumentsHandler(ctx context.Context, request mcp.CallToolRequest, input ListDocumentsInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	// Build the query
	query := "mimeType='application/vnd.google-apps.document'"
	if input.Query != "" {
		query += " and " + input.Query
	}

	// Set default max count if not provided
	maxCount := input.MaxCount
	if maxCount <= 0 {
		maxCount = 10
	}
	if maxCount > 100 {
		maxCount = 100
	}

	// List documents
	filesList, err := driveService.Files.List().
		Q(query).
		PageSize(maxCount).
		Fields("files(id,name,mimeType,createdTime,modifiedTime,owners,size,webViewLink)").
		Context(ctx).
		Do()

	if err != nil {
		return util.HandleGoogleAPIError("list documents", err), nil
	}

	if len(filesList.Files) == 0 {
		return mcp.NewToolResultText("No documents found matching the criteria."), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d Google Docs documents:\n\n", len(filesList.Files)))

	for i, file := range filesList.Files {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, util.FormatDriveFile(file)))
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func deleteDocumentHandler(ctx context.Context, request mcp.CallToolRequest, input DeleteDocumentInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	// Move the document to trash
	_, err := driveService.Files.Update(input.DocumentID, &drive.File{
		Trashed: true,
	}).Context(ctx).Do()

	if err != nil {
		return util.HandleGoogleAPIError("delete document", err), nil
	}

	result := fmt.Sprintf("Document moved to trash successfully!\n\nDocument ID: %s\n\nNote: The document can be restored from Google Drive trash if needed.",
		input.DocumentID)

	return mcp.NewToolResultText(result), nil
}

func copyDocumentHandler(ctx context.Context, request mcp.CallToolRequest, input CopyDocumentInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	// Copy the document
	copiedFile, err := driveService.Files.Copy(input.DocumentID, &drive.File{
		Name: input.NewTitle,
	}).Context(ctx).Do()

	if err != nil {
		return util.HandleGoogleAPIError("copy document", err), nil
	}

	result := fmt.Sprintf("Document copied successfully!\n\nOriginal Document ID: %s\nNew Document ID: %s\nNew Title: %s\nURL: https://docs.google.com/document/d/%s/edit",
		input.DocumentID, copiedFile.Id, copiedFile.Name, copiedFile.Id)

	return mcp.NewToolResultText(result), nil
}

func shareDocumentHandler(ctx context.Context, request mcp.CallToolRequest, input ShareDocumentInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	// Set default values
	role := input.Role
	if role == "" {
		role = "reader"
	}

	permissionType := input.Type
	if permissionType == "" {
		permissionType = "user"
	}

	// Validate role
	validRoles := map[string]bool{
		"reader":     true,
		"writer":     true,
		"commenter":  true,
	}
	if !validRoles[role] {
		return mcp.NewToolResultText("Error: Invalid role. Must be 'reader', 'writer', or 'commenter'."), nil
	}

	// Validate type
	validTypes := map[string]bool{
		"user":   true,
		"group":  true,
		"domain": true,
		"anyone": true,
	}
	if !validTypes[permissionType] {
		return mcp.NewToolResultText("Error: Invalid type. Must be 'user', 'group', 'domain', or 'anyone'."), nil
	}

	// Create permission
	permission := &drive.Permission{
		Role:         role,
		Type:         permissionType,
		EmailAddress: input.Email,
	}

	// Add the permission
	createdPermission, err := driveService.Permissions.Create(input.DocumentID, permission).
		SendNotificationEmail(true).
		Context(ctx).
		Do()

	if err != nil {
		return util.HandleGoogleAPIError("share document", err), nil
	}

	result := fmt.Sprintf("Document shared successfully!\n\nDocument ID: %s\nShared with: %s\nRole: %s\nPermission ID: %s",
		input.DocumentID, input.Email, role, createdPermission.Id)

	return mcp.NewToolResultText(result), nil
}
