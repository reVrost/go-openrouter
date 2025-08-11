package openrouter

import (
	"context"
	"net/http"
	"net/url"
)

const (
	getGenerationSuffix = "/generation"
)

type Generation struct {
	ID                     string   `json:"id"`
	TotalCost              float64  `json:"total_cost"`
	CreatedAt              string   `json:"created_at"`
	Model                  string   `json:"model"`
	Origin                 string   `json:"origin"`
	Usage                  float64  `json:"usage"`
	IsBYOK                 bool     `json:"is_byok"`
	UpstreamID             *string  `json:"upstream_id,omitempty"`
	CacheDiscount          *float64 `json:"cache_discount,omitempty"`
	UpstreamInferenceCost  *float64 `json:"upstream_inference_cost,omitempty"`
	AppID                  *int     `json:"app_id,omitempty"`
	Streamed               *bool    `json:"streamed,omitempty"`
	Cancelled              *bool    `json:"cancelled,omitempty"`
	ProviderName           *string  `json:"provider_name,omitempty"`
	Latency                *int     `json:"latency,omitempty"`
	ModerationLatency      *int     `json:"moderation_latency,omitempty"`
	GenerationTime         *int     `json:"generation_time,omitempty"`
	FinishReason           *string  `json:"finish_reason,omitempty"`
	NativeFinishReason     *string  `json:"native_finish_reason,omitempty"`
	TokensPrompt           *int     `json:"tokens_prompt,omitempty"`
	TokensCompletion       *int     `json:"tokens_completion,omitempty"`
	NativeTokensPrompt     *int     `json:"native_tokens_prompt,omitempty"`
	NativeTokensCompletion *int     `json:"native_tokens_completion,omitempty"`
	NativeTokensReasoning  *int     `json:"native_tokens_reasoning,omitempty"`
	NumMediaPrompt         *int     `json:"num_media_prompt,omitempty"`
	NumMediaCompletion     *int     `json:"num_media_completion,omitempty"`
	NumSearchResults       *int     `json:"num_search_results,omitempty"`
}

func (c *Client) GetGeneration(ctx context.Context, id string) (generation Generation, err error) {
	query := url.Values{}

	query.Set("id", id)

	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		c.fullURL(getGenerationSuffix, withQuery(query)),
	)
	if err != nil {
		return
	}

	var response struct {
		Data Generation `json:"data"`
	}

	err = c.sendRequest(req, &response)

	generation = response.Data
	return
}
