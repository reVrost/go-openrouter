package openrouter_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	"log/slog"

	openrouter "github.com/revrost/go-openrouter"
	"github.com/stretchr/testify/require"
)

// Test streaming with reasoning
func TestChatCompletionMessageMarshalJSON_StreamingWithReasoning(t *testing.T) {
	t.Skip("Only run this test locally")
	client := createTestClient(t)

	stream, err := client.CreateChatCompletionStream(
		context.Background(), openrouter.ChatCompletionRequest{
			Reasoning: &openrouter.ChatCompletionReasoning{
				Effort: openrouter.String("high"),
			},
			Model: "google/gemini-2.5-pro-preview",
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: openrouter.Content{Text: "Help me think whether i should make coffee with sugar ?"},
				},
			},
			Stream: true,
		},
	)
	require.NoError(t, err)
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil && err != io.EOF {
			require.NoError(t, err)
		}
		if errors.Is(err, io.EOF) {
			fmt.Println("EOF, stream finished")
			return
		}
		b, err := json.MarshalIndent(response, "", "  ")
		require.NoError(t, err)
		fmt.Println(string(b))
	}
}

// Test streaming
func TestChatCompletionMessageMarshalJSON_Streaming(t *testing.T) {
	client := createTestClient(t)

	stream, err := client.CreateChatCompletionStream(
		context.Background(), openrouter.ChatCompletionRequest{
			Model: FreeModel,
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: openrouter.Content{Text: "Help me think whether i should make coffee with sugar ?"},
				},
			},
			Stream: true,
		},
	)
	require.NoError(t, err)
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil && err != io.EOF {
			require.NoError(t, err)
		}
		if errors.Is(err, io.EOF) {
			fmt.Println("EOF, stream finished")
			return
		}
		b, err := json.MarshalIndent(response, "", "  ")
		require.NoError(t, err)
		slog.Debug(string(b))
	}
}
