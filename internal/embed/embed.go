package embed

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	openai "github.com/sashabaranov/go-openai"
)

// EmbedderInterface defines the interface for embedders
type EmbedderInterface interface {
	GetEmbedding(text string) ([]float32, error)
}

// OpenAI Embedder (costs money)
type Embedder struct {
	client *openai.Client
}

func NewEmbedder(apiKey string) *Embedder {
	client := openai.NewClient(apiKey)
	return &Embedder{client: client}
}

func (e *Embedder) GetEmbedding(text string) ([]float32, error) {
	resp, err := e.client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return resp.Data[0].Embedding, nil
}

// Local Embedder (free after setup)
type LocalEmbedder struct {
	modelName string
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
	Error     string    `json:"error,omitempty"`
}

func NewLocalEmbedder() *LocalEmbedder {
	return &LocalEmbedder{
		modelName: "all-MiniLM-L6-v2", // Fast, good quality model
	}
}

func (e *LocalEmbedder) GetEmbedding(text string) ([]float32, error) {
	pythonScript := fmt.Sprintf(`
import json
import sys
try:
    from sentence_transformers import SentenceTransformer
    model = SentenceTransformer('%s')
    text = %q
    embedding = model.encode(text).tolist()
    result = {"embedding": embedding}
    print(json.dumps(result))
except Exception as e:
    result = {"error": str(e)}
    print(json.dumps(result))
`, e.modelName, text)

	// Check if virtual environment exists
	pythonCmd := "python3"
	if _, err := os.Stat(".venv/bin/python"); err == nil {
		pythonCmd = ".venv/bin/python"
	}

	cmd := exec.Command(pythonCmd, "-c", pythonScript)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute python script: %w", err)
	}

	var response EmbeddingResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf("python embedding error: %s", response.Error)
	}

	return response.Embedding, nil
}

func (e *LocalEmbedder) CheckDependencies() error {
	checkScript := `
try:
    import sentence_transformers
    print("OK")
except ImportError:
    print("MISSING")
`

	// Check if virtual environment exists
	pythonCmd := "python3"
	if _, err := os.Stat(".venv/bin/python"); err == nil {
		pythonCmd = ".venv/bin/python"
	}

	cmd := exec.Command(pythonCmd, "-c", checkScript)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("python not available: %w", err)
	}

	if string(output) != "OK\n" {
		return fmt.Errorf("sentence-transformers not installed. Run: ./scripts/install_local_embeddings.sh")
	}

	return nil
}

// Factory function to create the right embedder
func CreateEmbedder(useLocal bool, openaiKey string) (EmbedderInterface, error) {
	if useLocal {
		localEmbedder := NewLocalEmbedder()
		if err := localEmbedder.CheckDependencies(); err != nil {
			return nil, fmt.Errorf("local embedder not available: %w", err)
		}
		return localEmbedder, nil
	}
	return NewEmbedder(openaiKey), nil
}
