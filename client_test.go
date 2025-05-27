package openrouter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	openrouter "github.com/revrost/go-openrouter"
	"github.com/stretchr/testify/require"
)

// Test client setup
func createTestClient(t *testing.T) *openrouter.Client {
	t.Helper()
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
				Model: "qwen/qwq-32b:free",
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

func TestExplicitPromptCachingApplies(t *testing.T) {
	t.Skip("Only run this test locally")
	client := createTestClient(t)

	message := openrouter.ChatCompletionMessage{
		Role: openrouter.ChatMessageRoleSystem,
		Content: openrouter.Content{
			Multi: []openrouter.ChatMessagePart{
				{
					Type:         openrouter.ChatMessagePartTypeText,
					Text:         testLongToken,
					CacheControl: &openrouter.CacheControl{Type: "ephemeral"},
				},
			},
		},
	}
	userMessage := openrouter.ChatCompletionMessage{
		Role: openrouter.ChatMessageRoleUser,
		Content: openrouter.Content{
			Multi: []openrouter.ChatMessagePart{
				{
					Type:         openrouter.ChatMessagePartTypeText,
					Text:         "Who was augustus based on the text?",
					CacheControl: &openrouter.CacheControl{Type: "ephemeral"},
				},
			},
		},
	}
	request := openrouter.ChatCompletionRequest{
		Model: "google/gemini-2.5-flash-preview-05-20",
		Messages: []openrouter.ChatCompletionMessage{
			message,
			userMessage,
		},
		Usage: &openrouter.IncludeUsage{
			Include: true,
		},
	}
	px, _ := json.MarshalIndent(request, "", "\t")
	fmt.Printf("request :\n %s\n", string(px))
	response, err := client.CreateChatCompletion(context.Background(), request)
	b, _ := json.MarshalIndent(response, "", "\t")
	fmt.Printf("response :\n %s\n", string(b))

	require.NoError(t, err)
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
