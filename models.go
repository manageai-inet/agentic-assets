package assetmanager

import "time"

const (
	AssetTypeDocument = "document"
	AssetTypePage     = "page"
	AssetTypeChunk    = "chunk"
	AssetTypeEntity   = "entity"
	AssetTypeRelation = "relation"
	AssetTypeVector	  = "vector"
)

const (
	AssetRefTypeParent = "parent" // for any asset, parent is asset that produce this asset
	AssetRefTypeChild  = "child" // for any asset, child is asset that is produced by this asset
	AssetRefTypeSource = "source" // for relation asset, source is entity
	AssetRefTypeTarget = "target" // for relation asset, target is entity
)

type UpdateStringOperation struct {
	// REPLACE
	Operation string
	Value     string
}

type UpdateListOperation struct {
	// INSERT, REPLACE, REMOVE
	Operation string
	Values    []any
	Index     *int
}

type UpdateMapOperation struct {
	// REPLACE, MERGE
	Operation string
	Values    map[string]any
}

type UpdatedContextualAsset struct {
	Content        *UpdateStringOperation
	EmbeddingModel *UpdateStringOperation
	Refs           *UpdateListOperation
	Labels         *UpdateListOperation
	Metadata       *UpdateMapOperation
}

type AssetRef struct {
	KbId      string	`json:"kb_id"`
	AssetType string	`json:"asset_type"`
	AssetId   string	`json:"asset_id"`
	RefType   string	`json:"ref_type"`
}

type ContextualAsset struct {
	IndexedBy      string			`json:"indexed_by"`
	KbId           string			`json:"kb_id"`
	AssetType      string			`json:"asset_type"`
	AssetId        string			`json:"asset_id"`
	Version        int				`json:"version"`
	Content        string			`json:"content"`
	EmbeddingModel *string			`json:"embedding_model"`
	Refs           *[]AssetRef		`json:"refs"`
	Labels         *[]string		`json:"labels"`
	Metadata       *map[string]any	`json:"metadata"`
}

type VectorAsset struct {
	KbId           string			`json:"kb_id"`
	AssetId        string			`json:"asset_id"`
	Version        int				`json:"version"`
	Content        string			`json:"content"`
	Refs           *[]AssetRef		`json:"refs"`
	Labels         *[]string		`json:"labels"`
	Metadata       *map[string]any	`json:"metadata"`

	EmbeddingModel 	*string		`json:"embedding_model"`
	EmbededVector	[]float32	`json:"embeded_vector"`
}

type VectorQueryFilter struct {
	// Filter vector that has KbId equal to KbId
	// if nil, no apply this filter
	KbId *string

	// Filter vector that has KbId in KbIdsIn
	// if empty, no apply this filter
	KbIdsIn []string

	// Filter vector that has Version equal to Version
	// if nil, get latest version
	Version *int
	
	// Filter vector that has Label equal to Label
	// if nil, no apply this filter
	Label *string
}

// KnowledgeSource is a struct that represents a knowledge source, which is input for indexing
type KnowledgeSource struct {
	// e.g. pdf, word, ppt, plain-text, image, audio, csv, excel, sql, mongodb
	// determined by which extractor is used, and it is used for mapping source to indexer.
	SourceType string		`json:"source_type"`
	// source name, generally, it is filename. but it can be other string for other types of sources.
	SourceName string		`json:"source_name"`
	// file data read from source file
	SourceData *[]byte		`json:"source_data"`
	// file url, used for identifying the source link or for downloading the source.
	SourceUrl *string		`json:"source_url"`
	// file content, generally, it is extracted from source file by extractor. e.g. by ocr or stt process.
	SourceContents *string		`json:"source_contents"`
	// metadata, used for storing metadata of the source. (optional)
	Metadata *map[string]any	`json:"metadata"`
}

type IndexingResult struct {
	KbId      string		`json:"kb_id"`
	Status    bool			`json:"status"`
	IndexedAt time.Time		`json:"indexed_at"`

	Error *string			`json:"error"`

	Version        *int				`json:"version"`
	EmbeddingModel *string			`json:"embedding_model"`
	EmbeddingDim   *int				`json:"embedding_dim"`
	AssetsCount    *map[string]int	`json:"assets_count"`
}

type RetrievedAsset struct {
	ContextualAsset
	Score *float32	`json:"score"`
}

type RetrievedVector struct {
	VectorAsset
	Score 	*float32 `json:"score"`
}

type RetrieveResult struct {
	KbIds  []string						`json:"kb_ids"`
	Query  string						`json:"query"`
	Config *map[string]any				`json:"config"`

	Status     bool						`json:"status"`
	RetrieveAt time.Time				`json:"retrieve_at"`

	Error           *string				`json:"error"`
	RetrievedAssets *[]RetrievedAsset	`json:"retrieved_assets"`
}
