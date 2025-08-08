package embed

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

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
