package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

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

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🔧 DebugGo - AI-Powered Error Analysis")
	fmt.Println("=====================================")

	// Initialize Qdrant client and connect
	vdb := vectordb.NewQdrantClient("localhost:6333")
	ctx := context.Background()
	if err := vdb.Connect(ctx, "localhost:6333"); err != nil {
		fmt.Printf("⚠️ Warning: Could not connect to Qdrant (%v)\n", err)
		fmt.Println("💡 To use full functionality, start Qdrant: docker run -p 6333:6333 qdrant/qdrant")
	} else {
		// Show database stats
		if stats, err := vdb.GetStats(); err == nil {
			fmt.Printf("📊 Database: %d stored errors\n", stats["total_embeddings"])
		}
	}

	fmt.Println("1. Log new error")
	fmt.Println("2. Ask for solution")
	fmt.Print("\nChoose an option (1 or 2): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		logNewError(reader, openaiKey, vdb)
	case "2":
		askForSolution(reader, openaiKey, vdb)
	default:
		fmt.Println("Invalid choice. Please run the program again and select 1 or 2.")
	}
}

func logNewError(reader *bufio.Reader, openaiKey string, vdb *vectordb.QdrantClient) {
	fmt.Print("\nEnter the error you want to log: ")
	errorText, _ := reader.ReadString('\n')
	errorText = strings.TrimSpace(errorText)

	fmt.Print("\nHow did you fix this error? (Describe the solution): ")
	solutionText, _ := reader.ReadString('\n')
	solutionText = strings.TrimSpace(solutionText)

	fmt.Println("\n📝 Logging error and generating embedding...")

	combinedText := fmt.Sprintf("Error: %s\nSolution: %s", errorText, solutionText)

	// Try local embedder first (free), fallback to OpenAI if needed
	embedder, err := embed.CreateEmbedder(true, openaiKey)
	if err != nil {
		fmt.Println("⚠️  Local embeddings not available, using OpenAI (costs money)")
		embedder, err = embed.CreateEmbedder(false, openaiKey)
		if err != nil {
			fmt.Printf("Error creating embedder: %v\n", err)
			return
		}
	} else {
		fmt.Println("✅ Using local embeddings (free!)")
	}

	// Store in vector database
	ctx := context.Background()

	metadata := map[string]interface{}{
		"error_type":     "user_logged",
		"has_solution":   true,
		"original_error": errorText,
		"solution":       solutionText,
	}

	err = vdb.StoreEmbedding(ctx, embedder, combinedText, metadata)
	if err != nil {
		fmt.Printf("Error storing embedding: %v\n", err)
		fmt.Println("💡 Make sure Qdrant is running: docker run -p 6333:6333 qdrant/qdrant")
		return
	}

	fmt.Println("✅ Error and solution stored successfully!")
	fmt.Println("📊 The error has been indexed and will be available for future searches.")
	fmt.Printf("📋 Summary:\n")
	fmt.Printf("   Problem: %s\n", errorText)
	fmt.Printf("   Solution: %s\n", solutionText)
}

func askForSolution(reader *bufio.Reader, openaiKey string, vdb *vectordb.QdrantClient) {

	fmt.Print("\nDescribe the error you need help with: ")
	queryText, _ := reader.ReadString('\n')
	queryText = strings.TrimSpace(queryText)

	fmt.Println("\n🔍 Searching for similar errors...")

	// Try local embedder first (free), fallback to OpenAI if needed
	embedder, err := embed.CreateEmbedder(true, openaiKey)
	if err != nil {
		fmt.Println("⚠️  Local embeddings not available, using OpenAI (costs money)")
		embedder, err = embed.CreateEmbedder(false, openaiKey)
		if err != nil {
			fmt.Printf("Error creating embedder: %v\n", err)
			return
		}
	} else {
		fmt.Println("✅ Using local embeddings for search (free!)")
	}

	vector, err := embedder.GetEmbedding(queryText)
	if err != nil {
		fmt.Printf("Error generating embedding: %v\n", err)
		return
	}

	similarDocs, err := vdb.SearchSimilar(vector, 3)
	if err != nil {
		fmt.Printf("Error searching database: %v\n", err)
		fmt.Println("💡 Make sure Qdrant is running: docker run -p 6333:6333 qdrant/qdrant")
		return
	}

	// Display similar errors found
	fmt.Printf("\n📋 Found %d similar error(s):\n", len(similarDocs))
	for i, doc := range similarDocs {
		fmt.Printf("\n--- Similar Error %d ---\n%s", i+1, doc)
	}

	if openaiKey == "" {
		fmt.Println("❌ OPENAI_API_KEY not found for generating AI solutions")
		fmt.Println("Refer similar docs above!")
		return
	}

	fmt.Println("\n🤖 Generating AI solution (costs money)...")

	gptClient := gpt.NewGPTClient(openaiKey)
	response, err := gptClient.GenerateFix(queryText, similarDocs)
	if err != nil {
		fmt.Printf("Error generating solution: %v\n", err)
		return
	}

	fmt.Println("\n🔥 AI Diagnosis:")
	fmt.Println("================")
	fmt.Println(response)
}
