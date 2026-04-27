package openrouter

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// APIError provides error information returned by the Openrouter API.
type APIError struct {
	Code     any       `json:"code,omitempty"`
	Message  string    `json:"message"`
	Metadata *Metadata `json:"metadata,omitempty"`

	// Internal fields
	HTTPStatusCode int            `json:"-"`
	ProviderError  *ProviderError `json:"-"`
}

// Metadata provides additional information about the error.
type Metadata map[string]any

// ProviderError provides the provider error (if available).
type ProviderError map[string]any

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

// HTTPStatusCode returns the HTTP status code carried by OpenRouter API errors.
func HTTPStatusCode(err error) (int, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.HTTPStatusCode, apiErr.HTTPStatusCode != 0
	}

	var reqErr *RequestError
	if errors.As(err, &reqErr) {
		return reqErr.HTTPStatusCode, reqErr.HTTPStatusCode != 0
	}

	return 0, false
}

// APIErrorCode returns the OpenRouter API error code carried by API errors.
func APIErrorCode(err error) (any, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code, apiErr.Code != nil
	}

	return nil, false
}

// IsHTTPStatus reports whether err carries the provided HTTP status code.
func IsHTTPStatus(err error, statusCode int) bool {
	code, ok := HTTPStatusCode(err)
	return ok && code == statusCode
}

// IsAPIErrorCode reports whether err carries the provided OpenRouter API error code.
func IsAPIErrorCode(err error, code any) bool {
	apiCode, ok := APIErrorCode(err)
	return ok && fmt.Sprint(apiCode) == fmt.Sprint(code)
}

// IsErrorCode reports whether err carries code as either its HTTP status code
// or OpenRouter API error code.
func IsErrorCode(err error, code int) bool {
	return IsHTTPStatus(err, code) || IsAPIErrorCode(err, code)
}

func (e *ProviderError) Message() any {
	// {"message": "string"}
	messageAny, ok := (*e)["message"]
	if ok {
		return messageAny
	}

	// {"error": {"message": "string"}}
	errAny, ok := (*e)["error"]
	if !ok {
		return nil
	}

	err, ok := errAny.(map[string]any)
	if !ok {
		return errAny
	}

	messageAny, ok = err["message"]
	if ok {
		return messageAny
	}

	return err
}

func (e *APIError) Error() string {
	// If it has a provider error
	if e.ProviderError != nil {
		if message := e.ProviderError.Message(); message != nil {
			return fmt.Sprintf("provider error, code: %v, message: %v", e.Code, message)
		}

		return fmt.Sprintf("provider error, code: %v, message: %s, error: %v", e.Code, e.Message, *e.ProviderError)
	}

	// If it has metadata
	if e.Metadata != nil {
		return fmt.Sprintf("error, code: %v, message: %s, metadata: %v", e.Code, e.Message, *e.Metadata)
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

	if meta, ok := rawMap["metadata"]; ok {
		err = json.Unmarshal(meta, &e.Metadata)
		if err != nil {
			return
		}
	}

	if e.Metadata != nil {
		raw, ok := (*e.Metadata)["raw"].(string)
		if ok {
			err = json.Unmarshal([]byte(raw), &e.ProviderError)
			if err != nil {
				return
			}
		}
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
