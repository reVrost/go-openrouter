package openrouter

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UserMessageWithAudioFromFile creates a user message with the given prompt text and audio file.
// It reads the audio file (mp3 or wav) and creates a message with the embedded audio data.
func UserMessageWithAudioFromFile(promptText, filePath string) (ChatCompletionMessage, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return ChatCompletionMessage{}, err
	}

	ext := filepath.Ext(filePath)
	var format AudioFormat
	switch strings.ToLower(ext) {
	case ".mp3":
		format = AudioFormatMp3
	case ".wav":
		format = AudioFormatWav
	default:
		return ChatCompletionMessage{}, fmt.Errorf("unsupported audio format: %s", ext)
	}

	msg := UserMessageWithAudio(promptText, fileData, format)

	return msg, nil
}

// UserMessageWithAudio creates a user message with the given prompt text and audio content.
// Creates a message with the embedded audio data.
func UserMessageWithAudio(promptText string, audio []byte, format AudioFormat) ChatCompletionMessage {
	msg := ChatCompletionMessage{
		Role: ChatMessageRoleUser,
		Content: Content{
			Multi: []ChatMessagePart{
				{
					Type: ChatMessagePartTypeText,
					Text: promptText,
				},
				chatMessagePartWithAudio(audio, format),
			},
		},
	}

	return msg
}

// chatMessagePartWithAudio creates a ChatMessagePart which contains the given audio content.
func chatMessagePartWithAudio(audio []byte, format AudioFormat) ChatMessagePart {
	audioEncoded := base64.StdEncoding.EncodeToString(audio)

	msg := ChatMessagePart{
		Type: ChatMessagePartTypeInputAudio,
		InputAudio: &ChatMessageInputAudio{
			Format: format,
			Data:   audioEncoded,
		},
	}

	return msg
}
