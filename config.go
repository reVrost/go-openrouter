package openrouter

import "net/http"

// ClientConfig is a configuration for the openrouter client.
type ClientConfig struct {
	authToken string

	BaseURL          string
	OrgID            string
	AssistantVersion string
	HTTPClient       HTTPDoer
	HttpReferer      string
	XTitle           string

	EmptyMessagesLimit uint
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

const defaultEmptyMessagesLimit = 10

func DefaultConfig(authToken string) *ClientConfig {
	return &ClientConfig{
		authToken:        authToken,
		XTitle:           "",
		HttpReferer:      "",
		BaseURL:          "https://openrouter.ai/api/v1",
		AssistantVersion: "",
		OrgID:            "",

		HTTPClient: &http.Client{},

		EmptyMessagesLimit: defaultEmptyMessagesLimit,
	}
}

type Option func(*ClientConfig)

func WithXTitle(title string) Option {
	return func(c *ClientConfig) {
		c.XTitle = title
	}
}

func WithHTTPReferer(referer string) Option {
	return func(c *ClientConfig) {
		c.HttpReferer = referer
	}
}
