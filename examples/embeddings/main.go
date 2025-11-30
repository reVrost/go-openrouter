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

	// Basic text embedding example
	request := openrouter.EmbeddingsRequest{
		Model: "openai/text-embedding-3-large",
		Input: []string{
			"Hello world",
			"OpenRouter embeddings example",
		},
		EncodingFormat: openrouter.EmbeddingsEncodingFormatFloat,
	}

	res, err := client.CreateEmbeddings(ctx, request)
	if err != nil {
		fmt.Println("error", err)
		return
	}

	b, _ := json.MarshalIndent(res, "", "\t")
	fmt.Printf("response :\n %s\n", string(b))
}


