package openrouter

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeHTTPClient struct {
	lastRequest *http.Request
	response    *http.Response
	err         error
}

func (f *fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	f.lastRequest = req
	if f.err != nil {
		return nil, f.err
	}
	return f.response, nil
}

func TestCreateEmbeddings_Basic(t *testing.T) {
	body := `{
		"id": "embd_123",
		"object": "list",
		"data": [
			{
				"object": "embedding",
				"embedding": [0.1, 0.2, 0.3],
				"index": 0
			}
		],
		"model": "test-embeddings-model",
		"usage": {
			"prompt_tokens": 5,
			"total_tokens": 5,
			"cost": 0.0001
		}
	}`

	fakeClient := &fakeHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		},
	}

	cfg := DefaultConfig("test-token")
	cfg.BaseURL = "https://example.com/api/v1"
	cfg.HTTPClient = fakeClient

	client := NewClientWithConfig(*cfg)

	req := EmbeddingsRequest{
		Model: "test-embeddings-model",
		Input: "hello world",
	}

	resp, err := client.CreateEmbeddings(context.Background(), req)
	require.NoError(t, err)

	require.NotNil(t, fakeClient.lastRequest)
	require.Equal(t, http.MethodPost, fakeClient.lastRequest.Method)
	require.True(t, strings.HasSuffix(fakeClient.lastRequest.URL.Path, "/embeddings"))

	require.Equal(t, "embd_123", resp.ID)
	require.Equal(t, "list", resp.Object)
	require.Equal(t, "test-embeddings-model", resp.Model)
	require.NotNil(t, resp.Usage)
	require.Equal(t, 5, resp.Usage.PromptTokens)
	require.Len(t, resp.Data, 1)
	require.Len(t, resp.Data[0].Embedding.Vector, 3)
}

func TestEmbeddingValue_UnmarshalJSON_Base64(t *testing.T) {
	var v EmbeddingValue

	err := v.UnmarshalJSON([]byte(`"dGVzdC1lbWJlZGRpbmc="`))
	require.NoError(t, err)
	require.Nil(t, v.Vector)
	require.Equal(t, "dGVzdC1lbWJlZGRpbmc=", v.Base64)
}


