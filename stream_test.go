package openrouter_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
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

// Test streaming with audio output
// Based on: https://openrouter.ai/docs/guides/overview/multimodal/audio#requesting-audio-output
func TestChatCompletionStreamingWithAudioOutput(t *testing.T) {
	client := createTestClient(t)

	stream, err := client.CreateChatCompletionStream(
		context.Background(), openrouter.ChatCompletionRequest{
			Model: "openai/gpt-4o-audio-preview",
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: openrouter.Content{Text: "Say Aldiwildan in a friendly tone."},
				},
			},
			Modalities: []openrouter.ChatCompletionModality{
				openrouter.ModalityText,
				openrouter.ModalityAudio,
			},
			AudioConfig: &openrouter.ChatCompletionAudioConfig{
				Voice:  openrouter.AudioVoiceAlloy,
				Format: openrouter.AudioFormatPcm16, // gpt-4o-audio-preview only supports pcm16 for now
			},
			Stream: true,
		},
	)
	require.NoError(t, err)
	defer stream.Close()

	var audioDataChunks []string
	var transcriptChunks []string

	for {
		response, err := stream.Recv()
		if err != nil && err != io.EOF {
			require.NoError(t, err)
		}
		if errors.Is(err, io.EOF) {
			slog.Info("EOF, stream finished")
			break
		}

		for _, choice := range response.Choices {
			if choice.Delta.Audio != nil {
				if choice.Delta.Audio.Data != "" {
					audioDataChunks = append(audioDataChunks, choice.Delta.Audio.Data)
				}
				if choice.Delta.Audio.Transcript != "" {
					transcriptChunks = append(transcriptChunks, choice.Delta.Audio.Transcript)
				}
			}
		}
	}

	transcript := strings.Join(transcriptChunks, "")
	fullAudioB64 := strings.Join(audioDataChunks, "")

	slog.Debug(fmt.Sprintf("Transcript: %s\n", transcript))
	slog.Debug(fmt.Sprintf("Audio data length (base64): %d\n", len(fullAudioB64)))

	require.NotEmpty(t, transcript, "transcript should not be empty")
	require.NotEmpty(t, fullAudioB64, "audio data should not be empty")
}
