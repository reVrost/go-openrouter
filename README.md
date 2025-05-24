# Go Openrouter

[![Go Reference](https://pkg.go.dev/badge/github.com/revrost/go-openrouter.svg)](https://pkg.go.dev/github.com/revrost/go-openrouter)
[![Go Report Card](https://goreportcard.com/badge/github.com/revrost/go-openrouter)](https://goreportcard.com/report/github.com/revrost/go-openrouter)
[![codecov](https://codecov.io/gh/revrost/go-openrouter/branch/master/graph/badge.svg?token=bCbIfHLIsW)](https://codecov.io/gh/revrost/go-openrouter)

This library provides unofficial Go client for [Openrouter API](https://openrouter.ai/docs/quick-start)

## Installation

```
go get github.com/revrost/go-openrouter
```

## Usage

### Deepseek V3 example usage:

```go
package main

import (
	"context"
	"fmt"
	openrouter "github.com/revrost/go-openrouter"
)

func main() {
	client := openrouter.NewClient(
		"your token",
		openrouter.WithXTitle("My App"),
		openrouter.WithHTTPReferer("https://myapp.com"),
	)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openrouter.ChatCompletionRequest{
			Model: openrouter.DeepseekV3,
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    openrouter.ChatMessageRoleUser,
					Content: openrouter.Content{Text: "Hello!"},
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

```

### Getting an Openrouter API Key:

1. Visit the openrouter website at [https://openrouter.ai/docs/quick-start](https://openrouter.ai/docs/quick-start).
2. If you don't have an account, click on "Sign Up" to create one. If you do, click "Log In".
3. Once logged in, navigate to your API key management page.
4. Click on "Create new secret key".
5. Enter a name for your new key, then click "Create secret key".
6. Your new API key will be displayed. Use this key to interact with the openrouter API.

**Note:** Your API key is sensitive information. Do not share it with anyone.

For deepseek models, sometimes its better to use openrouter integration feature and pass in your own API key into the control panel for better performance, as openrouter will use your API key to make requests to the underlying model which potentially avoids shared rate limits.

### Other examples:

<details>
<summary>Streaming Response</summary>

```go
func main() {
	ctx := context.Background()
	client := openrouter.NewClient(os.Getenv("OPENROUTER_API_KEY"))

	stream, err := client.CreateChatCompletionStream(
		context.Background(), openrouter.ChatCompletionRequest{
			Model: "qwen/qwq-32b:free",
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    "user",
					Content: openrouter.Content{Text: "Hello, how are you?"}},
			},
			Stream: true,
		},
	)
	require.NoError(t, err)
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil && err != io.EOF {
			require.NoError(t, err)
		}
		if errors.Is(err, io.EOF) {
			fmt.Println("EOF, stream finished")
			return
		}
		json, err := json.MarshalIndent(response, "", "  ")
		require.NoError(t, err)
		fmt.Println(string(json))
	}
}
```

</details>

<details>
<summary>JSON Schema for function calling</summary>

```json
{
  "name": "get_current_weather",
  "description": "Get the current weather in a given location",
  "parameters": {
    "type": "object",
    "properties": {
      "location": {
        "type": "string",
        "description": "The city and state, e.g. San Francisco, CA"
      },
      "unit": {
        "type": "string",
        "enum": ["celsius", "fahrenheit"]
      }
    },
    "required": ["location"]
  }
}
```

Using the `jsonschema` package, this schema could be created using structs as such:

```go
FunctionDefinition{
  Name: "get_current_weather",
  Parameters: jsonschema.Definition{
    Type: jsonschema.Object,
    Properties: map[string]jsonschema.Definition{
      "location": {
        Type: jsonschema.String,
        Description: "The city and state, e.g. San Francisco, CA",
      },
      "unit": {
        Type: jsonschema.String,
        Enum: []string{"celsius", "fahrenheit"},
      },
    },
    Required: []string{"location"},
  },
}
```

The `Parameters` field of a `FunctionDefinition` can accept either of the above styles, or even a nested struct from another library (as long as it can be marshalled into JSON).

</details>

<details>
<summary>Structured Outputs</summary>

```go
func main() {
	ctx := context.Background()
	client := openrouter.NewClient(os.Getenv("OPENROUTER_API_KEY"))

	type Result struct {
		Location    string  `json:"location"`
		Temperature float64 `json:"temperature"`
		Condition   string  `json:"condition"`
	}
	var result Result
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		log.Fatalf("GenerateSchemaForType error: %v", err)
	}

	request := openrouter.ChatCompletionRequest{
		Model: openrouter.DeepseekV3,
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role:    openrouter.ChatMessageRoleUser,
				Content: openrouter.Content{Text: "What's the weather like in London?"},
			},
		},
		ResponseFormat: &openrouter.ChatCompletionResponseFormat{
			Type: openrouter.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openrouter.ChatCompletionResponseFormatJSONSchema{
				Name:   "weather",
				Schema: schema,
				Strict: true,
			},
		},
	}

	pj, _ := json.MarshalIndent(request, "", "\t")
	fmt.Printf("request :\n %s\n", string(pj))

	res, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		fmt.Println("error", err)
	} else {
		b, _ := json.MarshalIndent(res, "", "\t")
		fmt.Printf("response :\n %s", string(b))
	}
}
```

</details>
More examples in `examples/` folder.

## Frequently Asked Questions

## Contributing

[Contributing Guidelines](https://github.com/revrost/go-openrouter/blob/master/CONTRIBUTING.md), we hope to see your contributions!
