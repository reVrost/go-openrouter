package openrouter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// APIError provides error information returned by the Openrouter API.
type APIError struct {
	Code     any       `json:"code,omitempty"`
	Message  string    `json:"message"`
	Metadata *Metadata `json:"metadata,omitempty"`

	// Internal fields
	HTTPStatusCode int `json:"-"`
}

// Metadata provides additional information about the error.
type Metadata struct {
	// Common fields
	ProviderName string `json:"provider_name,omitempty"`

	// Provider-specific fields
	Raw json.RawMessage `json:"raw,omitempty"` // Raw error from provider

	// Moderation-specific fields
	Reasons      []string `json:"reasons,omitempty"`       // Why input was flagged
	FlaggedInput string   `json:"flagged_input,omitempty"` // Truncated flagged text
	ModelSlug    string   `json:"model_slug,omitempty"`    // Model that flagged input
}

// ProviderError returns provider-specific error details
func (m *Metadata) ProviderError() (string, json.RawMessage) {
	return m.ProviderName, m.Raw
}

// ModerationError returns moderation-specific error details
func (m *Metadata) ModerationError() (string, []string, string, string) {
	return m.ProviderName, m.Reasons, m.FlaggedInput, m.ModelSlug
}

// IsProviderError checks if this is a provider error
func (m *Metadata) IsProviderError() bool {
	return m.Raw != nil
}

// IsModerationError checks if this is a moderation error
func (m *Metadata) IsModerationError() bool {
	return len(m.Reasons) > 0
}

// RequestError provides information about generic request errors.
type RequestError struct {
	HTTPStatus     string
	HTTPStatusCode int
	Err            error
	Body           []byte
}

type ErrorResponse struct {
	Error *APIError `json:"error,omitempty"`
}

func (e *APIError) Error() string {
	// If it has metadata
	if e.Metadata != nil {
		return fmt.Sprintf("error, code: %v, message: %s, reasons: %v, flagged_input: %s, provider_name: %s, model_slug: %s",
			e.Code, e.Message, e.Metadata.Reasons, e.Metadata.FlaggedInput, e.Metadata.ProviderName, e.Metadata.ModelSlug)
	}
	return e.Message
}

func (e *APIError) UnmarshalJSON(data []byte) (err error) {
	var rawMap map[string]json.RawMessage
	err = json.Unmarshal(data, &rawMap)
	if err != nil {
		return
	}

	err = json.Unmarshal(rawMap["message"], &e.Message)
	if err != nil {
		var messages []string
		err = json.Unmarshal(rawMap["message"], &messages)
		if err != nil {
			return
		}
		e.Message = strings.Join(messages, ", ")
	}

	if _, ok := rawMap["code"]; !ok {
		return nil
	}

	// if the api returned a number, we need to force an integer
	// since the json package defaults to float64
	var intCode int
	err = json.Unmarshal(rawMap["code"], &intCode)
	if err == nil {
		e.Code = intCode
		return nil
	}

	return json.Unmarshal(rawMap["code"], &e.Code)
}

func (e *RequestError) Error() string {
	return fmt.Sprintf(
		"error, status code: %d, status: %s, message: %s, body: %s",
		e.HTTPStatusCode, e.HTTPStatus, e.Err, e.Body,
	)
}

func (e *RequestError) Unwrap() error {
	return e.Err
}
