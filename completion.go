package openrouter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const completionsSuffix = "/completions"

var (
	ErrCompletionInvalidModel       = errors.New("this model is not supported with this method, please use CreateChatCompletion client method instead") //nolint:lll
	ErrCompletionStreamNotSupported = errors.New("streaming is not supported with this method, please use CreateCompletion")                            //nolint:lll
)

type CompletionRequest struct {
	Model string `json:"model,omitempty"`
	// The prompt to complete
	Prompt string `json:"prompt"`
	// Optional model fallbacks: https://openrouter.ai/docs/features/model-routing#the-models-parameter
	Models    []string                 `json:"models,omitempty"`
	Provider  *ChatProvider            `json:"provider,omitempty"`
	Reasoning *ChatCompletionReasoning `json:"reasoning,omitempty"`
	Usage     *IncludeUsage            `json:"usage,omitempty"`
	// Apply message transforms
	// https://openrouter.ai/docs/features/message-transforms
	Transforms []string `json:"transforms,omitempty"`
	Stream     bool     `json:"stream,omitempty"`
	// MaxTokens The maximum number of tokens that can be generated in the chat completion.
	// This value can be used to control costs for text generated via API.
	MaxTokens         int     `json:"max_tokens,omitempty"`
	Temperature       float32 `json:"temperature,omitempty"`
	Seed              *int    `json:"seed,omitempty"`
	TopP              float32 `json:"top_p,omitempty"`
	TopK              int     `json:"top_k,omitempty"`
	FrequencyPenalty  float32 `json:"frequency_penalty,omitempty"`
	PresencePenalty   float32 `json:"presence_penalty,omitempty"`
	RepetitionPenalty float32 `json:"repetition_penalty,omitempty"`
	// LogitBias is must be a token id string (specified by their token ID in the tokenizer), not a word string.
	// incorrect: `"logit_bias":{"You": 6}`, correct: `"logit_bias":{"1639": 6}`
	// refs: https://platform.openai.com/docs/api-reference/chat/create#chat/create-logit_bias
	LogitBias   map[string]int `json:"logit_bias,omitempty"`
	TopLogProbs int            `json:"top_logprobs,omitempty"`
	MinP        float32        `json:"min_p,omitempty"`
	TopA        float32        `json:"top_a,omitempty"`
	User        string         `json:"user,omitempty"`
}

type CompletionChoice struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
	// Reasoning Used by all the other models
	Reasoning *string `json:"reasoning,omitempty"`
	// FinishReason
	// stop: API returned complete message,
	// or a message terminated by one of the stop sequences provided via the stop parameter
	// length: Incomplete model output due to max_tokens parameter or token limit
	// function_call: The model decided to call a function
	// content_filter: Omitted content due to a flag from our content filters
	// null: API response still in progress or incomplete
	FinishReason FinishReason `json:"finish_reason"`
	LogProbs     *LogProbs    `json:"logprobs,omitempty"`
}

// CompletionResponse represents a response structure for completion API.
type CompletionResponse struct {
	ID                string             `json:"id"`
	Object            string             `json:"object"`
	Created           int64              `json:"created"`
	Model             string             `json:"model"`
	Choices           []CompletionChoice `json:"choices"`
	Citations         []string           `json:"citations"`
	Usage             *Usage             `json:"usage,omitempty"`
	SystemFingerprint string             `json:"system_fingerprint"`
}

// CreateCompletion — API call to Create a completion for the prompt.
func (c *Client) CreateCompletion(
	ctx context.Context,
	request CompletionRequest,
) (response CompletionResponse, err error) {
	if request.Stream {
		err = ErrCompletionStreamNotSupported
		return
	}

	if !isSupportingModel(completionsSuffix, request.Model) {
		err = ErrCompletionInvalidModel
		return
	}

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(completionsSuffix),
		withBody(request),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

type CompletionStream struct {
	stream   <-chan CompletionResponse
	done     chan struct{}
	response *http.Response
}

// CreateCompletionStream — API call to Create a completion for the prompt with streaming.
func (c *Client) CreateCompletionStream(
	ctx context.Context,
	request CompletionRequest,
) (*CompletionStream, error) {
	if !request.Stream {
		request.Stream = true
	}

	if !isSupportingModel(completionsSuffix, request.Model) {
		return nil, ErrCompletionInvalidModel
	}

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(completionsSuffix),
		withBody(request),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if isFailureStatusCode(resp) {
		return nil, c.handleErrorResp(resp)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.New("unexpected status code: " + resp.Status)
	}

	stream := make(chan CompletionResponse)
	done := make(chan struct{})

	go func() {
		defer close(stream)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				slog.Info("Stream stopped due to context cancellation")
				return
			default:
				line, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						return
					}
					slog.Error("failed to read completion stream", "error", err)
					return
				}
				// If stream ended with done, stop immediately
				if strings.HasSuffix(string(line), "[DONE]\n") {
					return
				}
				// Ignore openrouter comments, empty lines
				if strings.HasPrefix(string(line), ": OPENROUTER PROCESSING") || string(line) == "\n" {
					continue
				}
				// Trim everything before json object from line
				line = bytes.TrimPrefix(line, []byte("data:"))
				// Decode object into a CompletionResponse
				var chunk CompletionResponse
				if err := json.Unmarshal(line, &chunk); err != nil {
					slog.Error("failed to decode completion stream", "error", err, "line", string(line))
					return
				}
				stream <- chunk
			}
		}
	}()

	return &CompletionStream{
		stream:   stream,
		done:     done,
		response: resp,
	}, nil
}

// Recv reads the next chunk from the stream.
func (s *CompletionStream) Recv() (CompletionResponse, error) {
	select {
	case chunk, ok := <-s.stream:
		if !ok {
			return CompletionResponse{}, io.EOF
		}
		return chunk, nil
	case <-s.done:
		return CompletionResponse{}, io.EOF
	}
}

// Close terminates the stream and cleans up resources.
func (s *CompletionStream) Close() {
	close(s.done)
	if s.response != nil {
		s.response.Body.Close()
	}
}
