package util

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

// FormatGoogleDoc converts a Google Docs document to a formatted string representation
func FormatGoogleDoc(doc *docs.Document) string {
	var sb strings.Builder

	// Basic document information
	sb.WriteString(fmt.Sprintf("Document ID: %s\n", doc.DocumentId))
	sb.WriteString(fmt.Sprintf("Title: %s\n", doc.Title))
	
	if doc.DocumentStyle != nil && doc.DocumentStyle.DefaultHeaderId != "" {
		sb.WriteString(fmt.Sprintf("Default Header ID: %s\n", doc.DocumentStyle.DefaultHeaderId))
	}

	if doc.DocumentStyle != nil && doc.DocumentStyle.DefaultFooterId != "" {
		sb.WriteString(fmt.Sprintf("Default Footer ID: %s\n", doc.DocumentStyle.DefaultFooterId))
	}

	// Document content
	if doc.Body != nil && len(doc.Body.Content) > 0 {
		sb.WriteString("\n--- Document Content ---\n")
		sb.WriteString(FormatDocumentBody(doc.Body))
	}

	// Headers
	if len(doc.Headers) > 0 {
		sb.WriteString("\n--- Headers ---\n")
		for id, header := range doc.Headers {
			sb.WriteString(fmt.Sprintf("Header ID: %s\n", id))
			if header.Content != nil {
				sb.WriteString(FormatStructuralElements(header.Content))
			}
			sb.WriteString("\n")
		}
	}

	// Footers
	if len(doc.Footers) > 0 {
		sb.WriteString("\n--- Footers ---\n")
		for id, footer := range doc.Footers {
			sb.WriteString(fmt.Sprintf("Footer ID: %s\n", id))
			if footer.Content != nil {
				sb.WriteString(FormatStructuralElements(footer.Content))
			}
			sb.WriteString("\n")
		}
	}

	// Document revision information
	sb.WriteString(fmt.Sprintf("\nRevision ID: %s\n", doc.RevisionId))

	return sb.String()
}

// FormatDocumentBody formats the main body content of a document
func FormatDocumentBody(body *docs.Body) string {
	if body == nil || len(body.Content) == 0 {
		return "No content\n"
	}

	return FormatStructuralElements(body.Content)
}

// FormatStructuralElements formats a list of structural elements
func FormatStructuralElements(elements []*docs.StructuralElement) string {
	var sb strings.Builder

	for i, element := range elements {
		sb.WriteString(fmt.Sprintf("Element %d:\n", i+1))
		
		if element.Paragraph != nil {
			sb.WriteString(FormatParagraph(element.Paragraph))
		} else if element.Table != nil {
			sb.WriteString(FormatTable(element.Table))
		} else if element.SectionBreak != nil {
			sb.WriteString(FormatSectionBreak(element.SectionBreak))
		} else if element.TableOfContents != nil {
			sb.WriteString("Table of Contents\n")
		}
		
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatParagraph formats a paragraph element
func FormatParagraph(paragraph *docs.Paragraph) string {
	var sb strings.Builder

	// Paragraph style information
	if paragraph.ParagraphStyle != nil {
		if paragraph.ParagraphStyle.NamedStyleType != "" {
			sb.WriteString(fmt.Sprintf("Style: %s\n", paragraph.ParagraphStyle.NamedStyleType))
		}
		if paragraph.ParagraphStyle.Alignment != "" {
			sb.WriteString(fmt.Sprintf("Alignment: %s\n", paragraph.ParagraphStyle.Alignment))
		}
	}

	// Paragraph elements (text runs, etc.)
	if len(paragraph.Elements) > 0 {
		sb.WriteString("Content: ")
		for _, element := range paragraph.Elements {
			if element.TextRun != nil {
				sb.WriteString(FormatTextRun(element.TextRun))
			} else if element.InlineObjectElement != nil {
				sb.WriteString(fmt.Sprintf("[Inline Object: %s]", element.InlineObjectElement.InlineObjectId))
			} else if element.PageBreak != nil {
				sb.WriteString("[Page Break]")
			} else if element.FootnoteReference != nil {
				sb.WriteString(fmt.Sprintf("[Footnote: %s]", element.FootnoteReference.FootnoteId))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatTextRun formats a text run element
func FormatTextRun(textRun *docs.TextRun) string {
	content := textRun.Content
	
	// Add formatting information if available
	if textRun.TextStyle != nil {
		var formats []string
		
		if textRun.TextStyle.Bold {
			formats = append(formats, "bold")
		}
		if textRun.TextStyle.Italic {
			formats = append(formats, "italic")
		}
		if textRun.TextStyle.Underline {
			formats = append(formats, "underline")
		}
		if textRun.TextStyle.Strikethrough {
			formats = append(formats, "strikethrough")
		}
		
		if len(formats) > 0 {
			content = fmt.Sprintf("%s [%s]", content, strings.Join(formats, ", "))
		}
	}
	
	return content
}

// FormatTable formats a table element
func FormatTable(table *docs.Table) string {
	var sb strings.Builder
	sb.WriteString("Table:\n")
	
	if len(table.TableRows) > 0 {
		sb.WriteString(fmt.Sprintf("Rows: %d\n", len(table.TableRows)))
		
		for i, row := range table.TableRows {
			sb.WriteString(fmt.Sprintf("Row %d: ", i+1))
			
			for j, cell := range row.TableCells {
				if j > 0 {
					sb.WriteString(" | ")
				}
				
				// Get cell content
				cellContent := ""
				if len(cell.Content) > 0 {
					cellContent = strings.TrimSpace(FormatStructuralElements(cell.Content))
					// Remove newlines for table display
					cellContent = strings.ReplaceAll(cellContent, "\n", " ")
				}
				sb.WriteString(cellContent)
			}
			sb.WriteString("\n")
		}
	}
	
	return sb.String()
}

// FormatSectionBreak formats a section break element
func FormatSectionBreak(sectionBreak *docs.SectionBreak) string {
	breakType := "Section Break"
	if sectionBreak.SectionStyle != nil && sectionBreak.SectionStyle.SectionType != "" {
		breakType = fmt.Sprintf("Section Break (%s)", sectionBreak.SectionStyle.SectionType)
	}
	return fmt.Sprintf("%s\n", breakType)
}

// FormatDriveFile formats a Google Drive file for display
func FormatDriveFile(file *drive.File) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("ID: %s\n", file.Id))
	sb.WriteString(fmt.Sprintf("Name: %s\n", file.Name))
	sb.WriteString(fmt.Sprintf("MIME Type: %s\n", file.MimeType))
	
	if file.CreatedTime != "" {
		if createdTime, err := time.Parse(time.RFC3339, file.CreatedTime); err == nil {
			sb.WriteString(fmt.Sprintf("Created: %s\n", createdTime.Format("2006-01-02 15:04:05")))
		}
	}
	
	if file.ModifiedTime != "" {
		if modifiedTime, err := time.Parse(time.RFC3339, file.ModifiedTime); err == nil {
			sb.WriteString(fmt.Sprintf("Modified: %s\n", modifiedTime.Format("2006-01-02 15:04:05")))
		}
	}
	
	if len(file.Owners) > 0 {
		sb.WriteString("Owners: ")
		for i, owner := range file.Owners {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(owner.DisplayName)
			if owner.EmailAddress != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", owner.EmailAddress))
			}
		}
		sb.WriteString("\n")
	}
	
	if file.Size > 0 {
		sb.WriteString(fmt.Sprintf("Size: %d bytes\n", file.Size))
	}
	
	if file.WebViewLink != "" {
		sb.WriteString(fmt.Sprintf("View Link: %s\n", file.WebViewLink))
	}

	return sb.String()
}

// FormatDriveFileCompact returns a compact single-line representation of a Drive file
func FormatDriveFileCompact(file *drive.File) string {
	if file == nil {
		return ""
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("ID: %s", file.Id))
	parts = append(parts, fmt.Sprintf("Name: %s", file.Name))
	
	if file.ModifiedTime != "" {
		if modifiedTime, err := time.Parse(time.RFC3339, file.ModifiedTime); err == nil {
			parts = append(parts, fmt.Sprintf("Modified: %s", modifiedTime.Format("2006-01-02")))
		}
	}
	
	if len(file.Owners) > 0 {
		parts = append(parts, fmt.Sprintf("Owner: %s", file.Owners[0].DisplayName))
	}

	return strings.Join(parts, " | ")
}

// ExtractPlainText extracts plain text content from a Google Docs document
func ExtractPlainText(doc *docs.Document) string {
	var sb strings.Builder
	
	if doc.Body != nil && len(doc.Body.Content) > 0 {
		extractTextFromElements(doc.Body.Content, &sb)
	}
	
	return sb.String()
}

// extractTextFromElements recursively extracts text from structural elements
func extractTextFromElements(elements []*docs.StructuralElement, sb *strings.Builder) {
	for _, element := range elements {
		if element.Paragraph != nil {
			extractTextFromParagraph(element.Paragraph, sb)
		} else if element.Table != nil {
			extractTextFromTable(element.Table, sb)
		}
	}
}

// extractTextFromParagraph extracts text from a paragraph
func extractTextFromParagraph(paragraph *docs.Paragraph, sb *strings.Builder) {
	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			sb.WriteString(element.TextRun.Content)
		}
	}
	sb.WriteString("\n")
}

// extractTextFromTable extracts text from a table
func extractTextFromTable(table *docs.Table, sb *strings.Builder) {
	for _, row := range table.TableRows {
		for _, cell := range row.TableCells {
			extractTextFromElements(cell.Content, sb)
		}
	}
}
