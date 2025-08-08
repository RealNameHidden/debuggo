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
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	// Step 0: Load ENV vars (OPENAI key, Qdrant config, etc.)
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		panic("OPENAI_API_KEY not set")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter log error to embed: ")
	inputText, _ := reader.ReadString('\n')

	// Step 2: Embed the input text
	embedder := embed.NewEmbedder(openaiKey)
	vector, err := embedder.GetEmbedding(inputText)
	if err != nil {
		panic(fmt.Errorf("embedding failed: %w", err))
	}

	// Step 3: Query top-k similar from vector DB (e.g., Qdrant)
	vdb := vectordb.NewQdrantClient("localhost:6333") // TODO: Implement this
	similarDocs, err := vdb.SearchSimilar(vector, 3)
	if err != nil {
		panic(fmt.Errorf("vector search failed: %w", err))
	}

	// Step 4: Format GPT prompt using input + retrieved docs
	gptClient := gpt.NewGPTClient(openaiKey)
	response, err := gptClient.GenerateFix(inputText, similarDocs)
	if err != nil {
		panic(fmt.Errorf("gpt generation failed: %w", err))
	}

	// Step 5: Print the root cause + fix
	fmt.Println("\n🔥 GPT Diagnosis:")
	fmt.Println(response)
}
