package rag

import (
	"fmt"
	"math"
)

// VectorStore 向量存储接口
type VectorStore interface {
	Add(id string, embedding []float32, metadata map[string]interface{}) error
	Search(embedding []float32, topK int) ([]SearchResult, error)
	Delete(id string) error
}

// SearchResult 搜索结果
type SearchResult struct {
	ID       string
	Score    float64
	Metadata map[string]interface{}
}

// MemoryVectorStore 内存向量存储（简化实现）
type MemoryVectorStore struct {
	vectors  map[string][]float32
	metadata map[string]map[string]interface{}
}

// NewMemoryVectorStore 创建新的内存向量存储
func NewMemoryVectorStore() *MemoryVectorStore {
	return &MemoryVectorStore{
		vectors:  make(map[string][]float32),
		metadata: make(map[string]map[string]interface{}),
	}
}

// Add 添加向量
func (s *MemoryVectorStore) Add(id string, embedding []float32, metadata map[string]interface{}) error {
	if len(embedding) == 0 {
		return fmt.Errorf("empty embedding")
	}

	s.vectors[id] = embedding
	s.metadata[id] = metadata
	return nil
}

// Search 搜索相似向量
func (s *MemoryVectorStore) Search(embedding []float32, topK int) ([]SearchResult, error) {
	if len(embedding) == 0 {
		return nil, fmt.Errorf("empty query embedding")
	}

	var results []SearchResult

	for id, vec := range s.vectors {
		if len(vec) != len(embedding) {
			continue
		}

		score := cosineSimilarity(embedding, vec)
		results = append(results, SearchResult{
			ID:       id,
			Score:    score,
			Metadata: s.metadata[id],
		})
	}

	// 按相似度排序
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 取 Top-K
	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}

// Delete 删除向量
func (s *MemoryVectorStore) Delete(id string) error {
	delete(s.vectors, id)
	delete(s.metadata, id)
	return nil
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
