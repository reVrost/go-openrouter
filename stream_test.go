package openrouter_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	openrouter "github.com/revrost/go-openrouter"
	"github.com/stretchr/testify/require"
)

// Test streaming
func TestChatCompletionMessageMarshalJSON_Streaming(t *testing.T) {
	client := createTestClient(t)

	stream, err := client.CreateChatCompletionStream(
		context.Background(), openrouter.ChatCompletionRequest{
			Model: "qwen/qwq-32b:free",
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: openrouter.Content{Text: "Hello, how are you?"}},
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
		_, err = json.MarshalIndent(response, "", "  ")
		require.NoError(t, err)
		// fmt.Println(string(json))
	}
}
