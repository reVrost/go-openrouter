package openrouter_test

import (
	"encoding/json"
	"testing"

	openrouter "github.com/revrost/go-openrouter"
)

// ChatCompletionMessage json.Marshal tests

// Tests the case where MultiContent is not empty
func TestChatCompletionMessageMarshalJSON_MultiContent(t *testing.T) {
	parts := []openrouter.ChatMessagePart{
		{
			Type: openrouter.ChatMessagePartTypeText,
			Text: "What is in this image?",
		},
		{
			Type: openrouter.ChatMessagePartTypeImageURL,
			ImageURL: &openrouter.ChatMessageImageURL{
				URL: "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg",
			},
		},
	}
	message := openrouter.ChatCompletionMessage{
		Role:    openrouter.ChatMessageRoleUser,
		Content: openrouter.Content{Multi: parts},
	}

	expected := `{"role":"user","content":[{"type":"text","text":"What is in this image?"},{"type":"image_url","image_url":{"url":"https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"}}]}`
	marshalAndValidate(t, message, expected)
}

// Tests the case where Content is used (MultiContent is empty)
func TestChatCompletionMessageMarshalJSON_Content(t *testing.T) {
	message := openrouter.ChatCompletionMessage{
		Role:    openrouter.ChatMessageRoleUser,
		Content: openrouter.Content{Text: "This is a simple content"},
	}

	expected := `{"role":"user","content":"This is a simple content"}`
	marshalAndValidate(t, message, expected)
}

func marshalAndValidate(t *testing.T, message openrouter.ChatCompletionMessage, expected string) {
	// Calls MarshalJSON
	result, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Validates the resulting JSON
	if string(result) != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestUnmarshalChatCompletionMessage(t *testing.T) {
	input := `{"role":"user","content":"This is a simple content"}`
	var message openrouter.ChatCompletionMessage
	err := json.Unmarshal([]byte(input), &message)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if message.Role != openrouter.ChatMessageRoleUser {
		t.Errorf("expected %s, got %s", openrouter.ChatMessageRoleUser, message.Role)
	}
	if message.Content.Text != "This is a simple content" {
		t.Errorf("expected %s, got %s", "This is a simple content", message.Content.Text)
	}
}

func TestChatCompletionMessageMarshalJSON_MultiContent_WithPDF(t *testing.T) {
	parts := []openrouter.ChatMessagePart{
		{
			Type: openrouter.ChatMessagePartTypeText,
			Text: "Analyze this PDF document",
		},
		{
			Type: openrouter.ChatMessagePartTypeFile,
			File: &openrouter.FileContent{
				Filename: "document.pdf",
				FileData: "JVBERi0xLjQKJdPr6eEKMSAwIG9iago8PAovVGl0bGUgKFRlc3QgUERGKQo+PgplbmRvYmoKCjIgMCBvYmoKPDwKL1R5cGUgL0NhdGFsb2cKL1BhZ2VzIDMgMCBSCj4+CmVuZG9iagoKMyAwIG9iago8PAovVHlwZSAvUGFnZXMKL0NvdW50IDEKL0tpZHMgWzQgMCBSXQo+PgplbmRvYmoKCjQgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCAzIDAgUgovTWVkaWFCb3ggWzAgMCA2MTIgNzkyXQovUmVzb3VyY2VzIDw8Cj4+Ci9Db250ZW50cyA1IDAgUgo+PgplbmRvYmoKCjUgMCBvYmoKPDwKL0xlbmd0aCA0NAo+PgpzdHJlYW0KQlQKL0YxIDEyIFRmCjEwMCA3MDAgVGQKKEhlbGxvIFdvcmxkKSBUagoKRVQKZW5kc3RyZWFtCmVuZG9iagoKeHJlZgowIDYKMDAwMDAwMDAwMCA2NTUzNSBmIAowMDAwMDAwMDE2IDAwMDAwIG4gCjAwMDAwMDA2NSAwMDAwMCBuIAowMDAwMDAwMTEyIDAwMDAwIG4gCjAwMDAwMDAxNjkgMDAwMDAgbiAKMDAwMDAwMDMwMCAwMDAwMCBuIAp0cmFpbGVyCjw8Ci9TaXplIDYKL1Jvb3QgMiAwIFIKL0luZm8gMSAwIFIKPj4Kc3RhcnR4cmVmCjM5MwolJUVPRgo=",
			},
		},
	}
	message := openrouter.ChatCompletionMessage{
		Role:    openrouter.ChatMessageRoleUser,
		Content: openrouter.Content{Multi: parts},
	}

	expected := `{"role":"user","content":[{"type":"text","text":"Analyze this PDF document"},{"type":"file","file":{"filename":"document.pdf","file_data":"JVBERi0xLjQKJdPr6eEKMSAwIG9iago8PAovVGl0bGUgKFRlc3QgUERGKQo+PgplbmRvYmoKCjIgMCBvYmoKPDwKL1R5cGUgL0NhdGFsb2cKL1BhZ2VzIDMgMCBSCj4+CmVuZG9iagoKMyAwIG9iago8PAovVHlwZSAvUGFnZXMKL0NvdW50IDEKL0tpZHMgWzQgMCBSXQo+PgplbmRvYmoKCjQgMCBvYmoKPDwKL1R5cGUgL1BhZ2UKL1BhcmVudCAzIDAgUgovTWVkaWFCb3ggWzAgMCA2MTIgNzkyXQovUmVzb3VyY2VzIDw8Cj4+Ci9Db250ZW50cyA1IDAgUgo+PgplbmRvYmoKCjUgMCBvYmoKPDwKL0xlbmd0aCA0NAo+PgpzdHJlYW0KQlQKL0YxIDEyIFRmCjEwMCA3MDAgVGQKKEhlbGxvIFdvcmxkKSBUagoKRVQKZW5kc3RyZWFtCmVuZG9iagoKeHJlZgowIDYKMDAwMDAwMDAwMCA2NTUzNSBmIAowMDAwMDAwMDE2IDAwMDAwIG4gCjAwMDAwMDA2NSAwMDAwMCBuIAowMDAwMDAwMTEyIDAwMDAwIG4gCjAwMDAwMDAxNjkgMDAwMDAgbiAKMDAwMDAwMDMwMCAwMDAwMCBuIAp0cmFpbGVyCjw8Ci9TaXplIDYKL1Jvb3QgMiAwIFIKL0luZm8gMSAwIFIKPj4Kc3RhcnR4cmVmCjM5MwolJUVPRgo="}}]}`
	marshalAndValidate(t, message, expected)
}

func TestChatCompletionMessagePromptCachingApplies(t *testing.T) {
	message := openrouter.ChatCompletionMessage{
		Role: openrouter.ChatMessageRoleUser,
		Content: openrouter.Content{Multi: []openrouter.ChatMessagePart{
			{Text: "This is a simple content", CacheControl: &openrouter.CacheControl{
				Type: "ephemeral",
			}},
		},
		}}

	expected := `{"role":"user","content":[{"text":"This is a simple content","cache_control":{"type":"ephemeral"}}]}`
	marshalAndValidate(t, message, expected)
}
