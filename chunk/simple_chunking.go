package chunk

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	am "github.com/manageai-inet/agentic-assets"
)

const defautlChunkSize = 500
const defaultChunkOverlap = 50
const TokenToCharRatio = 4

var defaultSeparators = []string{
	// Paragraph and page separators
	"\n\n\n", "\n\n", "\r\n\r\n",
	// Sentence ending punctuation
	"。", "．", ".", "！", "!", "？", "?",
}

type positionedChunk struct {
	Content  string
	StartPos int
}

type chunkPart struct {
	Content       string
	ContentLength int
	StartPos      int
}

type SimpleChunkExtractor struct {
	Separators   []string
	ChunkSize    int
	ChunkOverlap int
	splitRegex   *regexp.Regexp
	am.LoggingCapacity
}

func NewSimpleChunkExtractor(chunkSize *int, chunkOverlap *int) *SimpleChunkExtractor {
	var escaped []string
	for _, s := range defaultSeparators {
		escaped = append(escaped, regexp.QuoteMeta(s))
	}
	pattern := "(" + strings.Join(escaped, "|") + ")"
	re := regexp.MustCompile(pattern)
	cs := defautlChunkSize
	if chunkSize != nil {
		cs = *chunkSize
	}
	co := defaultChunkOverlap
	if chunkOverlap != nil {
		co = *chunkOverlap
	}
	return &SimpleChunkExtractor{
		Separators:      defaultSeparators,
		ChunkSize:       cs * TokenToCharRatio,
		ChunkOverlap:    co * TokenToCharRatio,
		splitRegex:      re,
		LoggingCapacity: *am.GetDefaultLoggingCapacity(),
	}
}

func (s *SimpleChunkExtractor) Extract(ctx context.Context, kdId string, pages []am.KnowledgeSource) ([]ChunkImmediateAsset, error) {
	logger := am.GetLogger(s)
	logger.InfoContext(ctx, "Starting Chunck Extraction", slog.String("kbId", kdId), slog.Int("pages", len(pages)))
	docContent := ""
	pageBoundaries := make(map[[2]int]int)

	logger.Debug("Measure Page Boundaries", slog.Int("pages", len(pages)))
	for i, page := range pages {
		if page.SourceType != am.AssetTypePage {
			logger.ErrorContext(ctx, "Source type is not page", slog.String("sourceType", string(page.SourceType)))
			return []ChunkImmediateAsset{}, fmt.Errorf("Source type is not page, got %s", page.SourceType)
		}
		sourceContent := ""
		if page.SourceContents != nil {
			sourceContent = *page.SourceContents
		}
		docContent += sourceContent
		bound := [2]int{len(docContent) - len(sourceContent), len(docContent)}
		pageBoundaries[bound] = i
		logger.Debug("page boundary", slog.Int("page", i), slog.Int("start", bound[0]), slog.Int("end", bound[1]))
	}
	logger.Debug("Extract Chunks", slog.Int("docContentLength", len(docContent)))
	docChunks := s.extractChunks(docContent)

	var chunks []ChunkImmediateAsset
	for i, rawContent := range docChunks {
		pageNumber := -1
		for bound, page := range pageBoundaries {
			if rawContent.StartPos >= bound[0] && rawContent.StartPos < bound[1] {
				pageNumber = page
				break
			}
		}
		if pageNumber < 0 {
			logger.ErrorContext(ctx, "Start position of chunk is out of bound", slog.Int("chunkIndex", i), slog.Int("startPos", rawContent.StartPos))
			return []ChunkImmediateAsset{}, fmt.Errorf("Start position of %d-th chunk is out of bound, got %d", i, rawContent.StartPos)
		}
		page := pages[pageNumber]
		chunkId := fmt.Sprintf("%s[%d:%d]", page.SourceName, rawContent.StartPos, rawContent.StartPos+len(rawContent.Content))

		chunkMetadata := make(map[string]any)
		if page.Metadata != nil {
			for k, v := range *page.Metadata {
				chunkMetadata[k] = v
			}
		}
		chunkMetadata["page_number"] = pageNumber
		chunkMetadata["start_pos"] = rawContent.StartPos
		chunkMetadata["end_pos"] = rawContent.StartPos + len(rawContent.Content)

		chunks = append(chunks, ChunkImmediateAsset{
			KbId:       kdId,
			AssetId:    chunkId,
			Content:    rawContent.Content,
			PageNumber: pageNumber,
			StartPos:   rawContent.StartPos,
			Metadata:   &chunkMetadata,
		})
		logger.Debug(
			"chunk",
			slog.String("chunkId", chunkId),
			slog.Int("pageNumber", pageNumber),
			slog.Int("startPos", rawContent.StartPos),
			slog.Int("endPos", rawContent.StartPos+len(rawContent.Content)),
			slog.Int("chunkLength", len(rawContent.Content)),
		)
	}

	logger.InfoContext(ctx, "Finished Chunck Extraction", slog.Int("chunks", len(chunks)))
	return chunks, nil
}

// For simple chunking, we just return all new chunks, no compare logic
func (s *SimpleChunkExtractor) Compare(ctx context.Context, oldChunks, newChunks []ChunkImmediateAsset) ([]ChunkImmediateAsset, error) {
	logger := am.GetLogger(s)
	logger.InfoContext(ctx, "Comparing Chunks", slog.Int("oldChunks", len(oldChunks)), slog.Int("newChunks", len(newChunks)))
	logger.Debug("just return new chunk")

	return newChunks, nil
}

func (s *SimpleChunkExtractor) extractChunks(text string) []positionedChunk {
	if len(text) <= s.ChunkSize {
		return []positionedChunk{{Content: text, StartPos: 0}}
	}
	return s.splitText(text)
}

func (s *SimpleChunkExtractor) splitText(text string) []positionedChunk {
	// The Python re.split with capture groups returns [text, sep, text, sep, text]
	matches := s.splitRegex.FindAllStringIndex(text, -1)

	var splits []positionedChunk
	last := 0
	for _, match := range matches {
		start := match[0]
		end := match[1]
		// add text before separator
		splits = append(splits, positionedChunk{Content: text[last:start], StartPos: last})
		// add separator
		splits = append(splits, positionedChunk{Content: text[start:end], StartPos: start})
		last = end
	}
	splits = append(splits, positionedChunk{Content: text[last:], StartPos: last}) // add final text

	return s.mergeSplits(splits)
}

func (s *SimpleChunkExtractor) mergeSplits(splits []positionedChunk) []positionedChunk {
	if len(splits) == 0 {
		return []positionedChunk{}
	} else if len(splits) < 2 {
		return []positionedChunk{{Content: splits[0].Content, StartPos: splits[0].StartPos}}
	}
	splits = append(splits, positionedChunk{Content: "", StartPos: -1}) // add empty string at the end

	var mergedSplits [][]chunkPart
	var currentChunk []chunkPart
	var currentChunkLength int

	for i, split := range splits {
		splitLength := len(split.Content)
		overlapOffset := 0
		if i > 0 {
			overlapOffset = s.ChunkOverlap
		}

		isSeparator := i%2 == 1

		if isSeparator || (currentChunkLength+splitLength <= s.ChunkSize-overlapOffset) {
			currentChunk = append(currentChunk, chunkPart{Content: split.Content, ContentLength: splitLength, StartPos: split.StartPos})
			currentChunkLength += splitLength
		} else {
			mergedSplits = append(mergedSplits, currentChunk)
			currentChunk = []chunkPart{{Content: split.Content, ContentLength: splitLength, StartPos: split.StartPos}}
			currentChunkLength = splitLength
		}
	}

	if len(currentChunk) > 0 {
		mergedSplits = append(mergedSplits, currentChunk)
	}

	if len(mergedSplits) == 0 {
		singleChunkContent := ""
		startPos := -1
		for _, split := range splits {
			if startPos == -1 {
				startPos = split.StartPos
			}
			singleChunkContent += split.Content
		}
		return []positionedChunk{{Content: singleChunkContent, StartPos: startPos}}
	} else if len(mergedSplits[0]) == 0 {
		mergedSplits = mergedSplits[1:]
	}

	if s.ChunkOverlap > 0 {
		return s.enforceOverlap(mergedSplits)
	}

	var results []positionedChunk
	for _, chunk := range mergedSplits {
		partStr := ""
		startPos := -1
		for _, p := range chunk {
			if startPos == -1 {
				startPos = p.StartPos
			}
			partStr += p.Content
		}
		if partStr != "" {
			results = append(results, positionedChunk{Content: partStr, StartPos: startPos})
		}
	}
	return results
}

func (s *SimpleChunkExtractor) enforceOverlap(chunks [][]chunkPart) []positionedChunk {
	var result []positionedChunk
	for i, chunk := range chunks {
		if i == 0 {
			partStr := ""
			startPos := -1
			for _, p := range chunk {
				if startPos == -1 {
					startPos = p.StartPos
				}
				partStr += p.Content
			}
			result = append(result, positionedChunk{Content: partStr, StartPos: startPos})
		} else {
			overlapLength := 0
			var overlap []chunkPart

			// iterate backward on the previous chunk
			prevChunk := chunks[i-1]
			for j := len(prevChunk) - 1; j >= 0; j-- {
				p := prevChunk[j]
				if overlapLength+p.ContentLength > s.ChunkOverlap {
					break
				}
				overlapLength += p.ContentLength
				overlap = append(overlap, p)
			}

			// Reverse the collected overlap since we collected backwards
			partStr := ""
			startPos := -1
			for j := len(overlap) - 1; j >= 0; j-- {
				if startPos == -1 {
					startPos = overlap[j].StartPos
				}
				partStr += overlap[j].Content
			}

			for _, p := range chunk {
				if startPos == -1 && p.StartPos != -1 {
					startPos = p.StartPos
				}
				partStr += p.Content
			}
			if partStr != "" {
				result = append(result, positionedChunk{Content: partStr, StartPos: startPos})
			}
		}
	}
	return result
}
