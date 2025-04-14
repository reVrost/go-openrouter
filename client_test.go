package openrouter_test

import (
	"context"
	"os"
	"strings"
	"testing"

	openrouter "github.com/revrost/go-openrouter"
)

// Test client setup
func createTestClient(t *testing.T) *openrouter.Client {
	token := os.Getenv("OPENROUTER_API_KEY")
	if token == "" {
		t.Skip("Skipping integration test: OPENROUTER_API_KEY not set")
	}

	// Add optional headers if needed
	return openrouter.NewClient(token,
		openrouter.WithXTitle("Integration Tests"),
		openrouter.WithHTTPReferer("https://github.com/revrost/go-openrouter"),
	)
}

func TestCreateChatCompletion(t *testing.T) {
	client := createTestClient(t)

	tests := []struct {
		name     string
		request  openrouter.ChatCompletionRequest
		wantErr  bool
		validate func(*testing.T, openrouter.ChatCompletionResponse)
	}{
		{
			name: "basic completion",
			request: openrouter.ChatCompletionRequest{
				Model: openrouter.GeminiFlashExp,
				Messages: []openrouter.ChatCompletionMessage{
					{
						Role:    openrouter.ChatMessageRoleUser,
						Content: openrouter.Content{Text: "Hello! Respond with just 'world'"},
					},
				},
			},
			validate: func(t *testing.T, resp openrouter.ChatCompletionResponse) {
				if len(resp.Choices) == 0 {
					t.Error("Expected at least one choice in response")
				}
				if !strings.Contains(resp.Choices[0].Message.Content.Text, "world") {
					t.Errorf("Unexpected response: '%s' expected 'world'", resp.Choices[0].Message.Content.Text)
				}
			},
		},
		{
			name: "invalid model",
			request: openrouter.ChatCompletionRequest{
				Model: "invalid-model",
				Messages: []openrouter.ChatCompletionMessage{
					{Role: openrouter.ChatMessageRoleUser, Content: openrouter.Content{Text: "Hello"}},
				},
			},
			wantErr: true,
		},
		{
			name: "streaming not supported",
			request: openrouter.ChatCompletionRequest{
				Model:  openrouter.LiquidLFM7B,
				Stream: true,
				Messages: []openrouter.ChatCompletionMessage{
					{Role: openrouter.ChatMessageRoleUser, Content: openrouter.Content{Text: "Hello"}},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := client.CreateChatCompletion(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateChatCompletion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

func TestAuthFailure(t *testing.T) {
	// Test with invalid token
	client := openrouter.NewClient("invalid-token")

	_, err := client.CreateChatCompletion(context.Background(), openrouter.ChatCompletionRequest{
		Model: openrouter.LiquidLFM7B,
		Messages: []openrouter.ChatCompletionMessage{
			{Role: openrouter.ChatMessageRoleUser, Content: openrouter.Content{Text: "Hello"}},
		},
	})

	if err == nil {
		t.Error("Expected authentication error, got nil")
	}
}
