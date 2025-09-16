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
)

// Input types for content tools
type InsertTextInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	Text       string `json:"text" validate:"required"`
	Index      int64  `json:"index,omitempty"` // Position to insert text (default: end of document)
}

type ReplaceTextInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	StartIndex int64  `json:"start_index" validate:"required"`
	EndIndex   int64  `json:"end_index" validate:"required"`
	Text       string `json:"text" validate:"required"`
}

type DeleteTextInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	StartIndex int64  `json:"start_index" validate:"required"`
	EndIndex   int64  `json:"end_index" validate:"required"`
}

type AppendTextInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	Text       string `json:"text" validate:"required"`
}

type ReadTextInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	StartIndex int64  `json:"start_index,omitempty"`
	EndIndex   int64  `json:"end_index,omitempty"`
}

type FindReplaceInput struct {
	DocumentID  string `json:"document_id" validate:"required"`
	FindText    string `json:"find_text" validate:"required"`
	ReplaceText string `json:"replace_text" validate:"required"`
	MatchCase   bool   `json:"match_case,omitempty"`
	ReplaceAll  bool   `json:"replace_all,omitempty"`
}

func RegisterContentTools(s *server.MCPServer) {
	// Insert text tool
	insertTextTool := mcp.NewTool("insert_text",
		mcp.WithDescription("Insert text at a specific position in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithString("text", mcp.Required(), mcp.Description("The text to insert")),
		mcp.WithNumber("index", mcp.Description("Position to insert text (default: end of document)")),
	)
	s.AddTool(insertTextTool, mcp.NewTypedToolHandler(insertTextHandler))

	// Replace text tool
	replaceTextTool := mcp.NewTool("replace_text",
		mcp.WithDescription("Replace text in a specific range within a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("Start position of the text to replace")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("End position of the text to replace")),
		mcp.WithString("text", mcp.Required(), mcp.Description("The replacement text")),
	)
	s.AddTool(replaceTextTool, mcp.NewTypedToolHandler(replaceTextHandler))

	// Delete text tool
	deleteTextTool := mcp.NewTool("delete_text",
		mcp.WithDescription("Delete text in a specific range within a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("Start position of the text to delete")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("End position of the text to delete")),
	)
	s.AddTool(deleteTextTool, mcp.NewTypedToolHandler(deleteTextHandler))

	// Append text tool
	appendTextTool := mcp.NewTool("append_text",
		mcp.WithDescription("Append text to the end of a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithString("text", mcp.Required(), mcp.Description("The text to append")),
	)
	s.AddTool(appendTextTool, mcp.NewTypedToolHandler(appendTextHandler))

	// Read text tool
	readTextTool := mcp.NewTool("read_text",
		mcp.WithDescription("Read text content from a Google Docs document, optionally within a specific range"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("start_index", mcp.Description("Start position to read from (default: beginning of document)")),
		mcp.WithNumber("end_index", mcp.Description("End position to read to (default: end of document)")),
	)
	s.AddTool(readTextTool, mcp.NewTypedToolHandler(readTextHandler))

	// Find and replace tool
	findReplaceTool := mcp.NewTool("find_replace",
		mcp.WithDescription("Find and replace text in a Google Docs document with options for case sensitivity and replace all"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithString("find_text", mcp.Required(), mcp.Description("The text to find")),
		mcp.WithString("replace_text", mcp.Required(), mcp.Description("The replacement text")),
		mcp.WithBoolean("match_case", mcp.Description("Whether to match case when searching (default: false)")),
		mcp.WithBoolean("replace_all", mcp.Description("Whether to replace all occurrences (default: false, replaces first occurrence only)")),
	)
	s.AddTool(findReplaceTool, mcp.NewTypedToolHandler(findReplaceHandler))
}

func insertTextHandler(ctx context.Context, request mcp.CallToolRequest, input InsertTextInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	// Get document to determine insertion index if not provided
	doc, err := docsService.Documents.Get(input.DocumentID).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("get document for text insertion", err), nil
	}

	insertIndex := input.Index
	if insertIndex <= 0 {
		// Insert at the end of the document (before the last character which is usually a newline)
		insertIndex = int64(len(doc.Body.Content)) - 1
		if insertIndex < 1 {
			insertIndex = 1
		}
	}

	// Create the batch update request
	requests := []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Location: &docs.Location{
					Index: insertIndex,
				},
				Text: input.Text,
			},
		},
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err = docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("insert text", err), nil
	}

	result := fmt.Sprintf("Text inserted successfully!\n\nDocument ID: %s\nInsertion Index: %d\nText Length: %d characters",
		input.DocumentID, insertIndex, len(input.Text))

	return mcp.NewToolResultText(result), nil
}

func replaceTextHandler(ctx context.Context, request mcp.CallToolRequest, input ReplaceTextInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	if input.StartIndex >= input.EndIndex {
		return mcp.NewToolResultText("Error: Start index must be less than end index."), nil
	}

	// Create the batch update request
	requests := []*docs.Request{
		{
			ReplaceAllText: &docs.ReplaceAllTextRequest{
				ContainsText: &docs.SubstringMatchCriteria{
					Text: input.Text,
				},
				ReplaceText: input.Text,
			},
		},
	}

	// For more precise replacement, we'll use delete and insert
	requests = []*docs.Request{
		{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					StartIndex: input.StartIndex,
					EndIndex:   input.EndIndex,
				},
			},
		},
		{
			InsertText: &docs.InsertTextRequest{
				Location: &docs.Location{
					Index: input.StartIndex,
				},
				Text: input.Text,
			},
		},
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err := docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("replace text", err), nil
	}

	result := fmt.Sprintf("Text replaced successfully!\n\nDocument ID: %s\nRange: %d-%d\nReplacement Length: %d characters",
		input.DocumentID, input.StartIndex, input.EndIndex, len(input.Text))

	return mcp.NewToolResultText(result), nil
}

func deleteTextHandler(ctx context.Context, request mcp.CallToolRequest, input DeleteTextInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	if input.StartIndex >= input.EndIndex {
		return mcp.NewToolResultText("Error: Start index must be less than end index."), nil
	}

	// Create the batch update request
	requests := []*docs.Request{
		{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					StartIndex: input.StartIndex,
					EndIndex:   input.EndIndex,
				},
			},
		},
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err := docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("delete text", err), nil
	}

	deletedLength := input.EndIndex - input.StartIndex
	result := fmt.Sprintf("Text deleted successfully!\n\nDocument ID: %s\nDeleted Range: %d-%d\nDeleted Length: %d characters",
		input.DocumentID, input.StartIndex, input.EndIndex, deletedLength)

	return mcp.NewToolResultText(result), nil
}

func appendTextHandler(ctx context.Context, request mcp.CallToolRequest, input AppendTextInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	// Get document to determine the end index
	doc, err := docsService.Documents.Get(input.DocumentID).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("get document for text appending", err), nil
	}

	// Calculate the end index (usually the last index minus 1 for the final newline)
	endIndex := int64(1)
	if doc.Body != nil && len(doc.Body.Content) > 0 {
		// Find the actual end of content
		for _, element := range doc.Body.Content {
			if element.EndIndex > endIndex {
				endIndex = element.EndIndex
			}
		}
		endIndex-- // Insert before the final character
	}

	// Create the batch update request
	requests := []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Location: &docs.Location{
					Index: endIndex,
				},
				Text: input.Text,
			},
		},
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err = docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("append text", err), nil
	}

	result := fmt.Sprintf("Text appended successfully!\n\nDocument ID: %s\nAppended at Index: %d\nText Length: %d characters",
		input.DocumentID, endIndex, len(input.Text))

	return mcp.NewToolResultText(result), nil
}

func readTextHandler(ctx context.Context, request mcp.CallToolRequest, input ReadTextInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	doc, err := docsService.Documents.Get(input.DocumentID).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("get document for reading", err), nil
	}

	// Extract plain text from the document
	fullText := util.ExtractPlainText(doc)

	// If range is specified, extract the substring
	if input.StartIndex > 0 || input.EndIndex > 0 {
		startIdx := int(input.StartIndex)
		endIdx := int(input.EndIndex)

		if startIdx < 0 {
			startIdx = 0
		}
		if endIdx <= 0 || endIdx > len(fullText) {
			endIdx = len(fullText)
		}
		if startIdx >= endIdx {
			return mcp.NewToolResultText("Error: Invalid range. Start index must be less than end index."), nil
		}

		rangeText := fullText[startIdx:endIdx]
		result := fmt.Sprintf("Text content (range %d-%d):\n\n%s\n\n--- End of Content ---\nTotal characters in range: %d",
			startIdx, endIdx, rangeText, len(rangeText))
		return mcp.NewToolResultText(result), nil
	}

	// Return full document text
	result := fmt.Sprintf("Document: %s\nDocument ID: %s\n\n--- Content ---\n%s\n\n--- End of Content ---\nTotal characters: %d",
		doc.Title, doc.DocumentId, fullText, len(fullText))

	return mcp.NewToolResultText(result), nil
}

func findReplaceHandler(ctx context.Context, request mcp.CallToolRequest, input FindReplaceInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	// Create the batch update request
	requests := []*docs.Request{
		{
			ReplaceAllText: &docs.ReplaceAllTextRequest{
				ContainsText: &docs.SubstringMatchCriteria{
					Text:      input.FindText,
					MatchCase: input.MatchCase,
				},
				ReplaceText: input.ReplaceText,
			},
		},
	}

	// If not replace all, we need a different approach
	// The Google Docs API's ReplaceAllText always replaces all occurrences
	// To replace only the first occurrence, we would need to:
	// 1. Read the document
	// 2. Find the first occurrence
	// 3. Use replace text with specific indices
	if !input.ReplaceAll {
		// Get document content first
		doc, err := docsService.Documents.Get(input.DocumentID).Context(ctx).Do()
		if err != nil {
			return util.HandleGoogleAPIError("get document for find/replace", err), nil
		}

		fullText := util.ExtractPlainText(doc)
		
		// Find the first occurrence
		var findIndex int
		if input.MatchCase {
			findIndex = strings.Index(fullText, input.FindText)
		} else {
			findIndex = strings.Index(strings.ToLower(fullText), strings.ToLower(input.FindText))
		}

		if findIndex == -1 {
			return mcp.NewToolResultText(fmt.Sprintf("Text '%s' not found in the document.", input.FindText)), nil
		}

		// Replace only the first occurrence
		requests = []*docs.Request{
			{
				DeleteContentRange: &docs.DeleteContentRangeRequest{
					Range: &docs.Range{
						StartIndex: int64(findIndex + 1), // +1 because Google Docs uses 1-based indexing
						EndIndex:   int64(findIndex + 1 + len(input.FindText)),
					},
				},
			},
			{
				InsertText: &docs.InsertTextRequest{
					Location: &docs.Location{
						Index: int64(findIndex + 1),
					},
					Text: input.ReplaceText,
				},
			},
		}
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	response, err := docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("find and replace text", err), nil
	}

	replaceCount := "all occurrences"
	if !input.ReplaceAll {
		replaceCount = "first occurrence"
	}

	result := fmt.Sprintf("Find and replace completed successfully!\n\nDocument ID: %s\nFound Text: '%s'\nReplacement Text: '%s'\nReplaced: %s\nMatch Case: %t\nRevision ID: %s",
		input.DocumentID, input.FindText, input.ReplaceText, replaceCount, input.MatchCase, response.DocumentId)

	return mcp.NewToolResultText(result), nil
}
