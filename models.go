package openrouter

import (
	"context"
	"net/http"
)

const (
	listModelsSuffix           = "/models"
	listUserModelsSuffix       = "/models/user"
	listEmbeddingsModelsSuffix = "/embeddings/models"
	listImageModelsSuffix      = "/images/models"
)

type ModelArchitecture struct {
	InputModalities  []string `json:"input_modalities"`
	Modality         *string  `json:"modality"`
	OutputModalities []string `json:"output_modalities"`
	InstructType     *string  `json:"instruct_type"`
	Tokenizer        string   `json:"tokenizer"`
}

type ModelDefaultParameters struct {
	FrequencyPenalty  *float64 `json:"frequency_penalty"`
	PresencePenalty   *float64 `json:"presence_penalty"`
	RepetitionPenalty *float64 `json:"repetition_penalty"`
	Temperature       *float64 `json:"temperature"`
	TopK              *int64   `json:"top_k"`
	TopP              *float64 `json:"top_p"`
}

type ModelLinks struct {
	Details string `json:"details"`
}

type ModelPerRequestLimits struct {
	CompletionTokens float64 `json:"completion_tokens"`
	PromptTokens     float64 `json:"prompt_tokens"`
}

type ModelPricing struct {
	Completion        string  `json:"completion"`
	Prompt            string  `json:"prompt"`
	Audio             string  `json:"audio"`
	AudioOutput       string  `json:"audio_output"`
	Discount          float64 `json:"discount"`
	Image             string  `json:"image"`
	ImageOutput       string  `json:"image_output"`
	ImageToken        string  `json:"image_token"`
	InputAudioCache   string  `json:"input_audio_cache"`
	InputCacheRead    string  `json:"input_cache_read"`
	InputCacheWrite   string  `json:"input_cache_write"`
	InternalReasoning string  `json:"internal_reasoning"`
	Request           string  `json:"request"`
	WebSearch         string  `json:"web_search"`
}

type ModelTopProvider struct {
	IsModerated         bool   `json:"is_moderated"`
	ContextLength       *int64 `json:"context_length"`
	MaxCompletionTokens *int64 `json:"max_completion_tokens"`
}

type ModelDesignArenaBenchmark struct {
	Arena    string  `json:"arena"`
	Category string  `json:"category"`
	Elo      float64 `json:"elo"`
	Rank     int64   `json:"rank"`
	WinRate  float64 `json:"win_rate"`
}

type ModelArtificialAnalysisBenchmark struct {
	AgenticIndex      *float64 `json:"agentic_index"`
	CodingIndex       *float64 `json:"coding_index"`
	IntelligenceIndex *float64 `json:"intelligence_index"`
}

type ModelBenchmarks struct {
	DesignArena        []ModelDesignArenaBenchmark      `json:"design_arena"`
	ArtificialAnalysis ModelArtificialAnalysisBenchmark `json:"artificial_analysis"`
}

type ModelReasoning struct {
	DefaultEffort     *string   `json:"default_effort"`
	SupportedEfforts  *[]string `json:"supported_efforts"`
	Mandatory         bool      `json:"mandatory"`
	DefaultEnabled    bool      `json:"default_enabled"`
	SupportsMaxTokens bool      `json:"supports_max_tokens"`
}

type Model struct {
	Architecture        ModelArchitecture      `json:"architecture"`
	CanonicalSlug       string                 `json:"canonical_slug"`
	ContextLength       *int64                 `json:"context_length"`
	Created             int64                  `json:"created"`
	DefaultParameters   ModelDefaultParameters `json:"default_parameters"`
	ID                  string                 `json:"id"`
	Links               ModelLinks             `json:"links"`
	Name                string                 `json:"name"`
	PerRequestLimits    ModelPerRequestLimits  `json:"per_request_limits"`
	Pricing             ModelPricing           `json:"pricing"`
	SupportedParameters []string               `json:"supported_parameters"`
	SupportedVoices     *[]string              `json:"supported_voices"`
	TopProvider         ModelTopProvider       `json:"top_provider"`
	Benchmarks          *ModelBenchmarks       `json:"benchmarks,omitempty"`
	Description         string                 `json:"description"`
	ExpirationDate      *string                `json:"expiration_date"`
	HuggingFaceID       *string                `json:"hugging_face_id"`
	KnowledgeCutoff     *string                `json:"knowledge_cutoff"`
	Reasoning           *ModelReasoning        `json:"reasoning,omitempty"`
}

type ImageModelArchitecture struct {
	InputModalities  []string `json:"input_modalities"`
	OutputModalities []string `json:"output_modalities"`
}

type CapabilityDescriptor struct {
	Type   string   `json:"type"`
	Values []string `json:"values,omitempty"`
	Max    *float64 `json:"max,omitempty"`
	Min    *float64 `json:"min,omitempty"`
}

type ImageModel struct {
	ID                  string                          `json:"id"`
	Name                string                          `json:"name"`
	Created             int64                           `json:"created"`
	Description         string                          `json:"description"`
	Endpoints           string                          `json:"endpoints"`
	Architecture        ImageModelArchitecture          `json:"architecture"`
	SupportedParameters map[string]CapabilityDescriptor `json:"supported_parameters"`
	SupportsStreaming   bool                            `json:"supports_streaming"`
}

// ListModels lists all models and their properties.
func (c *Client) ListModels(ctx context.Context) (models []Model, err error) {
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(listModelsSuffix),
	)
	if err != nil {
		return
	}

	var response struct {
		Data []Model `json:"data"`
	}

	err = c.sendRequest(req, &response)

	models = response.Data
	return
}

// ListUserModels lists models filtered by user provider preferences, privacy settings and guardrails.
func (c *Client) ListUserModels(ctx context.Context) (models []Model, err error) {
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(listUserModelsSuffix),
	)
	if err != nil {
		return
	}

	var response struct {
		Data []Model `json:"data"`
	}

	err = c.sendRequest(req, &response)

	models = response.Data
	return
}

// ListEmbeddingsModels returns all available embeddings models and their properties.
func (c *Client) ListEmbeddingsModels(ctx context.Context) ([]Model, error) {
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(listEmbeddingsModelsSuffix),
	)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []Model `json:"data"`
	}

	if err := c.sendRequest(req, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// ListImageModels returns all available image generation models and their properties.
// API reference: https://openrouter.ai/docs/api/api-reference/images/list-image-models
func (c *Client) ListImageModels(ctx context.Context) ([]ImageModel, error) {
	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(listImageModelsSuffix),
	)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []ImageModel `json:"data"`
	}

	if err := c.sendRequest(req, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}
