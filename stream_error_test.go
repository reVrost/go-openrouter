package openrouter

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func newRateLimitedResponse() (*http.Response, *closeTrackingBody) {
	body := &closeTrackingBody{
		Reader: strings.NewReader(`{"error":{"message":"Rate limit exceeded","code":429}}`),
		closed: make(chan struct{}),
	}
	return &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     make(http.Header),
		Body:       body,
	}, body
}

func requireClosed(t *testing.T, body *closeTrackingBody) {
	t.Helper()
	select {
	case <-body.closed:
	default:
		t.Fatal("error response body was not closed")
	}
}

func TestCreateChatCompletionStreamHTTPError(t *testing.T) {
	resp, body := newRateLimitedResponse()
	client, _ := newImageTestClient(resp)

	stream, err := client.CreateChatCompletionStream(context.Background(), ChatCompletionRequest{})
	require.Nil(t, stream)
	require.True(t, IsErrorCode(err, http.StatusTooManyRequests))
	requireClosed(t, body)
}

func TestCreateCompletionStreamHTTPError(t *testing.T) {
	resp, body := newRateLimitedResponse()
	client, _ := newImageTestClient(resp)

	stream, err := client.CreateCompletionStream(context.Background(), CompletionRequest{})
	require.Nil(t, stream)
	require.True(t, IsErrorCode(err, http.StatusTooManyRequests))
	requireClosed(t, body)
}
