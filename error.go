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
type Metadata map[string]any

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
		return fmt.Sprintf("error, code: %v, message: %s, metadata: %v", e.Code, e.Message, e.Metadata)
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
