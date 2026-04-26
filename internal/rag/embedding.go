package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// EmbeddingClient embedding 客户端
type EmbeddingClient struct {
	BaseURL    string
	Model      string
	APIKey     string
	HTTPClient *http.Client
}

// NewEmbeddingClient 创建 embedding 客户端
func NewEmbeddingClient(baseURL, model, apiKey string) *EmbeddingClient {
	return &EmbeddingClient{
		BaseURL:    strings.TrimSuffix(baseURL, "/"),
		Model:      model,
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// EmbedText 对文本生成 embedding
func (c *EmbeddingClient) EmbedText(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"model": c.Model,
		"input": text,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/embeddings", strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request embeddings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embeddings API returned %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return result.Data[0].Embedding, nil
}

// EmbedTexts 批量生成 embedding
func (c *EmbeddingClient) EmbedTexts(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := c.EmbedText(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("embed text %d: %w", i, err)
		}
		results[i] = emb
	}
	return results, nil
}
