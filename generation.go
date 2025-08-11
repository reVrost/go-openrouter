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
	UpstreamID             *string  `json:"upstream_id"`
	CacheDiscount          *float64 `json:"cache_discount"`
	UpstreamInferenceCost  *float64 `json:"upstream_inference_cost"`
	AppID                  *int     `json:"app_id"`
	Streamed               *bool    `json:"streamed"`
	Cancelled              *bool    `json:"cancelled"`
	ProviderName           *string  `json:"provider_name"`
	Latency                *int     `json:"latency"`
	ModerationLatency      *int     `json:"moderation_latency"`
	GenerationTime         *int     `json:"generation_time"`
	FinishReason           *string  `json:"finish_reason"`
	NativeFinishReason     *string  `json:"native_finish_reason"`
	TokensPrompt           *int     `json:"tokens_prompt"`
	TokensCompletion       *int     `json:"tokens_completion"`
	NativeTokensPrompt     *int     `json:"native_tokens_prompt"`
	NativeTokensCompletion *int     `json:"native_tokens_completion"`
	NativeTokensReasoning  *int     `json:"native_tokens_reasoning"`
	NumMediaPrompt         *int     `json:"num_media_prompt"`
	NumMediaCompletion     *int     `json:"num_media_completion"`
	NumSearchResults       *int     `json:"num_search_results"`
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
