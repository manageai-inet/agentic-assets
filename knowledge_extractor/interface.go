package knowledgeextractor

import (
	"context"

	am "github.com/manageai-inet/agentic-assets"
)

type KnowledgeLoader interface {
	// return kind name of loader
	String() string
	// verify if source name, url, metadata can be loaded by this loader
	IsApplicable(sourceName, sourceUrl string, metadata *map[string]any) bool
	// load knowledge data from given url
	Load(ctx context.Context, sourceName, sourceUrl string, metadata *map[string]any) ([]byte, error)
	am.Loggable
}

type KnowledgeConverter interface {
	// return kind name of converter
	String() string
	// convert single part of knowledge data into knowledge source(s)
	Convert(ctx context.Context, kbId, sourceName, sourceUrl string, sourceData []byte, metadata *map[string]any) ([]am.KnowledgeSource, error)
	am.Loggable
}

type KnowledgeExtractor interface {
	// return kind name of extractor
	String() string
	Extract(ctx context.Context, kbId string, sourceName string, sourceUrl string, metadata *map[string]any) ([]am.KnowledgeSource, error)
	am.Loggable
}