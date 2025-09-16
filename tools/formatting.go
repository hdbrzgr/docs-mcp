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

// Input types for formatting tools
type FormatTextInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	StartIndex int64  `json:"start_index" validate:"required"`
	EndIndex   int64  `json:"end_index" validate:"required"`
	Bold       *bool  `json:"bold,omitempty"`
	Italic     *bool  `json:"italic,omitempty"`
	Underline  *bool  `json:"underline,omitempty"`
	FontSize   *int64 `json:"font_size,omitempty"`
	FontFamily string `json:"font_family,omitempty"`
}

type SetTextColorInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	StartIndex int64  `json:"start_index" validate:"required"`
	EndIndex   int64  `json:"end_index" validate:"required"`
	Color      string `json:"color" validate:"required"` // Hex color code (e.g., "#FF0000" for red)
}

type SetBackgroundColorInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	StartIndex int64  `json:"start_index" validate:"required"`
	EndIndex   int64  `json:"end_index" validate:"required"`
	Color      string `json:"color" validate:"required"` // Hex color code (e.g., "#FFFF00" for yellow)
}

type SetParagraphStyleInput struct {
	DocumentID string `json:"document_id" validate:"required"`
	StartIndex int64  `json:"start_index" validate:"required"`
	EndIndex   int64  `json:"end_index" validate:"required"`
	StyleType  string `json:"style_type" validate:"required"` // NORMAL_TEXT, HEADING_1, HEADING_2, etc.
	Alignment  string `json:"alignment,omitempty"`            // START, CENTER, END, JUSTIFY
}

type SetLineSpacingInput struct {
	DocumentID string  `json:"document_id" validate:"required"`
	StartIndex int64   `json:"start_index" validate:"required"`
	EndIndex   int64   `json:"end_index" validate:"required"`
	Spacing    float64 `json:"spacing" validate:"required"` // Line spacing (e.g., 1.0, 1.5, 2.0)
}

func RegisterFormattingTools(s *server.MCPServer) {
	// Format text tool
	formatTextTool := mcp.NewTool("format_text",
		mcp.WithDescription("Apply text formatting (bold, italic, underline, font size, font family) to a range of text in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("Start position of the text to format")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("End position of the text to format")),
		mcp.WithBoolean("bold", mcp.Description("Apply bold formatting (true/false)")),
		mcp.WithBoolean("italic", mcp.Description("Apply italic formatting (true/false)")),
		mcp.WithBoolean("underline", mcp.Description("Apply underline formatting (true/false)")),
		mcp.WithNumber("font_size", mcp.Description("Font size in points (e.g., 12, 14, 16)")),
		mcp.WithString("font_family", mcp.Description("Font family name (e.g., 'Arial', 'Times New Roman', 'Calibri')")),
	)
	s.AddTool(formatTextTool, mcp.NewTypedToolHandler(formatTextHandler))

	// Set text color tool
	setTextColorTool := mcp.NewTool("set_text_color",
		mcp.WithDescription("Set the text color for a range of text in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("Start position of the text to color")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("End position of the text to color")),
		mcp.WithString("color", mcp.Required(), mcp.Description("Hex color code (e.g., '#FF0000' for red, '#0000FF' for blue)")),
	)
	s.AddTool(setTextColorTool, mcp.NewTypedToolHandler(setTextColorHandler))

	// Set background color tool
	setBackgroundColorTool := mcp.NewTool("set_background_color",
		mcp.WithDescription("Set the background color for a range of text in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("Start position of the text to highlight")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("End position of the text to highlight")),
		mcp.WithString("color", mcp.Required(), mcp.Description("Hex color code (e.g., '#FFFF00' for yellow, '#00FF00' for green)")),
	)
	s.AddTool(setBackgroundColorTool, mcp.NewTypedToolHandler(setBackgroundColorHandler))

	// Set paragraph style tool
	setParagraphStyleTool := mcp.NewTool("set_paragraph_style",
		mcp.WithDescription("Set paragraph style and alignment for a range of text in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("Start position of the paragraph to style")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("End position of the paragraph to style")),
		mcp.WithString("style_type", mcp.Required(), mcp.Description("Style type: 'NORMAL_TEXT', 'HEADING_1', 'HEADING_2', 'HEADING_3', 'HEADING_4', 'HEADING_5', 'HEADING_6', 'TITLE', 'SUBTITLE'")),
		mcp.WithString("alignment", mcp.Description("Text alignment: 'START', 'CENTER', 'END', 'JUSTIFY'")),
	)
	s.AddTool(setParagraphStyleTool, mcp.NewTypedToolHandler(setParagraphStyleHandler))

	// Set line spacing tool
	setLineSpacingTool := mcp.NewTool("set_line_spacing",
		mcp.WithDescription("Set line spacing for a range of text in a Google Docs document"),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("The unique identifier of the document")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("Start position of the text to adjust spacing")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("End position of the text to adjust spacing")),
		mcp.WithNumber("spacing", mcp.Required(), mcp.Description("Line spacing value (e.g., 1.0 for single, 1.5 for 1.5x, 2.0 for double)")),
	)
	s.AddTool(setLineSpacingTool, mcp.NewTypedToolHandler(setLineSpacingHandler))
}

func formatTextHandler(ctx context.Context, request mcp.CallToolRequest, input FormatTextInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	if input.StartIndex >= input.EndIndex {
		return mcp.NewToolResultText("Error: Start index must be less than end index."), nil
	}

	var requests []*docs.Request

	// Build text style update
	textStyle := &docs.TextStyle{}
	var hasUpdates bool

	if input.Bold != nil {
		textStyle.Bold = *input.Bold
		hasUpdates = true
	}

	if input.Italic != nil {
		textStyle.Italic = *input.Italic
		hasUpdates = true
	}

	if input.Underline != nil {
		textStyle.Underline = *input.Underline
		hasUpdates = true
	}

	if input.FontSize != nil {
		textStyle.FontSize = &docs.Dimension{
			Magnitude: float64(*input.FontSize),
			Unit:      "PT",
		}
		hasUpdates = true
	}

	if input.FontFamily != "" {
		textStyle.WeightedFontFamily = &docs.WeightedFontFamily{
			FontFamily: input.FontFamily,
		}
		hasUpdates = true
	}

	if hasUpdates {
		requests = append(requests, &docs.Request{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				Range: &docs.Range{
					StartIndex: input.StartIndex,
					EndIndex:   input.EndIndex,
				},
				TextStyle: textStyle,
				Fields:    "*", // Update all specified fields
			},
		})
	}

	if len(requests) == 0 {
		return mcp.NewToolResultText("No formatting changes specified."), nil
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err := docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("format text", err), nil
	}

	result := fmt.Sprintf("Text formatting applied successfully!\n\nDocument ID: %s\nRange: %d-%d\nFormatting changes applied.",
		input.DocumentID, input.StartIndex, input.EndIndex)

	return mcp.NewToolResultText(result), nil
}

func setTextColorHandler(ctx context.Context, request mcp.CallToolRequest, input SetTextColorInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	if input.StartIndex >= input.EndIndex {
		return mcp.NewToolResultText("Error: Start index must be less than end index."), nil
	}

	// Parse hex color
	color, err := parseHexColor(input.Color)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error: Invalid color format. Use hex format like '#FF0000'. %v", err)), nil
	}

	requests := []*docs.Request{
		{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				Range: &docs.Range{
					StartIndex: input.StartIndex,
					EndIndex:   input.EndIndex,
				},
				TextStyle: &docs.TextStyle{
					ForegroundColor: &docs.OptionalColor{
						Color: color,
					},
				},
				Fields: "foregroundColor",
			},
		},
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err = docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("set text color", err), nil
	}

	result := fmt.Sprintf("Text color applied successfully!\n\nDocument ID: %s\nRange: %d-%d\nColor: %s",
		input.DocumentID, input.StartIndex, input.EndIndex, input.Color)

	return mcp.NewToolResultText(result), nil
}

func setBackgroundColorHandler(ctx context.Context, request mcp.CallToolRequest, input SetBackgroundColorInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	if input.StartIndex >= input.EndIndex {
		return mcp.NewToolResultText("Error: Start index must be less than end index."), nil
	}

	// Parse hex color
	color, err := parseHexColor(input.Color)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error: Invalid color format. Use hex format like '#FFFF00'. %v", err)), nil
	}

	requests := []*docs.Request{
		{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				Range: &docs.Range{
					StartIndex: input.StartIndex,
					EndIndex:   input.EndIndex,
				},
				TextStyle: &docs.TextStyle{
					BackgroundColor: &docs.OptionalColor{
						Color: color,
					},
				},
				Fields: "backgroundColor",
			},
		},
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err = docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("set background color", err), nil
	}

	result := fmt.Sprintf("Background color applied successfully!\n\nDocument ID: %s\nRange: %d-%d\nColor: %s",
		input.DocumentID, input.StartIndex, input.EndIndex, input.Color)

	return mcp.NewToolResultText(result), nil
}

func setParagraphStyleHandler(ctx context.Context, request mcp.CallToolRequest, input SetParagraphStyleInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	if input.StartIndex >= input.EndIndex {
		return mcp.NewToolResultText("Error: Start index must be less than end index."), nil
	}

	// Validate style type
	validStyles := map[string]bool{
		"NORMAL_TEXT": true,
		"HEADING_1":   true,
		"HEADING_2":   true,
		"HEADING_3":   true,
		"HEADING_4":   true,
		"HEADING_5":   true,
		"HEADING_6":   true,
		"TITLE":       true,
		"SUBTITLE":    true,
	}

	if !validStyles[input.StyleType] {
		return mcp.NewToolResultText("Error: Invalid style type. Must be one of: NORMAL_TEXT, HEADING_1, HEADING_2, HEADING_3, HEADING_4, HEADING_5, HEADING_6, TITLE, SUBTITLE"), nil
	}

	var requests []*docs.Request

	// Update paragraph style
	paragraphStyle := &docs.ParagraphStyle{
		NamedStyleType: input.StyleType,
	}

	// Set alignment if provided
	if input.Alignment != "" {
		validAlignments := map[string]bool{
			"START":   true,
			"CENTER":  true,
			"END":     true,
			"JUSTIFY": true,
		}

		if !validAlignments[input.Alignment] {
			return mcp.NewToolResultText("Error: Invalid alignment. Must be one of: START, CENTER, END, JUSTIFY"), nil
		}

		paragraphStyle.Alignment = input.Alignment
	}

	requests = append(requests, &docs.Request{
		UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
			Range: &docs.Range{
				StartIndex: input.StartIndex,
				EndIndex:   input.EndIndex,
			},
			ParagraphStyle: paragraphStyle,
			Fields:         "*",
		},
	})

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err := docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("set paragraph style", err), nil
	}

	result := fmt.Sprintf("Paragraph style applied successfully!\n\nDocument ID: %s\nRange: %d-%d\nStyle: %s",
		input.DocumentID, input.StartIndex, input.EndIndex, input.StyleType)

	if input.Alignment != "" {
		result += fmt.Sprintf("\nAlignment: %s", input.Alignment)
	}

	return mcp.NewToolResultText(result), nil
}

func setLineSpacingHandler(ctx context.Context, request mcp.CallToolRequest, input SetLineSpacingInput) (*mcp.CallToolResult, error) {
	docsService := services.GoogleDocsClient()

	if input.StartIndex >= input.EndIndex {
		return mcp.NewToolResultText("Error: Start index must be less than end index."), nil
	}

	if input.Spacing <= 0 {
		return mcp.NewToolResultText("Error: Line spacing must be greater than 0."), nil
	}

	requests := []*docs.Request{
		{
			UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
				Range: &docs.Range{
					StartIndex: input.StartIndex,
					EndIndex:   input.EndIndex,
				},
				ParagraphStyle: &docs.ParagraphStyle{
					LineSpacing: input.Spacing,
				},
				Fields: "lineSpacing",
			},
		},
	}

	batchUpdateRequest := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}

	_, err := docsService.Documents.BatchUpdate(input.DocumentID, batchUpdateRequest).Context(ctx).Do()
	if err != nil {
		return util.HandleGoogleAPIError("set line spacing", err), nil
	}

	result := fmt.Sprintf("Line spacing applied successfully!\n\nDocument ID: %s\nRange: %d-%d\nSpacing: %.1fx",
		input.DocumentID, input.StartIndex, input.EndIndex, input.Spacing)

	return mcp.NewToolResultText(result), nil
}

// parseHexColor parses a hex color string and returns a Google Docs Color object
func parseHexColor(hexColor string) (*docs.Color, error) {
	// Remove # if present
	if hexColor[0] == '#' {
		hexColor = hexColor[1:]
	}

	if len(hexColor) != 6 {
		return nil, fmt.Errorf("hex color must be 6 characters long")
	}

	// Parse RGB values
	var r, g, b int
	_, err := fmt.Sscanf(hexColor, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return nil, fmt.Errorf("invalid hex color format: %v", err)
	}

	return &docs.Color{
		RgbColor: &docs.RgbColor{
			Red:   float64(r) / 255.0,
			Green: float64(g) / 255.0,
			Blue:  float64(b) / 255.0,
		},
	}, nil
}
