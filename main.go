package main

import (
	"bufio"
	"fmt"
	"os"

	"debuggo/internal/embed"
	"debuggo/internal/gpt"
	"debuggo/internal/vectordb"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		panic("OPENAI_API_KEY not set")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter log error to embed: ")
	inputText, _ := reader.ReadString('\n')

	embedder := embed.NewEmbedder(openaiKey)
	vector, err := embedder.GetEmbedding(inputText)
	if err != nil {
		panic(fmt.Errorf("embedding failed: %w", err))
	}

	vdb := vectordb.NewQdrantClient("localhost:6333")
	similarDocs, err := vdb.SearchSimilar(vector, 3)
	if err != nil {
		panic(fmt.Errorf("vector search failed: %w", err))
	}

	gptClient := gpt.NewGPTClient(openaiKey)
	response, err := gptClient.GenerateFix(inputText, similarDocs)
	if err != nil {
		panic(fmt.Errorf("gpt generation failed: %w", err))
	}

	fmt.Println("\n🔥 GPT Diagnosis:")
	fmt.Println(response)
}
