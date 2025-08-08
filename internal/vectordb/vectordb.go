package vectordb

type QdrantClient struct {
	url string
}

func NewQdrantClient(url string) *QdrantClient {
	return &QdrantClient{url: url}
}

func (c *QdrantClient) SearchSimilar(vector []float32, k int) ([]string, error) {
	// TODO: Implement search functionality
	return []string{}, nil
}
