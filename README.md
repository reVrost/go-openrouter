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
					Content: "Hello!",
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

### Other examples:

<details>
<summary>ChatGPT streaming completion</summary>

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	openrouter "github.com/revrost/go-openrouter"
)

func main() {
	c := openrouter.NewClient("your token")
	ctx := context.Background()

	req := openrouter.ChatCompletionRequest{
		Model:     openrouter.GPT3Dot5Turbo,
		MaxTokens: 20,
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role:    openrouter.ChatMessageRoleUser,
				Content: "Lorem ipsum",
			},
		},
		Stream: true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}

		fmt.Printf(response.Choices[0].Delta.Content)
	}
}
```

</details>

<details>
<summary>GPT-3 completion</summary>

```go
package main

import (
	"context"
	"fmt"
	openrouter "github.com/revrost/go-openrouter"
)

func main() {
	c := openrouter.NewClient("your token")
	ctx := context.Background()

	req := openrouter.CompletionRequest{
		Model:     openrouter.GPT3Babbage002,
		MaxTokens: 5,
		Prompt:    "Lorem ipsum",
	}
	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return
	}
	fmt.Println(resp.Choices[0].Text)
}
```

</details>

<details>
<summary>GPT-3 streaming completion</summary>

```go
package main

import (
	"errors"
	"context"
	"fmt"
	"io"
	openrouter "github.com/revrost/go-openrouter"
)

func main() {
	c := openrouter.NewClient("your token")
	ctx := context.Background()

	req := openrouter.CompletionRequest{
		Model:     openrouter.GPT3Babbage002,
		MaxTokens: 5,
		Prompt:    "Lorem ipsum",
		Stream:    true,
	}
	stream, err := c.CreateCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("CompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream finished")
			return
		}

		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return
		}


		fmt.Printf("Stream response: %v\n", response)
	}
}
```

</details>

<details>
<summary>Audio Speech-To-Text</summary>

```go
package main

import (
	"context"
	"fmt"

	openrouter "github.com/revrost/go-openrouter"
)

func main() {
	c := openrouter.NewClient("your token")
	ctx := context.Background()

	req := openrouter.AudioRequest{
		Model:    openrouter.Whisper1,
		FilePath: "recording.mp3",
	}
	resp, err := c.CreateTranscription(ctx, req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}
	fmt.Println(resp.Text)
}
```

</details>

<details>
<summary>Audio Captions</summary>

```go
package main

import (
	"context"
	"fmt"
	"os"

	openrouter "github.com/revrost/go-openrouter"
)

func main() {
	c := openrouter.NewClient(os.Getenv("openrouter_KEY"))

	req := openrouter.AudioRequest{
		Model:    openrouter.Whisper1,
		FilePath: os.Args[1],
		Format:   openrouter.AudioResponseFormatSRT,
	}
	resp, err := c.CreateTranscription(context.Background(), req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}
	f, err := os.Create(os.Args[1] + ".srt")
	if err != nil {
		fmt.Printf("Could not open file: %v\n", err)
		return
	}
	defer f.Close()
	if _, err := f.WriteString(resp.Text); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
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
<summary>Error handling</summary>

Open-AI maintains clear documentation on how to [handle API errors](https://platform.openrouter.com/docs/guides/error-codes/api-errors)

example:

```
e := &openrouter.APIError{}
if errors.As(err, &e) {
  switch e.HTTPStatusCode {
    case 401:
      // invalid auth or key (do not retry)
    case 429:
      // rate limiting or engine overload (wait and retry)
    case 500:
      // openrouter server error (retry)
    default:
      // unhandled
  }
}

```

</details>

<details>
<summary>Structured Outputs</summary>

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/revrost/go-openrouter"
	"github.com/revrost/go-openrouter/jsonschema"
)

func main() {
	client := openrouter.NewClient("your token")
	ctx := context.Background()

	type Result struct {
		Steps []struct {
			Explanation string `json:"explanation"`
			Output      string `json:"output"`
		} `json:"steps"`
		FinalAnswer string `json:"final_answer"`
	}
	var result Result
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		log.Fatalf("GenerateSchemaForType error: %v", err)
	}
	resp, err := client.CreateChatCompletion(ctx, openrouter.ChatCompletionRequest{
		Model: openrouter.GPT4oMini,
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role:    openrouter.ChatMessageRoleSystem,
				Content: "You are a helpful math tutor. Guide the user through the solution step by step.",
			},
			{
				Role:    openrouter.ChatMessageRoleUser,
				Content: "how can I solve 8x + 7 = -23",
			},
		},
		ResponseFormat: &openrouter.ChatCompletionResponseFormat{
			Type: openrouter.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openrouter.ChatCompletionResponseFormatJSONSchema{
				Name:   "math_reasoning",
				Schema: schema,
				Strict: true,
			},
		},
	})
	if err != nil {
		log.Fatalf("CreateChatCompletion error: %v", err)
	}
	err = schema.Unmarshal(resp.Choices[0].Message.Content, &result)
	if err != nil {
		log.Fatalf("Unmarshal schema error: %v", err)
	}
	fmt.Println(result)
}
```

</details>
More examples in `examples/` folder.

## Frequently Asked Questions

## Contributing

[Contributing Guidelines](https://github.com/revrost/go-openrouter/blob/master/CONTRIBUTING.md), we hope to see your contributions!
