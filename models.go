package openrouter

import (
	"context"
	"net/http"
)

const (
	listModelsSuffix           = "/models"
	listUserModelsSuffix       = "/models/user"
	listEmbeddingsModelsSuffix = "/embeddings/models"
)

type ModelArchitecture struct {
	InputModalities  []string `json:"input_modalities"`
	OutputModalities []string `json:"output_modalities"`
	Tokenizer        string   `json:"tokenizer"`
	InstructType     *string  `json:"instruct_type,omitempty"`
}

type ModelTopProvider struct {
	IsModerated         bool   `json:"is_moderated"`
	ContextLength       *int64 `json:"context_length,omitempty"`
	MaxCompletionTokens *int64 `json:"max_completion_tokens,omitempty"`
}

type ModelPricing struct {
	Prompt            string  `json:"prompt"`
	Completion        string  `json:"completion"`
	Image             string  `json:"image"`
	Request           string  `json:"request"`
	WebSearch         string  `json:"web_search"`
	InternalReasoning string  `json:"internal_reasoning"`
	InputCacheRead    *string `json:"input_cache_read,omitempty"`
	InputCacheWrite   *string `json:"input_cache_write,omitempty"`
}

type Model struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Created             int64             `json:"created"`
	Description         string            `json:"description"`
	Architecture        ModelArchitecture `json:"architecture"`
	TopProvider         ModelTopProvider  `json:"top_provider"`
	Pricing             ModelPricing      `json:"pricing"`
	CanonicalSlug       *string           `json:"canonical_slug,omitempty"`
	ContextLength       *int64            `json:"context_length,omitempty"`
	HuggingFaceID       *string           `json:"hugging_face_id,omitempty"`
	PerRequestLimits    any               `json:"per_request_limits,omitempty"`
	SupportedParameters []string          `json:"supported_parameters,omitempty"`
}

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
// API reference: https://openrouter.ai/docs/api/api-reference/embeddings/list-embeddings-models
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
