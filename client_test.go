package openrouter_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	openrouter "github.com/revrost/go-openrouter"
	"github.com/stretchr/testify/require"
)

const FreeModel = "deepseek/deepseek-r1-0528-qwen3-8b:free"
const OSSFreeModel = "openai/gpt-oss-20b:free"

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
				Model: FreeModel,
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
				if len(resp.Choices[0].Message.Content.Text) > 10 {
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

func TestCreateCompletion(t *testing.T) {
	client := createTestClient(t)

	tests := []struct {
		name     string
		request  openrouter.CompletionRequest
		wantErr  bool
		validate func(*testing.T, openrouter.CompletionResponse)
	}{
		{
			name: "basic completion",
			request: openrouter.CompletionRequest{
				Model:  "nousresearch/hermes-4-70b",
				Prompt: "Hello! Respond with just 'world'",
			},
			validate: func(t *testing.T, resp openrouter.CompletionResponse) {
				if len(resp.Choices) == 0 {
					t.Error("Expected at least one choice in response")
				}
				if !strings.Contains(resp.Choices[0].Text, "world") {
					t.Errorf("Unexpected response: '%s' expected 'world'", resp.Choices[0].Text)
				}
			},
		},
		{
			name: "invalid model",
			request: openrouter.CompletionRequest{
				Model:  "invalid-model",
				Prompt: "Hello",
			},
			wantErr: true,
		},
		{
			name: "streaming not supported",
			request: openrouter.CompletionRequest{
				Model:  openrouter.LiquidLFM7B,
				Stream: true,
				Prompt: "Hello",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := client.CreateCompletion(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCompletion() error = %v, wantErr %v", err, tt.wantErr)
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

func TestUsageAccounting(t *testing.T) {
	client := createTestClient(t)
	request := openrouter.ChatCompletionRequest{
		Model: FreeModel,
		Messages: []openrouter.ChatCompletionMessage{
			openrouter.SystemMessage("You are a helpful assistant."),
			openrouter.UserMessage("How are you?"),
		},
		Usage: &openrouter.IncludeUsage{
			Include: true,
		},
	}

	response, err := client.CreateChatCompletion(context.Background(), request)
	require.NoError(t, err)

	usage := response.Usage
	require.NotNil(t, usage)
	require.NotNil(t, usage.PromptTokens)
	require.NotNil(t, usage.CompletionTokens)
	require.NotNil(t, usage.TotalTokens)
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

func TestProviderError(t *testing.T) {
	client := createTestClient(t)

	ctx := context.Background()
	_, err := client.CreateChatCompletion(ctx, openrouter.ChatCompletionRequest{
		Model: "openai/gpt-5-nano",
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role:    openrouter.ChatMessageRoleUser,
				Content: openrouter.Content{Text: "This will always fail with a provider error, because openai requires the message to contain the word j-s-o-n."},
			},
		},
		ResponseFormat: &openrouter.ChatCompletionResponseFormat{
			Type: openrouter.ChatCompletionResponseFormatTypeJSONObject,
		},
	})

	if err == nil {
		t.Error("Expected api error, got nil")
	}

	apiErr, ok := err.(*openrouter.APIError)
	if !ok {
		t.Errorf("Expected api error, got %T", err)
	}

	if msg := apiErr.Error(); msg != "provider error, code: 400, message: Response input messages must contain the word 'json' in some form to use 'text.format' of type 'json_object'." {
		t.Errorf("Expected provider error, got %v", msg)
	}
}

func TestGetGeneration(t *testing.T) {
	client := createTestClient(t)

	ctx := context.Background()

	request := openrouter.ChatCompletionRequest{
		Model: OSSFreeModel,
		Messages: []openrouter.ChatCompletionMessage{
			openrouter.SystemMessage("You are a helpful assistant."),
			openrouter.UserMessage("How are you?"),
		},
		Provider: &openrouter.ChatProvider{
			Only: []string{"atlas-cloud/fp8"},
		},
	}

	response, err := client.CreateChatCompletion(ctx, request)
	require.NoError(t, err)

	// openrouter takes a second to store it (removing this causes it to fail)
	time.Sleep(1 * time.Second)

	generation, err := client.GetGeneration(ctx, response.ID)
	require.NoError(t, err)

	require.Equal(t, generation.ID, response.ID)
}

func TestListModels(t *testing.T) {
	client := createTestClient(t)

	models, err := client.ListModels(context.Background())
	require.NoError(t, err)

	require.NotEmpty(t, models)
	require.NotEmpty(t, models[0].ID)
}

func TestListUserModels(t *testing.T) {
	client := createTestClient(t)

	models, err := client.ListUserModels(context.Background())
	require.NoError(t, err)

	require.NotEmpty(t, models)
	require.NotEmpty(t, models[0].ID)
}

func TestListEmbeddingsModels(t *testing.T) {
	client := createTestClient(t)

	models, err := client.ListEmbeddingsModels(context.Background())
	require.NoError(t, err)

	require.NotEmpty(t, models)
	require.NotEmpty(t, models[0].ID)
}

func TestGetCurrentAPIKey(t *testing.T) {
	client := createTestClient(t)

	resp, err := client.GetCurrentAPIKey(context.Background())
	require.NoError(t, err)

	require.NotNil(t, resp.Data.Label)
	require.NotEmpty(t, resp.Data.Label)

	require.GreaterOrEqual(t, resp.Data.Usage, float64(0))

	if resp.Data.RateLimit != nil {
		require.NotEmpty(t, resp.Data.RateLimit.Interval)
	}
}
