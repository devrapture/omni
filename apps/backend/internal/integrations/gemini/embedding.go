package gemini

import (
	"context"
	"fmt"
	"math"

	"google.golang.org/genai"
)

type EmbeddingClient struct {
	client         *genai.Client
	embeddingModel string
	dimension      int32
}

func NewEmbeddingClient(ctx context.Context, apiKey string) (*EmbeddingClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &EmbeddingClient{
		client:         client,
		embeddingModel: "gemini-embedding-001",
		dimension:      int32(768),
	}, nil
}

// Use this when indexing your business content.
func (c *EmbeddingClient) EmbedDocument(ctx context.Context, text string) ([]float32, error) {
	result, err := c.client.Models.EmbedContent(ctx, c.embeddingModel, genai.Text(text), &genai.EmbedContentConfig{
		TaskType:             "RETRIEVAL_DOCUMENT",
		OutputDimensionality: &c.dimension,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to embed content: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return normalize(result.Embeddings[0].Values), nil
}

// EmbedQuery embeds a user question for search.
// Use this when a user asks a question (slightly different task type from document embedding).
func (c *EmbeddingClient) EmbedQuestion(ctx context.Context, text string) ([]float32, error) {
	result, err := c.client.Models.EmbedContent(ctx, c.embeddingModel, genai.Text(text), &genai.EmbedContentConfig{
		TaskType:             "RETRIEVAL_QUERY",
		OutputDimensionality: &c.dimension,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to embed question: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return normalize(result.Embeddings[0].Values), nil
}

// EmbedBatch embeds multiple texts at once.
// Returns a slice of embeddings, one per input text.
func (c *EmbeddingClient) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	contents := make([]*genai.Content, len(texts))
	for i, text := range texts {
		contents[i] = genai.NewContentFromText(text, genai.RoleUser)
	}

	result, err := c.client.Models.EmbedContent(ctx, c.embeddingModel, contents, &genai.EmbedContentConfig{
		TaskType:             "RETRIEVAL_DOCUMENT",
		OutputDimensionality: &c.dimension,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to embed batch: %w", err)
	}

	embeddings := make([][]float32, len(result.Embeddings))
	for i, emb := range result.Embeddings {
		embeddings[i] = normalize(emb.Values)
	}

	return embeddings, nil
}

// normalize L2-normalizes the embedding vector.
// Required for gemini-embedding-001 when using dimensions < 3072.
func normalize(v []float32) []float32 {
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	norm := float32(math.Sqrt(sum))
	if norm == 0 {
		return v
	}
	for i := range v {
		v[i] /= norm
	}
	return v
}
