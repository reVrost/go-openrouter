package openrouter_test

import (
	"encoding/json"
	"testing"

	openrouter "github.com/revrost/go-openrouter"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalChatCompletionResponse(t *testing.T) {
	var resp openrouter.ChatCompletionResponse
	err := json.Unmarshal([]byte(exampleJsonResonse), &resp)
	require.NoError(t, err)

	require.Equal(t, "gen-1777603778-00000000000", resp.ID)
	require.Equal(t, "chat.completion", resp.Object)
	require.Equal(t, int64(1777603778), resp.Created)
	require.Equal(t, "anthropic/claude-4.5-sonnet-20250929", resp.Model)
	require.Equal(t, "Amazon Bedrock", resp.Provider)

	require.Len(t, resp.Choices, 1)
	choice := resp.Choices[0]
	require.Equal(t, 0, choice.Index)
	require.Equal(t, openrouter.FinishReasonStop, choice.FinishReason)
	require.Equal(t, "end_turn", choice.NativeFinishReason)
	require.Nil(t, choice.LogProbs)

	msg := choice.Message
	require.Equal(t, "assistant", msg.Role)
	require.Equal(t, "Hey! Need help with anything?", msg.Content.Text)
	require.NotNil(t, msg.Reasoning)
	require.Contains(t, *msg.Reasoning, "hello")

	require.Len(t, msg.ReasoningDetails, 1)
	rd := msg.ReasoningDetails[0]
	require.Equal(t, openrouter.ReasoningDetailsTypeText, rd.Type)
	require.Equal(t, "anthropic-claude-v1", rd.Format)
	require.Equal(t, 0, rd.Index)
	require.Contains(t, rd.Text, "hello")

	require.NotNil(t, resp.Usage)
	usage := resp.Usage
	require.Equal(t, 7325, usage.PromptTokens)
	require.Equal(t, 42, usage.CompletionTokens)
	require.Equal(t, 7367, usage.TotalTokens)
	require.InDelta(t, 0.0110382, usage.Cost, 1e-9)
	require.False(t, usage.IsBYOK)

	require.Equal(t, 29, usage.CompletionTokenDetails.ReasoningTokens)
	require.Equal(t, 0, usage.CompletionTokenDetails.ImageTokens)
	require.Equal(t, 0, usage.CompletionTokenDetails.AudioTokens)

	require.InDelta(t, 0.0110382, usage.CostDetails.UpstreamInferenceCost, 1e-9)
	require.InDelta(t, 0.0104082, usage.CostDetails.UpstreamInferencePromptCost, 1e-9)
	require.InDelta(t, 0.00063, usage.CostDetails.UpstreamInferenceCompletionCost, 1e-9)

	// NOTE: JSON field is "prompt_tokens_details" but struct tag is "prompt_token_details" — mismatch causes zero values here
	require.Equal(t, 4284, usage.PromptTokenDetails.CachedTokens)
	require.Equal(t, 0, usage.PromptTokenDetails.CacheWriteTokens)
}

// exampleJsonResonse is a real response from openrouter. Used for testing if the marshaller constructs it correctly
const exampleJsonResonse = `{
    "id": "gen-1777603778-00000000000",
    "object": "chat.completion",
    "created": 1777603778,
    "model": "anthropic/claude-4.5-sonnet-20250929",
    "provider": "Amazon Bedrock",
    "system_fingerprint": null,
    "choices":
    [
        {
            "index": 0,
            "logprobs": null,
            "finish_reason": "stop",
            "native_finish_reason": "end_turn",
            "message":
            {
                "role": "assistant",
                "content": "Hey! Need help with anything?",
                "refusal": null,
                "reasoning": "The user is just saying \"hello\" repeatedly. I should respond politely but briefly and ask if they need assistance.",
                "reasoning_details":
                [
                    {
                        "type": "reasoning.text",
                        "text": "The user is just saying \"hello\" repeatedly. I should respond politely but briefly and ask if they need assistance.",
                        "format": "anthropic-claude-v1",
                        "index": 0,
                        "signature": "Ep0CCkgIDRABGAIqQBwSUgzjiTJplyZOsH8pmHQU+3AC4n77QQa4i6TlwQsr+I7JkJuciaNZnopF9wYg5PWu31xHy8d6Zh+Mr/aYpwYSDPb1E7zmX4R2zwz57BoM3bj4as+jjD7EEgx0IjDdJapNZ8MSK/Ek3aVUYkp1pNLmZeTFaKeDfgG1g79gEhGqRoBeF3TV95DYjMvxsHMqggFTq6dkS911hd8881Y6QxEul5EZFtLUZxjgcCvXSsY1TUE7BK7DHEGd/vPVPMVcx9Qc8vjLk7kf8xF8O7f/lbkv/IAAGr+3znzN3rcuYsFN+mg9utr99i0m+RWho+Aq3m1AGi2Me0efMVQ+vPfldFNSANYT9ZfxvCG+KHjSU4AZ29puGAE="
                    }
                ]
            }
        }
    ],
    "usage":
    {
        "prompt_tokens": 7325,
        "completion_tokens": 42,
        "total_tokens": 7367,
        "cost": 0.0110382,
        "is_byok": false,
        "prompt_tokens_details":
        {
            "cached_tokens": 4284,
            "cache_write_tokens": 0,
            "audio_tokens": 0,
            "video_tokens": 0
        },
        "cost_details":
        {
            "upstream_inference_cost": 0.0110382,
            "upstream_inference_prompt_cost": 0.0104082,
            "upstream_inference_completions_cost": 0.00063
        },
        "completion_tokens_details":
        {
            "reasoning_tokens": 29,
            "image_tokens": 0,
            "audio_tokens": 0
        }
    }
}`
