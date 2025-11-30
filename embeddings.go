package openrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const embeddingsSuffix = "/embeddings"

// EmbeddingsEncodingFormat controls how embeddings are returned by the API.
// See: https://openrouter.ai/docs/api/api-reference/embeddings/create-embeddings
type EmbeddingsEncodingFormat string

const (
	EmbeddingsEncodingFormatFloat  EmbeddingsEncodingFormat = "float"
	EmbeddingsEncodingFormatBase64 EmbeddingsEncodingFormat = "base64"
)

// EmbeddingsRequest represents a request to the /embeddings endpoint.
//
// The input field is intentionally typed as any to support the flexible input
// types accepted by the OpenRouter API:
//   - string
//   - []string
//   - []float64
//   - [][]float64
//   - structured content blocks
//
// For examples, see: https://openrouter.ai/docs/api/api-reference/embeddings/create-embeddings
type EmbeddingsRequest struct {
	// Model is the model slug to use for embeddings.
	Model string `json:"model"`
	// Input is the content to embed. See the API docs for supported formats.
	Input any `json:"input"`

	// EncodingFormat controls how the embedding is returned: "float" or "base64".
	EncodingFormat EmbeddingsEncodingFormat `json:"encoding_format,omitempty"`
	// Dimensions optionally truncates the embedding to the given number of dimensions.
	Dimensions *int `json:"dimensions,omitempty"`
	// User is an optional identifier for the end-user making the request.
	User string `json:"user,omitempty"`
	// Provider configuration for provider routing. This reuses the same structure
	// as chat/completions provider routing, which is compatible with the embeddings API.
	Provider *ChatProvider `json:"provider,omitempty"`
	// InputType is an optional hint describing the type of input, e.g. "text" or "image".
	InputType string `json:"input_type,omitempty"`
}

// EmbeddingValue represents a single embedding, which can be returned either as
// a vector of floats or as a base64 string depending on encoding_format.
type EmbeddingValue struct {
	Vector []float64
	Base64 string
}

func (e *EmbeddingValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as []float64 first (encoding_format: "float").
	var vec []float64
	if err := json.Unmarshal(data, &vec); err == nil {
		e.Vector = vec
		e.Base64 = ""
		return nil
	}

	// Fallback to string (encoding_format: "base64").
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		e.Base64 = s
		e.Vector = nil
		return nil
	}

	return fmt.Errorf("embedding: invalid format, expected []float64 or string")
}

// EmbeddingData represents a single embedding entry in the response.
type EmbeddingData struct {
	Object    string         `json:"object"`
	Embedding EmbeddingValue `json:"embedding"`
	Index     int            `json:"index"`
}

// EmbeddingsUsage represents the token and cost statistics for an embeddings request.
type EmbeddingsUsage struct {
	PromptTokens int     `json:"prompt_tokens"`
	TotalTokens  int     `json:"total_tokens"`
	Cost         float64 `json:"cost"`
}

// EmbeddingsResponse represents the response from the /embeddings endpoint.
type EmbeddingsResponse struct {
	ID     string           `json:"id"`
	Object string           `json:"object"`
	Data   []EmbeddingData  `json:"data"`
	Model  string           `json:"model"`
	Usage  *EmbeddingsUsage `json:"usage,omitempty"`
}

// CreateEmbeddings submits an embedding request to the embeddings router.
//
// API reference: https://openrouter.ai/docs/api/api-reference/embeddings/create-embeddings
func (c *Client) CreateEmbeddings(
	ctx context.Context,
	request EmbeddingsRequest,
) (EmbeddingsResponse, error) {
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(embeddingsSuffix),
		withBody(request),
	)
	if err != nil {
		return EmbeddingsResponse{}, err
	}

	var response EmbeddingsResponse
	if err := c.sendRequest(req, &response); err != nil {
		return EmbeddingsResponse{}, err
	}

	return response, nil
}
