package chunk

import (
	"context"

	am "github.com/manageai-inet/agentic-assets"
)

const ChunkAssetType = "chunk"

type ChunkImmediateAsset struct {
	KbId string
	AssetId        string
	Content        string
	PageNumber     int
	StartPos       int
	Metadata       *map[string]any
}

type ChunkExtractor interface {
	// extract chunks from sources
	Extract(ctx context.Context, kbId string, sources []am.KnowledgeSource) ([]ChunkImmediateAsset, error)
	// return list of changed chunks, if no change return empty list
	Compare(ctx context.Context, oldChunks, newChunks []ChunkImmediateAsset) ([]ChunkImmediateAsset, error)
}