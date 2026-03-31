package main

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/revrost/go-openrouter"
)

// Here is an example of how to use the Go OpenRouter client to stream audio responses from a model that supports audio output (like "openai/gpt-4o-audio-preview"). The example sends a simple text prompt and receives both the transcript and the audio data in real-time, which it then saves as a WAV file.
// You can check the documentation from OpenRouter for more details
// https://openrouter.ai/docs/guides/overview/multimodal/audio#streaming-chunk-format
// Check the output result in the output.wav
func main() {
	client := openrouter.NewClient(os.Getenv("OPENROUTER_API_KEY"))

	stream, err := client.CreateChatCompletionStream(
		context.Background(), openrouter.ChatCompletionRequest{
			Model: "openai/gpt-4o-audio-preview",
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    openrouter.ChatMessageRoleUser,
					Content: openrouter.Content{Text: "Say hello `Aldiwildan` in a friendly tone."},
				},
			},
			Modalities: []openrouter.ChatCompletionModality{
				openrouter.ModalityText,
				openrouter.ModalityAudio,
			},
			AudioConfig: &openrouter.ChatCompletionAudioConfig{
				Voice:  openrouter.AudioVoiceAlloy,
				Format: openrouter.AudioFormatPcm16,
			},
			Stream: true,
		},
	)
	if err != nil {
		fmt.Println("Error creating stream:", err)
		os.Exit(1)
	}
	defer stream.Close()

	var audioDataChunks []string
	var transcriptChunks []string

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Println("Error receiving chunk:", err)
			os.Exit(1)
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

	fmt.Printf("Transcript: %s\n", transcript)

	// Decode base64 audio and write to WAV file
	pcmData, err := base64.StdEncoding.DecodeString(fullAudioB64)
	if err != nil {
		fmt.Println("Error decoding base64 audio:", err)
		os.Exit(1)
	}

	outputPath := "output.wav"
	if err := writePCM16ToWAV(outputPath, pcmData, 24000, 1); err != nil {
		fmt.Println("Error writing WAV file:", err)
		os.Exit(1)
	}

	fmt.Printf("Audio saved to %s (%d bytes)\n", outputPath, len(pcmData))
}

func writePCM16ToWAV(path string, pcmData []byte, sampleRate, numChannels int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dataSize := uint32(len(pcmData))
	bitsPerSample := uint16(16)
	blockAlign := uint16(numChannels) * bitsPerSample / 8
	byteRate := uint32(sampleRate) * uint32(blockAlign)

	// RIFF header
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(36+dataSize))
	f.Write([]byte("WAVE"))

	// fmt sub-chunk
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16))          // sub-chunk size
	binary.Write(f, binary.LittleEndian, uint16(1))           // PCM format
	binary.Write(f, binary.LittleEndian, uint16(numChannels)) // channels
	binary.Write(f, binary.LittleEndian, uint32(sampleRate))  // sample rate
	binary.Write(f, binary.LittleEndian, byteRate)            // byte rate
	binary.Write(f, binary.LittleEndian, blockAlign)          // block align
	binary.Write(f, binary.LittleEndian, bitsPerSample)       // bits per sample

	// data sub-chunk
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, dataSize)
	_, err = f.Write(pcmData)
	return err
}
