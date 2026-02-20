package openrouter

// Usage Represents the total token usage per request to OpenAI.
type Usage struct {
	PromptTokens           int                    `json:"prompt_tokens"`
	CompletionTokens       int                    `json:"completion_tokens"`
	CompletionTokenDetails CompletionTokenDetails `json:"completion_token_details"`
	TotalTokens            int                    `json:"total_tokens"`
	IsBYOK                 bool                   `json:"is_byok"`

	Cost        float64     `json:"cost"`
	CostDetails CostDetails `json:"cost_details"`

	PromptTokenDetails PromptTokenDetails `json:"prompt_token_details"`
}

type CostDetails struct {
	UpstreamInferenceCost           float64 `json:"upstream_inference_cost"`
	UpstreamInferencePromptCost     float64 `json:"upstream_inference_prompt_cost"`
	UpstreamInferenceCompletionCost float64 `json:"upstream_inference_completion_cost"`
}

type CompletionTokenDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
	ImageTokens     int `json:"image_tokens"`
}

type PromptTokenDetails struct {
	CachedTokens     int `json:"cached_tokens"`
	CacheWriteTokens int `json:"cache_write_tokens"`
	AudioTokens      int `json:"audio_tokens"`
	VideoTokens      int `json:"video_tokens"`
}
