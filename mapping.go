package main

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/analysis/token/keyword"
	"github.com/blevesearch/bleve/mapping"
)

func buildMapping() mapping.IndexMapping {
	enFieldMapping := bleve.NewTextFieldMapping()
	enFieldMapping.Analyzer = en.AnalyzerName

	eventMapping := bleve.NewDocumentMapping()
	eventMapping.AddFieldMappingsAt("summary", enFieldMapping)
	eventMapping.AddFieldMappingsAt("description", enFieldMapping)

	kwFieldMapping := bleve.NewTextFieldMapping()
	kwFieldMapping.Analyzer = keyword.Name

	eventMapping.AddFieldMappingsAt("url", kwFieldMapping)
	eventMapping.AddFieldMappingsAt("category", kwFieldMapping)

	mapping := bleve.NewIndexMapping()
	mapping.DefaultMapping = eventMapping
	mapping.DefaultAnalyzer = en.AnalyzerName

	return mapping
}
