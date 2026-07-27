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
	"sync"
)

const createImagesSuffix = "/images"

type ImageAspectRatio string

const (
	ImageAspectRatio1x1    ImageAspectRatio = "1:1"
	ImageAspectRatio1x2    ImageAspectRatio = "1:2"
	ImageAspectRatio1x4    ImageAspectRatio = "1:4"
	ImageAspectRatio1x8    ImageAspectRatio = "1:8"
	ImageAspectRatio2x1    ImageAspectRatio = "2:1"
	ImageAspectRatio2x3    ImageAspectRatio = "2:3"
	ImageAspectRatio3x2    ImageAspectRatio = "3:2"
	ImageAspectRatio3x4    ImageAspectRatio = "3:4"
	ImageAspectRatio4x1    ImageAspectRatio = "4:1"
	ImageAspectRatio4x3    ImageAspectRatio = "4:3"
	ImageAspectRatio4x5    ImageAspectRatio = "4:5"
	ImageAspectRatio5x4    ImageAspectRatio = "5:4"
	ImageAspectRatio8x1    ImageAspectRatio = "8:1"
	ImageAspectRatio9x16   ImageAspectRatio = "9:16"
	ImageAspectRatio16x9   ImageAspectRatio = "16:9"
	ImageAspectRatio9x19_5 ImageAspectRatio = "9:19.5"
	ImageAspectRatio19_5x9 ImageAspectRatio = "19.5:9"
	ImageAspectRatio9x20   ImageAspectRatio = "9:20"
	ImageAspectRatio20x9   ImageAspectRatio = "20:9"
	ImageAspectRatio9x21   ImageAspectRatio = "9:21"
	ImageAspectRatio21x9   ImageAspectRatio = "21:9"
	ImageAspectRatioAuto   ImageAspectRatio = "auto"
)

type ImageBackground string

const (
	ImageBackgroundAuto        ImageBackground = "auto"
	ImageBackgroundTransparent ImageBackground = "transparent"
	ImageBackgroundOpaque      ImageBackground = "opaque"
)

type ImageOutputFormat string

const (
	ImageOutputFormatPng  ImageOutputFormat = "png"
	ImageOutputFormatJpeg ImageOutputFormat = "jpeg"
	ImageOutputFormatWebp ImageOutputFormat = "webp"
	ImageOutputFormatSvg  ImageOutputFormat = "svg"
)

type ImageQuality string

const (
	ImageQualityAuto   ImageQuality = "auto"
	ImageQualityLow    ImageQuality = "low"
	ImageQualityMedium ImageQuality = "medium"
	ImageQualityHigh   ImageQuality = "high"
)

type ImageResolution string

const (
	ImageResolution512 ImageResolution = "512"
	ImageResolution1K  ImageResolution = "1K"
	ImageResolution2K  ImageResolution = "2K"
	ImageResolution4K  ImageResolution = "4K"
)

type ImageInputReferenceType string

const (
	ImageInputReferenceTypeImageURL ImageInputReferenceType = "image_url"
)

type ImageInputReference struct {
	Type     ImageInputReferenceType `json:"type"`
	ImageURL ImageURLRef             `json:"image_url"`
}

type ImageURLRef struct {
	URL string `json:"url"`
}

type ImageGenerationProvider struct {
	Options map[string]map[string]any `json:"options,omitempty"`
}

type ImageGenerationRequest struct {
	Model             string                   `json:"model"`
	Prompt            string                   `json:"prompt"`
	AspectRatio       ImageAspectRatio         `json:"aspect_ratio,omitempty"`
	Background        ImageBackground          `json:"background,omitempty"`
	InputReferences   []ImageInputReference    `json:"input_references,omitempty"`
	N                 *int                     `json:"n,omitempty"`
	OutputCompression *int                     `json:"output_compression,omitempty"`
	OutputFormat      ImageOutputFormat        `json:"output_format,omitempty"`
	Provider          *ImageGenerationProvider `json:"provider,omitempty"`
	Quality           ImageQuality             `json:"quality,omitempty"`
	Resolution        ImageResolution          `json:"resolution,omitempty"`
	Seed              *int                     `json:"seed,omitempty"`
	Size              string                   `json:"size,omitempty"`
	Stream            *bool                    `json:"stream,omitempty"`
}

type ImageGenerationData struct {
	B64JSON   string `json:"b64_json"`
	MediaType string `json:"media_type,omitempty"`
}

type ImageGenerationUsageCompletionTokensDetails struct {
	AudioTokens     *int `json:"audio_tokens"`
	ImageTokens     *int `json:"image_tokens"`
	ReasoningTokens *int `json:"reasoning_tokens"`
}

type ImageGenerationUsagePromptTokensDetails struct {
	AudioTokens      *int `json:"audio_tokens"`
	CacheWriteTokens *int `json:"cache_write_tokens"`
	CachedTokens     *int `json:"cached_tokens"`
	FileTokens       *int `json:"file_tokens"`
	VideoTokens      *int `json:"video_tokens"`
}

type ImageGenerationUsageServerToolUse struct {
	ToolCallsExecuted  *int `json:"tool_calls_executed"`
	ToolCallsRequested *int `json:"tool_calls_requested"`
	WebSearchRequests  *int `json:"web_search_requests"`
}

type SpeedType string

const (
	SpeedFast     SpeedType = "fast"
	SpeedStandard SpeedType = "standard"
)

type ImageGenerationUsage struct {
	CompletionTokens        int                                          `json:"completion_tokens"`
	CompletionTokensDetails *ImageGenerationUsageCompletionTokensDetails `json:"completion_tokens_details,omitempty"`
	Cost                    *float64                                     `json:"cost,omitempty"`
	CostDetails             *CostDetails                                 `json:"cost_details,omitempty"`
	IsBYOK                  bool                                         `json:"is_byok"`
	PromptTokens            int                                          `json:"prompt_tokens"`
	PromptTokensDetails     *ImageGenerationUsagePromptTokensDetails     `json:"prompt_tokens_details,omitempty"`
	ServerToolUse           *ImageGenerationUsageServerToolUse           `json:"server_tool_use,omitempty"`
	ServiceTier             *string                                      `json:"service_tier,omitempty"`
	Speed                   SpeedType                                    `json:"speed,omitempty"`
	TotalTokens             int                                          `json:"total_tokens"`
}

type ImageGenerationResponse struct {
	Created int64                 `json:"created"`
	Data    []ImageGenerationData `json:"data"`
	Usage   *ImageGenerationUsage `json:"usage,omitempty"`
}

type ImageGenerationStreamChunkType string

const (
	ImageStreamChunkTypePartialImage ImageGenerationStreamChunkType = "image_generation.partial_image"
	ImageStreamChunkTypeCompleted    ImageGenerationStreamChunkType = "image_generation.completed"
)

type ImageGenerationStreamChunk struct {
	Type              ImageGenerationStreamChunkType `json:"type"`
	PartialImageIndex *int                           `json:"partial_image_index,omitempty"`
	B64JSON           string                         `json:"b64_json,omitempty"`
	Created           int64                          `json:"created,omitempty"`
	Usage             *ImageGenerationUsage          `json:"usage,omitempty"`
}

type ImageGenerationStream struct {
	once     sync.Once
	stream   <-chan ImageGenerationStreamChunk
	done     chan struct{}
	response *http.Response
}

// CreateImages generates an image from a text prompt via the image generation router.
// API reference: https://openrouter.ai/docs/api/api-reference/images/create-images
func (c *Client) CreateImages(
	ctx context.Context,
	request ImageGenerationRequest,
) (response ImageGenerationResponse, err error) {
	if request.Stream != nil && *request.Stream {
		err = errors.New("streaming is not supported with this method, please use CreateImagesStream")
		return
	}

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(createImagesSuffix),
		withBody(request),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// CreateImagesStream generates an image with streaming support.
// CreateImagesStream generates an image with streaming support.
func (c *Client) CreateImagesStream(
	ctx context.Context,
	request ImageGenerationRequest,
) (*ImageGenerationStream, error) {
	if request.Stream == nil {
		b := true
		request.Stream = &b
	} else if !*request.Stream {
		b := true
		request.Stream = &b
	}

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(createImagesSuffix),
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

	stream := make(chan ImageGenerationStreamChunk)
	done := make(chan struct{})

	go func() {
		defer close(stream)
		defer resp.Body.Close()

		// If the model does not support streaming we get a regular application/json response
		contentType := strings.ToLower(resp.Header.Get("Content-Type"))
		if strings.Contains(contentType, "application/json") {
			var fallback ImageGenerationResponse
			if err := json.NewDecoder(resp.Body).Decode(&fallback); err != nil {
				slog.Error("failed to decode non-streaming image response", "error", err)
				return
			}

			for i, img := range fallback.Data {
				idx := i
				chunk := ImageGenerationStreamChunk{
					Type:              ImageStreamChunkTypeCompleted,
					PartialImageIndex: &idx,
					B64JSON:           img.B64JSON,
					Created:           fallback.Created,
					Usage:             fallback.Usage,
				}

				select {
				case <-done:
					return
				case <-ctx.Done():
					return
				case stream <- chunk:
				}
			}

			return
		}

		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				slog.Info("Image stream stopped due to context cancellation")
				return
			default:
				line, err := reader.ReadBytes('\n')
				if err != nil {
					if err == io.EOF {
						return
					}
					slog.Error("failed to read image stream", "error", err)
					return
				}

				if strings.HasSuffix(string(line), "[DONE]\n") {
					return
				}

				trimmed := strings.TrimSpace(string(line))
				if trimmed == "" || strings.HasPrefix(trimmed, ":") {
					continue
				}

				line = bytes.TrimPrefix(line, []byte("data:"))

				var chunk ImageGenerationStreamChunk
				if err := json.Unmarshal(line, &chunk); err != nil {
					slog.Error("failed to decode image stream", "error", err, "line", string(line))
					return
				}

				select {
				case <-done:
					return
				case <-ctx.Done():
					return
				case stream <- chunk:
				}
			}
		}
	}()

	return &ImageGenerationStream{
		stream:   stream,
		done:     done,
		response: resp,
	}, nil
}

// Recv reads the next chunk from the stream.
func (s *ImageGenerationStream) Recv() (ImageGenerationStreamChunk, error) {
	select {
	case chunk, ok := <-s.stream:
		if !ok {
			return ImageGenerationStreamChunk{}, io.EOF
		}
		return chunk, nil
	case <-s.done:
		return ImageGenerationStreamChunk{}, io.EOF
	}
}

// Close terminates the stream and cleans up resources.
func (s *ImageGenerationStream) Close() {
	s.once.Do(func() {
		close(s.done)
		if s.response != nil {
			s.response.Body.Close()
		}
	})
}
