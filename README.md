# agentic-assets

A Go library providing abstraction layers for asset management, indexing, vector storage, and knowledge extraction in AI-powered applications.

## Overview

`agentic-assets` is a comprehensive Go package that provides interfaces and abstractions for building AI-driven knowledge management systems. It supports:

- **Asset Management**: CRUD operations for various asset types (documents, pages, chunks, entities, relations)
- **Vector Storage**: Integration with vector databases for similarity search
- **Embedding**: Support for different embedding models
- **Knowledge Extraction**: Pluggable extractors for various document formats
- **Chunking**: Intelligent text chunking strategies
- **Indexing & Retrieval**: Full-text and semantic search capabilities

## Installation

```bash
go get github.com/manageai-inet/agentic-assets
```

## Core Interfaces

### AssetStorage

Manages CRUD operations for contextual assets:

```go
type AssetStorage interface {
    Setup(ctx context.Context) error
    GetVersions(ctx context.Context, kbId string) ([]int, error)
    GetAssets(ctx context.Context, kbId string, assetType string, assetIds *[]string, version *int, label *string) ([]ContextualAsset, error)
    InsertAsset(ctx context.Context, asset ContextualAsset) (bool, error)
    // ... more methods
}
```

### Embedder

Handles text embedding:

```go
type Embedder interface {
    GetEmbeddingModel() string
    GetEmbeddingDim() int
    Embed(ctx context.Context, content string) ([]float32, error)
    EmbedBatch(ctx context.Context, contents []string) ([][]float32, error)
}
```

### VectorStorage

Manages vector operations and similarity search:

```go
type VectorStorage interface {
    SetEmbedder(ctx context.Context, embedder Embedder) error
    EmbedContent(ctx context.Context, content string) ([]float32, error)
    QueryVectors(ctx context.Context, queryVector []float32, topK *int, threshold *float32, filter *VectorQueryFilter) ([]RetrievedVector, error)
    // ... more methods
}
```

### IndexerRepo

Provides high-level indexing and retrieval:

```go
type IndexerRepo interface {
    Index(ctx context.Context, kbId string, sources []KnowledgeSource, labels *[]string, metadata *map[string]any, config *map[string]any) (IndexingResult, error)
    Retrieve(ctx context.Context, query string, kbIds []string, seedAssets []ContextualAsset, config *map[string]any) (RetrieveResult, error)
    // ... more methods
}
```

## Asset Types

The library supports several predefined asset types:

- `AssetTypeDocument`: Full documents
- `AssetTypePage`: Document pages
- `AssetTypeChunk`: Text chunks
- `AssetTypeEntity`: Named entities
- `AssetTypeRelation`: Entity relationships

## Knowledge Extraction

The `knowledge_extractor` package provides interfaces for loading and converting various document formats:

```go
type KnowledgeExtractor interface {
    String() string
    Extract(ctx context.Context, kbId string, sourceName string, sourceUrl string, metadata *map[string]any) ([]KnowledgeSource, error)
}
```

## Chunking

The `chunk` package provides chunking strategies for text processing:

```go
type ChunkExtractor interface {
    Extract(ctx context.Context, kbId string, sources []KnowledgeSource) ([]ChunkImmediateAsset, error)
    Compare(ctx context.Context, oldChunks, newChunks []ChunkImmediateAsset) ([]ChunkImmediateAsset, error)
}
```

## License

See [LICENSE](LICENSE) file for details.

## Requirements

- Go 1.25.0 or later
