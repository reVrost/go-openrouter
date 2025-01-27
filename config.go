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

func DefaultConfig(authToken, xTitle, httpReferer string) ClientConfig {
	return ClientConfig{
		authToken:        authToken,
		XTitle:           "",
		HttpReferer:      "",
		BaseURL:          "",
		AssistantVersion: "",
		OrgID:            "",

		HTTPClient: &http.Client{},

		EmptyMessagesLimit: defaultEmptyMessagesLimit,
	}
}
