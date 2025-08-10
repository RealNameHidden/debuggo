package vectordb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"debuggo/internal/embed"
)

type QdrantClient struct {
	baseURL    string
	collection string
	httpClient *http.Client
}

type QdrantPoint struct {
	ID      int64                  `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

type QdrantUpsertRequest struct {
	Points []QdrantPoint `json:"points"`
}

type QdrantSearchRequest struct {
	Vector      []float32 `json:"vector"`
	Limit       int       `json:"limit"`
	WithPayload bool      `json:"with_payload"`
}

type QdrantSearchResponse struct {
	Result []QdrantSearchResult `json:"result"`
}

type QdrantSearchResult struct {
	ID      int64                  `json:"id"`
	Score   float32                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

type QdrantCollectionResponse struct {
	Result QdrantCollectionInfo `json:"result"`
}

type QdrantCollectionInfo struct {
	PointsCount uint64 `json:"points_count"`
	Config      struct {
		Params struct {
			VectorSize uint64 `json:"vector_size"`
		} `json:"params"`
	} `json:"config"`
}

func NewQdrantClient(url string) *QdrantClient {
	return &QdrantClient{
		baseURL:    fmt.Sprintf("http://%s", url),
		collection: "debug_errors",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *QdrantClient) Connect(ctx context.Context, url string) error {
	c.baseURL = fmt.Sprintf("http://%s", url)
	return nil
}

// ensureCollectionWithSize ensures a collection with the given name and vector size exists
func (c *QdrantClient) ensureCollectionWithSize(ctx context.Context, collection string, size int) error {
	collectionURL := fmt.Sprintf("%s/collections/%s", c.baseURL, collection)
	req, err := http.NewRequestWithContext(ctx, "GET", collectionURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil
	}
	if resp.StatusCode != 404 {
		return fmt.Errorf("unexpected status checking collection: %d", resp.StatusCode)
	}

	createReq := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     size,
			"distance": "Cosine",
		},
	}
	jsonData, _ := json.Marshal(createReq)
	req, err = http.NewRequestWithContext(ctx, "PUT", collectionURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to create collection, status: %d", resp.StatusCode)
	}
	return nil
}

func (c *QdrantClient) StoreEmbedding(ctx context.Context, embedder embed.EmbedderInterface, text string, metadata map[string]interface{}) error {
	vector, err := embedder.GetEmbedding(text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Choose collection per vector dimension to avoid size mismatch
	coll := fmt.Sprintf("%s_%d", c.collection, len(vector))
	if err := c.ensureCollectionWithSize(ctx, coll, len(vector)); err != nil {
		return fmt.Errorf("failed to ensure collection: %w", err)
	}

	pointID := time.Now().UnixNano()
	payload := map[string]interface{}{
		"text":      text,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	for k, v := range metadata {
		payload[k] = v
	}
	point := QdrantPoint{ID: pointID, Vector: vector, Payload: payload}

	upsertReq := QdrantUpsertRequest{Points: []QdrantPoint{point}}
	jsonData, err := json.Marshal(upsertReq)
	if err != nil {
		return err
	}
	// upsertURL := fmt.Sprintf("%s/collections/%s/points", c.baseURL, coll)
	upsertURL := fmt.Sprintf("%s/collections/%s/points", c.baseURL, coll)
	fmt.Println(string(jsonData))
	fmt.Println(upsertURL)
	req, err := http.NewRequestWithContext(ctx, "PUT", upsertURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to store embedding, status: %d", resp.StatusCode)
	}
	return nil
}

func (c *QdrantClient) SearchSimilar(vector []float32, k int) ([]string, error) {
	ctx := context.Background()
	coll := fmt.Sprintf("%s_%d", c.collection, len(vector))
	if err := c.ensureCollectionWithSize(ctx, coll, len(vector)); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	searchReq := QdrantSearchRequest{Vector: vector, Limit: k, WithPayload: true}
	jsonData, err := json.Marshal(searchReq)
	if err != nil {
		return nil, err
	}
	searchURL := fmt.Sprintf("%s/collections/%s/points/search", c.baseURL, coll)
	req, err := http.NewRequestWithContext(ctx, "POST", searchURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("search failed, status: %d", resp.StatusCode)
	}
	var searchResp QdrantSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}
	if len(searchResp.Result) == 0 {
		return []string{"No similar errors found in the database."}, nil
	}
	var similarDocs []string
	for _, result := range searchResp.Result {
		text := "Unknown content"
		timestamp := "Unknown time"
		if textVal, ok := result.Payload["text"].(string); ok {
			text = textVal
		}
		if tsVal, ok := result.Payload["timestamp"].(string); ok {
			timestamp = tsVal
		}
		doc := fmt.Sprintf("Similarity: %.2f\nContent: %s\nTimestamp: %s\n", result.Score, text, timestamp)
		similarDocs = append(similarDocs, doc)
	}
	return similarDocs, nil
}

func (c *QdrantClient) GetStats() (map[string]interface{}, error) {
	ctx := context.Background()
	// Default to 384 collection for stats; if not present, return zero
	coll := fmt.Sprintf("%s_%d", c.collection, 384)
	infoURL := fmt.Sprintf("%s/collections/%s", c.baseURL, coll)
	req, err := http.NewRequestWithContext(ctx, "GET", infoURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return map[string]interface{}{
			"total_embeddings": 0,
			"collection":       coll,
			"status":           "collection_not_found",
		}, nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get stats, status: %d", resp.StatusCode)
	}
	var collectionResp QdrantCollectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&collectionResp); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_embeddings": int(collectionResp.Result.PointsCount),
		"collection":       coll,
		"vector_size":      int(collectionResp.Result.Config.Params.VectorSize),
	}, nil
}

func (c *QdrantClient) Close() error { return nil }
