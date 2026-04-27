package assetmanager

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"math/rand"
)

type FakeEmbedder struct {
	ModelName string
	Dimension int
}

func NewFakeEmbedder(model string, dim int) *FakeEmbedder {
	return &FakeEmbedder{
		ModelName: model,
		Dimension: dim,
	}
}

func (m *FakeEmbedder) GetEmbeddingModel() string {
	return m.ModelName
}

func (m *FakeEmbedder) GetEmbeddingDim() int {
	return m.Dimension
}

// Embed generates a deterministic pseudo-random vector based on the content string
func (m *FakeEmbedder) Embed(ctx context.Context, content string) ([]float32, error) {
	// 1. Create a deterministic seed from the content string using MD5
	hash := md5.Sum([]byte(content))
	seed := int64(binary.BigEndian.Uint64(hash[:8]))
	
	// 2. Initialize a local random generator with that seed
	rng := rand.New(rand.NewSource(seed))
	
	vec := make([]float32, m.Dimension)
	for i := 0; i < m.Dimension; i++ {
		// Generating values between -1 and 1
		vec[i] = rng.Float32()*2 - 1
	}
	
	return vec, nil
}

func (m *FakeEmbedder) EmbedBatch(ctx context.Context, contents []string) ([][]float32, error) {
	results := make([][]float32, len(contents))
	for i, content := range contents {
		vec, err := m.Embed(ctx, content)
		if err != nil {
			return nil, err
		}
		results[i] = vec
	}
	return results, nil
}