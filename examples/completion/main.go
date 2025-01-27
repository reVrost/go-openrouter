package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/revrost/go-openrouter"
)

func main() {
	ctx := context.Background()
	client := openrouter.NewClient(os.Getenv("OPENROUTER_API_KEY"))
	request := openrouter.ChatCompletionRequest{
		Model: openrouter.DeepseekV3,
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role:    openrouter.ChatMessageRoleUser,
				Content: "Hello! Respond with just 'world'",
			},
		},
	}

	res, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		fmt.Println("error", err)
	} else {
		b, _ := json.MarshalIndent(res, "", "\t")
		fmt.Printf("request :\n %s", string(b))
	}
}
