package openrouter

import (
	"os"
	"strings"
)

// CreatePDFPlugin creates a completion plugin to process PDFs using the specified engine.
// The engine can be: "mistral-ocr" (for scanned documents/PDFs with images),
// "pdf-text" (for well-structured PDFs - free), or "native" (only for models that support file input).
func CreatePDFPlugin(engine PDFEngine) ChatCompletionPlugin {
	return ChatCompletionPlugin{
		ID: PluginIDFileParser,
		PDF: &PDFPlugin{
			Engine: string(engine),
		},
	}
}

// UserMessageWithPDFFromFile creates a user message with text and PDF content from a file.
// It reads the PDF file and creates a message with the embedded PDF data.
func UserMessageWithPDFFromFile(text, filePath string) (ChatCompletionMessage, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return ChatCompletionMessage{}, err
	}

	filename := filePath
	if idx := strings.LastIndex(filePath, "\\"); idx != -1 {
		filename = filePath[idx+1:]
	}
	if idx := strings.LastIndex(filename, "/"); idx != -1 {
		filename = filename[idx+1:]
	}

	return ChatCompletionMessage{
		Role: ChatMessageRoleUser,
		Content: Content{
			Multi: []ChatMessagePart{
				{
					Type: ChatMessagePartTypeText,
					Text: text,
				},
				{
					Type: ChatMessagePartTypeFile,
					File: &FileContent{
						Filename: filename,
						FileData: string(fileData),
					},
				},
			},
		},
	}, nil
}
