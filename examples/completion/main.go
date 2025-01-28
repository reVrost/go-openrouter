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
				Role:    openrouter.ChatMessageRoleSystem,
				Content: "You are a helfpul assistant.",
			},
			{
				Role:    openrouter.ChatMessageRoleUser,
				Content: "Hello!",
			},
		},
		Stream: false,
	}

	res, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		fmt.Println("error", err)
	} else {
		b, _ := json.MarshalIndent(res, "", "\t")
		fmt.Printf("request :\n %s", string(b))
	}
}
