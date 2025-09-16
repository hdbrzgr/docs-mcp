package tools

import (
	"context"
	"fmt"

	"github.com/hdbrzgr/docs-mcp/services"
	"github.com/hdbrzgr/docs-mcp/util"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/api/docs/v1"
)

// Input types for structure tools
type InsertTableInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	Index      int64  `json:"index" validate:"required"`
	Rows       int64  `json:"rows" validate:"required"`
	Columns    int64  `json:"columns" validate:"required"`
}

type InsertListInput struct {
	DocumentID string   `json:"document_id" validate:"required"`
	Index      int64    `json:"index" validate:"required"`
	Items      []string `json:"items" validate:"required"`
	Ordered    bool     `json:"ordered,omitempty"` // true for numbered list, false for bullet list
}

type InsertPageBreakInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	Index      int64  `json:"index" validate:"required"`
}

type InsertHorizontalRuleInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	Index      int64  `json:"index" validate:"required"`
}

type CreateTableOfContentsInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	Index      int64  `json:"index" validate:"required"`
}

type UpdateTableCellInput struct {
	DocumentID  string `json:"document_id" validate:"required"`
	TableIndex  int64  `json:"table_index" validate:"required"`
	RowIndex    int64  `json:"row_index" validate:"required"`
	ColumnIndex int64  `json:"column_index" validate:"required"`
	Text        string `json:"text" validate:"required"`
}

type InsertImageInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	Index      int64  `json:"index" validate:"required"`
	ImageURL   string `json:"image_url" validate:"required"`
	Width      int64  `json:"width,omitempty"`  // Width in points
	Height     int64  `json:"height,omitempty"` // Height in points
}

func RegisterStructureTools(s *server.MCPServer) {
	// Insert table tool
	insertTableTool := mcp.NewTool("insert_table",
		mcp.WithDescription("Insert a table with specified rows and columns at a specific position in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("Position to insert the table")),
		mcp.WithNumber("rows", mcp.Required(), mcp.Description("Number of rows in the table")),
		mcp.WithNumber("columns", mcp.Required(), mcp.Description("Number of columns in the table")),
	)
	s.AddTool(insertTableTool, mcp.NewTypedToolHandler(insertTableHandler))

	// Insert list tool
	insertListTool := mcp.NewTool("insert_list",
		mcp.WithDescription("Insert a bulleted or numbered list at a specific position in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("Position to insert the list")),
		mcp.WithArray("items", mcp.Required(), mcp.Description("Array of text items for the list")),
		mcp.WithBoolean("ordered", mcp.Description("Whether to create a numbered list (true) or bullet list (false, default)")),
	)
	s.AddTool(insertListTool, mcp.NewTypedToolHandler(insertListHandler))

	// Insert page break tool
	insertPageBreakTool := mcp.NewTool("insert_page_break",
		mcp.WithDescription("Insert a page break at a specific position in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("Position to insert the page break")),
	)
	s.AddTool(insertPageBreakTool, mcp.NewTypedToolHandler(insertPageBreakHandler))

	// Insert horizontal rule tool
	insertHorizontalRuleTool := mcp.NewTool("insert_horizontal_rule",
		mcp.WithDescription("Insert a horizontal rule (line) at a specific position in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("Position to insert the horizontal rule")),
	)
	s.AddTool(insertHorizontalRuleTool, mcp.NewTypedToolHandler(insertHorizontalRuleHandler))

	// Create table of contents tool
	createTOCTool := mcp.NewTool("create_table_of_contents",
		mcp.WithDescription("Create a table of contents based on document headings at a specific position in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("Position to insert the table of contents")),
	)
	s.AddTool(createTOCTool, mcp.NewTypedToolHandler(createTableOfContentsHandler))

	// Update table cell tool
	updateTableCellTool := mcp.NewTool("update_table_cell",
		mcp.WithDescription("Update the content of a specific cell in a table within a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("table_index", mcp.Required(), mcp.Description("Index of the table in the document (0-based)")),
		mcp.WithNumber("row_index", mcp.Required(), mcp.Description("Row index in the table (0-based)")),
		mcp.WithNumber("column_index", mcp.Required(), mcp.Description("Column index in the table (0-based)")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Text content to insert in the cell")),
	)
	s.AddTool(updateTableCellTool, mcp.NewTypedToolHandler(updateTableCellHandler))

	// Insert image tool
	insertImageTool := mcp.NewTool("insert_image",
		mcp.WithDescription("Insert an image from a URL at a specific position in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("Position to insert the image")),
		mcp.WithString("image_url", mcp.Required(), mcp.Description("URL of the image to insert")),
		mcp.WithNumber("width", mcp.Description("Image width in points (optional)")),
		mcp.WithNumber("height", mcp.Description("Image height in points (optional)")),
	)
	s.AddTool(insertImageTool, mcp.NewTypedToolHandler(insertImageHandler))
}

func insertTableHandler(ctx context.Context, request mcp.CallToolRequest, input InsertTableInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	if input.Rows <= 0 || input.Columns <= 0 {
		return mcp.NewToolResultText("Error: Rows and columns must be greater than 0."), nil
	}

	if input.Rows > 20 || input.Columns > 20 {
		return mcp.NewToolResultText("Error: Maximum 20 rows and 20 columns allowed."), nil
	}

	requests := []*docs.Request{
		{
			InsertTable: &docs.InsertTableRequest{
				Location: &docs.Location{
					Index: input.Index,
				},
				Rows:    input.Rows,
				Columns: input.Columns,
			},
		},
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err := docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("insert table", err), nil
	}

	result := fmt.Sprintf("Table inserted successfully!\n\nDocument ID: %s\nPosition: %d\nSize: %dx%d (rows x columns)",
		input.DocumentID, input.Index, input.Rows, input.Columns)

	return mcp.NewToolResultText(result), nil
}

func insertListHandler(ctx context.Context, request mcp.CallToolRequest, input InsertListInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	if len(input.Items) == 0 {
		return mcp.NewToolResultText("Error: List must contain at least one item."), nil
	}

	var requests []*docs.Request
	currentIndex := input.Index

	// Insert each list item
	for i, item := range input.Items {
		// Insert the text
		requests = append(requests, &docs.Request{
			InsertText: &docs.InsertTextRequest{
				Location: &docs.Location{
					Index: currentIndex,
				},
				Text: item + "\n",
			},
		})

		// Apply list formatting to the paragraph
		listType := "BULLET_DISC_CIRCLE_SQUARE"
		if input.Ordered {
			listType = "DECIMAL_ALPHA_ROMAN"
		}

		requests = append(requests, &docs.Request{
			CreateParagraphBullets: &docs.CreateParagraphBulletsRequest{
				Range: &docs.Range{
					StartIndex: currentIndex,
					EndIndex:   currentIndex + int64(len(item)) + 1,
				},
				BulletPreset: listType,
			},
		})

		currentIndex += int64(len(item)) + 1 // +1 for the newline

		// Add some spacing between requests to avoid conflicts
		if i < len(input.Items)-1 {
			currentIndex += 1
		}
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err := docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("insert list", err), nil
	}

	listType := "bullet"
	if input.Ordered {
		listType = "numbered"
	}

	result := fmt.Sprintf("List inserted successfully!\n\nDocument ID: %s\nPosition: %d\nType: %s list\nItems: %d",
		input.DocumentID, input.Index, listType, len(input.Items))

	return mcp.NewToolResultText(result), nil
}

func insertPageBreakHandler(ctx context.Context, request mcp.CallToolRequest, input InsertPageBreakInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	requests := []*docs.Request{
		{
			InsertPageBreak: &docs.InsertPageBreakRequest{
				Location: &docs.Location{
					Index: input.Index,
				},
			},
		},
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err := docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("insert page break", err), nil
	}

	result := fmt.Sprintf("Page break inserted successfully!\n\nDocument ID: %s\nPosition: %d",
		input.DocumentID, input.Index)

	return mcp.NewToolResultText(result), nil
}

func insertHorizontalRuleHandler(ctx context.Context, request mcp.CallToolRequest, input InsertHorizontalRuleInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	// Insert a horizontal rule by inserting text and formatting it
	requests := []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Location: &docs.Location{
					Index: input.Index,
				},
				Text: "___\n", // Horizontal line representation
			},
		},
		{
			UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
				Range: &docs.Range{
					StartIndex: input.Index,
					EndIndex:   input.Index + 4, // Length of "___\n"
				},
				ParagraphStyle: &docs.ParagraphStyle{
					Alignment: "CENTER",
				},
				Fields: "alignment",
			},
		},
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err := docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("insert horizontal rule", err), nil
	}

	result := fmt.Sprintf("Horizontal rule inserted successfully!\n\nDocument ID: %s\nPosition: %d",
		input.DocumentID, input.Index)

	return mcp.NewToolResultText(result), nil
}

func createTableOfContentsHandler(ctx context.Context, request mcp.CallToolRequest, input CreateTableOfContentsInput) (*mcp.CallToolResult, error) {
	// Note: The Google Docs API does not currently support programmatically inserting a table of contents.
	// This functionality must be done manually through the Google Docs UI:
	// 1. Open your document in Google Docs
	// 2. Click where you want to insert the table of contents
	// 3. Go to Insert > Table of contents
	// 4. Choose your preferred style

	result := fmt.Sprintf("❌ Table of Contents Creation Not Supported\n\nThe Google Docs API does not currently support programmatically inserting a table of contents.\n\nTo add a table of contents to your document:\n1. Open the document in Google Docs: https://docs.google.com/document/d/%s/edit\n2. Click at position %d (or where you want the table of contents)\n3. Go to Insert → Table of contents\n4. Choose your preferred style\n\nThe table of contents will automatically update based on headings in your document.",
		input.DocumentID, input.Index)

	return mcp.NewToolResultText(result), nil
}

func updateTableCellHandler(ctx context.Context, request mcp.CallToolRequest, input UpdateTableCellInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	// Get the document to find the table
	doc, err := docsService.Documents.Get(input.DocumentID).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("get document for table update", err), nil
	}

	// Find the table and cell
	var targetTable *docs.Table
	tableCount := int64(0)

	if doc.Body != nil {
		for _, element := range doc.Body.Content {
			if element.Table != nil {
				if tableCount == input.TableIndex {
					targetTable = element.Table
					break
				}
				tableCount++
			}
		}
	}

	if targetTable == nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error: Table with index %d not found in document.", input.TableIndex)), nil
	}

	if input.RowIndex >= int64(len(targetTable.TableRows)) {
		return mcp.NewToolResultText(fmt.Sprintf("Error: Row index %d is out of range. Table has %d rows.", input.RowIndex, len(targetTable.TableRows))), nil
	}

	row := targetTable.TableRows[input.RowIndex]
	if input.ColumnIndex >= int64(len(row.TableCells)) {
		return mcp.NewToolResultText(fmt.Sprintf("Error: Column index %d is out of range. Row has %d columns.", input.ColumnIndex, len(row.TableCells))), nil
	}

	cell := row.TableCells[input.ColumnIndex]

	// Clear existing content and insert new text
	requests := []*docs.Request{
		{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					StartIndex: cell.StartIndex,
					EndIndex:   cell.EndIndex - 1, // -1 to preserve the cell structure
				},
			},
		},
		{
			InsertText: &docs.InsertTextRequest{
				Location: &docs.Location{
					Index: cell.StartIndex,
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
		return util.HandleGoogleAPIError("update table cell", err), nil
	}

	result := fmt.Sprintf("Table cell updated successfully!\n\nDocument ID: %s\nTable: %d\nCell: Row %d, Column %d\nContent: %s",
		input.DocumentID, input.TableIndex, input.RowIndex, input.ColumnIndex, input.Text)

	return mcp.NewToolResultText(result), nil
}

func insertImageHandler(ctx context.Context, request mcp.CallToolRequest, input InsertImageInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	// Create inline object properties for the image
	inlineObjectProperties := &docs.InlineObjectProperties{
		EmbeddedObject: &docs.EmbeddedObject{
			ImageProperties: &docs.ImageProperties{},
		},
	}

	// Set dimensions if provided
	if input.Width > 0 || input.Height > 0 {
		inlineObjectProperties.EmbeddedObject.Size = &docs.Size{}

		if input.Width > 0 {
			inlineObjectProperties.EmbeddedObject.Size.Width = &docs.Dimension{
				Magnitude: float64(input.Width),
				Unit:      "PT",
			}
		}

		if input.Height > 0 {
			inlineObjectProperties.EmbeddedObject.Size.Height = &docs.Dimension{
				Magnitude: float64(input.Height),
				Unit:      "PT",
			}
		}
	}

	requests := []*docs.Request{
		{
			InsertInlineImage: &docs.InsertInlineImageRequest{
				Location: &docs.Location{
					Index: input.Index,
				},
				Uri:        input.ImageURL,
				ObjectSize: inlineObjectProperties.EmbeddedObject.Size,
			},
		},
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err := docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("insert image", err), nil
	}

	result := fmt.Sprintf("Image inserted successfully!\n\nDocument ID: %s\nPosition: %d\nImage URL: %s",
		input.DocumentID, input.Index, input.ImageURL)

	if input.Width > 0 || input.Height > 0 {
		result += fmt.Sprintf("\nDimensions: %dx%d points", input.Width, input.Height)
	}

	return mcp.NewToolResultText(result), nil
}
