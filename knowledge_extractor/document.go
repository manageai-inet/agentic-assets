package knowledgeextractor

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	am "github.com/manageai-inet/agentic-assets"
)

type HttpAuthentication struct {
	Type  string
	Token string
}

func (a *HttpAuthentication) ApplyHeader(headers *http.Header) error {
	if a.Type == "bearer" {
		headers.Set("Authorization", "Bearer "+a.Token)
	} else if a.Type == "basic" {
		headers.Set("Authorization", "Basic "+a.Token)
	} else if a.Type == "apikey" {
		headers.Set("Authorization", "Apikey "+a.Token)
	} else {
		return fmt.Errorf("unknown authentication type: %s", a.Type)
	}
	return nil
}

type HttpFileLoader struct {
	auth *HttpAuthentication
	am.LoggingCapacity
}

func NewHttpFileLoader(auth *HttpAuthentication) *HttpFileLoader {
	return &HttpFileLoader{auth: auth, LoggingCapacity: *am.GetDefaultLoggingCapacity()}
}

func (p *HttpFileLoader) String() string {
	return "HttpFileLoader"
}

func (p *HttpFileLoader) IsApplicable(sourceName, sourceUrl string, metadata *map[string]any) bool {
	return (strings.HasPrefix(sourceUrl, "http://") || strings.HasPrefix(sourceUrl, "https://"))
}

func (p *HttpFileLoader) Load(ctx context.Context, sourceName, sourceUrl string, metadata *map[string]any) ([]byte, error) {
	logger := am.GetLogger(p)
	logger.InfoContext(ctx, "Loading File with HTTP Loader", slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceUrl, nil)
	if err != nil {
		return nil, err
	}

	if p.auth != nil {
		if err := p.auth.ApplyHeader(&req.Header); err != nil {
			return nil, err
		}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to load file: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	logger.InfoContext(ctx, "File loaded successfully", slog.Int("dataLength", len(data)))
	return data, nil
}

type LocalFileLoader struct {
	rootPath string
	am.LoggingCapacity
}

func NewLocalFileLoader(rootPath string) *LocalFileLoader {
	return &LocalFileLoader{rootPath: rootPath, LoggingCapacity: *am.GetDefaultLoggingCapacity()}
}

func (p *LocalFileLoader) String() string {
	return "LocalFileLoader"
}

func (p *LocalFileLoader) IsApplicable(sourceName, sourceUrl string, metadata *map[string]any) bool {
	return strings.HasPrefix(sourceUrl, "file://")
}

func (p *LocalFileLoader) Load(ctx context.Context, sourceName, sourceUrl string, metadata *map[string]any) ([]byte, error) {
	logger := am.GetLogger(p)
	logger.InfoContext(ctx, "Loading File with Local Loader", slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl))
	path := strings.TrimPrefix(sourceUrl, "file://")
	if p.rootPath != "" {
		path = p.rootPath + "/" + path
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	logger.InfoContext(ctx, "File loaded successfully", slog.Int("dataLength", len(data)))
	return data, nil
}

type DocumentExtractor struct {
	// for convert pdf to knowledge source
	LoaderToConverter map[KnowledgeLoader]KnowledgeConverter
	am.LoggingCapacity
}

func NewDocumentExtractor(l2c map[KnowledgeLoader]KnowledgeConverter) *DocumentExtractor {
	return &DocumentExtractor{LoaderToConverter: l2c}
}

func (ext *DocumentExtractor) String() string {
	return "DocumentExtractor"
}

func (ext *DocumentExtractor) SetLogger(logger *slog.Logger) {
	ext.LoggingCapacity.SetLogger(logger)
	for loader, converter := range ext.LoaderToConverter {
		loader.SetLogger(logger)
		converter.SetLogger(logger)
	}
}

func (ext *DocumentExtractor) Extract(ctx context.Context, kbId string, sourceName string, sourceUrl string, metadata *map[string]any) ([]am.KnowledgeSource, error) {
	logger := am.GetLogger(ext)
	logger.InfoContext(ctx, "Extracting Document", slog.String("extractor", ext.String()), slog.String("kbId", kbId), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl))
	for loader, converter := range ext.LoaderToConverter {
		logger.Debug("Checking if loader is applicable", slog.String("loader", loader.String()), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl))
		if loader.IsApplicable(sourceName, sourceUrl, metadata) {
			logger.Debug("Loader is applicable", slog.String("loader", loader.String()))
			logger.Debug("Loading Document with loader", slog.String("loader", loader.String()), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl))
			data, err := loader.Load(ctx, sourceName, sourceUrl, metadata)
			if err != nil {
				// if error try next loader
				logger.ErrorContext(ctx, "Failed to load Document with loader, let's try next loader", slog.String("loader", loader.String()), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl), slog.String("error", err.Error()))
				continue
			}
			logger.Debug("Document loaded successfully", slog.String("loader", loader.String()), slog.Int("dataLength", len(data)))
			logger.Debug("Converting Document to knowledge source", slog.String("loader", loader.String()), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl))
			converted, err := converter.Convert(ctx, kbId, sourceName, sourceUrl, data, metadata)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to convert Document to knowledge source", slog.String("loader", loader.String()), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl), slog.String("error", err.Error()))
				return nil, err
			}

			logger.InfoContext(ctx, "Document extracted successfully", slog.String("loader", loader.String()), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl), slog.Int("knowledgeSourceCount", len(converted)))
			return converted, nil
		}
	}
	logger.ErrorContext(ctx, "No applicable loader found", slog.String("kbId", kbId), slog.String("sourceName", sourceName), slog.String("sourceUrl", sourceUrl))
	return nil, fmt.Errorf("no applicable loader found")
}