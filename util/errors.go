package util

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/mark3labs/mcp-go/mcp"
)

// ErrorGuard wraps a tool handler function to catch and handle panics gracefully
func ErrorGuard(handler func(ctx interface{}, request interface{}) (*mcp.CallToolResult, error)) func(ctx interface{}, request interface{}) (*mcp.CallToolResult, error) {
	return func(ctx interface{}, request interface{}) (result *mcp.CallToolResult, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in tool handler: %v\n%s", r, debug.Stack())
				err = fmt.Errorf("internal error: %v", r)
				result = mcp.NewToolResultText(fmt.Sprintf("Error: %v", err))
			}
		}()

		return handler(ctx, request)
	}
}

// WrapError creates a consistent error format for Google Docs API errors
func WrapError(operation string, err error) error {
	return fmt.Errorf("Google Docs API error in %s: %v", operation, err)
}

// HandleGoogleAPIError processes Google API errors and returns user-friendly messages
func HandleGoogleAPIError(operation string, err error) *mcp.CallToolResult {
	if err == nil {
		return nil
	}

	errorMsg := fmt.Sprintf("Failed to %s: %v", operation, err)
	log.Printf("Google API Error: %s", errorMsg)
	
	// Check for common Google API error patterns and provide helpful messages
	errorStr := err.Error()
	
	if contains(errorStr, "403") || contains(errorStr, "Forbidden") {
		errorMsg += "\n\nThis might be a permissions issue. Please check:"
		errorMsg += "\n- The service account has access to the document"
		errorMsg += "\n- The document is shared with the service account email"
		errorMsg += "\n- The required APIs are enabled in Google Cloud Console"
	} else if contains(errorStr, "404") || contains(errorStr, "Not Found") {
		errorMsg += "\n\nThe document was not found. Please check:"
		errorMsg += "\n- The document ID is correct"
		errorMsg += "\n- The document exists and hasn't been deleted"
		errorMsg += "\n- You have access to the document"
	} else if contains(errorStr, "401") || contains(errorStr, "Unauthorized") {
		errorMsg += "\n\nAuthentication failed. Please check:"
		errorMsg += "\n- Your credentials are valid and not expired"
		errorMsg += "\n- The service account key is properly configured"
		errorMsg += "\n- The required scopes are included in the credentials"
	} else if contains(errorStr, "429") || contains(errorStr, "quota") {
		errorMsg += "\n\nRate limit exceeded. Please:"
		errorMsg += "\n- Wait a moment before retrying"
		errorMsg += "\n- Check your API quota in Google Cloud Console"
		errorMsg += "\n- Consider implementing exponential backoff"
	}

	return mcp.NewToolResultText(errorMsg)
}

// contains checks if a string contains a substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
