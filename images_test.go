package openrouter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newImageTestClient(response *http.Response) (*Client, *fakeHTTPClient) {
	httpClient := &fakeHTTPClient{response: response}
	cfg := DefaultConfig("test-token")
	cfg.BaseURL = "https://example.com/api/v1"
	cfg.HTTPClient = httpClient
	return NewClientWithConfig(*cfg), httpClient
}

type closeTrackingBody struct {
	io.Reader
	closed chan struct{}
	once   sync.Once
}

func (b *closeTrackingBody) Close() error {
	b.once.Do(func() {
		close(b.closed)
	})
	return nil
}

func TestCreateImages(t *testing.T) {
	client, httpClient := newImageTestClient(&http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body: io.NopCloser(strings.NewReader(`{
			"created": 1748372400,
			"data": [{"b64_json": "aW1hZ2U=", "media_type": "image/png"}],
			"usage": {"prompt_tokens": 16, "completion_tokens": 272, "total_tokens": 288, "cost": 0.011}
		}`)),
	})

	allowFallbacks := false
	response, err := client.CreateImages(context.Background(), ImageGenerationRequest{
		Model:        "openai/gpt-image-1",
		Prompt:       "a red panda astronaut",
		OutputFormat: ImageOutputFormatPng,
		Provider: &ImageGenerationProvider{
			Only:           []string{"openai"},
			Order:          []string{"openai"},
			Ignore:         []string{"example"},
			Sort:           "price",
			AllowFallbacks: &allowFallbacks,
			Options: map[string]map[string]any{
				"openai": {"quality": "high"},
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, int64(1748372400), response.Created)
	require.Len(t, response.Data, 1)
	require.Equal(t, "aW1hZ2U=", response.Data[0].B64JSON)
	require.Equal(t, "image/png", response.Data[0].MediaType)
	require.NotNil(t, response.Usage)
	require.InDelta(t, 0.011, *response.Usage.Cost, 1e-12)

	require.Equal(t, http.MethodPost, httpClient.lastRequest.Method)
	require.Equal(t, "/api/v1/images", httpClient.lastRequest.URL.Path)
	require.Equal(t, "Bearer test-token", httpClient.lastRequest.Header.Get("Authorization"))

	var body map[string]any
	require.NoError(t, json.NewDecoder(httpClient.lastRequest.Body).Decode(&body))
	require.Equal(t, "openai/gpt-image-1", body["model"])
	require.Equal(t, "a red panda astronaut", body["prompt"])
	require.Equal(t, "png", body["output_format"])
	require.NotContains(t, body, "stream")

	provider := body["provider"].(map[string]any)
	require.Equal(t, []any{"openai"}, provider["only"])
	require.Equal(t, []any{"openai"}, provider["order"])
	require.Equal(t, []any{"example"}, provider["ignore"])
	require.Equal(t, "price", provider["sort"])
	require.Equal(t, false, provider["allow_fallbacks"])
	require.Contains(t, provider, "options")
}

func TestCreateImagesRejectsStreaming(t *testing.T) {
	client, _ := newImageTestClient(nil)
	stream := true

	_, err := client.CreateImages(context.Background(), ImageGenerationRequest{Stream: &stream})
	require.EqualError(t, err, "streaming is not supported with this method, please use CreateImagesStream")
}

func TestCreateImagesStream(t *testing.T) {
	body := strings.Join([]string{
		": OPENROUTER PROCESSING",
		"",
		`data: {"type":"image_generation.partial_image","partial_image_index":0,"b64_json":"cGFydGlhbA=="}`,
		`data: {"type":"image_generation.completed","b64_json":"ZmluYWw=","media_type":"image/svg+xml","created":1748372400,"usage":{"total_tokens":288}}`,
		"data: [DONE]",
		"",
	}, "\r\n")
	client, httpClient := newImageTestClient(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	})

	stream, err := client.CreateImagesStream(context.Background(), ImageGenerationRequest{
		Model:  "openai/gpt-image-1",
		Prompt: "a landscape",
	})
	require.NoError(t, err)
	t.Cleanup(stream.Close)

	partial, err := stream.Recv()
	require.NoError(t, err)
	require.Equal(t, ImageStreamChunkTypePartialImage, partial.Type)
	require.Equal(t, 0, *partial.PartialImageIndex)
	require.Equal(t, "cGFydGlhbA==", partial.B64JSON)

	completed, err := stream.Recv()
	require.NoError(t, err)
	require.Equal(t, ImageStreamChunkTypeCompleted, completed.Type)
	require.Equal(t, "ZmluYWw=", completed.B64JSON)
	require.Equal(t, "image/svg+xml", completed.MediaType)
	require.Equal(t, int64(1748372400), completed.Created)
	require.NotNil(t, completed.Usage)
	require.Equal(t, 288, completed.Usage.TotalTokens)

	_, err = stream.Recv()
	require.ErrorIs(t, err, io.EOF)

	require.Equal(t, "text/event-stream", httpClient.lastRequest.Header.Get("Accept"))
	var request ImageGenerationRequest
	require.NoError(t, json.NewDecoder(httpClient.lastRequest.Body).Decode(&request))
	require.NotNil(t, request.Stream)
	require.True(t, *request.Stream)
}

func TestCreateImagesStreamReturnsAPIError(t *testing.T) {
	client, _ := newImageTestClient(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(
			"data: {\"type\":\"error\",\"error\":{\"message\":\"Generation failed\",\"code\":\"server_error\"}}\n\n",
		)),
	})

	stream, err := client.CreateImagesStream(context.Background(), ImageGenerationRequest{})
	require.NoError(t, err)
	t.Cleanup(stream.Close)

	_, err = stream.Recv()
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, "Generation failed", apiErr.Message)
	require.Equal(t, "server_error", apiErr.Code)
}

func TestCreateImagesStreamHTTPError(t *testing.T) {
	responseBody := &closeTrackingBody{
		Reader: strings.NewReader(`{"error":{"message":"Rate limit exceeded","code":429}}`),
		closed: make(chan struct{}),
	}
	client, _ := newImageTestClient(&http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     make(http.Header),
		Body:       responseBody,
	})

	stream, err := client.CreateImagesStream(context.Background(), ImageGenerationRequest{})
	require.Nil(t, stream)
	require.True(t, IsErrorCode(err, http.StatusTooManyRequests))
	select {
	case <-responseBody.closed:
	default:
		t.Fatal("error response body was not closed")
	}
}

func TestCreateImagesStreamFallsBackToJSON(t *testing.T) {
	client, _ := newImageTestClient(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
		Body: io.NopCloser(strings.NewReader(`{
			"created": 1748372400,
			"data": [
				{"b64_json": "Zmlyc3Q=", "media_type": "image/png"},
				{"b64_json": "c2Vjb25k", "media_type": "image/webp"}
			],
			"usage": {"total_tokens": 288}
		}`)),
	})

	stream, err := client.CreateImagesStream(context.Background(), ImageGenerationRequest{})
	require.NoError(t, err)

	first, err := stream.Recv()
	require.NoError(t, err)
	require.Equal(t, ImageStreamChunkTypeCompleted, first.Type)
	require.Equal(t, 0, *first.PartialImageIndex)
	require.Equal(t, "image/png", first.MediaType)

	second, err := stream.Recv()
	require.NoError(t, err)
	require.Equal(t, 1, *second.PartialImageIndex)
	require.Equal(t, "image/webp", second.MediaType)
	require.Same(t, first.Usage, second.Usage)

	_, err = stream.Recv()
	require.ErrorIs(t, err, io.EOF)

	require.NotPanics(t, stream.Close)
	require.NotPanics(t, stream.Close)
}

func TestCreateImagesStreamCloseIsIdempotent(t *testing.T) {
	client, _ := newImageTestClient(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader("data: [DONE]\n\n")),
	})

	stream, err := client.CreateImagesStream(context.Background(), ImageGenerationRequest{})
	require.NoError(t, err)
	require.NotPanics(t, stream.Close)
	require.NotPanics(t, stream.Close)

	_, err = stream.Recv()
	require.ErrorIs(t, err, io.EOF)
}

func TestCreateImagesStreamCancellationUnblocksSend(t *testing.T) {
	body := strings.Join([]string{
		`data: {"type":"image_generation.partial_image","partial_image_index":0,"b64_json":"Zmlyc3Q="}`,
		`data: {"type":"image_generation.partial_image","partial_image_index":1,"b64_json":"c2Vjb25k"}`,
		"data: [DONE]",
		"",
	}, "\n")
	responseBody := &closeTrackingBody{
		Reader: strings.NewReader(body),
		closed: make(chan struct{}),
	}
	client, _ := newImageTestClient(&http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       responseBody,
	})
	ctx, cancel := context.WithCancel(context.Background())

	stream, err := client.CreateImagesStream(ctx, ImageGenerationRequest{})
	require.NoError(t, err)
	t.Cleanup(stream.Close)

	_, err = stream.Recv()
	require.NoError(t, err)
	cancel()

	select {
	case <-responseBody.closed:
	case <-time.After(time.Second):
		t.Fatal("stream did not stop after context cancellation")
	}
}

func TestListImageModels(t *testing.T) {
	client, httpClient := newImageTestClient(&http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body: io.NopCloser(strings.NewReader(`{
			"data": [{
				"id": "bytedance-seed/seedream-4.5",
				"name": "Seedream 4.5",
				"description": "A text-to-image model.",
				"created": 1692901234,
				"architecture": {
					"input_modalities": ["text", "image"],
					"output_modalities": ["image"]
				},
				"supported_parameters": {
					"resolution": {"type": "enum", "values": ["1K", "2K", "4K"]}
				},
				"supports_streaming": false,
				"endpoints": "/api/v1/images/models/bytedance-seed/seedream-4.5/endpoints"
			}]
		}`)),
	})

	models, err := client.ListImageModels(context.Background())
	require.NoError(t, err)
	require.Len(t, models, 1)
	require.Equal(t, "bytedance-seed/seedream-4.5", models[0].ID)
	require.Equal(t, []string{"1K", "2K", "4K"}, models[0].SupportedParameters["resolution"].Values)
	require.Equal(t, http.MethodGet, httpClient.lastRequest.Method)
	require.Equal(t, "/api/v1/images/models", httpClient.lastRequest.URL.Path)
}
