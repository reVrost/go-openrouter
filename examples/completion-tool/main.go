package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/revrost/go-openrouter"
	"github.com/revrost/go-openrouter/jsonschema"
)

func main() {
	ctx := context.Background()
	client := openrouter.NewClient(os.Getenv("OPENROUTER_API_KEY"))
	var provider *openrouter.ChatProvider

	// describe the function & its inputs
	params := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"location": {
				Type:        jsonschema.String,
				Description: "The city and state, e.g. San Francisco, CA",
			},
			"unit": {
				Type: jsonschema.String,
				Enum: []string{"celsius", "fahrenheit"},
			},
		},
		Required: []string{"location"},
	}
	f := openrouter.FunctionDefinition{
		Name:        "get_current_weather",
		Description: "Get the current weather in a given location",
		Parameters:  params,
	}
	t := openrouter.Tool{
		Type:     openrouter.ToolTypeFunction,
		Function: &f,
	}

	// simulate user asking a question that requires the function
	dialogue := []openrouter.ChatCompletionMessage{
		{Role: openrouter.ChatMessageRoleUser, Content: "What is the weather in Boston today?"},
	}
	fmt.Printf("Asking openrouter '%v' and providing it a '%v()' function...\n",
		dialogue[0].Content, f.Name)

	resp, err := client.CreateChatCompletion(ctx,
		openrouter.ChatCompletionRequest{
			Model:    openrouter.GeminiFlash8B,
			Provider: provider,
			Messages: dialogue,
			Tools:    []openrouter.Tool{t},
		},
	)
	if err != nil || len(resp.Choices) != 1 {
		fmt.Printf("Completion error: err:%v len(choices):%v\n", err,
			len(resp.Choices))
		b, _ := json.MarshalIndent(resp, "", "\t")
		fmt.Printf("resp :\n %s\n", string(b))
		return
	}

	type Argument struct {
		Location string `json:"location"`
		Unit     string `json:"unit"`
	}
	b, _ := json.MarshalIndent(resp, "", "\t")
	fmt.Printf("resp :\n %s\n", string(b))
	msg := resp.Choices[0].Message
	for len(msg.ToolCalls) > 0 {
		dialogue = append(dialogue, msg)
		fmt.Printf("openrouter called us back wanting to invoke our function '%v' with params '%v'\n",
			msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments)

		args := Argument{}
		if err := json.Unmarshal([]byte(msg.ToolCalls[0].Function.Arguments), &args); err != nil {
			fmt.Printf("Error unmarshalling arguments: %v\n", err)
			return
		}
		content := ""
		if args.Unit == "celsius" {
			content = "Sunny and 26 degrees."
		} else {
			content = "Sunny and 80 degrees."
		}
		dialogue = append(dialogue, openrouter.ChatCompletionMessage{
			Role:       openrouter.ChatMessageRoleTool,
			Content:    content,
			ToolCallID: msg.ToolCalls[0].ID,
		})

		// simulate calling the function & responding to openrouter
		fmt.Println("Sending openrouter function's response and requesting the reply to the original question...")
		resp, err = client.CreateChatCompletion(ctx,
			openrouter.ChatCompletionRequest{
				Model:    openrouter.GeminiFlash8B,
				Provider: provider,
				Messages: dialogue,
				Tools:    []openrouter.Tool{t},
			},
		)
		if err != nil || len(resp.Choices) != 1 {
			fmt.Printf("Tool completion error: err:%v len(choices):%v\n", err,
				len(resp.Choices))
			return
		}

		msg = resp.Choices[0].Message
	}
	fmt.Printf("openrouter answered the original request with: %v\n", msg.Content)
}
