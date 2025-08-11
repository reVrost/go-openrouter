package openrouter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	GPT4o                  = "openai/chatgpt-4o-latest"
	DeepseekV3             = "deepseek/deepseek-chat"
	DeepseekR1             = "deepseek/deepseek-r1"
	DeepseekR1DistillLlama = "deepseek/deepseek-r1-distill-llama-70b"
	LiquidLFM7B            = "liquid/lfm-7b"
	Phi3Mini               = "microsoft/phi-3-mini-128k-instruct:free"
	GeminiFlashExp         = "google/gemini-2.0-flash-exp:free"
	GeminiProExp           = "google/gemini-pro-1.5-exp"
	GeminiFlash8B          = "google/gemini-flash-1.5-8b"
	GPT4oMini              = "openai/gpt-4o-mini"
)

// Chat message role defined by the Openrouter API.
const (
	ChatMessageRoleSystem    = "system"
	ChatMessageRoleUser      = "user"
	ChatMessageRoleAssistant = "assistant"
	ChatMessageRoleFunction  = "function"
	ChatMessageRoleTool      = "tool"
)
const chatCompletionsSuffix = "/chat/completions"

var (
	ErrChatCompletionInvalidModel       = errors.New("this model is not supported with this method, please use CreateCompletion client method instead") //nolint:lll
	ErrChatCompletionStreamNotSupported = errors.New("streaming is not supported with this method, please use CreateChatCompletionStream")              //nolint:lll
	ErrContentFieldsMisused             = errors.New("can't use both Content and MultiContent properties simultaneously")
)

type ChatCompletionReasoning struct {
	// Effort The prompt that was used to generate the reasoning. [high, medium, low]
	Effort *string `json:"prompt,omitempty"`

	// MaxTokens cannot be simultaneously used with effort.
	MaxTokens *int `json:"max_tokens,omitempty"`

	// Exclude defaults to false.
	Exclude *bool `json:"exclude,omitempty"`
}

type PluginID string

const (
	// Processing PDFs: https://openrouter.ai/docs/features/images-and-pdfs#processing-pdfs
	PluginIDFileParser PluginID = "file-parser"
	// Web search plugin: https://openrouter.ai/docs/features/web-search
	PluginIDWeb PluginID = "web"
)

type PDFEngine string

const (
	// Best for scanned documents or PDFs with images ($2 per 1,000 pages).
	PDFEngineMistralOCR PDFEngine = "mistral-ocr"
	// Best for well-structured PDFs with clear text content (Free).
	PDFEnginePDFText PDFEngine = "pdf-text"
	// Only available for models that support file input natively (charged as input tokens).
	PDFEngineNative PDFEngine = "native"
)

type ChatCompletionPlugin struct {
	ID  PluginID   `json:"id"`
	PDF *PDFPlugin `json:"pdf,omitempty"`
}

type PDFPlugin struct {
	Engine string `json:"engine"`
}

type ChatCompletionRequest struct {
	Model string `json:"model,omitempty"`
	// Optional model fallbacks: https://openrouter.ai/docs/features/model-routing#the-models-parameter
	Models   []string                `json:"models,omitempty"`
	Provider *ChatProvider           `json:"provider,omitempty"`
	Messages []ChatCompletionMessage `json:"messages"`

	Reasoning *ChatCompletionReasoning `json:"reasoning,omitempty"`

	Plugins []ChatCompletionPlugin `json:"plugins,omitempty"`

	// MaxTokens The maximum number of tokens that can be generated in the chat completion.
	// This value can be used to control costs for text generated via API.
	MaxTokens int `json:"max_tokens,omitempty"`
	// MaxCompletionTokens Upper bound for completion tokens, supported for OpenAI API compliance.
	// Prefer "max_tokens" for limiting output in new integrations.
	// refs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-max_completion_tokens
	MaxCompletionTokens int                           `json:"max_completion_tokens,omitempty"`
	Temperature         float32                       `json:"temperature,omitempty"`
	TopP                float32                       `json:"top_p,omitempty"`
	TopK                int                           `json:"top_k,omitempty"`
	TopA                float32                       `json:"top_a,omitempty"`
	N                   int                           `json:"n,omitempty"`
	Stream              bool                          `json:"stream,omitempty"`
	Stop                []string                      `json:"stop,omitempty"`
	PresencePenalty     float32                       `json:"presence_penalty,omitempty"`
	RepetitionPenalty   float32                       `json:"repetition_penalty,omitempty"`
	ResponseFormat      *ChatCompletionResponseFormat `json:"response_format,omitempty"`
	Seed                *int                          `json:"seed,omitempty"`
	MinP                float32                       `json:"min_p,omitempty"`
	FrequencyPenalty    float32                       `json:"frequency_penalty,omitempty"`
	// LogitBias is must be a token id string (specified by their token ID in the tokenizer), not a word string.
	// incorrect: `"logit_bias":{"You": 6}`, correct: `"logit_bias":{"1639": 6}`
	// refs: https://platform.openai.com/docs/api-reference/chat/create#chat/create-logit_bias
	LogitBias map[string]int `json:"logit_bias,omitempty"`
	// LogProbs indicates whether to return log probabilities of the output tokens or not.
	// If true, returns the log probabilities of each output token returned in the content of message.
	// This option is currently not available on the gpt-4-vision-preview model.
	LogProbs bool `json:"logprobs,omitempty"`
	// TopLogProbs is an integer between 0 and 5 specifying the number of most likely tokens to return at each
	// token position, each with an associated log probability.
	// logprobs must be set to true if this parameter is used.
	TopLogProbs int    `json:"top_logprobs,omitempty"`
	User        string `json:"user,omitempty"`
	// Deprecated: use Tools instead.
	Functions []FunctionDefinition `json:"functions,omitempty"`
	// Deprecated: use ToolChoice instead.
	FunctionCall any    `json:"function_call,omitempty"`
	Tools        []Tool `json:"tools,omitempty"`
	// This can be either a string or an ToolChoice object.
	ToolChoice any `json:"tool_choice,omitempty"`
	// Options for streaming response. Only set this when you set stream: true.
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`
	// Disable the default behavior of parallel tool calls by setting it: false.
	ParallelToolCalls any `json:"parallel_tool_calls,omitempty"`
	// Store can be set to true to store the output of this completion request for use in distillations and evals.
	// https://platform.openai.com/docs/api-reference/chat/create#chat-create-store
	Store bool `json:"store,omitempty"`
	// Metadata to store with the completion.
	Metadata map[string]string `json:"metadata,omitempty"`
	// Apply message transforms
	// https://openrouter.ai/docs/features/message-transforms
	Transforms []string `json:"transforms,omitempty"`
	// Optional web search options
	// https://openrouter.ai/docs/features/web-search#specifying-search-context-size
	WebSearchOptions *WebSearchOptions `json:"web_search_options,omitempty"`

	Usage *IncludeUsage `json:"usage,omitempty"`
}

type SearchContextSize string

const (
	SearchContextSizeLow    SearchContextSize = "low"
	SearchContextSizeMedium SearchContextSize = "medium"
	SearchContextSizeHigh   SearchContextSize = "high"
)

type WebSearchOptions struct {
	SearchContextSize SearchContextSize `json:"search_context_size"`
}

type IncludeUsage struct {
	Include bool `json:"include"`
}

type DataCollection string

const (
	DataCollectionAllow DataCollection = "allow"
	DataCollectionDeny  DataCollection = "deny"
)

type ProviderSorting string

const (
	ProviderSortingPrice      ProviderSorting = "price"
	ProviderSortingThroughput ProviderSorting = "throughput"
	ProviderSortingLatency    ProviderSorting = "latency"
)

// Provider Routing: https://openrouter.ai/docs/features/provider-routing
type ChatProvider struct {
	// The order of the providers in the list determines the order in which they are called.
	Order []string `json:"order,omitempty"`
	// Allow fallbacks to other providers if the primary provider fails.
	AllowFallbacks bool `json:"allow_fallbacks,omitempty"`
	// Only use providers that support all parameters in your request.
	RequireParameters bool `json:"require_parameters,omitempty"`
	// Control whether to use providers that may store data.
	DataCollection DataCollection `json:"data_collection,omitempty"`
	// List of provider slugs to allow for this request.
	Only []string `json:"only,omitempty"`
	// List of provider slugs to skip for this request.
	Ignore []string `json:"ignore,omitempty"`
	// List of quantization levels to filter by (e.g. ["int4", "int8"]).
	Quantizations []string `json:"quantizations,omitempty"`
	// Sort providers by price or throughput. (e.g. "price" or "throughput").
	Sort ProviderSorting `json:"sort,omitempty"`
}

// ChatCompletionResponse represents a response structure for chat completion API.
type ChatCompletionResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Citations         []string               `json:"citations"`
	Usage             *Usage                 `json:"usage,omitempty"`
	SystemFingerprint string                 `json:"system_fingerprint"`

	// http.Header
}

type TopLogProbs struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
	Bytes   []byte  `json:"bytes,omitempty"`
}

// LogProb represents the probability information for a token.
type LogProb struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
	Bytes   []byte  `json:"bytes,omitempty"` // Omitting the field if it is null
	// TopLogProbs is a list of the most likely tokens and their log probability, at this token position.
	// In rare cases, there may be fewer than the number of requested top_logprobs returned.
	TopLogProbs []TopLogProbs `json:"top_logprobs"`
}

// LogProbs is the top-level structure containing the log probability information.
type LogProbs struct {
	// Content is a list of message content tokens with log probability information.
	Content []LogProb `json:"content"`
}

type FinishReason string

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonLength        FinishReason = "length"
	FinishReasonFunctionCall  FinishReason = "function_call"
	FinishReasonToolCalls     FinishReason = "tool_calls"
	FinishReasonContentFilter FinishReason = "content_filter"
	FinishReasonNull          FinishReason = "null"
)

func (r FinishReason) MarshalJSON() ([]byte, error) {
	if r == FinishReasonNull || r == "" {
		return []byte("null"), nil
	}
	return []byte(`"` + string(r) + `"`), nil // best effort to not break future API changes
}

type ChatCompletionChoice struct {
	Index   int                   `json:"index"`
	Message ChatCompletionMessage `json:"message"`
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

type PromptAnnotation struct {
	PromptIndex          int                  `json:"prompt_index,omitempty"`
	ContentFilterResults ContentFilterResults `json:"content_filter_results,omitempty"`
}
type ContentFilterResults struct {
	Hate      Hate      `json:"hate,omitempty"`
	SelfHarm  SelfHarm  `json:"self_harm,omitempty"`
	Sexual    Sexual    `json:"sexual,omitempty"`
	Violence  Violence  `json:"violence,omitempty"`
	JailBreak JailBreak `json:"jailbreak,omitempty"`
	Profanity Profanity `json:"profanity,omitempty"`
}
type PromptFilterResult struct {
	Index                int                  `json:"index"`
	ContentFilterResults ContentFilterResults `json:"content_filter_results,omitempty"`
}
type Hate struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity,omitempty"`
}
type SelfHarm struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity,omitempty"`
}
type Sexual struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity,omitempty"`
}
type Violence struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity,omitempty"`
}

type JailBreak struct {
	Filtered bool `json:"filtered"`
	Detected bool `json:"detected"`
}
type Profanity struct {
	Filtered bool `json:"filtered"`
	Detected bool `json:"detected"`
}

type StreamOptions struct {
	// If set, an additional chunk will be streamed before the data: [DONE] message.
	// The usage field on this chunk shows the token usage statistics for the entire request,
	// and the choices field will always be an empty array.
	// All other chunks will also include a usage field, but with a null value.
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type ChatCompletionResponseFormatType string

const (
	ChatCompletionResponseFormatTypeJSONObject ChatCompletionResponseFormatType = "json_object"
	ChatCompletionResponseFormatTypeJSONSchema ChatCompletionResponseFormatType = "json_schema"
	ChatCompletionResponseFormatTypeText       ChatCompletionResponseFormatType = "text"
)

type ChatCompletionResponseFormat struct {
	Type       ChatCompletionResponseFormatType        `json:"type,omitempty"`
	JSONSchema *ChatCompletionResponseFormatJSONSchema `json:"json_schema,omitempty"`
}

type ChatCompletionResponseFormatJSONSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Schema      json.Marshaler `json:"schema"`
	Strict      bool           `json:"strict"`
}

type FunctionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Strict      bool   `json:"strict,omitempty"`
	// Parameters is an object describing the function.
	// You can pass json.RawMessage to describe the schema,
	// or you can pass in a struct which serializes to the proper JSON schema.
	// The jsonschema package is provided for convenience, but you should
	// consider another specialized library if you require more complex schemas.
	Parameters any `json:"parameters"`
}

type ChatMessagePartType string

const (
	ChatMessagePartTypeText     ChatMessagePartType = "text"
	ChatMessagePartTypeImageURL ChatMessagePartType = "image_url"
	ChatMessagePartTypeFile     ChatMessagePartType = "file"
)

type ChatMessagePart struct {
	Type ChatMessagePartType `json:"type,omitempty"`
	Text string              `json:"text,omitempty"`
	// Prompt caching
	// https://openrouter.ai/docs/features/prompt-caching
	CacheControl *CacheControl `json:"cache_control,omitempty"`

	ImageURL *ChatMessageImageURL `json:"image_url,omitempty"`
	File     *FileContent         `json:"file,omitempty"`
}

type ImageURLDetail string

const (
	ImageURLDetailHigh ImageURLDetail = "high"
	ImageURLDetailLow  ImageURLDetail = "low"
	ImageURLDetailAuto ImageURLDetail = "auto"
)

type ChatMessageImageURL struct {
	URL    string         `json:"url,omitempty"`
	Detail ImageURLDetail `json:"detail,omitempty"`
}

// FileContent represents file content for PDF processing
type FileContent struct {
	Filename string `json:"filename"`
	FileData string `json:"file_data"`
}

// Content handles both string and multi-part content.
type Content struct {
	Text  string
	Multi []ChatMessagePart
}

type Annotation struct {
	Type        AnnotationType `json:"type"`
	URLCitation URLCitation    `json:"url_citation"`
}

type AnnotationType string

const (
	AnnotationTypeUrlCitation AnnotationType = "url_citation"
)

type URLCitation struct {
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	URL        string `json:"url"`
}

type CacheControl struct {
	// Type only supports "ephemeral" for now.
	Type string `json:"type"`
	// TTL in  format of "5m" | "1h"
	TTL *string `json:"ttl,omitempty"`
}

type ChatCompletionMessage struct {
	Role    string  `json:"role"`
	Content Content `json:"content,omitzero"`
	Refusal string  `json:"refusal,omitempty"`

	// This property is used for the "reasoning" feature supported by deepseek-reasoner
	// - https://api-docs.deepseek.com/api/create-chat-completion#responses
	ReasoningContent *string `json:"reasoning_content,omitempty"`

	// Reasoning Used by all the other models
	Reasoning *string `json:"reasoning,omitempty"`

	FunctionCall *FunctionCall `json:"function_call,omitempty"`

	// For Role=assistant prompts this may be set to the tool calls generated by the model, such as function calls.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// For Role=tool prompts this should be set to the ID given in the assistant's prior request to call a tool.
	ToolCallID string `json:"tool_call_id,omitempty"`

	// Web Search Annotations
	Annotations []Annotation `json:"annotations,omitempty"`
}

// MarshalJSON serializes ContentType as a string or array.
func (c Content) MarshalJSON() ([]byte, error) {
	if c.Text != "" && len(c.Multi) == 0 {
		return json.Marshal(c.Text)
	}
	if len(c.Multi) > 0 && c.Text == "" {
		return json.Marshal(c.Multi)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON deserializes ContentType from a string or array.
func (c *Content) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil && s != "" {
		c.Text = s
		c.Multi = nil
		return nil
	}

	var parts []ChatMessagePart
	if err := json.Unmarshal(data, &parts); err == nil && len(parts) > 0 {
		c.Text = ""
		c.Multi = parts
		return nil
	}

	c.Text = ""
	c.Multi = nil
	return nil
}

type Tool struct {
	Type     ToolType            `json:"type"`
	Function *FunctionDefinition `json:"function,omitempty"`
}

type ToolType string

const (
	ToolTypeFunction ToolType = "function"
)

type ToolCall struct {
	// Index is not nil only in chat completion chunk object
	Index    *int         `json:"index,omitempty"`
	ID       string       `json:"id,omitempty"`
	Type     ToolType     `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name string `json:"name,omitempty"`
	// call function with arguments in JSON format
	Arguments string `json:"arguments,omitempty"`
}

func isSupportingModel(suffix, model string) bool {
	return true
}

// CreateChatCompletion — API call to Create a completion for the chat message.
func (c *Client) CreateChatCompletion(
	ctx context.Context,
	request ChatCompletionRequest,
) (response ChatCompletionResponse, err error) {
	if request.Stream {
		err = ErrChatCompletionStreamNotSupported
		return
	}

	if !isSupportingModel(chatCompletionsSuffix, request.Model) {
		err = ErrChatCompletionInvalidModel
		return
	}

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(chatCompletionsSuffix),
		withBody(request),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

type ChatCompletionStream struct {
	stream   <-chan ChatCompletionStreamResponse
	done     chan struct{}
	response *http.Response
}

// CreateChatCompletionStream — API call to Create a completion for the chat message with streaming.
func (c *Client) CreateChatCompletionStream(
	ctx context.Context,
	request ChatCompletionRequest,
) (*ChatCompletionStream, error) {
	if !request.Stream {
		request.Stream = true
	}

	if !isSupportingModel(chatCompletionsSuffix, request.Model) {
		return nil, ErrChatCompletionInvalidModel
	}

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(chatCompletionsSuffix),
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

	stream := make(chan ChatCompletionStreamResponse)
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
				log.Info().Msg("Stream stopped due to context cancellation")
				return
			default:
				line, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						return
					}
					log.Error().Err(err).Msg("failed to read chat completion stream")
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
				// Decode object into a ChatCompletionResponse
				var chunk ChatCompletionStreamResponse
				if err := json.Unmarshal(line, &chunk); err != nil {
					log.Error().Err(err).
						Str("line", string(line)).
						Msg("failed to decode chat completion stream")
					return
				}
				stream <- chunk
			}
		}
	}()

	return &ChatCompletionStream{
		stream:   stream,
		done:     done,
		response: resp,
	}, nil
}

type ChatCompletionStreamChoiceDelta struct {
	Content      string        `json:"content,omitempty"`
	Role         string        `json:"role,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
	ToolCalls    []ToolCall    `json:"tool_calls,omitempty"`
	Refusal      string        `json:"refusal,omitempty"`
	Reasoning    *string       `json:"reasoning,omitempty"`

	// This property is used for the "reasoning" feature supported by deepseek-reasoner
	// which is not in the official documentation.
	// the doc from deepseek:
	// - https://api-docs.deepseek.com/api/create-chat-completion#responses
	ReasoningContent string `json:"reasoning_content,omitempty"`
}
type ChatCompletionStreamChoiceLogprobs struct {
	Content []ChatCompletionTokenLogprob `json:"content,omitempty"`
	Refusal []ChatCompletionTokenLogprob `json:"refusal,omitempty"`
}
type ChatCompletionTokenLogprob struct {
	Token       string                                 `json:"token"`
	Bytes       []int64                                `json:"bytes,omitempty"`
	Logprob     float64                                `json:"logprob,omitempty"`
	TopLogprobs []ChatCompletionTokenLogprobTopLogprob `json:"top_logprobs"`
}
type ChatCompletionTokenLogprobTopLogprob struct {
	Token   string  `json:"token"`
	Bytes   []int64 `json:"bytes"`
	Logprob float64 `json:"logprob"`
}
type ChatCompletionStreamChoice struct {
	Index                int                                 `json:"index"`
	Delta                ChatCompletionStreamChoiceDelta     `json:"delta"`
	Logprobs             *ChatCompletionStreamChoiceLogprobs `json:"logprobs,omitempty"`
	FinishReason         FinishReason                        `json:"finish_reason"`
	ContentFilterResults *ContentFilterResults               `json:"content_filter_results,omitempty"`
}

type ChatCompletionStreamResponse struct {
	ID                  string                       `json:"id"`
	Object              string                       `json:"object"`
	Created             int64                        `json:"created"`
	Model               string                       `json:"model"`
	Choices             []ChatCompletionStreamChoice `json:"choices"`
	SystemFingerprint   string                       `json:"system_fingerprint"`
	PromptAnnotations   []PromptAnnotation           `json:"prompt_annotations,omitempty"`
	PromptFilterResults []PromptFilterResult         `json:"prompt_filter_results,omitempty"`
	// An optional field that will only be present when you set stream_options: {"include_usage": true} in your request.
	// When present, it contains a null value except for the last chunk which contains the token usage statistics
	// for the entire request.
	Usage *Usage `json:"usage,omitempty"`
}

// Recv reads the next chunk from the stream.
func (s *ChatCompletionStream) Recv() (ChatCompletionStreamResponse, error) {
	select {
	case chunk, ok := <-s.stream:
		if !ok {
			return ChatCompletionStreamResponse{}, io.EOF
		}
		return chunk, nil
	case <-s.done:
		return ChatCompletionStreamResponse{}, io.EOF
	}
}

// Close terminates the stream and cleans up resources.
func (s *ChatCompletionStream) Close() {
	close(s.done)
	if s.response != nil {
		s.response.Body.Close()
	}
}

// String is a helper function returns a pointer to the string value passed in.
func String(s string) *string {
	return &s
}

// SetLogLevel sets the minimum log level for the internally used logger.
func SetLogLevel(level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
}

// DisableLogs disables the internally used logger.
func DisableLogs() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}
