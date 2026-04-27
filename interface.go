package assetmanager

import (
	"context"
)

type AssetStorage interface {
	Setup(ctx context.Context) error
	GetVersions(ctx context.Context, kbId string) ([]int, error)
	GetAssets(ctx context.Context, kbId string, assetType string, assetIds *[]string, version *int, label *string) ([]ContextualAsset, error)
	// Get assets with given kbId, assetType
	// Filter assets if its refs hold AssetRef with assetIds in `refIds` and refTypes in `refTypes`
	// if `refIds` or `refTypes` are not provided, mean any refIds or refTypes
	GetAssetsByRefs(ctx context.Context, kbId string, assetType string, refIds *[]string, refTypes *[]string, version *int, label *string) ([]ContextualAsset, error)
	GetAssetsByKbId(ctx context.Context, kbId string) ([]ContextualAsset, error)
	CheckAssetsExist(ctx context.Context, kbId string, assetType string, assetIds []string) (map[string]bool, error)
	InsertAsset(ctx context.Context, asset ContextualAsset) (bool, error)
	InsertBatchAssets(ctx context.Context, assets []ContextualAsset) (int, error)
	UpdateAsset(ctx context.Context, kbId string, assetType string, assetId string, version *int, updatedAsset UpdatedContextualAsset) (*ContextualAsset, error)
	DeleteAsset(ctx context.Context, kbId string, assetType string, assetId string, version *int) (bool, error)
	DeleteAssetsByKbId(ctx context.Context, kbId string) (int, error)
	DeleteAssetsByKbIdAndVersion(ctx context.Context, kbId string, version *int) (int, error)
	DeleteAssetsByKbIdAndAssetType(ctx context.Context, kbId string, assetType string) (int, error)
}

// Embedder is an interface that represents an embedder, which is used to embed content into vectors
// You should define EmbeddingModel and EmbeddingDim in the struct that implements this interface
type Embedder interface {
	// GetEmbeddingModel returns the name of the embedding model
	GetEmbeddingModel() string
	// GetEmbeddingDim returns the dimension of the embedding model
	GetEmbeddingDim() int
	// Embed returns the vector of the given content
	Embed(ctx context.Context, content string) ([]float32, error)
	// EmbedBatch returns the vectors of the given contents
	EmbedBatch(ctx context.Context, contents []string) ([][]float32, error)
}

type Reranker interface {
	GetRerankingModel() string
	// return the same order as references
	Rerank(ctx context.Context, query string, references []string) ([]float32, error)
}

const DefaultQueryTopK = 50
const DefaultQueryThreshold = float32(0.3)

type VectorStorage interface {
	// used at initialization step
	SetEmbedder(ctx context.Context, embedder Embedder) error
	// used after initialization step
	Setup(ctx context.Context) error

	// Vector related Methods
	// Call Embed() of Embedder to get vector of given content
	EmbedContent(ctx context.Context, content string) ([]float32, error)
	// Call Embed() of Embedder to get vector of given asset content
	// returned VectorAsset must have the same KbId, AssetId, Version, Labels, and Metadata as input asset
	// But it should reference to the input asset, also set EmbeddingModel the one used for embedding
	//
	// If contentConstructorFn is nil,
	// 	use asset.Content
	// else
	// 	call contentConstructorFn(asset) to get content
	// **Note:** DO NOT store the vector to repository or vector database, just return
	//
	// **Note:** Refs and EmbeddingModel of input asset do not need to be set
	EmbedAsset(ctx context.Context, asset *ContextualAsset, contentConstructorFn *func(ContextualAsset) string) (VectorAsset, error)
	// Query vectors from vector database similar to queryVector
	// Always filter only vector assets that have the same EmbeddingModel as the embedder
	//
	// - If topK is nil, use DefaultQueryTopK (50)
	//
	// - If threshold is nil, use DefaultQueryThreshold (0.3)
	//
	// **Note:** To set similarity metric you should set it in the constructor of VectorStorage if it is supported.
	// because different vector databases may support different similarity metrics.
	QueryVectors(ctx context.Context, queryVector []float32, topK *int, threshold *float32, filter *VectorQueryFilter) ([]RetrievedVector, error)

	// CRUD Methods
	GetVersions(ctx context.Context, kbId string) ([]int, error)
	InsertVector(ctx context.Context, vectorAsset VectorAsset) (bool, error)
	InsertBatchVectors(ctx context.Context, vectorAssets []VectorAsset) (int, error)
	DeleteVector(ctx context.Context, kbId string, assetId string, version *int) (bool, error)
	DeleteVectorsByKbId(ctx context.Context, kbId string) (int, error)
	DeleteVectorsByKbIdAndVersion(ctx context.Context, kbId string, version *int) (int, error)
}

type IndexerRepo interface {
	// Index knowledge
	Index(ctx context.Context, kbId string, sources []KnowledgeSource, labels *[]string, metadata *map[string]any, config *map[string]any) (IndexingResult, error)
	// Retrieve knowledge
	Retrieve(ctx context.Context, query string, kbIds []string, seedAssets []ContextualAsset, config *map[string]any) (RetrieveResult, error)
	// Filter supported sources
	FilterSupportedSources(sources []KnowledgeSource) []KnowledgeSource
	// Get retrieval seed assets
	GetRetrievalSeedAssets(ctx context.Context, query string, kbIds []string, version *int, label *string) []ContextualAsset
}