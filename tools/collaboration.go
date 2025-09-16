package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/hdbrzgr/docs-mcp/services"
	"github.com/hdbrzgr/docs-mcp/util"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/api/drive/v3"
)

// Input types for collaboration tools
type CreateCommentInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	StartIndex int64  `json:"start_index" validate:"required"`
	EndIndex   int64  `json:"end_index" validate:"required"`
	Comment    string `json:"comment" validate:"required"`
}

type ReplyToCommentInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	CommentID  string `json:"comment_id" validate:"required"`
	Reply      string `json:"reply" validate:"required"`
}

type ListCommentsInput struct {
	DocumentID string `json:"document_id" validate:"required"`
}

type ResolveCommentInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	CommentID  string `json:"comment_id" validate:"required"`
}

type GetPermissionsInput struct {
	DocumentID string `json:"document_id" validate:"required"`
}

type UpdatePermissionInput struct {
	DocumentID   string `json:"document_id" validate:"required"`
	PermissionID string `json:"permission_id" validate:"required"`
	Role         string `json:"role" validate:"required"` // reader, writer, commenter
}

type RemovePermissionInput struct {
	DocumentID   string `json:"document_id" validate:"required"`
	PermissionID string `json:"permission_id" validate:"required"`
}

type CreateSuggestionInput struct {
	DocumentID      string `json:"document_id" validate:"required"`
	StartIndex      int64  `json:"start_index" validate:"required"`
	EndIndex        int64  `json:"end_index" validate:"required"`
	SuggestedText   string `json:"suggested_text" validate:"required"`
	SuggestionType  string `json:"suggestion_type,omitempty"` // REPLACE_TEXT, DELETE_TEXT, INSERT_TEXT
}

func RegisterCollaborationTools(s *server.MCPServer) {
	// Create comment tool
	createCommentTool := mcp.NewTool("create_comment",
		mcp.WithDescription("Create a comment on a specific range of text in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("Start position of the text to comment on")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("End position of the text to comment on")),
		mcp.WithString("comment", mcp.Required(), mcp.Description("The comment text")),
	)
	s.AddTool(createCommentTool, mcp.NewTypedToolHandler(createCommentHandler))

	// Reply to comment tool
	replyToCommentTool := mcp.NewTool("reply_to_comment",
		mcp.WithDescription("Reply to an existing comment in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithString("comment_id", mcp.Required(), mcp.Description("The ID of the comment to reply to")),
		mcp.WithString("reply", mcp.Required(), mcp.Description("The reply text")),
	)
	s.AddTool(replyToCommentTool, mcp.NewTypedToolHandler(replyToCommentHandler))

	// List comments tool
	listCommentsTool := mcp.NewTool("list_comments",
		mcp.WithDescription("List all comments in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
	)
	s.AddTool(listCommentsTool, mcp.NewTypedToolHandler(listCommentsHandler))

	// Resolve comment tool
	resolveCommentTool := mcp.NewTool("resolve_comment",
		mcp.WithDescription("Resolve (mark as done) a comment in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithString("comment_id", mcp.Required(), mcp.Description("The ID of the comment to resolve")),
	)
	s.AddTool(resolveCommentTool, mcp.NewTypedToolHandler(resolveCommentHandler))

	// Get permissions tool
	getPermissionsTool := mcp.NewTool("get_permissions",
		mcp.WithDescription("Get all permissions (sharing settings) for a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
	)
	s.AddTool(getPermissionsTool, mcp.NewTypedToolHandler(getPermissionsHandler))

	// Update permission tool
	updatePermissionTool := mcp.NewTool("update_permission",
		mcp.WithDescription("Update the role of an existing permission for a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithString("permission_id", mcp.Required(), mcp.Description("The ID of the permission to update")),
		mcp.WithString("role", mcp.Required(), mcp.Description("New role: 'reader', 'writer', or 'commenter'")),
	)
	s.AddTool(updatePermissionTool, mcp.NewTypedToolHandler(updatePermissionHandler))

	// Remove permission tool
	removePermissionTool := mcp.NewTool("remove_permission",
		mcp.WithDescription("Remove a permission (stop sharing) from a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithString("permission_id", mcp.Required(), mcp.Description("The ID of the permission to remove")),
	)
	s.AddTool(removePermissionTool, mcp.NewTypedToolHandler(removePermissionHandler))

	// Create suggestion tool
	createSuggestionTool := mcp.NewTool("create_suggestion",
		mcp.WithDescription("Create a suggestion for text changes in a Google Docs document (suggestion mode)"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("Start position of the text to suggest changes for")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("End position of the text to suggest changes for")),
		mcp.WithString("suggested_text", mcp.Required(), mcp.Description("The suggested replacement text")),
		mcp.WithString("suggestion_type", mcp.Description("Type of suggestion: 'REPLACE_TEXT', 'DELETE_TEXT', or 'INSERT_TEXT' (default: 'REPLACE_TEXT')")),
	)
	s.AddTool(createSuggestionTool, mcp.NewTypedToolHandler(createSuggestionHandler))
}

func createCommentHandler(ctx context.Context, request mcp.CallToolRequest, input CreateCommentInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	if input.StartIndex >= input.EndIndex {
		return mcp.NewToolResultText("Error: Start index must be less than end index."), nil
	}

	// Create a comment with an anchor to the specified range
	comment := &drive.Comment{
		Content: input.Comment,
		Anchor: fmt.Sprintf("kix.%d:%d", input.StartIndex, input.EndIndex),
	}

	createdComment, err := driveService.Comments.Create(input.DocumentID, comment).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("create comment", err), nil
	}

	result := fmt.Sprintf("Comment created successfully!\n\nDocument ID: %s\nComment ID: %s\nRange: %d-%d\nComment: %s\nAuthor: %s",
		input.DocumentID, createdComment.Id, input.StartIndex, input.EndIndex, input.Comment, createdComment.Author.DisplayName)

	return mcp.NewToolResultText(result), nil
}

func replyToCommentHandler(ctx context.Context, request mcp.CallToolRequest, input ReplyToCommentInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	// Create a reply to the comment
	reply := &drive.Reply{
		Content: input.Reply,
	}

	createdReply, err := driveService.Replies.Create(input.DocumentID, input.CommentID, reply).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("reply to comment", err), nil
	}

	result := fmt.Sprintf("Reply created successfully!\n\nDocument ID: %s\nComment ID: %s\nReply ID: %s\nReply: %s\nAuthor: %s",
		input.DocumentID, input.CommentID, createdReply.Id, input.Reply, createdReply.Author.DisplayName)

	return mcp.NewToolResultText(result), nil
}

func listCommentsHandler(ctx context.Context, request mcp.CallToolRequest, input ListCommentsInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	commentsList, err := driveService.Comments.List(input.DocumentID).
		Fields("comments(id,content,author,createdTime,resolved,anchor,replies)").
		Context(ctx).
		Do()

	if err != nil {
		return util.HandleGoogleAPIError("list comments", err), nil
	}

	if len(commentsList.Comments) == 0 {
		return mcp.NewToolResultText("No comments found in this document."), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d comments in the document:\n\n", len(commentsList.Comments)))

	for i, comment := range commentsList.Comments {
		result.WriteString(fmt.Sprintf("%d. Comment ID: %s\n", i+1, comment.Id))
		result.WriteString(fmt.Sprintf("   Author: %s\n", comment.Author.DisplayName))
		result.WriteString(fmt.Sprintf("   Created: %s\n", comment.CreatedTime))
		result.WriteString(fmt.Sprintf("   Content: %s\n", comment.Content))
		result.WriteString(fmt.Sprintf("   Resolved: %t\n", comment.Resolved))
		
		if comment.Anchor != "" {
			result.WriteString(fmt.Sprintf("   Anchor: %s\n", comment.Anchor))
		}

		if len(comment.Replies) > 0 {
			result.WriteString(fmt.Sprintf("   Replies (%d):\n", len(comment.Replies)))
			for j, reply := range comment.Replies {
				result.WriteString(fmt.Sprintf("     %d. %s (%s): %s\n", j+1, reply.Author.DisplayName, reply.CreatedTime, reply.Content))
			}
		}
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func resolveCommentHandler(ctx context.Context, request mcp.CallToolRequest, input ResolveCommentInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	// Update the comment to mark it as resolved
	comment := &drive.Comment{
		Resolved: true,
	}

	updatedComment, err := driveService.Comments.Update(input.DocumentID, input.CommentID, comment).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("resolve comment", err), nil
	}

	result := fmt.Sprintf("Comment resolved successfully!\n\nDocument ID: %s\nComment ID: %s\nResolved: %t",
		input.DocumentID, input.CommentID, updatedComment.Resolved)

	return mcp.NewToolResultText(result), nil
}

func getPermissionsHandler(ctx context.Context, request mcp.CallToolRequest, input GetPermissionsInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	permissionsList, err := driveService.Permissions.List(input.DocumentID).
		Fields("permissions(id,type,role,emailAddress,displayName,domain)").
		Context(ctx).
		Do()

	if err != nil {
		return util.HandleGoogleAPIError("get permissions", err), nil
	}

	if len(permissionsList.Permissions) == 0 {
		return mcp.NewToolResultText("No permissions found for this document."), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Document permissions (%d total):\n\n", len(permissionsList.Permissions)))

	for i, permission := range permissionsList.Permissions {
		result.WriteString(fmt.Sprintf("%d. Permission ID: %s\n", i+1, permission.Id))
		result.WriteString(fmt.Sprintf("   Type: %s\n", permission.Type))
		result.WriteString(fmt.Sprintf("   Role: %s\n", permission.Role))
		
		if permission.EmailAddress != "" {
			result.WriteString(fmt.Sprintf("   Email: %s\n", permission.EmailAddress))
		}
		
		if permission.DisplayName != "" {
			result.WriteString(fmt.Sprintf("   Name: %s\n", permission.DisplayName))
		}
		
		if permission.Domain != "" {
			result.WriteString(fmt.Sprintf("   Domain: %s\n", permission.Domain))
		}
		
		result.WriteString("\n")
	}

	return mcp.NewToolResultText(result.String()), nil
}

func updatePermissionHandler(ctx context.Context, request mcp.CallToolRequest, input UpdatePermissionInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	// Validate role
	validRoles := map[string]bool{
		"reader":    true,
		"writer":    true,
		"commenter": true,
	}
	if !validRoles[input.Role] {
		return mcp.NewToolResultText("Error: Invalid role. Must be 'reader', 'writer', or 'commenter'."), nil
	}

	// Update the permission
	permission := &drive.Permission{
		Role: input.Role,
	}

	updatedPermission, err := driveService.Permissions.Update(input.DocumentID, input.PermissionID, permission).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("update permission", err), nil
	}

	result := fmt.Sprintf("Permission updated successfully!\n\nDocument ID: %s\nPermission ID: %s\nNew Role: %s",
		input.DocumentID, input.PermissionID, updatedPermission.Role)

	return mcp.NewToolResultText(result), nil
}

func removePermissionHandler(ctx context.Context, request mcp.CallToolRequest, input RemovePermissionInput) (*mcp.CallToolResult, error) {
	driveService := services.GoogleDriveClient()

	err := driveService.Permissions.Delete(input.DocumentID, input.PermissionID).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("remove permission", err), nil
	}

	result := fmt.Sprintf("Permission removed successfully!\n\nDocument ID: %s\nRemoved Permission ID: %s",
		input.DocumentID, input.PermissionID)

	return mcp.NewToolResultText(result), nil
}

func createSuggestionHandler(ctx context.Context, request mcp.CallToolRequest, input CreateSuggestionInput) (*mcp.CallToolResult, error) {
	// Note: Google Docs API doesn't directly support creating suggestions via API
	// This is a limitation of the current API. Suggestions are typically created
	// through the web interface when in "Suggesting" mode.
	
	// As a workaround, we can create a comment that describes the suggested change
	driveService := services.GoogleDriveClient()

	if input.StartIndex >= input.EndIndex {
		return mcp.NewToolResultText("Error: Start index must be less than end index."), nil
	}

	suggestionType := input.SuggestionType
	if suggestionType == "" {
		suggestionType = "REPLACE_TEXT"
	}

	// Validate suggestion type
	validTypes := map[string]bool{
		"REPLACE_TEXT": true,
		"DELETE_TEXT":  true,
		"INSERT_TEXT":  true,
	}
	if !validTypes[suggestionType] {
		return mcp.NewToolResultText("Error: Invalid suggestion type. Must be 'REPLACE_TEXT', 'DELETE_TEXT', or 'INSERT_TEXT'."), nil
	}

	// Create a comment that describes the suggestion
	var commentText string
	switch suggestionType {
	case "REPLACE_TEXT":
		commentText = fmt.Sprintf("SUGGESTION: Replace with '%s'", input.SuggestedText)
	case "DELETE_TEXT":
		commentText = "SUGGESTION: Delete this text"
	case "INSERT_TEXT":
		commentText = fmt.Sprintf("SUGGESTION: Insert '%s' here", input.SuggestedText)
	}

	comment := &drive.Comment{
		Content: commentText,
		Anchor:  fmt.Sprintf("kix.%d:%d", input.StartIndex, input.EndIndex),
	}

	createdComment, err := driveService.Comments.Create(input.DocumentID, comment).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("create suggestion comment", err), nil
	}

	result := fmt.Sprintf("Suggestion created successfully!\n\nNote: Google Docs API doesn't directly support suggestions, so this was created as a comment.\n\nDocument ID: %s\nComment ID: %s\nRange: %d-%d\nSuggestion Type: %s\nSuggested Text: %s\nComment: %s",
		input.DocumentID, createdComment.Id, input.StartIndex, input.EndIndex, suggestionType, input.SuggestedText, commentText)

	return mcp.NewToolResultText(result), nil
}
