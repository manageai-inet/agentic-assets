package knowledgeextractor

import (
	"context"
	"log/slog"
	"strings"

	am "github.com/manageai-inet/agentic-assets"
)

var plainTextApplicableExts = []string{"txt", "text", "md", "markdown", "json", "html", "htm", "xml", "yaml", "yml", "toml"}

type HttpPlainTextLoader struct {
	httpLoader *HttpFileLoader
	am.LoggingCapacity
}

func NewHttpPlainTextLoader(auth *HttpAuthentication) *HttpPlainTextLoader {
	return &HttpPlainTextLoader{httpLoader: NewHttpFileLoader(auth), LoggingCapacity: *am.GetDefaultLoggingCapacity()}
}

func (p *HttpPlainTextLoader) String() string {
	return "HttpPlainTextLoader"
}

func (p *HttpPlainTextLoader) IsApplicable(sourceName, sourceUrl string, metadata *map[string]any) bool {
	if !p.httpLoader.IsApplicable(sourceName, sourceUrl, metadata) {
		return false
	}
	sourceSplited := strings.Split(sourceName, ".")
	sourceExt := strings.ToLower(sourceSplited[len(sourceSplited)-1])
	for _, ext := range plainTextApplicableExts {
		if ext == sourceExt {
			return true
		}
	}
	return false
}

func (p *HttpPlainTextLoader) Load(ctx context.Context, sourceName, sourceUrl string, metadata *map[string]any) ([]byte, error) {
	return p.httpLoader.Load(ctx, sourceName, sourceUrl, metadata)
}

type LocalPlainTextLoader struct {
	localLoader *LocalFileLoader
	am.LoggingCapacity
}
	
func NewLocalPlainTextLoader(rootPath string) *LocalPlainTextLoader {
	return &LocalPlainTextLoader{localLoader: NewLocalFileLoader(rootPath), LoggingCapacity: *am.GetDefaultLoggingCapacity()}
}

func (p *LocalPlainTextLoader) String() string {
	return "LocalPlainTextLoader"
}

func (p *LocalPlainTextLoader) IsApplicable(sourceName, sourceUrl string, metadata *map[string]any) bool {
	if !p.localLoader.IsApplicable(sourceName, sourceUrl, metadata) {
		return false
	}
	sourceSplited := strings.Split(sourceName, ".")
	sourceExt := strings.ToLower(sourceSplited[len(sourceSplited)-1])
	for _, ext := range plainTextApplicableExts {
		if ext == sourceExt {
			return true
		}
	}
	return false
}

func (p *LocalPlainTextLoader) Load(ctx context.Context, sourceName, sourceUrl string, metadata *map[string]any) ([]byte, error) {
	return p.localLoader.Load(ctx, sourceName, sourceUrl, metadata)
}

type PlainTextConverter struct {
	am.LoggingCapacity
}

func NewPlainTextConverter() *PlainTextConverter {
	return &PlainTextConverter{LoggingCapacity: *am.GetDefaultLoggingCapacity()}
}

func (p *PlainTextConverter) String() string {
	return "PlainTextConverter"
}

func (p *PlainTextConverter) Convert(ctx context.Context, kbId string, sourceName string, sourceUrl string, data []byte, metadata *map[string]any) ([]am.KnowledgeSource, error) {
	logger := am.GetLogger(p)
	logger.InfoContext(ctx, "Converting PlainText to knowledge source", slog.String("kbId", kbId), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl))
	// convert byte data to string
	text := string(data)
	logger.InfoContext(ctx, "PlainText converted successfully", slog.String("kbId", kbId), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl))
	return []am.KnowledgeSource{
		{
			SourceType: am.AssetTypePage,
			SourceName: kbId + ":" + sourceName + ":1",
			SourceUrl:  &sourceUrl,
			SourceData: &data,
			SourceContents: &text,
			Metadata: metadata,
		},
	}, nil
}
