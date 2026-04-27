package openrouter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type sequenceHTTPClient struct {
	requests  []ChatCompletionRequest
	responses []*http.Response
}

func (s *sequenceHTTPClient) Do(req *http.Request) (*http.Response, error) {
	var chatReq ChatCompletionRequest
	if req.Body != nil {
		err := json.NewDecoder(req.Body).Decode(&chatReq)
		if err != nil {
			return nil, err
		}
	}
	s.requests = append(s.requests, chatReq)

	resp := s.responses[0]
	s.responses = s.responses[1:]
	return resp, nil
}

func TestErrorCodeHelpers(t *testing.T) {
	t.Parallel()

	require.True(t, IsHTTPStatus(&APIError{HTTPStatusCode: http.StatusPaymentRequired}, http.StatusPaymentRequired))
	require.True(t, IsHTTPStatus(&RequestError{HTTPStatusCode: http.StatusTooManyRequests}, http.StatusTooManyRequests))
	require.False(t, IsHTTPStatus(io.EOF, http.StatusPaymentRequired))
	require.True(t, IsAPIErrorCode(&APIError{Code: http.StatusPaymentRequired}, http.StatusPaymentRequired))
	require.True(t, IsAPIErrorCode(&APIError{Code: "402"}, http.StatusPaymentRequired))
	require.True(t, IsErrorCode(&APIError{HTTPStatusCode: http.StatusPaymentRequired}, http.StatusPaymentRequired))
	require.True(t, IsErrorCode(&APIError{Code: http.StatusPaymentRequired}, http.StatusPaymentRequired))
	require.False(t, IsErrorCode(&APIError{HTTPStatusCode: http.StatusBadRequest}, http.StatusPaymentRequired))
}

func TestDefaultChatCompletionFallbackErrorCodes(t *testing.T) {
	t.Parallel()

	codes := DefaultChatCompletionFallbackErrorCodes()
	require.Contains(t, codes, http.StatusPaymentRequired)
	require.Contains(t, codes, http.StatusBadGateway)
	require.Contains(t, codes, http.StatusGatewayTimeout)
	require.Contains(t, codes, StatusProviderOverloaded)

	codes[0] = http.StatusTeapot
	require.NotEqual(t, http.StatusTeapot, DefaultChatCompletionFallbackErrorCodes()[0])
}

func TestCreateChatCompletionWithFallbackRetriesOnDefaultHTTPStatus(t *testing.T) {
	t.Parallel()

	httpClient := &sequenceHTTPClient{
		responses: []*http.Response{
			jsonResponse(http.StatusPaymentRequired, `{"error":{"code":402,"message":"insufficient funds"}}`),
			jsonResponse(http.StatusOK, `{
				"id":"chatcmpl_1",
				"object":"chat.completion",
				"model":"xiaomi/mimo-v2-flash",
				"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]
			}`),
		},
	}
	cfg := DefaultConfig("test-token")
	cfg.HTTPClient = httpClient
	cfg.BaseURL = "https://example.com/api/v1"
	client := NewClientWithConfig(*cfg)

	resp, err := client.CreateChatCompletionWithFallback(context.Background(), ChatCompletionRequest{
		Model:    "deepseek/deepseek-v4-flash",
		Messages: []ChatCompletionMessage{UserMessage("hello")},
	}, "xiaomi/mimo-v2-flash")

	require.NoError(t, err)
	require.Equal(t, "xiaomi/mimo-v2-flash", resp.Model)
	require.Len(t, httpClient.requests, 2)
	require.Equal(t, "deepseek/deepseek-v4-flash", httpClient.requests[0].Model)
	require.Equal(t, "xiaomi/mimo-v2-flash", httpClient.requests[1].Model)
}

func TestCreateChatCompletionWithFallbackRetriesOnDefaultAPIErrorCode(t *testing.T) {
	t.Parallel()

	httpClient := &sequenceHTTPClient{
		responses: []*http.Response{
			jsonResponse(http.StatusBadRequest, `{"error":{"code":402,"message":"Insufficient Balance"}}`),
			jsonResponse(http.StatusOK, `{
				"id":"chatcmpl_1",
				"object":"chat.completion",
				"model":"xiaomi/mimo-v2-flash",
				"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]
			}`),
		},
	}
	cfg := DefaultConfig("test-token")
	cfg.HTTPClient = httpClient
	cfg.BaseURL = "https://example.com/api/v1"
	client := NewClientWithConfig(*cfg)

	resp, err := client.CreateChatCompletionWithFallback(context.Background(), ChatCompletionRequest{
		Model:    "deepseek/deepseek-v4-flash",
		Messages: []ChatCompletionMessage{UserMessage("hello")},
	}, "xiaomi/mimo-v2-flash")

	require.NoError(t, err)
	require.Equal(t, "xiaomi/mimo-v2-flash", resp.Model)
	require.Len(t, httpClient.requests, 2)
}

func TestCreateChatCompletionWithFallbackTriesMultipleModels(t *testing.T) {
	t.Parallel()

	httpClient := &sequenceHTTPClient{
		responses: []*http.Response{
			jsonResponse(http.StatusPaymentRequired, `{"error":{"code":402,"message":"insufficient funds"}}`),
			jsonResponse(StatusProviderOverloaded, `{"error":{"code":529,"message":"provider overloaded"}}`),
			jsonResponse(http.StatusOK, `{
				"id":"chatcmpl_1",
				"object":"chat.completion",
				"model":"google/gemini-flash-1.5-8b",
				"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]
			}`),
		},
	}
	cfg := DefaultConfig("test-token")
	cfg.HTTPClient = httpClient
	cfg.BaseURL = "https://example.com/api/v1"
	client := NewClientWithConfig(*cfg)

	resp, err := client.CreateChatCompletionWithFallback(context.Background(), ChatCompletionRequest{
		Model:    "deepseek/deepseek-v4-flash",
		Messages: []ChatCompletionMessage{UserMessage("hello")},
	}, "xiaomi/mimo-v2-flash", "google/gemini-flash-1.5-8b")

	require.NoError(t, err)
	require.Equal(t, "google/gemini-flash-1.5-8b", resp.Model)
	require.Len(t, httpClient.requests, 3)
	require.Equal(t, "deepseek/deepseek-v4-flash", httpClient.requests[0].Model)
	require.Equal(t, "xiaomi/mimo-v2-flash", httpClient.requests[1].Model)
	require.Equal(t, "google/gemini-flash-1.5-8b", httpClient.requests[2].Model)
}

func TestCreateChatCompletionWithFallbackUsesCustomErrorCodes(t *testing.T) {
	t.Parallel()

	httpClient := &sequenceHTTPClient{
		responses: []*http.Response{
			jsonResponse(http.StatusBadRequest, `{"error":{"code":400,"message":"bad request"}}`),
			jsonResponse(http.StatusOK, `{
				"id":"chatcmpl_1",
				"object":"chat.completion",
				"model":"xiaomi/mimo-v2-flash",
				"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]
			}`),
		},
	}
	cfg := DefaultConfig("test-token")
	cfg.HTTPClient = httpClient
	cfg.BaseURL = "https://example.com/api/v1"
	client := NewClientWithConfig(*cfg)

	resp, err := client.CreateChatCompletionWithFallbackPolicy(context.Background(), ChatCompletionRequest{
		Model:    "deepseek/deepseek-v4-flash",
		Messages: []ChatCompletionMessage{UserMessage("hello")},
	}, ChatCompletionFallbackPolicy{
		Models:     []string{"xiaomi/mimo-v2-flash"},
		ErrorCodes: []int{http.StatusBadRequest},
	})

	require.NoError(t, err)
	require.Equal(t, "xiaomi/mimo-v2-flash", resp.Model)
	require.Len(t, httpClient.requests, 2)
}

func TestCreateChatCompletionWithFallbackDoesNotRetryOnNonFallbackableError(t *testing.T) {
	t.Parallel()

	httpClient := &sequenceHTTPClient{
		responses: []*http.Response{
			jsonResponse(http.StatusBadRequest, `{"error":{"code":400,"message":"bad request"}}`),
		},
	}
	cfg := DefaultConfig("test-token")
	cfg.HTTPClient = httpClient
	cfg.BaseURL = "https://example.com/api/v1"
	client := NewClientWithConfig(*cfg)

	_, err := client.CreateChatCompletionWithFallback(context.Background(), ChatCompletionRequest{
		Model:    "deepseek/deepseek-v4-flash",
		Messages: []ChatCompletionMessage{UserMessage("hello")},
	}, "xiaomi/mimo-v2-flash")

	require.Error(t, err)
	require.True(t, IsHTTPStatus(err, http.StatusBadRequest))
	require.Len(t, httpClient.requests, 1)
}

func TestCreateChatCompletionStreamWithFallbackRetriesOnInitialError(t *testing.T) {
	t.Parallel()

	httpClient := &sequenceHTTPClient{
		responses: []*http.Response{
			jsonResponse(http.StatusPaymentRequired, `{"error":{"code":402,"message":"insufficient funds"}}`),
			jsonResponse(http.StatusOK, strings.Join([]string{
				`data: {"id":"chatcmpl_1","model":"xiaomi/mimo-v2-flash","choices":[{"delta":{"content":"ok"}}]}`,
				`data: [DONE]`,
				``,
			}, "\n")),
		},
	}
	cfg := DefaultConfig("test-token")
	cfg.HTTPClient = httpClient
	cfg.BaseURL = "https://example.com/api/v1"
	client := NewClientWithConfig(*cfg)

	stream, err := client.CreateChatCompletionStreamWithFallback(context.Background(), ChatCompletionRequest{
		Model:    "deepseek/deepseek-v4-flash",
		Messages: []ChatCompletionMessage{UserMessage("hello")},
	}, "xiaomi/mimo-v2-flash")

	require.NoError(t, err)
	chunk, err := stream.Recv()
	require.NoError(t, err)
	require.Equal(t, "xiaomi/mimo-v2-flash", chunk.Model)
	require.Len(t, httpClient.requests, 2)
	require.True(t, httpClient.requests[0].Stream)
	require.True(t, httpClient.requests[1].Stream)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
