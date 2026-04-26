package rag

import (
	"testing"
)

func TestMemoryVectorStore(t *testing.T) {
	store := NewMemoryVectorStore()

	// Add vectors
	vec1 := []float32{1.0, 0.0, 0.0}
	vec2 := []float32{0.0, 1.0, 0.0}
	vec3 := []float32{0.9, 0.1, 0.0}

	store.Add("v1", vec1, map[string]interface{}{"name": "vector1"})
	store.Add("v2", vec2, map[string]interface{}{"name": "vector2"})
	store.Add("v3", vec3, map[string]interface{}{"name": "vector3"})

	// Search
	query := []float32{1.0, 0.0, 0.0}
	results, err := store.Search(query, 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// First result should be v1 (exact match)
	if results[0].ID != "v1" {
		t.Errorf("expected v1, got %s", results[0].ID)
	}
	if results[0].Score < 0.99 {
		t.Errorf("expected score close to 1.0, got %f", results[0].Score)
	}

	// Delete
	store.Delete("v1")
	results, _ = store.Search(query, 2)
	if len(results) != 2 {
		t.Errorf("expected 2 results after delete, got %d", len(results))
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []float32
		expected float64
	}{
		{"identical", []float32{1, 0, 0}, []float32{1, 0, 0}, 1.0},
		{"orthogonal", []float32{1, 0, 0}, []float32{0, 1, 0}, 0.0},
		{"opposite", []float32{1, 0, 0}, []float32{-1, 0, 0}, -1.0},
		{"empty", []float32{}, []float32{}, 0.0},
		{"diff_len", []float32{1, 0}, []float32{1, 0, 0}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := cosineSimilarity(tt.a, tt.b)
			if abs(score-tt.expected) > 0.001 {
				t.Errorf("expected %f, got %f", tt.expected, score)
			}
		})
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
