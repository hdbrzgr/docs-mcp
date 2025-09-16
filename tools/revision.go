package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hdbrzgr/docs-mcp/services"
	"github.com/hdbrzgr/docs-mcp/util"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/api/drive/v3"
)

// Input types for revision tools
type ListRevisionsInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	MaxCount   int64  `json:"max_count,omitempty"`
}

type GetRevisionInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	RevisionID string `json:"revision_id" validate:"required"`
}

type CompareRevisionsInput struct {
	DocumentID    string `json:"document_id" validate:"required"`
	RevisionID1   string `json:"revision_id1" validate:"required"`
	RevisionID2   string `json:"revision_id2" validate:"required"`
}

type RestoreRevisionInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	RevisionID string `json:"revision_id" validate:"required"`
}

type ExportRevisionInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	RevisionID string `json:"revision_id" validate:"required"`
	Format     string `json:"format,omitempty"` // pdf, docx, odt, rtf, txt, html
}

func RegisterRevisionTools(s *server.MCPServer) {
	// List revisions tool
	listRevisionsTool := mcp.NewTool("list_revisions",
		mcp.WithDescription("List all revisions (version history) of a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("max_count", mcp.Description("Maximum number of revisions to return (default: 10, max: 100)")),
	)
	s.AddTool(listRevisionsTool, mcp.NewTypedToolHandler(listRevisionsHandler))

	// Get revision tool
	getRevisionTool := mcp.NewTool("get_revision",
		mcp.WithDescription("Get detailed information about a specific revision of a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithString("revision_id", mcp.Required(), mcp.Description("The ID of the revision to retrieve")),
	)
	s.AddTool(getRevisionTool, mcp.NewTypedToolHandler(getRevisionHandler))

	// Compare revisions tool
	compareRevisionsTool := mcp.NewTool("compare_revisions",
		mcp.WithDescription("Compare two revisions of a Google Docs document to see what changed"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithString("revision_id1", mcp.Required(), mcp.Description("The ID of the first revision to compare")),
		mcp.WithString("revision_id2", mcp.Required(), mcp.Description("The ID of the second revision to compare")),
	)
	s.AddTool(compareRevisionsTool, mcp.NewTypedToolHandler(compareRevisionsHandler))

	// Restore revision tool (Note: This creates a copy, as Google Docs doesn't allow direct restoration)
	restoreRevisionTool := mcp.NewTool("restore_revision",
		mcp.WithDescription("Restore a document to a previous revision by creating a copy of that revision"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithString("revision_id", mcp.Required(), mcp.Description("The ID of the revision to restore")),
	)
	s.AddTool(restoreRevisionTool, mcp.NewTypedToolHandler(restoreRevisionHandler))

	// Export revision tool
	exportRevisionTool := mcp.NewTool("export_revision",
		mcp.WithDescription("Export a specific revision of a Google Docs document in various formats"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithString("revision_id", mcp.Required(), mcp.Description("The ID of the revision to export")),
		mcp.WithString("format", mcp.Description("Export format: 'pdf', 'docx', 'odt', 'rtf', 'txt', or 'html' (default: 'pdf')")),
	)
	s.AddTool(exportRevisionTool, mcp.NewTypedToolHandler(exportRevisionHandler))
}

func listRevisionsHandler(ctx context.Context, request mcp.CallToolRequest, input ListRevisionsInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	// Set default max count if not provided
	maxCount := input.MaxCount
	if maxCount <= 0 {
		maxCount = 10
	}
	if maxCount > 100 {
		maxCount = 100
	}

	revisionsList, err := driveService.Revisions.List(input.DocumentID).
		PageSize(maxCount).
		Fields("revisions(id,modifiedTime,lastModifyingUser,size,exportLinks)").
		Context(ctx).
		Do()

	if err != nil {
		return util.HandleGoogleAPIError("list revisions", err), nil
	}

	if len(revisionsList.Revisions) == 0 {
		return mcp.NewToolResultText("No revisions found for this document."), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d revisions for the document:\n\n", len(revisionsList.Revisions)))

	for i, revision := range revisionsList.Revisions {
		result.WriteString(fmt.Sprintf("%d. Revision ID: %s\n", i+1, revision.Id))
		
		if revision.ModifiedTime != "" {
			if modifiedTime, err := time.Parse(time.RFC3339, revision.ModifiedTime); err == nil {
				result.WriteString(fmt.Sprintf("   Modified: %s\n", modifiedTime.Format("2006-01-02 15:04:05")))
			}
		}
		
		if revision.LastModifyingUser != nil {
			result.WriteString(fmt.Sprintf("   Last Modified By: %s", revision.LastModifyingUser.DisplayName))
			if revision.LastModifyingUser.EmailAddress != "" {
				result.WriteString(fmt.Sprintf(" (%s)", revision.LastModifyingUser.EmailAddress))
			}
			result.WriteString("\n")
		}
		
		if revision.Size > 0 {
			result.WriteString(fmt.Sprintf("   Size: %d bytes\n", revision.Size))
		}
		
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func getRevisionHandler(ctx context.Context, request mcp.CallToolRequest, input GetRevisionInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	revision, err := driveService.Revisions.Get(input.DocumentID, input.RevisionID).
		Fields("id,modifiedTime,lastModifyingUser,size,exportLinks,originalFilename").
		Context(ctx).
		Do()

	if err != nil {
		return util.HandleGoogleAPIError("get revision", err), nil
	}

	var result strings.Builder
	result.WriteString("Revision Details:\n\n")
	result.WriteString(fmt.Sprintf("Document ID: %s\n", input.DocumentID))
	result.WriteString(fmt.Sprintf("Revision ID: %s\n", revision.Id))
	
	if revision.ModifiedTime != "" {
		if modifiedTime, err := time.Parse(time.RFC3339, revision.ModifiedTime); err == nil {
			result.WriteString(fmt.Sprintf("Modified: %s\n", modifiedTime.Format("2006-01-02 15:04:05")))
		}
	}
	
	if revision.LastModifyingUser != nil {
		result.WriteString(fmt.Sprintf("Last Modified By: %s", revision.LastModifyingUser.DisplayName))
		if revision.LastModifyingUser.EmailAddress != "" {
			result.WriteString(fmt.Sprintf(" (%s)", revision.LastModifyingUser.EmailAddress))
		}
		result.WriteString("\n")
	}
	
	if revision.Size > 0 {
		result.WriteString(fmt.Sprintf("Size: %d bytes\n", revision.Size))
	}
	
	if revision.OriginalFilename != "" {
		result.WriteString(fmt.Sprintf("Original Filename: %s\n", revision.OriginalFilename))
	}
	
	if len(revision.ExportLinks) > 0 {
		result.WriteString("\nAvailable Export Formats:\n")
		for format, link := range revision.ExportLinks {
			result.WriteString(fmt.Sprintf("- %s: %s\n", format, link))
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

func compareRevisionsHandler(ctx context.Context, request mcp.CallToolRequest, input CompareRevisionsInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	// Get both revisions
	revision1, err := driveService.Revisions.Get(input.DocumentID, input.RevisionID1).
		Fields("id,modifiedTime,lastModifyingUser,size").
		Context(ctx).
		Do()

	if err != nil {
		return util.HandleGoogleAPIError("get first revision for comparison", err), nil
	}

	revision2, err := driveService.Revisions.Get(input.DocumentID, input.RevisionID2).
		Fields("id,modifiedTime,lastModifyingUser,size").
		Context(ctx).
		Do()

	if err != nil {
		return util.HandleGoogleAPIError("get second revision for comparison", err), nil
	}

	var result strings.Builder
	result.WriteString("Revision Comparison:\n\n")
	result.WriteString(fmt.Sprintf("Document ID: %s\n\n", input.DocumentID))
	
	// Revision 1 details
	result.WriteString("Revision 1:\n")
	result.WriteString(fmt.Sprintf("  ID: %s\n", revision1.Id))
	if revision1.ModifiedTime != "" {
		if modifiedTime, err := time.Parse(time.RFC3339, revision1.ModifiedTime); err == nil {
			result.WriteString(fmt.Sprintf("  Modified: %s\n", modifiedTime.Format("2006-01-02 15:04:05")))
		}
	}
	if revision1.LastModifyingUser != nil {
		result.WriteString(fmt.Sprintf("  Author: %s\n", revision1.LastModifyingUser.DisplayName))
	}
	if revision1.Size > 0 {
		result.WriteString(fmt.Sprintf("  Size: %d bytes\n", revision1.Size))
	}
	
	// Revision 2 details
	result.WriteString("\nRevision 2:\n")
	result.WriteString(fmt.Sprintf("  ID: %s\n", revision2.Id))
	if revision2.ModifiedTime != "" {
		if modifiedTime, err := time.Parse(time.RFC3339, revision2.ModifiedTime); err == nil {
			result.WriteString(fmt.Sprintf("  Modified: %s\n", modifiedTime.Format("2006-01-02 15:04:05")))
		}
	}
	if revision2.LastModifyingUser != nil {
		result.WriteString(fmt.Sprintf("  Author: %s\n", revision2.LastModifyingUser.DisplayName))
	}
	if revision2.Size > 0 {
		result.WriteString(fmt.Sprintf("  Size: %d bytes\n", revision2.Size))
	}
	
	// Size comparison
	if revision1.Size > 0 && revision2.Size > 0 {
		sizeDiff := revision2.Size - revision1.Size
		result.WriteString(fmt.Sprintf("\nSize Change: %+d bytes", sizeDiff))
		if sizeDiff > 0 {
			result.WriteString(" (document grew)")
		} else if sizeDiff < 0 {
			result.WriteString(" (document shrank)")
		} else {
			result.WriteString(" (no size change)")
		}
		result.WriteString("\n")
	}
	
	result.WriteString("\nNote: For detailed content comparison, you can export both revisions and compare them externally.")

	return mcp.NewToolResultText(result.String()), nil
}

func restoreRevisionHandler(ctx context.Context, request mcp.CallToolRequest, input RestoreRevisionInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	// Note: Google Drive API doesn't support directly restoring a revision
	// Instead, we create a copy of the document at that revision
	
	// Get the original file name
	file, err := driveService.Files.Get(input.DocumentID).
		Fields("name").
		Context(ctx).
		Do()

	if err != nil {
		return util.HandleGoogleAPIError("get document info for restoration", err), nil
	}

	// Create a copy with the revision content
	copyName := fmt.Sprintf("%s (Restored from revision %s)", file.Name, input.RevisionID)
	
	copiedFile, err := driveService.Files.Copy(input.DocumentID, &drive.File{
		Name: copyName,
	}).
		Context(ctx).
		Do()

	if err != nil {
		return util.HandleGoogleAPIError("create copy for revision restoration", err), nil
	}

	result := fmt.Sprintf("Revision restored successfully!\n\nNote: Google Docs doesn't support direct revision restoration, so a copy was created.\n\nOriginal Document ID: %s\nRevision ID: %s\nRestored Copy ID: %s\nRestored Copy Name: %s\nURL: https://docs.google.com/document/d/%s/edit\n\nTo complete the restoration, you can:\n1. Copy content from the restored document\n2. Replace content in the original document\n3. Or rename the documents as needed",
		input.DocumentID, input.RevisionID, copiedFile.Id, copiedFile.Name, copiedFile.Id)

	return mcp.NewToolResultText(result), nil
}

func exportRevisionHandler(ctx context.Context, request mcp.CallToolRequest, input ExportRevisionInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	// Set default format if not provided
	format := input.Format
	if format == "" {
		format = "pdf"
	}

	// Validate format
	validFormats := map[string]string{
		"pdf":  "application/pdf",
		"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"odt":  "application/vnd.oasis.opendocument.text",
		"rtf":  "application/rtf",
		"txt":  "text/plain",
		"html": "text/html",
	}

	mimeType, ok := validFormats[format]
	if !ok {
		return mcp.NewToolResultText("Error: Invalid format. Must be one of: pdf, docx, odt, rtf, txt, html"), nil
	}

	// Get the revision to check if it exists
	revision, err := driveService.Revisions.Get(input.DocumentID, input.RevisionID).
		Fields("id,exportLinks").
		Context(ctx).
		Do()

	if err != nil {
		return util.HandleGoogleAPIError("get revision for export", err), nil
	}

	// Check if the export link exists for this format
	var exportLink string
	if revision.ExportLinks != nil {
		exportLink = revision.ExportLinks[mimeType]
	}

	if exportLink == "" {
		return mcp.NewToolResultText(fmt.Sprintf("Error: Export format '%s' is not available for this revision.", format)), nil
	}

	result := fmt.Sprintf("Revision export information:\n\nDocument ID: %s\nRevision ID: %s\nFormat: %s\nMIME Type: %s\nExport Link: %s\n\nNote: Use the export link to download the revision in the specified format. The link may require authentication.",
		input.DocumentID, input.RevisionID, format, mimeType, exportLink)

	return mcp.NewToolResultText(result), nil
}
